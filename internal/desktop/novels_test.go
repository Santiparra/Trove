package desktop

import (
	"path/filepath"
	"testing"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/novels"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
)

func TestNovelsRejectWritesDuringShutdown(t *testing.T) {
	store, err := sqlite.Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	library := NewLibrary(books.NewService(store), "test", "synthetic", "contract")
	adapter := NewNovels(novels.NewService(store), library)
	zero := 0
	n, err := adapter.Save(novels.Input{Title: "Novela", Status: "to-read", Completed: &zero})
	if err != nil {
		t.Fatal(err)
	}
	library.closing = true
	if _, err = adapter.StartVolume(n.ID); err == nil {
		t.Fatal("aceptó inicio durante cierre")
	}
	if _, err = adapter.FinishVolume(n.ID); err == nil {
		t.Fatal("aceptó avance durante cierre")
	}
	if _, err = adapter.ChangeStatus(n.ID, "read"); err == nil {
		t.Fatal("aceptó cambio de estado durante cierre")
	}
	if _, err = adapter.SaveVolume(n.ID, novels.VolumeInput{Number: 1}); err == nil {
		t.Fatal("aceptó volumen durante cierre")
	}
	if err = adapter.Delete(n.ID); err == nil {
		t.Fatal("aceptó borrado durante cierre")
	}
	stored, err := adapter.Get(n.ID)
	if err != nil || *stored.Completed != 0 || len(stored.Volumes) != 0 {
		t.Fatal("cierre modificó datos")
	}
}
