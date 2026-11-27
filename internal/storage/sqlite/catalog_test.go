package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
)

func selectedSnapshot() catalog.Snapshot {
	return catalog.Snapshot{Provider: catalog.Hardcover, FetchedAt: "2026-01-01T00:00:00Z", Work: catalog.Work{ID: "10", Title: "Título externo", Authors: []string{"Autora externa"}, Series: []catalog.Series{{ID: "40", Name: "Serie externa"}}}, Edition: catalog.Edition{ID: "20", WorkID: "10", Title: "Edición externa", Language: "en", Format: "digital", Declared: []catalog.Identifier{{Scheme: "isbn13", Value: "9780547928227", Source: catalog.Hardcover}}}, Identifiers: catalog.NormalizeIdentifier("9780547928227", catalog.Hardcover), Description: "Descripción externa", CoverURL: "https://assets.hardcover.app/cover.jpg", Sources: map[string]catalog.Source{"description": {Provider: catalog.Hardcover, Entity: "work", ID: "10", FetchedAt: "2026-01-01T00:00:00Z"}}, Warnings: []string{}}
}

func TestLinkAndArchivePreservePersonalDataAndCatalogIdentity(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	service := books.NewService(s)
	b, err := service.Save(ctx, books.Input{Title: "Título manual", Authors: []string{"Autor local"}, Status: "reading", StartedOn: "2020-01-01", Identifier: "978-0-547-92822-7", Notes: "No modificar", SeriesName: "Serie propia", Favorite: true, Owned: true})
	if err != nil {
		t.Fatal(err)
	}
	linked, err := service.LinkCatalog(ctx, b.ID, selectedSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	if linked.Title != b.Title || linked.Notes != b.Notes || !reflect.DeepEqual(linked.Cycle, b.Cycle) || linked.Edition.ID != b.Edition.ID || linked.Series.ID != b.Series.ID || !linked.Favorite || !linked.Edition.Owned || linked.Edition.Catalog == nil {
		t.Fatal("vincular alteró datos propios")
	}
	result, err := service.AddCatalog(ctx, selectedSnapshot(), "", "to-read")
	if err != nil || result.Kind != "existing" || result.Book.ID != b.ID {
		t.Fatal("se duplicó una edición ya vinculada")
	}
	if err = os.WriteFile(filepath.Join(dir, "catalog-credentials.bin"), []byte("synthetic-key-never-backup"), 0600); err != nil {
		t.Fatal(err)
	}
	doc, err := s.Capture(ctx, filepath.Join(dir, "backup.sqlite"), "catalog-generation", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := json.Marshal(doc)
	if err != nil || strings.Contains(string(bytes), "synthetic-key-never-backup") {
		t.Fatal("credenciales incluidas en respaldo")
	}
	prepared := filepath.Join(dir, "prepared.sqlite")
	if err = s.Prepare(ctx, doc, prepared); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Install(ctx, prepared); err != nil {
		t.Fatal(err)
	}
	restored, err := service.Get(ctx, b.ID)
	if err != nil || restored.Edition.Catalog.Edition.ID != "20" || len(restored.Edition.Identifiers) != 2 || !reflect.DeepEqual(restored.Cycle, b.Cycle) {
		t.Fatal("recuperación perdió snapshot/identificadores/historial")
	}
	unlinked, err := service.UnlinkCatalog(ctx, b.ID)
	if err != nil || unlinked.Edition.Catalog != nil || unlinked.Title != b.Title || len(unlinked.Edition.Identifiers) != 2 {
		t.Fatal("desvincular perdió datos propios")
	}
}

func TestCatalogAdditionAndLinkingDoNotSilentlyMergeEditions(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	service := books.NewService(s)
	snapshot := selectedSnapshot()
	manual, err := service.Save(ctx, books.Input{Title: "Manual", Status: "to-read", Identifier: "054792822X"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.AddCatalog(ctx, snapshot, "", "to-read")
	if err != nil || result.Kind != "review" || result.Book.ID != manual.ID {
		t.Fatal("el ISBN local se duplicó o vinculó sin revisión")
	}
	other, err := service.Save(ctx, books.Input{Title: "Otra edición", Status: "to-read", Identifier: "9780131103627"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.LinkCatalog(ctx, other.ID, snapshot); err == nil {
		t.Fatal("vinculó ISBN incompatibles")
	}
	if _, err = service.LinkCatalog(ctx, manual.ID, snapshot); err != nil {
		t.Fatal(err)
	}
	changed := snapshot
	changed.Edition.ID = "21"
	changed.Edition.Declared = nil
	changed.Identifiers = nil
	if result, err = service.AddCatalog(ctx, changed, "", "to-read"); err != nil || result.Kind != "review" {
		t.Fatal("otra edición de la misma obra se fusionó/duplicó automáticamente")
	}
	novels, _ := service.ForMedium(books.MediumLightNovels)
	if _, err = novels.AddCatalog(ctx, snapshot, "", "to-read"); err == nil {
		t.Fatal("se habilitó proveedor de novelas fuera de alcance")
	}
}

func TestSchemaThreeUpgradeNormalizesIdentifiersWithoutChangingRawData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "old.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"migrations/001_books.sql", "migrations/002_personal_flags.sql", "migrations/003_reading_media.sql"} {
		ddl, err := migrations.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(string(ddl)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Exec(`INSERT INTO books(id,title,authors_json,notes,added_on) VALUES('b','Manual','[]','Conservar',''); INSERT INTO editions(id,book_id,identifier) VALUES('e','b','978-0-547-92822-7'); INSERT INTO reading_cycles(id,edition_id,status) VALUES('c','e','read'); PRAGMA user_version=3;`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	b, err := s.Get(context.Background(), "b")
	if err != nil || b.Edition.Identifier != "978-0-547-92822-7" || len(b.Edition.Identifiers) != 2 || b.Cycle.FinishedOn != "" || b.Notes != "Conservar" {
		t.Fatal("migración cambió datos propios")
	}
	paths, err := filepath.Glob(filepath.Join(dir, "before-schema-8-*.sqlite"))
	if err != nil || len(paths) != 1 {
		t.Fatal("falta copia previa")
	}
}

func TestCatalogInspectionPrefersExactEditionAndDoesNotWrite(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	service := books.NewService(s)
	other, err := service.Save(ctx, books.Input{Title: "A otra edición", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	otherSnapshot := selectedSnapshot()
	otherSnapshot.Edition.ID = "21"
	otherSnapshot.Identifiers = nil
	if _, err = service.LinkCatalog(ctx, other.ID, otherSnapshot); err != nil {
		t.Fatal(err)
	}
	exact, err := service.Save(ctx, books.Input{Title: "Z edición exacta", Status: "read", Notes: "Conservar"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.LinkCatalog(ctx, exact.ID, selectedSnapshot()); err != nil {
		t.Fatal(err)
	}
	match, err := service.MatchCatalog(ctx, selectedSnapshot())
	if err != nil || match == nil || match.Kind != "existing" || match.Book.ID != exact.ID {
		t.Fatalf("no priorizó edición exacta: %+v %v", match, err)
	}
	items, err := service.List(ctx, "")
	if err != nil || len(items) != 2 || match.Book.Notes != "Conservar" || match.Book.Cycle.FinishedOn != "" {
		t.Fatal("inspeccionar modificó datos")
	}
}

func TestCatalogAddsChosenStatusAndSeriesWithoutInventingReadDates(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	service := books.NewService(s)
	ctx := context.Background()
	result, err := service.AddCatalog(ctx, selectedSnapshot(), "40", "read")
	if err != nil {
		t.Fatal(err)
	}
	b := result.Book
	if b.Cycle.Status != "read" || b.Cycle.StartedOn != "" || b.Cycle.FinishedOn != "" || b.Series == nil || b.Series.Name != "Serie externa" {
		t.Fatalf("alta incorrecta: %+v", b)
	}
	if _, err = service.AddCatalog(ctx, selectedSnapshot(), "40", "invalid"); err == nil {
		t.Fatal("aceptó estado inválido")
	}
}

func TestLocalSearchFindsAuthorSeriesAndEquivalentISBN(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	service := books.NewService(s)
	b, err := service.Save(ctx, books.Input{Title: "Título sintético", Authors: []string{"Autora sintética"}, SeriesName: "Serie inventada", Identifier: "0-547-92822-X", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{"Autora sintética", "Serie inventada", "978-0-547-92822-7"} {
		items, err := service.Search(ctx, query)
		if err != nil || len(items.Matches) != 1 || items.Matches[0].ID != b.ID {
			t.Fatalf("búsqueda local %q: %+v %v", query, items, err)
		}
	}
}
