package novels_test

import (
	"github.com/Santiparra/Trove/internal/novelcatalog"
	"github.com/Santiparra/Trove/internal/novels"
	"testing"
)

func TestDecimalLabelsDoNotChangeCompletedCountsOrStableIdentity(t *testing.T) {
	ctx, s, _ := fixture(t)
	series := seriesSnapshot()
	series.Books = []novelcatalog.BookRef{{ID: "20", Number: 1, Order: 6, Title: "Novel, Vol. 6"}, {ID: "21", Number: 2, Order: 6.5, Title: "Novel, Vol. 6.5"}, {ID: "22", Number: 3, Order: 7, Title: "Novel, Vol. 7"}}
	pub := publicationSnapshot()
	created, err := s.AddCatalog(ctx, series, pub, "up-next", count(0))
	if err != nil {
		t.Fatal(err)
	}
	n := *created.Novel
	in := input(n)
	in.ProgressBookID = "21"
	n, err = s.Save(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if *n.Completed != 1 || n.NextNumber() != 2 || n.Label(2) != "6.5" {
		t.Fatal("confundió etiqueta decimal y cantidad")
	}
	ids := []string{n.Volumes[0].ID}
	n, err = s.SetCompleted(ctx, n.ID, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if *n.Completed != 2 || n.Label(n.NextNumber()) != "7" || n.Volumes[0].ID != ids[0] {
		t.Fatal("el contador o la identidad se desplazaron")
	}
	for _, v := range n.Volumes {
		if v.Cycle.StartedOn != "" || v.Cycle.FinishedOn != "" {
			t.Fatal("ajuste rápido inventó fechas")
		}
	}
}

func TestQuickProgressPreservesKnownDatesAndCanRetractKnownCompletion(t *testing.T) {
	ctx, s, _ := fixture(t)
	n, err := s.Save(ctx, novels.Input{Title: "Novel", Status: "up-next", Completed: count(1), Total: count(2)})
	if err != nil {
		t.Fatal(err)
	}
	v := n.Volumes[0]
	n, err = s.SaveVolume(ctx, n.ID, novels.VolumeInput{ID: v.ID, Number: 1, StartedOn: "2020-01-01", FinishedOn: "2020-01-02", Notes: "Keep"})
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.SetCompleted(ctx, n.ID, 2, true)
	if err != nil || n.Status != "read" {
		t.Fatal("no completó con consentimiento")
	}
	n, err = s.SetCompleted(ctx, n.ID, 1, false)
	if err != nil || n.Status != "up-next" || *n.Completed != 1 {
		t.Fatal("no permitió corregir progreso de Leída")
	}
	if n.Volumes[0].Cycle.StartedOn != "2020-01-01" || n.Volumes[0].Cycle.FinishedOn != "2020-01-02" || n.Volumes[0].Notes != "Keep" || n.Volumes[1].Cycle.FinishedOn != "" {
		t.Fatal("el ajuste alteró fechas o notas")
	}
}

func TestNovelDashboardCountsParentsAndLimitsRecentAndQueue(t *testing.T) {
	ctx, s, _ := fixture(t)
	for index := 0; index < 7; index++ {
		n, err := s.Save(ctx, novels.Input{Title: "Novel", Status: "up-next", Completed: count(1)})
		if err != nil {
			t.Fatal(err)
		}
		v := n.Volumes[0]
		if _, err = s.SaveVolume(ctx, n.ID, novels.VolumeInput{ID: v.ID, Number: 1, FinishedOn: "2020-01-01"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Save(ctx, novels.Input{Title: "Unknown dates", Status: "read", Completed: count(2)}); err != nil {
		t.Fatal(err)
	}
	d, err := s.Dashboard(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if d.Total != 8 || len(d.UpNext) != 5 || len(d.Recent) != 5 {
		t.Fatal("portal contó volúmenes como novelas o excedió límites")
	}
	for _, item := range d.Recent {
		if item.FinishedOn == "" {
			t.Fatal("ordenó lectura sin fecha como reciente")
		}
	}
}
