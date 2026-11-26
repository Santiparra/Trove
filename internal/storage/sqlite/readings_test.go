package sqlite

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/novels"
)

func TestBookRereadPreservesHistoryAndHistoricalEditsDoNotReplaceCurrent(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := books.NewService(store)
	b, err := s.Save(ctx, books.Input{Title: "Synthetic book", Status: "read", StartedOn: "2020-01-01", FinishedOn: "2020-01-03", Language: "en", Owned: true, Notes: "Keep"})
	if err != nil {
		t.Fatal(err)
	}
	rating := 8.25
	b, err = s.SaveReading(ctx, b.ID, books.ReadingInput{ID: b.Cycle.ID, Status: "read", StartedOn: b.Cycle.StartedOn, FinishedOn: b.Cycle.FinishedOn, Rating: &rating})
	if err != nil {
		t.Fatal(err)
	}
	previous := b.Cycle
	b, err = s.ChangeStatus(ctx, b.ID, "reading")
	if err != nil {
		t.Fatal(err)
	}
	if b.Cycle.ID == previous.ID || b.Cycle.Rating != nil || b.Cycle.FinishedOn != "" || len(b.History) != 1 || !reflect.DeepEqual(b.History[0], previous) {
		t.Fatal("reread reset the previous reading")
	}
	currentID := b.Cycle.ID
	b, err = s.ChangeStatus(ctx, b.ID, "paused")
	if err != nil {
		t.Fatal(err)
	}
	b, err = s.ChangeStatus(ctx, b.ID, "reading")
	if err != nil || b.Cycle.ID != currentID || len(b.History) != 1 {
		t.Fatal("resume created another reading", err)
	}
	b, err = s.SaveReading(ctx, b.ID, books.ReadingInput{Status: "read", FinishedOn: "2019-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	if b.Cycle.ID != currentID || b.Cycle.Status != "reading" || len(b.History) != 2 || b.History[1].StartedOn != "" || b.History[1].Edition != nil {
		t.Fatal("historical insertion changed current or invented dates")
	}
	b, err = s.SaveReading(ctx, b.ID, books.ReadingInput{ID: previous.ID, Status: "read", StartedOn: "2020-01-02", FinishedOn: "2020-01-04", Rating: &rating})
	if err != nil {
		t.Fatal(err)
	}
	if b.Cycle.ID != currentID || len(b.History) != 2 || b.History[0].ID != previous.ID || b.History[0].StartedOn != "2020-01-02" {
		t.Fatal("correction created another reading")
	}
	bad := 8.1
	if _, err = s.SaveReading(ctx, b.ID, books.ReadingInput{ID: previous.ID, Status: "read", Rating: &bad}); err == nil {
		t.Fatal("invalid rating accepted")
	}
	foreign, _ := s.ForMedium(books.MediumLightNovels)
	if _, err = foreign.SaveReading(ctx, b.ID, books.ReadingInput{Status: "read"}); err == nil {
		t.Fatal("cross-medium history write accepted")
	}
	items, err := s.List(ctx, "")
	if err != nil || len(items) != 1 || len(items[0].History) != 2 {
		t.Fatal("cycles duplicated the library entry", err)
	}
	doc, err := store.Capture(ctx, filepath.Join(dir, "native.sqlite"), "readings", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Catalog.Cycles) != 3 {
		t.Fatal("backup dropped history")
	}
	jsonPath := filepath.Join(dir, "library.json")
	if err = backup.WriteJSON(jsonPath, doc); err != nil {
		t.Fatal(err)
	}
	logical, err := backup.ReadJSON(jsonPath)
	if err != nil {
		t.Fatal(err)
	}
	prepared := filepath.Join(dir, "prepared.sqlite")
	if err = store.Prepare(ctx, logical, prepared); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Install(ctx, prepared); err != nil {
		t.Fatal(err)
	}
	restored, err := s.Get(ctx, b.ID)
	if err != nil || !reflect.DeepEqual(restored, b) {
		t.Fatal("round trip changed current/history", err)
	}
	b, err = s.ChangeStatus(ctx, b.ID, "read")
	if err != nil {
		t.Fatal(err)
	}
	finishedID := b.Cycle.ID
	b, err = s.Save(ctx, books.Input{ID: b.ID, Title: b.Title, Status: "reading", Language: "es", StartedOn: b.Cycle.StartedOn, FinishedOn: b.Cycle.FinishedOn, Owned: true, Notes: b.Notes})
	if err != nil {
		t.Fatal(err)
	}
	if b.Cycle.ID == finishedID || b.Cycle.FinishedOn != "" || b.Cycle.Rating != nil || b.Cycle.Edition.Language != "es" || b.History[len(b.History)-1].Edition.Language != "en" {
		t.Fatal("general editor reused prior dates or edition for a new reading")
	}
}

func TestNovelSelectedRereadAndIndividualRereadKeepIndependentHistory(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := novels.NewService(store)
	total := 3
	n, err := s.Save(ctx, novels.Input{Title: "Synthetic novel", Status: "read", Total: &total, Completed: &total})
	if err != nil {
		t.Fatal(err)
	}
	first := n.Volumes[0].Cycle
	second := n.Volumes[1].Cycle
	ids := []string{n.Volumes[1].ID, n.Volumes[2].ID}
	n, err = s.RereadNovel(ctx, n.ID, ids)
	if err != nil {
		t.Fatal(err)
	}
	if n.NextNumber() != 2 || *n.Completed != 0 || len(n.Readings) != 2 || n.Volumes[1].Cycle.ID != second.ID || n.Volumes[1].Cycle.StartedOn != "" {
		t.Fatal("selecting journey started/reset volumes")
	}
	runID := n.ActiveRunID
	n, err = s.Start(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Volumes[1].History) != 1 || n.Volumes[1].History[0].ID != second.ID || n.Volumes[1].Cycle.RunID != runID || n.Volumes[1].Cycle.StartedOn == "" {
		t.Fatal("starting lost prior reading")
	}
	cycleID := n.Volumes[1].Cycle.ID
	n, err = s.ChangeStatus(ctx, n.ID, "paused")
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.Start(ctx, n.ID)
	if err != nil || n.Volumes[1].Cycle.ID != cycleID {
		t.Fatal("resume duplicated volume cycle", err)
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil || *n.Completed != 1 || n.NextNumber() != 3 || n.Status != "up-next" {
		t.Fatal("journey did not advance by selection", err)
	}
	if n.Volumes[2].Cycle.StartedOn != "" || len(n.Volumes[2].History) != 0 {
		t.Fatal("finishing started next volume")
	}
	n, err = s.Start(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil || *n.Completed != 2 || n.NextNumber() != 0 || n.Status != "read" {
		t.Fatal("selected journey did not finish", err)
	}
	if n.Volumes[0].Cycle.ID != first.ID || len(n.Volumes[0].History) != 0 {
		t.Fatal("unselected volume reset")
	}
	previousRun := *n.Run()
	n, err = s.RereadVolume(ctx, n.ID, n.Volumes[1].ID)
	if err != nil {
		t.Fatal(err)
	}
	if *n.Completed != 2 || len(n.Readings) != 2 || n.ActiveVolumeID == "" {
		t.Fatal("individual reread restarted journey")
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil || n.ActiveVolumeID != "" || n.Status != "read" || !reflect.DeepEqual(*n.Run(), previousRun) {
		t.Fatal("individual finish changed journey history", err)
	}
	n, err = s.SaveReading(ctx, n.ID, n.Volumes[0].ID, books.ReadingInput{Status: "read", FinishedOn: "2018-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	if n.Volumes[0].Cycle.ID != first.ID || len(n.Volumes[0].History) != 1 {
		t.Fatal("historical volume replaced tracking")
	}
	doc, err := store.Capture(ctx, filepath.Join(dir, "native.sqlite"), "novel-readings", "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "prepared.sqlite")
	if err = store.Prepare(ctx, doc, path); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Install(ctx, path); err != nil {
		t.Fatal(err)
	}
	restored, err := s.Get(ctx, n.ID)
	if err != nil || !reflect.DeepEqual(restored, n) {
		t.Fatal("novel history round trip changed data", err)
	}
	if _, err = s.RereadNovel(ctx, n.ID, []string{"foreign-volume"}); err == nil {
		t.Fatal("foreign volume selected")
	}
	saved, _ := s.Get(ctx, n.ID)
	if !reflect.DeepEqual(saved, n) {
		t.Fatal("failed journey left partial state")
	}
	n, err = s.RereadVolume(ctx, n.ID, n.Volumes[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	n, err = s.ChangeStatus(ctx, n.ID, "up-next")
	if err != nil {
		t.Fatal("could not queue the same reread", err)
	}
	cycleID = n.Volumes[0].Cycle.ID
	n, err = s.Save(ctx, novels.Input{ID: n.ID, Title: n.Title, Status: "paused", Completed: n.Completed, Total: n.Total, AddedOn: n.AddedOn, Notes: "Edited while rereading"})
	if err != nil || n.Volumes[0].Cycle.ID != cycleID || n.ActiveVolumeID == "" {
		t.Fatal("general editor lost the individual reread", err)
	}
	n, err = s.Start(ctx, n.ID)
	if err != nil || n.Volumes[0].Cycle.ID != cycleID {
		t.Fatal("queued reread created another cycle", err)
	}
	n, err = s.ChangeStatus(ctx, n.ID, "did-not-finish")
	if err != nil || n.ActiveVolumeID != "" || n.Status != "read" || *n.Completed != 2 {
		t.Fatal("abandoned reread destroyed completed tracking", err)
	}
}

func TestNovelReadToReadingCreatesJourneyAtomically(t *testing.T) {
	ctx := context.Background()
	store, err := Open(filepath.Join(t.TempDir(), "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	s := novels.NewService(store)
	one := 1
	n, err := s.Save(ctx, novels.Input{Title: "Novel", Status: "read", Total: &one})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Save(ctx, novels.Input{ID: n.ID, Title: "", Status: "reading", Total: &one, Completed: &one}); err == nil {
		t.Fatal("invalid edit accepted")
	}
	saved, _ := s.Get(ctx, n.ID)
	if len(saved.Readings) != 0 || saved.Status != "read" {
		t.Fatal("failed edit persisted a new journey")
	}
	n, err = s.ChangeStatus(ctx, n.ID, "reading")
	if err != nil || n.Status != "reading" || *n.Completed != 0 || n.Volumes[0].Cycle.Status != "reading" {
		t.Fatal("read-to-reading did not create a reread", err)
	}
	n, err = s.Finish(ctx, n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.RereadVolume(ctx, n.ID, n.Volumes[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ChangeStatus(ctx, n.ID, "read"); err != nil {
		t.Fatal("individual status completion failed", err)
	}
}
