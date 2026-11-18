package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestISBNEquivalencesAndContradictoryAliases(t *testing.T) {
	for _, pair := range []struct{ short, long string }{{"054792822X", "9780547928227"}, {"0131103628", "9780131103627"}, {"0441172717", "9780441172719"}} {
		if ISBN13(pair.short) != pair.long || ISBN10(pair.long) != pair.short {
			t.Fatal("equivalencia incorrecta")
		}
	}
	if !ValidISBN13("9791234567896") || ISBN10("9791234567896") != "" {
		t.Fatal("se inventó ISBN-10 de prefijo 979")
	}
	if ValidISBN13("9780547928228") || ValidISBN10("0547928221") {
		t.Fatal("se aceptó checksum inválido")
	}
	ids, warnings := AcceptDeclared([]Identifier{{Scheme: "isbn10", Value: "3423026189"}, {Scheme: "isbn13", Value: "9780441172719"}})
	if PreferredISBN(ids) != "9780441172719" || len(warnings) == 0 {
		t.Fatal("no se detectó contradicción")
	}
	for _, id := range ids {
		if id.Value == "3423026189" {
			t.Fatal("se promovió un alias contradictorio")
		}
	}
	if PreferredISBN(NormalizeIdentifier(" 978-0-547-92822-7 ", "manual")) != "9780547928227" {
		t.Fatal("no normaliza ISBN")
	}
	if got := NormalizeIdentifier("custom-uid", "manual"); len(got) != 1 || got[0].Scheme != "other" || got[0].Value != "custom-uid" {
		t.Fatal("se inventó identificador externo")
	}
}

type primaryStub struct {
	work    Work
	edition Edition
	err     error
}

func (p primaryStub) Search(context.Context, string) ([]Work, error) { return []Work{p.work}, p.err }
func (p primaryStub) Editions(context.Context, string) ([]Edition, error) {
	return []Edition{p.edition}, p.err
}
func (p primaryStub) Resolve(context.Context, string) (Work, Edition, error) {
	return p.work, p.edition, p.err
}

type secondaryStub struct {
	calls int
	value Supplement
	err   error
}

func (p *secondaryStub) Lookup(context.Context, string) (Supplement, error) {
	p.calls++
	return p.value, p.err
}

func TestCascadeOnlyFillsMissingSelectedFieldsWithProvenance(t *testing.T) {
	p := primaryStub{work: Work{ID: "10", Title: "Obra", Description: "Principal"}, edition: Edition{ID: "20", WorkID: "10", Title: "Edición", Declared: []Identifier{{Scheme: "isbn13", Value: "9780547928227"}}}}
	secondary := &secondaryStub{value: Supplement{EditionID: "OL1M", WorkID: "OL2W", CoverURL: "https://covers.openlibrary.org/b/id/1-M.jpg", Description: "No sustituir"}}
	s := New(p, secondary)
	if _, err := s.Search(context.Background(), "Obra"); err != nil || secondary.calls != 0 {
		t.Fatal("buscar consultó el secundario")
	}
	snapshot, err := s.Resolve(context.Background(), "20")
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Description != "Principal" || snapshot.Sources["description"].Provider != Hardcover || snapshot.Sources["cover"].Provider != OpenLibrary || secondary.calls != 1 {
		t.Fatal("se mezcló/sustituyó procedencia")
	}
	p.work.CoverURL = "https://assets.hardcover.app/cover.jpg"
	s = New(p, secondary)
	if _, err = s.Resolve(context.Background(), "20"); err != nil || secondary.calls != 1 {
		t.Fatal("consulta secundaria sin huecos")
	}
	p.work.CoverURL = ""
	secondary.err = errors.New("offline")
	s = New(p, secondary)
	snapshot, err = s.Resolve(context.Background(), "20")
	if err != nil || snapshot.Description != "Principal" || len(snapshot.Warnings) == 0 {
		t.Fatal("el fallo secundario invalidó información principal")
	}
	p.edition.Declared = nil
	s = New(p, secondary)
	before := secondary.calls
	if _, err = s.Resolve(context.Background(), "20"); err != nil || secondary.calls != before {
		t.Fatal("se buscó secundario sin identificación segura")
	}
}

func TestRefreshRetainsMissingMetadataAndRejectsUnsafeCover(t *testing.T) {
	pages := 123
	old := Snapshot{Provider: Hardcover, FetchedAt: "2026-01-01T00:00:00Z", Work: Work{ID: "10"}, Edition: Edition{ID: "20", Pages: &pages}, Description: "Conservar", CoverURL: "https://assets.hardcover.app/old.jpg", Sources: map[string]Source{"description": {Hardcover, "work", "10", "2026-01-01T00:00:00Z"}}}
	fresh := KeepMissing(old, Snapshot{Sources: map[string]Source{}})
	if fresh.Description != old.Description || fresh.CoverURL != old.CoverURL || fresh.Edition.Pages == nil || fresh.Sources["description"].FetchedAt != old.FetchedAt {
		t.Fatal("se borró metadata válida")
	}
	for _, url := range []string{"http://assets.hardcover.app/a.jpg", "https://localhost/a.jpg", "https://assets.hardcover.app.evil.test/a.jpg", "https://user@assets.hardcover.app/a.jpg", "https://covers.openlibrary.org:123/a.jpg"} {
		if ValidCoverURL(url) {
			t.Fatal("URL de portada insegura admitida")
		}
	}
}
