package sqlite

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/novels"
)

func TestNovelAndVolumesRoundTripAcrossNativeAndJSONBackups(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	service := novels.NewService(s)
	completed := 7
	n, err := service.Save(ctx, novels.Input{Title: "Novela sintética", Status: "up-next", Completed: &completed, Notes: "Nota de novela", Favorite: true})
	if err != nil {
		t.Fatal(err)
	}
	v := n.Volumes[1]
	n, err = service.SaveVolume(ctx, n.ID, novels.VolumeInput{ID: v.ID, Number: 2, Notes: "Nota propia", Owned: true, Identifier: "9780547928227", StartedOn: "2020-01-01", FinishedOn: "2020-01-03"})
	if err != nil {
		t.Fatal(err)
	}
	native := filepath.Join(dir, "native.sqlite")
	doc, err := s.Capture(ctx, native, "novel-backup", time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	logical := filepath.Join(dir, "logical.json")
	if err = backup.WriteJSON(logical, doc); err != nil {
		t.Fatal(err)
	}
	fromJSON, err := backup.ReadJSON(logical)
	if err != nil {
		t.Fatal(err)
	}
	fromNative, err := s.ReadNative(ctx, native)
	if err != nil {
		t.Fatal(err)
	}
	for index, source := range []backup.Document{fromJSON, fromNative} {
		if source.ContentHash() != doc.ContentHash() {
			t.Fatal("hash de novelas divergente")
		}
		target, err := Open(filepath.Join(t.TempDir(), "target.sqlite"))
		if err != nil {
			t.Fatal(err)
		}
		prepared := filepath.Join(t.TempDir(), "prepared.sqlite")
		if err = target.Prepare(ctx, source, prepared); err != nil {
			target.Close()
			t.Fatal(err)
		}
		if _, err = target.Install(ctx, prepared); err != nil {
			target.Close()
			t.Fatal(err)
		}
		got, err := target.GetNovel(ctx, n.ID)
		target.Close()
		if err != nil || !reflect.DeepEqual(got, n) {
			t.Fatalf("representación %d perdió identidades o historial: %v", index, err)
		}
	}
	damaged := fromJSON
	damaged.Catalog.Novels = append([]novels.Novel(nil), fromJSON.Catalog.Novels...)
	damaged.Catalog.Novels[0].Volumes = nil
	if err := damaged.Validate(); err == nil {
		t.Fatal("aceptó colección de volúmenes omitida")
	}
}

func TestNovelOnlyBackupPreviewIsNotReportedAsEmpty(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "profile", "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	n, err := novels.NewService(s).Save(ctx, novels.Input{Title: "Cantidad desconocida", Status: "read"})
	if err != nil {
		t.Fatal(err)
	}
	m := backup.New(s, filepath.Join(dir, "profile"), "test")
	defer m.Close()
	state, err := m.Configure(ctx, filepath.Join(dir, "backups"))
	if err != nil {
		t.Fatal(err)
	}
	preview, err := m.PreviewGeneration(ctx, state.Generation)
	if err != nil || preview.Works != 1 || preview.Editions != 0 || preview.Cycles != 0 {
		t.Fatalf("vista previa vacía incorrecta: %+v %v", preview, err)
	}
	if _, err = m.ConfirmRestore(ctx, preview.Token); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetNovel(ctx, n.ID)
	if err != nil || got.Completed != nil || got.Total != nil || len(got.Volumes) != 0 {
		t.Fatal("restauración inventó datos")
	}
}
