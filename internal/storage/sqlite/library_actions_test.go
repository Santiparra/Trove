package sqlite

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Santiparra/Trove/internal/books"
)

func TestQuickStatusChangesOnlyStatusAndUpdatesDashboard(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	service := books.NewService(s)
	b, err := service.Save(ctx, books.Input{Title: "Obra sintética", Status: "reading", Notes: "Conservar", SeriesName: "Serie", Identifier: "Referencia", StartedOn: "2026-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	changed, err := service.ChangeStatus(ctx, b.ID, "read")
	if err != nil {
		t.Fatal(err)
	}
	if changed.Cycle.Status != "read" || changed.Notes != b.Notes || changed.Edition != b.Edition || changed.Series.ID != b.Series.ID || changed.Cycle.ID != b.Cycle.ID || changed.Cycle.StartedOn != b.Cycle.StartedOn || changed.Cycle.FinishedOn != "" {
		t.Fatal("cambiar estado alteró otros datos o inventó fechas")
	}
	if _, err = service.ChangeStatus(ctx, b.ID, "invalid"); err == nil {
		t.Fatal("estado inválido aceptado")
	}
	if _, err = service.ChangeStatus(ctx, "missing", "read"); !errors.Is(err, books.ErrNotFound) {
		t.Fatal("ID inexistente aceptado")
	}
	d, err := service.Dashboard(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Total != 1 || d.ReadWithoutDate != 1 || len(d.RecentlyFinished) != 0 || len(d.CurrentlyReading) != 0 {
		t.Fatalf("portal inconsistente: %+v", d)
	}
	list, err := service.ListByStatus(ctx, "sintética", "read")
	if err != nil || len(list) != 1 {
		t.Fatal("filtro por estado y título incorrecto")
	}
	list, err = service.ListByStatus(ctx, "", "reading")
	if err != nil || len(list) != 0 {
		t.Fatal("el registro quedó en el estado anterior")
	}
}

func TestDashboardUsesCompletionDatesAndLimitsRecentBooks(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	service := books.NewService(s)
	for i := 1; i <= 7; i++ {
		_, err = service.Save(ctx, books.Input{Title: fmt.Sprintf("Ejemplo %d", i), Status: "read", FinishedOn: fmt.Sprintf("2026-01-%02d", i)})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = service.Save(ctx, books.Input{Title: "Sin fecha", Status: "read"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Save(ctx, books.Input{Title: "Actual", Status: "reading"}); err != nil {
		t.Fatal(err)
	}
	d, err := service.Dashboard(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Total != 9 || len(d.RecentlyFinished) != 5 || d.RecentlyFinished[0].Cycle.FinishedOn != "2026-01-07" || d.RecentlyFinished[4].Cycle.FinishedOn != "2026-01-03" || d.ReadWithoutDate != 1 || len(d.CurrentlyReading) != 1 {
		t.Fatalf("resumen incorrecto: %+v", d)
	}
	sum := 0
	for _, count := range d.Counts {
		sum += count.Count
	}
	if sum != d.Total {
		t.Fatal("los estados no suman el total")
	}
}

func TestDeletePreservesOtherBooksAndRollsBackOnFailure(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	service := books.NewService(s)
	a, err := service.Save(ctx, books.Input{Title: "Quitar", Status: "read", SeriesName: "Compartida"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := service.Save(ctx, books.Input{Title: "Conservar", Status: "reading", SeriesName: "Compartida"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.db.Exec(`CREATE TRIGGER reject_delete BEFORE DELETE ON books BEGIN SELECT RAISE(ABORT,'synthetic failure'); END`); err != nil {
		t.Fatal(err)
	}
	if err = service.Delete(ctx, a.ID); err == nil {
		t.Fatal("se esperaba fallo")
	}
	if saved, err := service.Get(ctx, a.ID); err != nil || saved.Cycle.ID != a.Cycle.ID || saved.Edition.ID != a.Edition.ID {
		t.Fatal("el fallo borró parcialmente el libro")
	}
	if _, err = s.db.Exec("DROP TRIGGER reject_delete"); err != nil {
		t.Fatal(err)
	}
	if err = service.Delete(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Get(ctx, a.ID); !errors.Is(err, books.ErrNotFound) {
		t.Fatal("el libro no se eliminó")
	}
	if saved, err := service.Get(ctx, b.ID); err != nil || saved.Series.ID != b.Series.ID {
		t.Fatal("se alteró otro libro o su serie")
	}
	if err = service.Delete(ctx, a.ID); !errors.Is(err, books.ErrNotFound) {
		t.Fatal("eliminar inexistente debe informar el conflicto")
	}
	for _, table := range []string{"books", "editions", "reading_cycles"} {
		var count int
		if err = s.db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("filas residuales en %s", table)
		}
	}
}
