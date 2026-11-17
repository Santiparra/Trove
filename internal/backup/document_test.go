package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func validDocument() Document {
	return Document{Format: Format, Version: Version, Generation: "synthetic", CreatedAt: "2026-01-01T00:00:00Z", Catalog: Catalog{
		Works:    []Work{{ID: "work", Medium: "books", Title: "Obra sintética"}},
		Editions: []Edition{{ID: "edition", WorkID: "work"}},
		Cycles:   []Cycle{{ID: "cycle", EditionID: "edition", Status: "read"}},
	}}
}

func TestValidationRejectsBrokenRelationshipsAndFutureFormats(t *testing.T) {
	d := validDocument()
	if err := d.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]func(*Document){
		"future format":  func(d *Document) { d.Version = 999 },
		"orphan edition": func(d *Document) { d.Catalog.Editions[0].WorkID = "missing" },
		"orphan cycle":   func(d *Document) { d.Catalog.Cycles[0].EditionID = "missing" },
		"unknown medium": func(d *Document) { d.Catalog.Works[0].Medium = "future-medium" },
		"reversed dates": func(d *Document) {
			d.Catalog.Cycles[0].StartedOn = "2026-02-01"
			d.Catalog.Cycles[0].FinishedOn = "2026-01-01"
		},
		"duplicate identity": func(d *Document) { d.Catalog.Works = append(d.Catalog.Works, d.Catalog.Works[0]) },
		"missing cycle":      func(d *Document) { d.Catalog.Cycles = nil },
	} {
		t.Run(name, func(t *testing.T) {
			d := validDocument()
			change(&d)
			if err := d.Validate(); err == nil {
				t.Fatal("aceptó un respaldo inválido")
			}
		})
	}
}

func TestJSONRejectsUnknownFieldsAndTrailingData(t *testing.T) {
	for _, data := range []string{`{"format":"trove-library","version":1,"unknown":true}`, `{} {}`, `not json`, `{"format":"trove-library","version":1,"generation":"synthetic","createdAt":"2026-01-01T00:00:00Z"}`, `{"format":"trove-library","version":1,"generation":"synthetic","createdAt":"2026-01-01T00:00:00Z","catalog":{}}`} {
		path := filepath.Join(t.TempDir(), "test.json")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadJSON(path); err == nil {
			t.Fatal("JSON inválido aceptado")
		}
	}
}
