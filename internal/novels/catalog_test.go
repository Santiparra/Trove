package novels_test

import (
	"github.com/Santiparra/Trove/internal/catalog"
	"reflect"
	"testing"

	"github.com/Santiparra/Trove/internal/novelcatalog"
	"github.com/Santiparra/Trove/internal/novels"
)

func seriesSnapshot() novelcatalog.Series {
	return novelcatalog.Series{ID: "10", Title: "Remote novel", Authors: []novelcatalog.Author{{ID: "40", Name: "Remote author"}}, Language: "en", Format: "print", FetchedAt: "2026-01-01T00:00:00Z", Books: []novelcatalog.BookRef{{ID: "20", Number: 1, Order: 1, Title: "Volume 1"}, {ID: "21", Number: 2, Order: 2.5, Title: "Part 2, Volume 1"}}}
}
func publicationSnapshot() novelcatalog.Publication {
	pages := 240
	return novelcatalog.Publication{SeriesID: "10", BookID: "20", ReleaseID: "30", Number: 1, Order: 1, Title: "Volume 1", ReleaseTitle: "English release", Language: "en", ReleaseFormat: "print", Format: "paperback", ISBN: "9780547928227", Description: "Source synopsis", CoverURL: "https://images.ranobedb.org/example.jpg", Pages: &pages, FetchedAt: "2026-01-01T00:00:00Z", Evidence: "ranobedb-main", Sources: map[string]novelcatalog.Source{"description": {Provider: novelcatalog.RanobeDB, Entity: "book", ID: "20", FetchedAt: "2026-01-01T00:00:00Z"}}, Warnings: []string{}}
}

func TestCatalogNovelCreationDeduplicatesAndDoesNotInferTotal(t *testing.T) {
	ctx, s, _ := fixture(t)
	result, err := s.AddCatalog(ctx, seriesSnapshot(), publicationSnapshot(), "to-read", count(0))
	if err != nil {
		t.Fatal(err)
	}
	if result.Kind != "created" || result.Novel == nil || result.Novel.Total != nil || *result.Novel.Completed != 0 || result.Novel.Volumes[0].Cycle.Status != "to-read" {
		t.Fatal("catálogo modificó seguimiento o total")
	}
	duplicate, err := s.AddCatalog(ctx, seriesSnapshot(), publicationSnapshot(), "read", count(2))
	if err != nil || duplicate.Kind != "existing" || duplicate.Novel.ID != result.Novel.ID {
		t.Fatal("duplicó novela vinculada")
	}
	items, _ := s.List(ctx, "", "")
	if len(items) != 1 || items[0].Status != "to-read" {
		t.Fatal("duplicado alteró estado personal")
	}
	local, err := s.List(ctx, "0-547-92822-X", "")
	if err != nil || len(local) != 1 {
		t.Fatal("ISBN equivalente no encuentra copia local")
	}
}

func TestLinkAndRefreshPreserveManualNovelAndVolumeData(t *testing.T) {
	ctx, s, _ := fixture(t)
	n, err := s.Save(ctx, novels.Input{Title: "Título personal", Authors: []string{"Autor personal"}, Status: "paused", Completed: count(1), Notes: "Notas personales", Favorite: true})
	if err != nil {
		t.Fatal(err)
	}
	v := n.Volumes[0]
	n, err = s.SaveVolume(ctx, n.ID, novels.VolumeInput{ID: v.ID, Number: 1, Title: "Título de volumen propio", Description: "Sinopsis propia", Identifier: "9780547928227", Owned: true, Favorite: true, Notes: "Notas del volumen", StartedOn: "2020-01-01", FinishedOn: "2020-02-01"})
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.LinkCatalog(ctx, n.ID, seriesSnapshot(), publicationSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if n.Title != "Título personal" || n.Authors[0] != "Autor personal" || n.Notes != "Notas personales" || !n.Favorite || n.Status != "paused" || n.Total != nil || *n.Completed != 1 {
		t.Fatal("vinculación pisó novela personal")
	}
	v = n.Volumes[0]
	if v.Title != "Título de volumen propio" || v.Description != "Sinopsis propia" || v.ID == "" || v.Notes != "Notas del volumen" || !v.Edition.Owned || !v.Favorite || v.Cycle.StartedOn != "2020-01-01" || v.Cycle.FinishedOn != "2020-02-01" {
		t.Fatal("vinculación pisó datos del volumen")
	}
	fresh := publicationSnapshot()
	fresh.Description = "Refreshed source"
	fresh.Pages = nil
	fresh.CoverURL = ""
	n, err = s.RefreshPublication(ctx, n.ID, *n.Catalog, seriesSnapshot(), fresh)
	if err != nil {
		t.Fatal(err)
	}
	if n.Volumes[0].Catalog.Pages == nil || n.Volumes[0].Catalog.CoverURL == "" || n.Volumes[0].Description != "Sinopsis propia" {
		t.Fatal("refresco perdió último dato válido o sinopsis personal")
	}
	before := n.Volumes[0]
	n, err = s.UnlinkCatalog(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if n.Catalog != nil || n.Volumes[0].Catalog != nil || n.Volumes[0].ID != before.ID || n.Volumes[0].Edition.ID != before.Edition.ID || !reflect.DeepEqual(n.Volumes[0].Cycle, before.Cycle) || n.Volumes[0].Notes != before.Notes {
		t.Fatal("desvincular destruyó identidad o historial")
	}
}

func TestCatalogRefreshRejectsReorderedBooksAndChangedISBN(t *testing.T) {
	ctx, s, _ := fixture(t)
	created, err := s.AddCatalog(ctx, seriesSnapshot(), publicationSnapshot(), "to-read", count(0))
	if err != nil {
		t.Fatal(err)
	}
	n := *created.Novel
	reordered := seriesSnapshot()
	reordered.Books[0].ID, reordered.Books[1].ID = reordered.Books[1].ID, reordered.Books[0].ID
	if _, err = s.RefreshPublication(ctx, n.ID, *n.Catalog, reordered, publicationSnapshot()); err == nil {
		t.Fatal("reasignó números a identidades diferentes")
	}
	changed := publicationSnapshot()
	changed.ISBN = "9780131103627"
	if _, err = s.RefreshPublication(ctx, n.ID, *n.Catalog, seriesSnapshot(), changed); err == nil {
		t.Fatal("cambió ISBN silenciosamente")
	}
	stored, _ := s.Get(ctx, n.ID)
	if stored.Volumes[0].Catalog.ISBN != "9780547928227" || stored.Catalog.Books[0].ID != "20" {
		t.Fatal("refresco rechazado alteró snapshot")
	}
}

func TestCatalogLinkKeepsNonISBNPersonalIdentifierAlongsidePublicationISBN(t *testing.T) {
	ctx, s, _ := fixture(t)
	n, err := s.Save(ctx, novels.Input{Title: "Personal", Status: "to-read", Completed: count(0)})
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.SaveVolume(ctx, n.ID, novels.VolumeInput{Number: 1, Identifier: "shelf-identifier", Notes: "Keep"})
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.LinkCatalog(ctx, n.ID, seriesSnapshot(), publicationSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	v := n.Volumes[0]
	if v.Edition.Identifier != "shelf-identifier" || catalog.PreferredISBN(v.Edition.Identifiers) != "9780547928227" || v.Notes != "Keep" {
		t.Fatal("pisó identificador personal o perdió ISBN seguro")
	}
	n, err = s.SaveVolume(ctx, n.ID, novels.VolumeInput{ID: v.ID, Number: 1, Identifier: "shelf-updated", Notes: "Keep"})
	if err != nil {
		t.Fatal(err)
	}
	if catalog.PreferredISBN(n.Volumes[0].Edition.Identifiers) != "9780547928227" {
		t.Fatal("editar un identificador personal eliminó el ISBN de catálogo")
	}
}
