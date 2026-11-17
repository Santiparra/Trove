package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
)

func TestArchiveRoundTripPreservesIDsAllCyclesAndUnknownDates(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	position := 1.5
	b, err := books.NewService(s).Save(ctx, books.Input{Title: "Obra sintética", Authors: []string{"Autor de ejemplo"}, Status: "read", SeriesName: "Serie sintética", SeriesPosition: &position, Notes: "Nota\ncon Unicode ñ", Favorite: true, Owned: true})
	if err != nil {
		t.Fatal(err)
	}
	// Additional historical data must not disappear through the current list view.
	if _, err = s.db.Exec(`INSERT INTO editions(id,book_id,language,format,identifier,owned) VALUES('extra-edition',?,'en','digital','test',0);
INSERT INTO reading_cycles(id,edition_id,status,started_on,finished_on) VALUES('extra-cycle','extra-edition','paused','2020-01-01','');`, b.ID); err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(dir, "space # backup.sqlite")
	doc, err := s.Capture(ctx, native, "synthetic-generation", time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Catalog.Editions) != 2 || len(doc.Catalog.Cycles) != 2 {
		t.Fatal("se perdieron ediciones/ciclos")
	}
	logical := filepath.Join(dir, "trove.json")
	if err = backup.WriteJSON(logical, doc); err != nil {
		t.Fatal(err)
	}
	fromJSON, err := backup.ReadJSON(logical)
	if err != nil {
		t.Fatal(err)
	}
	fromSQLite, err := s.ReadNative(ctx, native)
	if err != nil {
		t.Fatal(err)
	}
	if fromJSON.ContentHash() != fromSQLite.ContentHash() {
		t.Fatal("las representaciones no son equivalentes")
	}
	prepared := filepath.Join(dir, "prepared.sqlite")
	if err = s.Prepare(ctx, fromJSON, prepared); err != nil {
		t.Fatal(err)
	}
	if err = s.Delete(ctx, b.ID); err != nil {
		t.Fatal(err)
	}
	other, err := books.NewService(s).Save(ctx, books.Input{Title: "Antes de restaurar", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	safety, err := s.Install(ctx, prepared)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err = s.db.QueryRow("SELECT count(*) FROM reading_cycles").Scan(&count); err != nil || count != 2 {
		t.Fatal("no se reconstruyó todo el historial")
	}
	recovered, err := s.Get(ctx, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !recovered.Favorite || !recovered.Edition.Owned || recovered.Series.ID != b.Series.ID || recovered.Notes != b.Notes || recovered.Cycle.StartedOn != "" || recovered.Cycle.FinishedOn != "" {
		t.Fatal("identidad, propiedades o fechas alteradas")
	}
	if _, err = s.Get(ctx, other.ID); !errors.Is(err, books.ErrNotFound) {
		t.Fatal("restauración mezcló bibliotecas")
	}
	previous, err := s.ReadNative(ctx, safety)
	if err != nil {
		t.Fatal(err)
	}
	if len(previous.Catalog.Works) != 1 || previous.Catalog.Works[0].ID != other.ID {
		t.Fatal("copia previa incompleta")
	}
}

func TestInstallRollbackLeavesPreviousLibraryReadable(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	b, err := books.NewService(s).Save(ctx, books.Input{Title: "Conservar", Status: "up-next"})
	if err != nil {
		t.Fatal(err)
	}
	doc := backup.Document{Format: backup.Format, Version: backup.Version, Generation: "empty", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	prepared := filepath.Join(dir, "prepared.sqlite")
	if err = s.Prepare(ctx, doc, prepared); err != nil {
		t.Fatal(err)
	}
	calls := 0
	s.rename = func(a, b string) error {
		calls++
		if calls == 2 {
			return errors.New("fallo sintético de reemplazo")
		}
		return os.Rename(a, b)
	}
	safety, err := s.Install(ctx, prepared)
	if err == nil {
		t.Fatal("se esperaba fallo")
	}
	if _, err = s.Get(ctx, b.ID); err != nil {
		t.Fatal("biblioteca anterior no recuperada")
	}
	if _, err = s.ReadNative(ctx, safety); err != nil {
		t.Fatal("copia previa no legible")
	}
}

func TestUnidentifiedSQLiteCannotBeRestored(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err = s.ReadNative(context.Background(), s.path); err == nil {
		t.Fatal("una base activa no es un respaldo identificado")
	}
}
