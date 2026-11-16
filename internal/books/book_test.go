package books

import "testing"

func TestValidationKeepsUnknownDatesAndRejectsInvalidInput(t *testing.T) {
	n, err := (Input{Title: " Libro manual ", Status: "read", Authors: []string{"", "  Autora "}}).Normalize("2026-11-15")
	if err != nil {
		t.Fatal(err)
	}
	if n.StartedOn != "" || n.FinishedOn != "" || n.AddedOn != "2026-11-15" || len(n.Authors) != 1 || n.Title != "Libro manual" {
		t.Fatal("normalización inventa o pierde datos")
	}
	negative := -1.0
	one := 1.0
	for name, input := range map[string]Input{
		"empty title":       {Status: "to-read"},
		"unknown status":    {Title: "Test", Status: "unknown"},
		"invalid date":      {Title: "Test", Status: "reading", StartedOn: "2026-02-30"},
		"reversed dates":    {Title: "Test", Status: "read", StartedOn: "2026-01-10", FinishedOn: "2026-01-01"},
		"orphan position":   {Title: "Test", Status: "to-read", SeriesPosition: &one},
		"negative position": {Title: "Test", Status: "to-read", SeriesName: "Serie", SeriesPosition: &negative},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := input.Normalize("2026-11-15"); err == nil {
				t.Fatal("debe rechazar datos inválidos")
			}
		})
	}
}
