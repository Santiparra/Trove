package desktop

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
)

type delayedRepository struct {
	books.Repository
	entered, release chan struct{}
}

func (r *delayedRepository) Save(ctx context.Context, b books.Book) (books.Book, error) {
	close(r.entered)
	<-r.release
	return r.Repository.Save(ctx, b)
}

func TestShutdownWaitsForAcceptedWritesAndBacksThemUp(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "data", "test")
	store, err := sqlite.Open(filepath.Join(dir, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	manager := backup.New(store, dir, "test")
	defer manager.Close()
	store.SetOnChange(manager.MarkDirty)
	if _, err = manager.Configure(context.Background(), filepath.Join(root, "cloud")); err != nil {
		t.Fatal(err)
	}
	repository := &delayedRepository{Repository: store, entered: make(chan struct{}), release: make(chan struct{})}
	library := NewLibrary(books.NewService(repository), "test", dir, "synthetic")
	service := NewBackups(manager, library, nil, nil)
	written := make(chan error, 1)
	go func() {
		_, err := library.SaveBook(books.Input{Title: "Antes de cerrar", Status: "to-read"})
		written <- err
	}()
	<-repository.entered
	done := make(chan error, 1)
	go func() { done <- service.ServiceShutdown() }()
	select {
	case <-done:
		close(repository.release)
		t.Fatal("cerró sin esperar el guardado aceptado")
	case <-time.After(80 * time.Millisecond):
	}
	close(repository.release)
	if err = <-written; err != nil {
		t.Fatal(err)
	}
	select {
	case err = <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("el cierre quedó bloqueado")
	}
	state := manager.State()
	doc, err := backup.ReadJSON(filepath.Join(state.Folder, "generation-"+state.Generation, "trove.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Catalog.Works) != 1 || doc.Catalog.Works[0].Title != "Antes de cerrar" {
		t.Fatal("el respaldo de cierre perdió el último cambio")
	}
	if _, err = library.SaveBook(books.Input{Title: "Después de cerrar", Status: "to-read"}); err == nil {
		t.Fatal("aceptó escrituras después del cierre")
	}
}
