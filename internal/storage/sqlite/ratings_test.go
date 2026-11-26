package sqlite

import (
	"context"
	"math"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/novels"
)

func TestRatingsModesPrecisionFrozenHistoryAndPortableRestore(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := novels.NewService(store)
	total := 4
	n, err := s.Save(ctx, novels.Input{Title: "Rated synthetic novel", Status: "read", Total: &total})
	if err != nil {
		t.Fatal(err)
	}
	if n.Rating != nil || n.RatingMode != "" {
		t.Fatal("novel was rated by default")
	}
	n, err = s.SetRating(ctx, n.ID, "auto", nil)
	if err != nil || n.Rating != nil {
		t.Fatal("empty automatic rating became zero", err)
	}
	for i, value := range []float64{7, 8.25, 10} {
		v := n.Volumes[i]
		n, err = s.SaveReading(ctx, n.ID, v.ID, books.ReadingInput{ID: v.Cycle.ID, Status: "read", Rating: &value})
		if err != nil {
			t.Fatal(err)
		}
	}
	if math.Abs(*n.Rating-25.25/3) > 1e-12 || *n.Rating == 8.5 || n.Volumes[3].Cycle.Rating != nil {
		t.Fatal("automatic score rounded to quarters or counted unrated volumes")
	}
	n, err = s.SetRating(ctx, n.ID, "auto", nil)
	if err != nil {
		t.Fatal(err)
	}
	frozen := *n.Readings[0].Cycle.Rating
	value := 9.5
	v := n.Volumes[1]
	n, err = s.SaveReading(ctx, n.ID, v.ID, books.ReadingInput{ID: v.Cycle.ID, Status: "read", Rating: &value})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(*n.Rating-26.5/3) > 1e-12 || *n.Readings[0].Cycle.Rating != frozen {
		t.Fatal("live average rewrote completed reading")
	}
	// Date corrections keep an automatic snapshot's full precision.
	c := n.Readings[0].Cycle
	n, err = s.SaveRun(ctx, n.ID, books.ReadingInput{ID: c.ID, Status: c.Status, StartedOn: "2020-01-01", FinishedOn: "2020-01-02", Rating: c.Rating})
	if err != nil || *n.Readings[0].Cycle.Rating != frozen {
		t.Fatal("could not correct an automatic historical reading", err)
	}
	manual := 8.25
	n, err = s.SetRating(ctx, n.ID, "manual", &manual)
	if err != nil {
		t.Fatal(err)
	}
	value = 1
	v = n.Volumes[0]
	n, err = s.SaveReading(ctx, n.ID, v.ID, books.ReadingInput{ID: v.Cycle.ID, Status: "read", Rating: &value})
	if err != nil || *n.Rating != manual {
		t.Fatal("volume rating changed manual novel rating", err)
	}
	for _, bad := range []float64{0, 0.75, 10.25, 8.1, math.NaN(), math.Inf(1)} {
		if _, err = s.SetRating(ctx, n.ID, "manual", &bad); err == nil {
			t.Fatal("invalid personal score accepted", bad)
		}
	}
	if _, err = s.SetRating(ctx, n.ID, "manual", nil); err == nil {
		t.Fatal("manual null accepted instead of explicit none")
	}
	if _, err = s.SetRating(ctx, n.ID, "provider", nil); err == nil {
		t.Fatal("unknown rating mode accepted")
	}
	saved, _ := s.Get(ctx, n.ID)
	if !reflect.DeepEqual(saved, n) {
		t.Fatal("failed rating write altered data")
	}
	n, err = s.SetRating(ctx, n.ID, "auto", nil)
	if err != nil {
		t.Fatal(err)
	}
	last := n.Readings[0].Cycle
	n, err = s.RereadNovel(ctx, n.ID, []string{n.Volumes[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if n.RatingMode != "" || n.Rating != nil || !reflect.DeepEqual(n.Readings[0].Cycle, last) {
		t.Fatal("reread reused or changed the previous novel rating")
	}
	n, err = s.Start(ctx, n.ID)
	if err != nil || n.Volumes[0].Cycle.Rating != nil {
		t.Fatal("volume reread reused score", err)
	}
	n, err = s.SetRating(ctx, n.ID, "auto", nil)
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !n.Run().RatingFrozen || n.Run().Cycle.Rating == nil || !n.Run().Cycle.RatingAutomatic {
		t.Fatal("automatic journey score was not snapshotted")
	}
	doc, err := store.Capture(ctx, filepath.Join(dir, "native.sqlite"), "ratings", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	logical := filepath.Join(dir, "library.json")
	if err = backup.WriteJSON(logical, doc); err != nil {
		t.Fatal(err)
	}
	fromJSON, err := backup.ReadJSON(logical)
	if err != nil {
		t.Fatal(err)
	}
	fromNative, err := store.ReadNative(ctx, filepath.Join(dir, "native.sqlite"))
	if err != nil || fromNative.ContentHash() != fromJSON.ContentHash() {
		t.Fatal("ratings diverged in backups", err)
	}
	prepared := filepath.Join(dir, "prepared.sqlite")
	if err = store.Prepare(ctx, fromJSON, prepared); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Install(ctx, prepared); err != nil {
		t.Fatal(err)
	}
	restored, err := s.Get(ctx, n.ID)
	if err != nil || !reflect.DeepEqual(restored, n) {
		t.Fatal("ratings or frozen snapshots changed during restore", err)
	}
}

func TestBookExternalRefreshNeverReplacesPersonalRating(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := books.NewService(store)
	b, err := s.Save(ctx, books.Input{Title: "Synthetic rated book", Status: "read"})
	if err != nil {
		t.Fatal(err)
	}
	value := 1.0
	b, err = s.SaveReading(ctx, b.ID, books.ReadingInput{ID: b.Cycle.ID, Status: "read", Rating: &value})
	if err != nil {
		t.Fatal("lower bound rejected", err)
	}
	value = 10
	b, err = s.SaveReading(ctx, b.ID, books.ReadingInput{ID: b.Cycle.ID, Status: "read", Rating: &value})
	if err != nil {
		t.Fatal("upper bound rejected", err)
	}
	before := b.Cycle
	snapshot := selectedSnapshot()
	votes := 12
	snapshot.Work.Rating = &catalog.Rating{Provider: catalog.Hardcover, Entity: "work", ID: snapshot.Work.ID, Value: 4.135, Scale: 5, Votes: &votes, FetchedAt: snapshot.FetchedAt}
	b, err = s.LinkCatalog(ctx, b.ID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(b.Cycle, before) || b.Edition.Catalog.Work.Rating.Value != 4.135 || b.Edition.Catalog.Work.Rating.Scale != 5 {
		t.Fatal("external score overwrote personal or original scale")
	}
	snapshot.Work.Rating.Value = 3.513457
	b, err = s.LinkCatalog(ctx, b.ID, snapshot)
	if err != nil || !reflect.DeepEqual(b.Cycle, before) {
		t.Fatal("refresh changed personal score", err)
	}
	b, err = s.Reread(ctx, b.ID)
	if err != nil || b.Cycle.Rating != nil || *b.History[0].Rating != 10 {
		t.Fatal("reread failed to preserve and clear personal rating", err)
	}
	doc, err := store.Capture(ctx, filepath.Join(dir, "native.sqlite"), "external-ratings", "2026-01-01T00:00:00Z")
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
	restored, err := s.Get(ctx, b.ID)
	if err != nil || !reflect.DeepEqual(restored, b) || restored.Edition.Catalog.Work.Rating.Value != 3.513457 {
		t.Fatal("external aggregate, personal history or original precision changed during restore", err)
	}
}

func TestRatingModeReturnsWithInterruptedNovelProgress(t *testing.T) {
	ctx := context.Background()
	store, err := Open(filepath.Join(t.TempDir(), "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := novels.NewService(store)
	count := 2
	n, err := s.Save(ctx, novels.Input{Title: "Unfinished rated novel", Status: "reading", Completed: &count})
	if err != nil {
		t.Fatal(err)
	}
	value := 8.25
	n, err = s.SetRating(ctx, n.ID, "manual", &value)
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.RereadNovel(ctx, n.ID, []string{n.Volumes[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	if n.Rating != nil || n.RatingMode != "" {
		t.Fatal("new journey did not start without a rating")
	}
	n, err = s.Start(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if n.Status != "reading" || n.ActiveRunID != "" || n.RatingMode != "manual" || n.Rating == nil || *n.Rating != value {
		t.Fatal("interrupted progress lost its rating context")
	}
}
