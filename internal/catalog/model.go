// Package catalog defines bibliographic contracts, independent of provider JSON and SQL.
package catalog

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const Hardcover = "hardcover"
const OpenLibrary = "openlibrary"

type Identifier struct {
	Scheme  string `json:"scheme"`
	Value   string `json:"value"`
	Source  string `json:"source"`
	Derived bool   `json:"derived"`
}
type Source struct {
	Provider  string `json:"provider"`
	Entity    string `json:"entity"`
	ID        string `json:"id"`
	FetchedAt string `json:"fetchedAt"`
}
type Series struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position *float64 `json:"position"`
}
type Work struct {
	MatchedEditionID string   `json:"matchedEditionId,omitempty"`
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Authors          []string `json:"authors"`
	Description      string   `json:"description"`
	CoverURL         string   `json:"coverURL"`
	Series           []Series `json:"series"`
}
type Edition struct {
	ID          string       `json:"id"`
	WorkID      string       `json:"workId"`
	Title       string       `json:"title"`
	Language    string       `json:"language"`
	Format      string       `json:"format"`
	Publisher   string       `json:"publisher"`
	PublishedOn string       `json:"publishedOn"`
	Pages       *int         `json:"pages"`
	CoverURL    string       `json:"coverURL"`
	Declared    []Identifier `json:"declared"`
}
type Snapshot struct {
	Provider    string            `json:"provider"`
	FetchedAt   string            `json:"fetchedAt"`
	Work        Work              `json:"work"`
	Edition     Edition           `json:"edition"`
	Identifiers []Identifier      `json:"identifiers"`
	Description string            `json:"description"`
	CoverURL    string            `json:"coverURL"`
	Sources     map[string]Source `json:"sources"`
	Warnings    []string          `json:"warnings"`
}
type SearchResult struct {
	Works    []Work   `json:"works"`
	Warnings []string `json:"warnings"`
}

var numericID = regexp.MustCompile(`^[1-9][0-9]{0,17}$`)
var olID = regexp.MustCompile(`^OL[0-9]+[MW]$`)

func ValidID(provider, id string) bool {
	if provider == Hardcover {
		return numericID.MatchString(id)
	}
	return provider == OpenLibrary && olID.MatchString(id)
}
func ValidCoverURL(value string) bool {
	if value == "" {
		return true
	}
	u, err := url.Parse(value)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.Fragment != "" {
		return false
	}
	switch strings.ToLower(u.Hostname()) {
	case "assets.hardcover.app", "images.hardcover.app", "covers.openlibrary.org":
		return true
	}
	return false
}
func (s Snapshot) Validate() error {
	if s.Provider != Hardcover || !ValidID(Hardcover, s.Work.ID) || !ValidID(Hardcover, s.Edition.ID) || s.Edition.WorkID != s.Work.ID {
		return errors.New("referencias de catálogo inválidas")
	}
	if strings.TrimSpace(s.Work.Title) == "" || strings.TrimSpace(s.Edition.Title) == "" {
		return errors.New("el catálogo no aporta título de obra/edición")
	}
	if _, err := time.Parse(time.RFC3339Nano, s.FetchedAt); err != nil {
		return errors.New("fecha de snapshot inválida")
	}
	if !ValidCoverURL(s.CoverURL) || !ValidCoverURL(s.Work.CoverURL) || !ValidCoverURL(s.Edition.CoverURL) {
		return errors.New("destino de portada no permitido")
	}
	if s.Edition.Pages != nil && *s.Edition.Pages < 0 {
		return errors.New("páginas inválidas")
	}
	for _, v := range s.Identifiers {
		if !ValidIdentifier(v) {
			return errors.New("identificador bibliográfico inválido")
		}
	}
	for _, v := range s.Sources {
		if !ValidID(v.Provider, v.ID) {
			return errors.New("origen de metadata inválido")
		}
		if v.Entity != "work" && v.Entity != "edition" {
			return errors.New("tipo de origen inválido")
		}
		if _, err := time.Parse(time.RFC3339Nano, v.FetchedAt); err != nil {
			return errors.New("fecha de origen inválida")
		}
	}
	return nil
}
