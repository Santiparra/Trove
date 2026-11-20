package providers

import (
	"context"
	"strings"
	"time"

	"github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/novelcatalog"
)

// NovelPublication never falls back to text search or to Open Library.
func (p *Hardcover) NovelPublication(ctx context.Context, isbn, language, format string) (novelcatalog.Supplement, error) {
	result := novelcatalog.Supplement{Warnings: []string{}}
	if !catalog.ValidISBN13(isbn) {
		return result, nil
	}
	var data struct {
		Categories []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"book_categories"`
		Editions []struct {
			hcEdition
			Book struct {
				CategoryID  *int   `json:"book_category_id"`
				ID          int64  `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Canonical   *int64 `json:"canonical_id"`
			} `json:"book"`
		} `json:"editions"`
	}
	fields := `id title isbn_10 isbn_13 asin pages edition_format release_date image { url } language { code2 } publisher { name } book { id title description canonical_id book_category_id }`
	variables := map[string]string{"long": isbn}
	query := `query TroveNovelISBN($long:String!){book_categories{id name} editions(limit:21,where:{isbn_13:{_eq:$long}}){` + fields + `}}`
	if short := catalog.ISBN10(isbn); short != "" {
		variables["short"] = short
		query = `query TroveNovelISBN($long:String!,$short:String!){book_categories{id name} editions(limit:21,where:{_or:[{isbn_13:{_eq:$long}},{isbn_10:{_eq:$short}}]}){` + fields + `}}`
	}
	if err := p.query(ctx, query, variables, &data); err != nil {
		return result, err
	}
	if len(data.Editions) > 20 {
		result.Warnings = append(result.Warnings, "Hardcover devolvió demasiadas coincidencias de ISBN; no se mezcló metadata.")
		return result, nil
	}
	matches := []novelcatalog.Supplement{}
	categories := map[int]string{}
	for _, category := range data.Categories {
		categories[category.ID] = category.Name
	}
	for _, raw := range data.Editions {
		if raw.Book.Canonical != nil {
			continue
		}
		if novelcatalog.ExplicitManga(raw.Title) || novelcatalog.ExplicitManga(raw.Book.Title) {
			result.Warnings = append(result.Warnings, "Hardcover identifica una coincidencia como manga; se descartó su metadata.")
			continue
		}
		category := ""
		if raw.Book.CategoryID != nil {
			category = categories[*raw.Book.CategoryID]
			if !novelcatalog.CompatibleHardcoverCategory(category) {
				result.Warnings = append(result.Warnings, "La categoría Hardcover no corresponde a una novela verificable; no se mezcló metadata.")
				continue
			}
		}
		confirmed, conflict := false, false
		for _, declared := range []string{raw.ISBN13, raw.ISBN10} {
			if declared == "" {
				continue
			}
			value := catalog.ISBN13(declared)
			if value == "" || value != isbn {
				conflict = true
			} else {
				confirmed = true
			}
		}
		if !confirmed || conflict {
			result.Warnings = append(result.Warnings, "Hardcover devolvió identificadores no equivalentes; no se mezcló metadata.")
			continue
		}
		ed := edition(raw.hcEdition)
		if ed.Pages != nil && *ed.Pages <= 0 {
			ed.Pages = nil
		}
		if ed.Language != language || (format == "print" && ed.Format != "paperback" && ed.Format != "hardcover") || (format != "print" && ed.Format != format) {
			continue
		}
		if raw.ID <= 0 || raw.Book.ID <= 0 {
			continue
		}
		date := strings.TrimSpace(ed.PublishedOn)
		if parsed, err := time.Parse(time.RFC3339Nano, date); err == nil {
			date = parsed.Format("2006-01-02")
		}
		if !novelcatalog.ValidPublicationDate(date) {
			date = ""
		}
		matches = append(matches, novelcatalog.Supplement{Category: category, WorkID: fmtID(raw.Book.ID), EditionID: fmtID(raw.ID), Description: plain(raw.Book.Description), CoverURL: ed.CoverURL, Publisher: ed.Publisher, PublishedOn: date, Pages: ed.Pages, Format: ed.Format})
	}
	if len(matches) == 1 {
		extra := matches[0]
		extra.Warnings = result.Warnings
		return extra, nil
	}
	if len(matches) > 1 {
		result.Warnings = append(result.Warnings, "Hay varias publicaciones Hardcover compatibles con ese ISBN; se mantuvo RanobeDB sin elegir una arbitrariamente.")
	}
	return result, nil
}
