// Package novelcatalog defines provider-neutral novel identities and publication snapshots.
package novelcatalog

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
)

const RanobeDB = "ranobedb"
const MaxBooks = 10000

var numericID = regexp.MustCompile(`^[1-9][0-9]{0,17}$`)
var languageCode = regexp.MustCompile(`^[a-z]{2,3}(?:-[A-Za-z]{2,8})?$`)

// A disambiguation annotation is evidence; a word inside a genuine title is not.
var mangaMarker = regexp.MustCompile(`(?i)[(\[（【]\s*(?:the\s+)?(?:manga|graphic\s+novel|comic|漫画)(?:\s+(?:edition|adaptation|version|series))?\s*[)\]）】]`)

func ValidID(id string) bool          { return numericID.MatchString(id) }
func ExplicitManga(title string) bool { return mangaMarker.MatchString(title) }
func CompatibleHardcoverCategory(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "book", "light novel", "novella", "short story", "web novel", "collection":
		return true
	}
	return false
}

type Author struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type Candidate struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	OriginalTitle    string `json:"originalTitle"`
	CoverURL         string `json:"coverURL"`
	MatchedBookID    string `json:"matchedBookID"`
	MatchedReleaseID string `json:"matchedReleaseID"`
}
type SearchResult struct {
	Items    []Candidate `json:"items"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	Pages    int         `json:"pages"`
	Warnings []string    `json:"warnings"`
}
type BookRef struct {
	ID            string  `json:"id"`
	Number        int     `json:"number"`
	Order         float64 `json:"order"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"originalTitle"`
	CoverURL      string  `json:"coverURL"`
}
type Series struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	OriginalTitle string    `json:"originalTitle"`
	Description   string    `json:"description"`
	CoverURL      string    `json:"coverURL"`
	Authors       []Author  `json:"authors"`
	Books         []BookRef `json:"books"`
	FetchedAt     string    `json:"fetchedAt"`
	Language      string    `json:"language"`
	Format        string    `json:"format"`
}
type Release struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Language    string `json:"language"`
	Format      string `json:"format"`
	ISBN        string `json:"isbn"`
	PublishedOn string `json:"publishedOn"`
}
type Book struct {
	ID            string
	SeriesID      string
	Title         string
	OriginalTitle string
	Description   string
	CoverURL      string
	Releases      []Release
}
type Source struct {
	Provider  string `json:"provider"`
	Entity    string `json:"entity"`
	ID        string `json:"id"`
	FetchedAt string `json:"fetchedAt"`
}
type Publication struct {
	HardcoverCategory string            `json:"hardcoverCategory,omitempty"`
	ReleaseTitle      string            `json:"releaseTitle"`
	SeriesID          string            `json:"seriesID"`
	BookID            string            `json:"bookID"`
	ReleaseID         string            `json:"releaseID"`
	Number            int               `json:"number"`
	Order             float64           `json:"order"`
	Title             string            `json:"title"`
	OriginalTitle     string            `json:"originalTitle"`
	Language          string            `json:"language"`
	ReleaseFormat     string            `json:"releaseFormat"`
	Format            string            `json:"format"`
	ISBN              string            `json:"isbn"`
	Description       string            `json:"description"`
	CoverURL          string            `json:"coverURL"`
	Publisher         string            `json:"publisher"`
	PublishedOn       string            `json:"publishedOn"`
	Pages             *int              `json:"pages"`
	FetchedAt         string            `json:"fetchedAt"`
	Evidence          string            `json:"evidence"`
	Sources           map[string]Source `json:"sources"`
	Warnings          []string          `json:"warnings"`
}
type Supplement struct {
	Category    string
	WorkID      string
	EditionID   string
	Description string
	CoverURL    string
	Publisher   string
	PublishedOn string
	Pages       *int
	Format      string
	Warnings    []string
}

func validTime(value string) bool { _, err := time.Parse(time.RFC3339Nano, value); return err == nil }
func ValidPublicationDate(value string) bool {
	if value == "" {
		return true
	}
	layout := ""
	switch len(value) {
	case 4:
		layout = "2006"
	case 7:
		layout = "2006-01"
	case 10:
		layout = "2006-01-02"
	default:
		return false
	}
	if strings.HasPrefix(value, "0000") || strings.HasPrefix(value, "9999") {
		return false
	}
	_, err := time.Parse(layout, value)
	return err == nil
}
func (s Series) Validate() error {
	if s.Language != "" && !languageCode.MatchString(s.Language) {
		return errors.New("idioma de publicación inválido")
	}
	if !ValidID(s.ID) || strings.TrimSpace(s.Title) == "" || ExplicitManga(s.Title) || ExplicitManga(s.OriginalTitle) || !validTime(s.FetchedAt) || !catalog.ValidCoverURL(s.CoverURL) || len(s.Books) == 0 || len(s.Books) > MaxBooks {
		return errors.New("identidad de novela RanobeDB no verificable")
	}
	ids := map[string]bool{}
	for index, b := range s.Books {
		if !ValidID(b.ID) || ids[b.ID] || b.Number != index+1 || math.IsNaN(b.Order) || math.IsInf(b.Order, 0) || b.Order < 0 || (index > 0 && b.Order <= s.Books[index-1].Order) || ExplicitManga(b.Title) || ExplicitManga(b.OriginalTitle) || !catalog.ValidCoverURL(b.CoverURL) {
			return errors.New("orden o identidad de volúmenes ambiguos")
		}
		ids[b.ID] = true
	}
	for _, a := range s.Authors {
		if !ValidID(a.ID) || strings.TrimSpace(a.Name) == "" {
			return errors.New("autoría RanobeDB inválida")
		}
	}
	if s.Format != "" && s.Format != "print" && s.Format != "digital" && s.Format != "audio" {
		return errors.New("formato de publicación inválido")
	}
	return nil
}
func (p Publication) Validate() error {
	if p.HardcoverCategory != "" && !CompatibleHardcoverCategory(p.HardcoverCategory) {
		return errors.New("categoría secundaria incompatible con novela")
	}
	if !ValidPublicationDate(p.PublishedOn) || math.IsNaN(p.Order) || math.IsInf(p.Order, 0) || p.Order < 0 {
		return errors.New("fecha u orden bibliográfico inválido")
	}
	if p.ReleaseID != "" && (!languageCode.MatchString(p.Language) || (p.ReleaseFormat != "print" && p.ReleaseFormat != "digital" && p.ReleaseFormat != "audio")) {
		return errors.New("idioma o formato de publicación no verificables")
	}
	if p.ReleaseID == "" && (p.ISBN != "" || p.Language != "" || p.ReleaseFormat != "") {
		return errors.New("no se puede atribuir una edición sin publicación")
	}
	if !ValidID(p.SeriesID) || !ValidID(p.BookID) || (p.ReleaseID != "" && !ValidID(p.ReleaseID)) || p.Number < 1 || p.Number > MaxBooks || p.Evidence != "ranobedb-main" || !validTime(p.FetchedAt) || !catalog.ValidCoverURL(p.CoverURL) || ExplicitManga(p.Title) || ExplicitManga(p.OriginalTitle) || ExplicitManga(p.ReleaseTitle) || !books.ValidFormat(p.Format) || (p.ISBN != "" && !catalog.ValidISBN13(p.ISBN)) || (p.Pages != nil && *p.Pages <= 0) {
		return errors.New("publicación de novela no verificable")
	}
	for _, s := range p.Sources {
		if (s.Provider != RanobeDB && s.Provider != catalog.Hardcover) || !ValidID(s.ID) || !validTime(s.FetchedAt) {
			return errors.New("procedencia bibliográfica inválida")
		}
		if s.Provider == RanobeDB && s.Entity != "series" && s.Entity != "book" && s.Entity != "release" {
			return errors.New("tipo de origen RanobeDB inválido")
		}
		if s.Provider == catalog.Hardcover && s.Entity != "work" && s.Entity != "edition" {
			return errors.New("tipo de origen Hardcover inválido")
		}
	}
	return nil
}
func (s Series) Ref(number int) (BookRef, bool) {
	if number < 1 || number > len(s.Books) {
		return BookRef{}, false
	}
	return s.Books[number-1], true
}
func SameOrder(old, fresh Series) bool {
	if old.ID != fresh.ID || len(fresh.Books) < len(old.Books) {
		return false
	}
	for i, b := range old.Books {
		if fresh.Books[i].ID != b.ID {
			return false
		}
	}
	return true
}
