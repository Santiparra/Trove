package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Santiparra/Trove/internal/books"
)

func TestSchemaTwoMigrationKeepsBookAndSeriesIdentities(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"migrations/001_books.sql", "migrations/002_personal_flags.sql"} {
		ddl, err := migrations.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(string(ddl)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = db.Exec(`INSERT INTO series(id,name,name_key) VALUES('series','Saga','saga');
INSERT INTO books(id,title,authors_json,notes,added_on,series_id,favorite) VALUES('work','Original','[]','Nota','2020-01-01','series',1);
INSERT INTO editions(id,book_id,owned) VALUES('edition','work',1);
INSERT INTO reading_cycles(id,edition_id,status) VALUES('cycle','edition','read'); PRAGMA user_version=2;`); err != nil {
		t.Fatal(err)
	}
	legacyNative := filepath.Join(dir, "legacy-backup.sqlite")
	if err = nativeCopy(context.Background(), db, legacyNative, "legacy-generation", "2020-01-01T00:00:00Z"); err != nil {
		t.Fatal(err)
	}
	copy, err := sql.Open("sqlite", legacyNative)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = copy.Exec("UPDATE trove_backup SET version=1"); err != nil {
		t.Fatal(err)
	}
	copy.Close()
	db.Close()
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	b, err := s.Get(context.Background(), "work")
	if err != nil || b.Medium != books.MediumBooks || b.Series.ID != "series" || !b.Favorite || !b.Edition.Owned || b.Notes != "Nota" || b.Cycle.FinishedOn != "" {
		t.Fatal("migración alteró el registro anterior")
	}
	bookService := books.NewService(s)
	second, err := bookService.Save(context.Background(), books.Input{Title: "Segundo libro", SeriesName: "Saga", Status: "to-read"})
	if err != nil || second.Series.ID != "series" {
		t.Fatal("la serie anterior perdió su identidad de búsqueda")
	}
	novels, err := bookService.ForMedium(books.MediumLightNovels)
	if err != nil {
		t.Fatal(err)
	}
	volume, err := novels.Save(context.Background(), books.Input{Medium: books.MediumLightNovels, Title: "Volumen", SeriesName: "Saga", Status: "to-read"})
	if err != nil || volume.Series.ID == "series" {
		t.Fatal("se mezclaron series de distintos medios")
	}
	paths, err := filepath.Glob(filepath.Join(dir, "before-schema-5-*.sqlite"))
	if err != nil || len(paths) != 1 {
		t.Fatal("falta copia previa a migración")
	}
	doc, err := s.ReadNative(context.Background(), legacyNative)
	if err != nil || doc.Version != 1 || doc.Catalog.Works[0].Medium != books.MediumBooks || doc.Catalog.Series[0].Medium != "" {
		t.Fatal("no se admite el respaldo SQLite anterior")
	}
	prepared := filepath.Join(dir, "prepared.sqlite")
	if err = s.Prepare(context.Background(), doc, prepared); err != nil {
		t.Fatal(err)
	}
	from, err := Open(prepared)
	if err != nil {
		t.Fatal(err)
	}
	defer from.Close()
	recovered, err := from.Get(context.Background(), "work")
	if err != nil || recovered.Edition.ID != "edition" || recovered.Cycle.ID != "cycle" {
		t.Fatal("recuperación anterior no preservó IDs")
	}
}
