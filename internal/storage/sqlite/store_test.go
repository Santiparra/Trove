package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/Santiparra/Trove/internal/books"
)

func TestBooksSurviveReopenAndShareSeries(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "trove.sqlite")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	service := books.NewService(store)
	position := 1.5
	first, err := service.Save(ctx, books.Input{Title: "Árbol de invierno", Authors: []string{" Autor de ejemplo "}, Status: "to-read", SeriesName: "Ciclo de prueba", SeriesPosition: &position})
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Save(ctx, books.Input{Title: "Segunda obra", Status: "reading", SeriesName: " ciclo   DE prueba "})
	if err != nil {
		t.Fatal(err)
	}
	if first.Series.ID != second.Series.ID {
		t.Fatal("los libros no comparten identidad de serie")
	}
	edited, err := service.Save(ctx, books.Input{ID: first.ID, Title: "Árbol de invierno", Authors: first.Authors, Status: "read", Notes: "Opinión sintética", AddedOn: first.AddedOn, StartedOn: "2026-01-02", FinishedOn: "2026-01-07", SeriesName: first.Series.Name, SeriesPosition: &position, Language: "en", Format: "digital"})
	if err != nil {
		t.Fatal(err)
	}
	if edited.Edition.ID != first.Edition.ID || edited.Cycle.ID != first.Cycle.ID {
		t.Fatal("editar cambió la identidad de edición o ciclo")
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	found, err := reopened.List(ctx, "ÁRBOL", books.MediumBooks)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 {
		t.Fatalf("búsqueda: %d resultados", len(found))
	}
	b := found[0]
	if b.ID != first.ID || b.Notes != "Opinión sintética" || b.Cycle.FinishedOn != "2026-01-07" || b.Edition.Format != "digital" || *b.Series.Position != position {
		t.Fatalf("reapertura incompleta: %+v", b)
	}
	all, err := reopened.List(ctx, "", books.MediumBooks)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatal("editar creó duplicados")
	}
	if _, err = books.NewService(reopened).Save(ctx, books.Input{ID: "inexistente", Title: "No crear", Status: "to-read"}); !errors.Is(err, books.ErrNotFound) {
		t.Fatal("editar ID inexistente debe fallar")
	}
}

func TestSaveRollsBackAllTablesOnFailure(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_, err = s.db.Exec(`CREATE TRIGGER reject_cycle BEFORE INSERT ON reading_cycles BEGIN SELECT RAISE(ABORT,'synthetic failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = books.NewService(s).Save(context.Background(), books.Input{Title: "No debe persistir", Status: "reading", SeriesName: "Tampoco persistir"})
	if err == nil {
		t.Fatal("se esperaba un fallo")
	}
	for _, table := range []string{"books", "editions", "reading_cycles", "series"} {
		var count int
		if err = s.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s quedó parcialmente escrito", table)
		}
	}
}

func TestFutureSchemaIsRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "future.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec("PRAGMA user_version=999"); err != nil {
		t.Fatal(err)
	}
	s.Close()
	if s, err = Open(path); err == nil {
		s.Close()
		t.Fatal("una base de esquema futuro debe rechazarse")
	}
}

func TestFailedEditPreservesSavedBookAndSeriesCanBeCleared(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	service := books.NewService(s)
	ctx := context.Background()
	b, err := service.Save(ctx, books.Input{Title: "Libro conservado", Status: "to-read", SeriesName: "Serie sintética"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec(`CREATE TRIGGER reject_update BEFORE UPDATE ON reading_cycles BEGIN SELECT RAISE(ABORT,'synthetic failure'); END`)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Save(ctx, books.Input{ID: b.ID, Title: "Cambio que debe revertirse", Status: "read"})
	if err == nil {
		t.Fatal("se esperaba fallo al editar")
	}
	saved, err := service.Get(ctx, b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Title != b.Title || saved.Series == nil || saved.Series.ID != b.Series.ID || saved.Cycle.Status != b.Cycle.Status {
		t.Fatal("un guardado fallido dejó cambios parciales")
	}
	if _, err = s.db.Exec("DROP TRIGGER reject_update"); err != nil {
		t.Fatal(err)
	}
	saved, err = service.Save(ctx, books.Input{ID: b.ID, Title: b.Title, Status: "reading"})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Series != nil || saved.Edition.ID != b.Edition.ID || saved.Cycle.ID != b.Cycle.ID {
		t.Fatal("quitar serie no debe cambiar identidades ni retener el vínculo")
	}
	groups, err := service.ListSeries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 0 {
		t.Fatal("no deben sugerirse series sin libros vinculados")
	}
}
