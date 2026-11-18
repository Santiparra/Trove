package desktop

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Santiparra/Trove/internal/books"
	cat "github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
)

type catalogRemote struct{ err error }

func (r *catalogRemote) Search(context.Context, string) ([]cat.Work, error) {
	return []cat.Work{}, r.err
}
func (r *catalogRemote) Editions(context.Context, string) ([]cat.Edition, error) {
	return []cat.Edition{}, r.err
}
func (r *catalogRemote) Resolve(context.Context, string) (cat.Work, cat.Edition, error) {
	return cat.Work{ID: "10", Title: "Remoto", Authors: []string{"Autora"}, Description: "Conservar descripción"}, cat.Edition{ID: "20", WorkID: "10", Title: "Edición", Declared: []cat.Identifier{{Scheme: "isbn13", Value: "9780547928227"}}}, r.err
}

type catalogKey struct{ value string }

func (k *catalogKey) Get() (string, error) { return k.value, nil }
func (k *catalogKey) Save(v string) error  { k.value = v; return nil }
func (k *catalogKey) Clear() error         { k.value = ""; return nil }

type catalogImages struct{}

func (catalogImages) Get(context.Context, string) (string, error) { return "", nil }

type catalogChoices struct {
	catalogRemote
	selected string
}

func (r *catalogChoices) Editions(context.Context, string) ([]cat.Edition, error) {
	return []cat.Edition{{ID: "21", WorkID: "10", Title: "Vacía"}, {ID: "22", WorkID: "10", Title: "Informativa", Language: "en", Publisher: "Sintética"}}, nil
}
func (r *catalogChoices) Resolve(ctx context.Context, id string) (cat.Work, cat.Edition, error) {
	r.selected = id
	work, edition, err := r.catalogRemote.Resolve(ctx, id)
	edition.ID = id
	return work, edition, err
}

func TestCatalogWorkOpensInformativeEditionThenRecognizesLocalCopy(t *testing.T) {
	s, err := sqlite.Open(filepath.Join(t.TempDir(), "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	library := NewLibrary(books.NewService(s), "test", "synthetic", "contract")
	remote := &catalogChoices{}
	adapter := NewCatalog(cat.New(remote, nil), &catalogKey{}, catalogImages{}, library, nil)
	preview, err := adapter.OpenWork("10")
	if err != nil || remote.selected != "22" || preview.Match != nil {
		t.Fatalf("no abrió edición informativa: %+v %v", preview, err)
	}
	created, err := adapter.Add(preview.Token, "", "read")
	if err != nil {
		t.Fatal(err)
	}
	remote.selected = ""
	preview, err = adapter.OpenWork("10")
	if err != nil || preview.Match == nil || preview.Match.Kind != "existing" || preview.Match.Book.ID != created.Book.ID || remote.selected != "" {
		t.Fatal("no reconoció copia local sin nuevas consultas")
	}
}

func TestCatalogPreviewIsTemporaryAndFailedRefreshKeepsLastSnapshot(t *testing.T) {
	s, err := sqlite.Open(filepath.Join(t.TempDir(), "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	library := NewLibrary(books.NewService(s), "test", "synthetic", "contract")
	remote := &catalogRemote{}
	adapter := NewCatalog(cat.New(remote, nil), &catalogKey{}, catalogImages{}, library, nil)
	preview, err := adapter.Preview("20")
	if err != nil {
		t.Fatal(err)
	}
	items, err := library.ListBooks("", "", books.MediumBooks)
	if err != nil || len(items) != 0 {
		t.Fatal("consultar creó biblioteca")
	}
	if _, err = adapter.Add("forged-token", "", "to-read"); err == nil {
		t.Fatal("aceptó selección no validada")
	}
	adapter.mu.Lock()
	entry := adapter.previews[preview.Token]
	entry.at = time.Now().Add(-11 * time.Minute)
	adapter.previews[preview.Token] = entry
	adapter.mu.Unlock()
	if _, err = adapter.Add(preview.Token, "", "to-read"); err == nil {
		t.Fatal("aceptó selección vencida")
	}
	preview, err = adapter.Preview("20")
	if err != nil {
		t.Fatal(err)
	}
	created, err := adapter.Add(preview.Token, "", "to-read")
	if err != nil || created.Kind != "created" {
		t.Fatal(err)
	}
	if _, err = library.SaveBookNotes(created.Book.ID, "Nota local", books.MediumBooks); err != nil {
		t.Fatal(err)
	}
	remote.err = errors.New("offline")
	if _, err = adapter.Refresh(created.Book.ID); err == nil {
		t.Fatal("se esperaba fallo de refresco")
	}
	stored, err := library.GetBook(created.Book.ID, books.MediumBooks)
	if err != nil || stored.Notes != "Nota local" || stored.Edition.Catalog.Description != "Conservar descripción" || stored.Edition.Catalog.FetchedAt != preview.Snapshot.FetchedAt {
		t.Fatal("un fallo perdió datos personales o snapshot válido")
	}
	if _, err = adapter.Unlink(created.Book.ID); err != nil {
		t.Fatal(err)
	}
	stored, err = library.GetBook(created.Book.ID, books.MediumBooks)
	if err != nil || stored.Edition.Catalog != nil || stored.Notes != "Nota local" {
		t.Fatal("desvincular alteró seguimiento")
	}
}
