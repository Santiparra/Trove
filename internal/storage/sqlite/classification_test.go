package sqlite

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/classification"
	"github.com/Santiparra/Trove/internal/novelcatalog"
	"github.com/Santiparra/Trove/internal/novels"
)

func importedTerms(provider string, terms ...classification.ProviderTerm) *classification.Set {
	return classification.ProviderSet(provider, terms)
}

func TestBookClassificationManualProvenanceSuppressionRefreshAndRestore(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := books.NewService(store)
	b, err := s.Save(ctx, books.Input{Title: "Synthetic classified book", Status: "reading", StartedOn: "2020-01-01", Notes: "Keep notes", Favorite: true, Owned: true})
	if err != nil {
		t.Fatal(err)
	}
	cycle := b.Cycle
	b, err = s.SetClassification(ctx, b.ID, classification.Input{Genres: []string{"Romance", " fantasy "}, Tags: []string{" Cozy ", "cozy"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Classification.Tags) != 1 || b.Classification.Tags[0] != "Cozy" {
		t.Fatal("manual normalization failed")
	}
	snapshot := selectedSnapshot()
	snapshot.Work.Classification = importedTerms(catalog.Hardcover, classification.ProviderTerm{Name: "Fantasy", ID: "4", Category: "genre"}, classification.ProviderTerm{Name: "Calm", ID: "5", Category: "mood"})
	b, err = s.LinkCatalog(ctx, b.ID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.EffectiveClassification.Genres) != 2 || len(b.EffectiveClassification.Genres[0].Origins) != 2 || len(b.EffectiveClassification.Tags) != 2 {
		t.Fatal("manual/provider identities were lost or duplicates not combined")
	}
	b, err = s.SetClassification(ctx, b.ID, classification.Input{Genres: []string{"Romance"}, Tags: []string{"Cozy"}, Hidden: []classification.Hidden{{Kind: "genre", Provider: catalog.Hardcover, ID: "4"}}})
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Work.Classification = importedTerms(catalog.Hardcover, classification.ProviderTerm{Name: "Fantasy", ID: "4", Category: "genre"}, classification.ProviderTerm{Name: "Drama", ID: "6", Category: "genre"}, classification.ProviderTerm{Name: "Joyful", ID: "5", Category: "mood"})
	b, err = s.LinkCatalog(ctx, b.ID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if len(b.EffectiveClassification.Genres) != 2 || b.EffectiveClassification.Genres[0].Name != "Drama" || b.EffectiveClassification.Genres[1].Name != "Romance" || b.EffectiveClassification.Tags[0].Name != "Cozy" {
		t.Fatal("refresh overwrote manual terms or restored a suppressed term")
	}
	if !reflect.DeepEqual(cycle, b.Cycle) || b.Notes != "Keep notes" || !b.Favorite || !b.Edition.Owned {
		t.Fatal("classification changed reading or personal data")
	}
	foreign, _ := s.ForMedium(books.MediumLightNovels)
	if _, err = foreign.SetClassification(ctx, b.ID, classification.Input{Tags: []string{"Wrong medium"}}); err == nil {
		t.Fatal("classification crossed media")
	}
	if _, err = s.SetClassification(ctx, b.ID, classification.Input{Tags: []string{"bad\x00term"}}); err == nil {
		t.Fatal("invalid term accepted")
	}
	saved, _ := s.Get(ctx, b.ID)
	if !reflect.DeepEqual(saved, b) {
		t.Fatal("failed classification left partial changes")
	}
	doc, err := store.Capture(ctx, filepath.Join(dir, "native.sqlite"), "classification", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	logical := filepath.Join(dir, "library.json")
	if err = backup.WriteJSON(logical, doc); err != nil {
		t.Fatal(err)
	}
	decoded, err := backup.ReadJSON(logical)
	if err != nil {
		t.Fatal(err)
	}
	native, err := store.ReadNative(ctx, filepath.Join(dir, "native.sqlite"))
	if err != nil || decoded.ContentHash() != native.ContentHash() {
		t.Fatal("classification differs across native and JSON", err)
	}
	prepared := filepath.Join(dir, "prepared.sqlite")
	if err = store.Prepare(ctx, decoded, prepared); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Install(ctx, prepared); err != nil {
		t.Fatal(err)
	}
	restored, err := s.Get(ctx, b.ID)
	if err != nil || !reflect.DeepEqual(restored, b) {
		t.Fatal("classification provenance or overrides were lost on restore", err)
	}
	b, err = s.UnlinkCatalog(ctx, b.ID)
	if err != nil || len(b.EffectiveClassification.Genres) != 1 || b.EffectiveClassification.Genres[0].Name != "Romance" || len(b.EffectiveClassification.Tags) != 1 {
		t.Fatal("unlink removed manual terms or retained provider terms", err)
	}
}

func TestNovelClassificationDoesNotPropagateBetweenNovelAndVolumes(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := novels.NewService(store)
	series := novelcatalog.Series{ID: "10", Title: "Synthetic novel", Language: "en", Format: "print", FetchedAt: "2026-01-01T00:00:00Z", Books: []novelcatalog.BookRef{{ID: "20", Number: 1, Order: 1, Title: "Volume 1"}}}
	series.Classification = importedTerms(catalog.RanobeDB, classification.ProviderTerm{Name: "Fantasy", ID: "7", Category: "genre"}, classification.ProviderTerm{Name: "Found family", ID: "26", Category: "tag"})
	pub := novelcatalog.Publication{SeriesID: "10", BookID: "20", ReleaseID: "30", Number: 1, Order: 1, Title: "Volume 1", Language: "en", ReleaseFormat: "print", Format: "other", Evidence: "ranobedb-main", FetchedAt: "2026-01-01T00:00:00Z", Sources: map[string]novelcatalog.Source{}}
	zero := 0
	created, err := s.AddCatalog(ctx, series, pub, "to-read", &zero)
	if err != nil {
		t.Fatal(err)
	}
	n := *created.Novel
	if n.EffectiveClassification == nil || n.Volumes[0].EffectiveClassification != nil {
		t.Fatal("series classification leaked into its volume")
	}
	cycle := n.Volumes[0].Cycle
	n, err = s.SetClassification(ctx, n.ID, n.Volumes[0].ID, classification.Input{Genres: []string{"Mystery"}, Tags: []string{"Cozy"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(n.EffectiveClassification.Genres) != 1 || n.EffectiveClassification.Genres[0].Name != "Fantasy" || n.Volumes[0].EffectiveClassification.Genres[0].Name != "Mystery" {
		t.Fatal("volume classification changed novel classification")
	}
	n, err = s.SetClassification(ctx, n.ID, "", classification.Input{Tags: []string{"Personal series tag"}})
	if err != nil {
		t.Fatal(err)
	}
	fresh := series
	fresh.Classification = importedTerms(catalog.RanobeDB, classification.ProviderTerm{Name: "Romance", ID: "11", Category: "genre"})
	n, err = s.RefreshPublication(ctx, n.ID, series, fresh, pub)
	if err != nil {
		t.Fatal(err)
	}
	if n.EffectiveClassification.Tags[0].Name != "Personal series tag" || n.EffectiveClassification.Genres[0].Name != "Romance" || n.Volumes[0].EffectiveClassification.Tags[0].Name != "Cozy" || !reflect.DeepEqual(n.Volumes[0].Cycle, cycle) {
		t.Fatal("refresh lost manual classification or changed reading")
	}
	if _, err = s.SetClassification(ctx, n.ID, "foreign-volume", classification.Input{Tags: []string{"No"}}); err == nil {
		t.Fatal("foreign volume accepted")
	}
	raw, err := store.GetNovel(ctx, n.ID)
	if err != nil || raw.EffectiveClassification != nil || raw.Volumes[0].EffectiveClassification != nil {
		t.Fatal("derived API projection was stored in library data", err)
	}
	doc, err := store.Capture(ctx, filepath.Join(dir, "native.sqlite"), "novel-classification", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	logical := filepath.Join(dir, "library.json")
	if err = backup.WriteJSON(logical, doc); err != nil {
		t.Fatal(err)
	}
	decoded, err := backup.ReadJSON(logical)
	if err != nil {
		t.Fatal(err)
	}
	prepared := filepath.Join(dir, "prepared.sqlite")
	if err = store.Prepare(ctx, decoded, prepared); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Install(ctx, prepared); err != nil {
		t.Fatal(err)
	}
	restored, err := s.Get(ctx, n.ID)
	if err != nil || !reflect.DeepEqual(restored, n) {
		t.Fatal("novel/volume classification changed on restore", err)
	}
}
