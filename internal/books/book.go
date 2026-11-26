// Package books defines the local book model without UI, provider or SQL dependencies.
package books

import (
	"errors"
	"github.com/Santiparra/Trove/internal/catalog"
	"math"
	"strings"
	"time"
)

var ErrNotFound = errors.New("el registro no existe en este medio; actualizá el listado")

const MediumBooks = "books"
const MediumLightNovels = "light-novels"

func ValidMedium(medium string) bool { return medium == MediumBooks || medium == MediumLightNovels }

func ValidStatus(status string) bool {
	switch status {
	case "to-read", "up-next", "reading", "read", "paused", "did-not-finish":
		return true
	}
	return false
}

func ValidFormat(format string) bool {
	switch format {
	case "", "paperback", "hardcover", "digital", "audio", "other":
		return true
	}
	return false
}

type Series struct {
	Medium   string   `json:"medium"`
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position *float64 `json:"position"`
}

type Edition struct {
	Identifiers []catalog.Identifier `json:"identifiers"`
	Catalog     *catalog.Snapshot    `json:"catalog"`
	ID          string               `json:"id"`
	Language    string               `json:"language"`
	Format      string               `json:"format"`
	Identifier  string               `json:"identifier"`
	Owned       bool                 `json:"owned"`
}

// Cycle is a single reading. Rereading creates a new cycle rather than resetting it.
type Cycle struct {
	ID              string          `json:"id"`
	Status          string          `json:"status"`
	StartedOn       string          `json:"startedOn"`
	FinishedOn      string          `json:"finishedOn"`
	Rating          *float64        `json:"rating,omitempty"`
	RatingAutomatic bool            `json:"ratingAutomatic,omitempty"`
	Edition         *ReadingEdition `json:"edition,omitempty"`
	RunID           string          `json:"runID,omitempty"`
}

// ReadingEdition is the edition used for this reading, independent of ownership.
type ReadingEdition struct {
	Language   string `json:"language"`
	Format     string `json:"format"`
	Identifier string `json:"identifier"`
}

func (e Edition) ReadingEdition() *ReadingEdition {
	return &ReadingEdition{Language: e.Language, Format: e.Format, Identifier: e.Identifier}
}

func (c Cycle) Validate() error {
	if c.ID == "" || !ValidStatus(c.Status) {
		return errors.New("identidad o estado de lectura inválidos")
	}
	for _, d := range []string{c.StartedOn, c.FinishedOn} {
		if d != "" {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return errors.New("fecha de lectura inválida")
			}
		}
	}
	if c.StartedOn != "" && c.FinishedOn != "" && c.FinishedOn < c.StartedOn {
		return errors.New("la fecha de finalización no puede ser anterior al inicio")
	}
	if c.Rating != nil {
		v := *c.Rating
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 1 || v > 10 || (!c.RatingAutomatic && math.Round(v*4) != v*4) {
			return errors.New("el puntaje debe estar entre 1 y 10 en pasos de 0,25")
		}
	}
	if c.Edition != nil && !ValidFormat(c.Edition.Format) {
		return errors.New("formato de lectura inválido")
	}
	return nil
}

func (c Cycle) HasReading() bool {
	return c.StartedOn != "" || c.FinishedOn != "" || c.Rating != nil || c.Status == "read" || c.Status == "reading" || c.Status == "paused" || c.Status == "did-not-finish"
}

// ReadingInput edits a particular cycle. An empty ID adds history, never replaces current.
type ReadingInput struct {
	ID         string          `json:"id"`
	Status     string          `json:"status"`
	StartedOn  string          `json:"startedOn"`
	FinishedOn string          `json:"finishedOn"`
	Rating     *float64        `json:"rating"`
	Edition    *ReadingEdition `json:"edition"`
}

func (in ReadingInput) Cycle(id string) (Cycle, error) {
	c := Cycle{ID: id, Status: in.Status, StartedOn: in.StartedOn, FinishedOn: in.FinishedOn, Rating: in.Rating, Edition: in.Edition}
	return c, c.Validate()
}

type Book struct {
	Medium   string   `json:"medium"`
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Authors  []string `json:"authors"`
	Series   *Series  `json:"series"`
	Edition  Edition  `json:"edition"`
	Cycle    Cycle    `json:"cycle"`
	History  []Cycle  `json:"history,omitempty"`
	Notes    string   `json:"notes"`
	AddedOn  string   `json:"addedOn"`
	Favorite bool     `json:"favorite"`
}

// Input intentionally excludes database, edition and cycle identifiers.
type Input struct {
	Medium           string   `json:"medium"`
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Authors          []string `json:"authors"`
	SeriesName       string   `json:"seriesName"`
	SeriesPosition   *float64 `json:"seriesPosition"`
	Language         string   `json:"language"`
	Format           string   `json:"format"`
	Identifier       string   `json:"identifier"`
	Status           string   `json:"status"`
	Notes            string   `json:"notes"`
	AddedOn          string   `json:"addedOn"`
	StartedOn        string   `json:"startedOn"`
	FinishedOn       string   `json:"finishedOn"`
	Favorite         bool     `json:"favorite"`
	Owned            bool     `json:"owned"`
	StartedOnEdited  bool     `json:"startedOnEdited"`
	FinishedOnEdited bool     `json:"finishedOnEdited"`
}

// Transition is for live local actions only, not imported/restored historical states.
// Explicit date edits take precedence, including deliberately clearing a date.
func (c Cycle) Transition(next, today string, editStart, editFinish bool) (Cycle, error) {
	if !ValidStatus(next) {
		return c, errors.New("el estado de lectura no es válido")
	}
	if c.Status != next {
		if next == "reading" && c.StartedOn == "" && !editStart {
			c.StartedOn = today
		}
		if next == "read" && c.StartedOn != "" && c.FinishedOn == "" && !editFinish {
			c.FinishedOn = today
		}
	}
	c.Status = next
	if c.StartedOn != "" && c.FinishedOn != "" && c.FinishedOn < c.StartedOn {
		return c, errors.New("la fecha de finalización no puede ser anterior al inicio")
	}
	return c, nil
}

func (in Input) Normalize(today string) (Input, error) {
	if in.Medium == "" {
		in.Medium = MediumBooks
	}
	if !ValidMedium(in.Medium) {
		return in, errors.New("el medio de lectura no es válido")
	}
	in.ID = strings.TrimSpace(in.ID)
	in.Title = strings.TrimSpace(in.Title)
	in.SeriesName = strings.Join(strings.Fields(in.SeriesName), " ")
	in.Language = strings.TrimSpace(in.Language)
	in.Identifier = strings.TrimSpace(in.Identifier)
	authors := make([]string, 0, len(in.Authors))
	for _, author := range in.Authors {
		if a := strings.TrimSpace(author); a != "" {
			authors = append(authors, a)
		}
	}
	in.Authors = authors
	if in.Title == "" {
		return in, errors.New("el título es obligatorio")
	}
	if in.AddedOn == "" {
		in.AddedOn = today
	}
	if !ValidStatus(in.Status) {
		return in, errors.New("el estado de lectura no es válido")
	}
	if !ValidFormat(in.Format) {
		return in, errors.New("el formato de edición no es válido")
	}
	for _, value := range []string{in.AddedOn, in.StartedOn, in.FinishedOn} {
		if value != "" {
			if _, err := time.Parse("2006-01-02", value); err != nil {
				return in, errors.New("las fechas deben ser fechas válidas en formato AAAA-MM-DD")
			}
		}
	}
	if in.StartedOn != "" && in.FinishedOn != "" && in.StartedOn > in.FinishedOn {
		return in, errors.New("la fecha de finalización no puede ser anterior al inicio")
	}
	if in.SeriesPosition != nil {
		if in.SeriesName == "" {
			return in, errors.New("indicá una serie antes de su número de orden")
		}
		if n := *in.SeriesPosition; math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
			return in, errors.New("el número de serie debe ser un número no negativo")
		}
	}
	return in, nil
}
