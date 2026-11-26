package desktop

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/novelcatalog"
	"github.com/Santiparra/Trove/internal/novels"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
)

type novelRemote struct{ entered, release chan struct{} }

func (r *novelRemote) Search(context.Context, string, string, int) (novelcatalog.SearchResult, error) {
	return novelcatalog.SearchResult{}, nil
}
func (r *novelRemote) Series(context.Context, string) (novelcatalog.Series, error) {
	return novelcatalog.Series{ID: "10", Title: "Remote", Language: "en", Format: "print", FetchedAt: "2026-01-01T00:00:00Z", Books: []novelcatalog.BookRef{{ID: "20", Number: 1, Order: 1, Title: "Vol. 1"}}}, nil
}
func (r *novelRemote) Book(context.Context, string, string) (novelcatalog.Book, error) {
	return novelcatalog.Book{ID: "20", SeriesID: "10", Title: "Vol. 1", Description: "Source description", Releases: []novelcatalog.Release{{ID: "30", Language: "en", Format: "print", ISBN: "9780547928227"}}}, nil
}
func (r *novelRemote) Release(context.Context, string, string) (novelcatalog.Publication, error) {
	if r.entered != nil {
		close(r.entered)
		<-r.release
	}
	return novelcatalog.Publication{ReleaseID: "30", Language: "en", ReleaseFormat: "print", Format: "other", ISBN: "9780547928227"}, nil
}

func TestNovelCatalogPreviewRequiresTokenAndDoesNotCreateRecords(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(filepath.Join(dir, "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	library := NewLibrary(books.NewService(store), "test", dir, "contract")
	remote := &novelRemote{}
	adapter := NewNovels(novels.NewService(store), library, novelcatalog.New(remote, nil), nil)
	preview, err := adapter.OpenCatalog("10", 1, "", "")
	if err != nil {
		t.Fatal(err)
	}
	items, _ := adapter.List("", "")
	if len(items) != 0 {
		t.Fatal("consultar creó registros")
	}
	zero := 0
	if _, err = adapter.AddCatalog("forged", "to-read", &zero); err == nil {
		t.Fatal("aceptó snapshot inventado")
	}
	created, err := adapter.AddCatalog(preview.Token, "to-read", &zero)
	if err != nil || created.Novel == nil {
		t.Fatal(err)
	}
	opened, err := adapter.OpenCatalog("10", 1, "", "")
	if err != nil || opened.Match.Kind != "existing" || opened.Match.Novel.ID != created.Novel.ID || opened.Token != "" {
		t.Fatal("no reconoció novela local")
	}
}

func TestNovelMetadataStartedBeforeRestoreCannotOverwriteRecoveredSnapshot(t *testing.T) {
	dir := t.TempDir()
	store, err := sqlite.Open(filepath.Join(dir, "profile", "live.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	library := NewLibrary(books.NewService(store), "test", dir, "contract")
	remote := &novelRemote{}
	adapter := NewNovels(novels.NewService(store), library, novelcatalog.New(remote, nil), nil)
	preview, err := adapter.OpenCatalog("10", 1, "", "")
	if err != nil {
		t.Fatal(err)
	}
	zero := 0
	created, err := adapter.AddCatalog(preview.Token, "to-read", &zero)
	if err != nil {
		t.Fatal(err)
	}
	m := backup.New(store, filepath.Join(dir, "profile"), "test")
	defer m.Close()
	state, err := m.Configure(context.Background(), filepath.Join(dir, "backups"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := m.PreviewGeneration(context.Background(), state.Generation)
	if err != nil {
		t.Fatal(err)
	}
	remote.entered, remote.release = make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() { _, err := adapter.RefreshCatalog(created.Novel.ID, 1, false); done <- err }()
	<-remote.entered
	backups := NewBackups(m, library, nil, nil)
	if _, err = backups.ConfirmRestore(p.Token); err != nil {
		close(remote.release)
		t.Fatal(err)
	}
	close(remote.release)
	if err = <-done; err == nil {
		t.Fatal("refresco anterior a restauración escribió sobre datos recuperados")
	}
	stored, err := adapter.Get(created.Novel.ID)
	if err != nil || stored.Volumes[0].Catalog.FetchedAt != preview.Publication.FetchedAt {
		t.Fatal("perdió snapshot restaurado")
	}
}
