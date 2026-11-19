package novels_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/novels"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
)

func count(value int) *int { return &value }
func fixture(t *testing.T) (context.Context, *novels.Service, *sqlite.Store) {
	t.Helper()
	s, err := sqlite.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return context.Background(), novels.NewService(s), s
}
func input(n novels.Novel) novels.Input {
	return novels.Input{ID: n.ID, Title: n.Title, Authors: n.Authors, Status: n.Status, Total: n.Total, Completed: n.Completed, Notes: n.Notes, Favorite: n.Favorite, AddedOn: n.AddedOn}
}

func TestHistoricalProgressCreatesOnlyKnownCompletionsWithoutDates(t *testing.T) {
	ctx, s, _ := fixture(t)
	n, err := s.Save(ctx, novels.Input{Title: "Novela sintética", Status: "up-next", Completed: count(7)})
	if err != nil {
		t.Fatal(err)
	}
	if n.NextNumber() != 8 || n.Total != nil || len(n.Volumes) != 7 {
		t.Fatalf("progreso incorrecto: %+v", n)
	}
	for _, v := range n.Volumes {
		if v.Cycle.Status != "read" || v.Cycle.StartedOn != "" || v.Cycle.FinishedOn != "" || v.Title != "" {
			t.Fatal("inventó fechas o títulos")
		}
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil || *n.Completed != 8 || n.Status != "up-next" || len(n.Volumes) != 8 || n.NextNumber() != 9 {
		t.Fatal("no avanzó sin iniciar el siguiente")
	}
	v := n.Volumes[7]
	if v.Cycle.StartedOn != "" || v.Cycle.FinishedOn != "" {
		t.Fatal("terminar sin inicio inventó fechas")
	}
}

func TestLiveVolumeFlowKeepsIdentityAndStopsAtKnownTotal(t *testing.T) {
	ctx, s, _ := fixture(t)
	n, err := s.Save(ctx, novels.Input{Title: "Novela", Status: "to-read", Completed: count(0), Total: count(2)})
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.Start(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	first := n.Volumes[0]
	if first.Cycle.StartedOn != time.Now().Format("2006-01-02") || n.Status != "reading" {
		t.Fatal("no inició volumen")
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if n.Status != "up-next" || *n.Completed != 1 || len(n.Volumes) != 1 || n.Volumes[0].ID != first.ID || n.Volumes[0].Cycle.ID != first.Cycle.ID || n.Volumes[0].Cycle.FinishedOn == "" {
		t.Fatal("perdió historial o inició siguiente automáticamente")
	}
	n, err = s.Start(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if n.Status != "read" || *n.Completed != 2 || n.NextNumber() != 0 || len(n.Volumes) != 2 {
		t.Fatal("no terminó en N/N")
	}
	if _, err = s.Start(ctx, n.ID); err == nil {
		t.Fatal("creó volumen N+1")
	}
	if _, err = s.Finish(ctx, n.ID); err == nil {
		t.Fatal("completó volumen inexistente")
	}
}

func TestReadWithUnknownQuantityDoesNotMaterializeVolumes(t *testing.T) {
	ctx, s, _ := fixture(t)
	n, err := s.Save(ctx, novels.Input{Title: "Leída sin cantidades", Status: "read"})
	if err != nil || n.Total != nil || n.Completed != nil || len(n.Volumes) != 0 {
		t.Fatal("inventó una cantidad desconocida")
	}
	if _, err = s.Start(ctx, n.ID); err == nil {
		t.Fatal("eligió un volumen sin progreso conocido")
	}
	n, err = s.Save(ctx, novels.Input{Title: "Leída con cantidad conocida", Status: "read", Completed: count(3)})
	if err != nil || *n.Completed != 3 || n.Total != nil || len(n.Volumes) != 3 {
		t.Fatal("no conservó n/?")
	}
}

func TestVolumeMetadataAndDatesSurviveBulkProgressAndOtherNovelCannotEditIt(t *testing.T) {
	ctx, s, _ := fixture(t)
	n, err := s.Save(ctx, novels.Input{Title: "Principal", Status: "paused", Completed: count(2)})
	if err != nil {
		t.Fatal(err)
	}
	v := n.Volumes[0]
	volumeID, editionID, cycleID := v.ID, v.Edition.ID, v.Cycle.ID
	n, err = s.SaveVolume(ctx, n.ID, novels.VolumeInput{ID: v.ID, Number: 1, Owned: true, Favorite: true, Notes: "Nota del volumen", Language: "ja", StartedOn: "2020-01-01", FinishedOn: "2020-01-02"})
	if err != nil {
		t.Fatal(err)
	}
	in := input(n)
	in.Completed = count(5)
	n, err = s.Save(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	v = n.Volumes[0]
	if v.ID != volumeID || v.Edition.ID != editionID || v.Cycle.ID != cycleID || v.Notes != "Nota del volumen" || !v.Edition.Owned || !v.Favorite || v.Cycle.StartedOn != "2020-01-01" || v.Cycle.FinishedOn != "2020-01-02" {
		t.Fatal("progreso pisó identidades o datos personales")
	}
	in = input(n)
	in.Completed = count(0)
	n, err = s.Save(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if n.Volumes[0].Cycle.StartedOn != "2020-01-01" || n.Volumes[0].Cycle.FinishedOn != "2020-01-02" {
		t.Fatal("reducir progreso destruyó fechas conocidas")
	}
	other, err := s.Save(ctx, novels.Input{Title: "Otra", Status: "to-read", Completed: count(0)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.SaveVolume(ctx, other.ID, novels.VolumeInput{ID: v.ID, Number: 1}); err == nil {
		t.Fatal("editó volumen ajeno")
	}
	bad := input(n)
	bad.Total = count(1)
	if _, err = s.Save(ctx, bad); err == nil {
		t.Fatal("ocultó volúmenes fuera del nuevo total")
	}
	saved, _ := s.Get(ctx, n.ID)
	if saved.Total != nil || len(saved.Volumes) != 5 {
		t.Fatal("escritura fallida dejó estado parcial")
	}
}

func TestNovelLibraryDoesNotConvertOrDeleteOldVolumesOrBooks(t *testing.T) {
	ctx, s, store := fixture(t)
	book, err := books.NewService(store).Save(ctx, books.Input{Title: "Libro", Status: "read", Notes: "Conservar"})
	if err != nil {
		t.Fatal(err)
	}
	oldService, _ := books.NewService(store).ForMedium(books.MediumLightNovels)
	old, err := oldService.Save(ctx, books.Input{Medium: books.MediumLightNovels, Title: "Volumen del modelo anterior", Status: "paused", Notes: "No convertir"})
	if err != nil {
		t.Fatal(err)
	}
	items, err := s.List(ctx, "", "")
	if err != nil || len(items) != 0 {
		t.Fatal("convirtió volúmenes anteriores")
	}
	n, err := s.Save(ctx, novels.Input{Title: "Nueva novela", Status: "to-read", Completed: count(0)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Get(ctx, book.ID); err == nil {
		t.Fatal("confundió libro con novela")
	}
	if err = s.Delete(ctx, n.ID); err != nil {
		t.Fatal(err)
	}
	stored, err := store.Get(ctx, old.ID)
	if err != nil || stored.Notes != "No convertir" {
		t.Fatal("tocó volumen anterior")
	}
	stored, err = store.Get(ctx, book.ID)
	if err != nil || stored.Notes != "Conservar" {
		t.Fatal("tocó libros")
	}
}
