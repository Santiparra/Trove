package catalog

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Primary interface {
	Search(context.Context, string) ([]Work, error)
	Editions(context.Context, string) ([]Edition, error)
	Resolve(context.Context, string) (Work, Edition, error)
}
type Supplement struct {
	EditionID   string
	WorkID      string
	CoverURL    string
	Description string
	Warnings    []string
}
type Secondary interface {
	Lookup(context.Context, string) (Supplement, error)
}
type Service struct {
	primary   Primary
	secondary Secondary
	now       func() time.Time
}

func New(primary Primary, secondary Secondary) *Service {
	return &Service{primary: primary, secondary: secondary, now: time.Now}
}

func (s *Service) Search(ctx context.Context, query string) (SearchResult, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 || len(query) > 300 {
		return SearchResult{}, errors.New("ingresá entre 2 y 300 caracteres")
	}
	items, err := s.primary.Search(ctx, query)
	return SearchResult{Works: items, Warnings: []string{}}, err
}
func (s *Service) Editions(ctx context.Context, id string) ([]Edition, error) {
	if !ValidID(Hardcover, id) {
		return nil, errors.New("obra de catálogo inválida")
	}
	return s.primary.Editions(ctx, id)
}
func (s *Service) Resolve(ctx context.Context, id string) (Snapshot, error) {
	if !ValidID(Hardcover, id) {
		return Snapshot{}, errors.New("edición de catálogo inválida")
	}
	work, edition, err := s.primary.Resolve(ctx, id)
	if err != nil {
		return Snapshot{}, err
	}
	at := s.now().UTC().Format(time.RFC3339Nano)
	ids, warnings := AcceptDeclared(edition.Declared)
	result := Snapshot{Provider: Hardcover, FetchedAt: at, Work: work, Edition: edition, Identifiers: ids, Description: work.Description, CoverURL: edition.CoverURL, Sources: map[string]Source{}, Warnings: warnings}
	for _, field := range []string{"language", "format", "publisher", "publishedOn", "pages"} {
		result.Sources[field] = Source{Hardcover, "edition", edition.ID, at}
	}
	if result.CoverURL == "" {
		result.CoverURL = work.CoverURL
		if result.CoverURL != "" {
			result.Warnings = append(result.Warnings, "La portada es de la obra; no se confirmó una portada específica de esta edición.")
		}
	}
	if result.Description != "" {
		result.Sources["description"] = Source{Hardcover, "work", work.ID, at}
	}
	if result.CoverURL != "" {
		entity, key := "edition", edition.ID
		if edition.CoverURL == "" {
			entity, key = "work", work.ID
		}
		result.Sources["cover"] = Source{Hardcover, entity, key, at}
	}
	// Only selected details cascade. Searching cards never calls the secondary provider.
	if s.secondary != nil && (result.Description == "" || result.CoverURL == "") {
		isbn := PreferredISBN(ids)
		if isbn != "" {
			extra, lookupErr := s.secondary.Lookup(ctx, isbn)
			if lookupErr != nil {
				result.Warnings = append(result.Warnings, "Open Library no está disponible; se conservó la información principal.")
			} else {
				result.Warnings = append(result.Warnings, extra.Warnings...)
				if result.CoverURL == "" && extra.CoverURL != "" {
					result.CoverURL = extra.CoverURL
					result.Sources["cover"] = Source{OpenLibrary, "edition", extra.EditionID, at}
				}
				if result.Description == "" && extra.Description != "" {
					result.Description = extra.Description
					result.Sources["description"] = Source{OpenLibrary, "work", extra.WorkID, at}
				}
			}
		} else {
			result.Warnings = append(result.Warnings, "No se consultó el secundario porque no hay un ISBN válido para una coincidencia segura.")
		}
	}
	if err = result.Validate(); err != nil {
		return Snapshot{}, err
	}
	return result, nil
}

func KeepMissing(old, fresh Snapshot) Snapshot {
	if fresh.Sources == nil {
		fresh.Sources = map[string]Source{}
	}
	retain := func(field string) {
		v, ok := old.Sources[field]
		if !ok {
			v = Source{old.Provider, "edition", old.Edition.ID, old.FetchedAt}
		}
		fresh.Sources[field] = v
	}
	if fresh.Description == "" && old.Description != "" {
		fresh.Description = old.Description
		retain("description")
	}
	if fresh.CoverURL == "" && old.CoverURL != "" {
		fresh.CoverURL = old.CoverURL
		retain("cover")
	}
	if fresh.Edition.Pages == nil && old.Edition.Pages != nil {
		fresh.Edition.Pages = old.Edition.Pages
		retain("pages")
	}
	if fresh.Edition.Publisher == "" {
		fresh.Edition.Publisher = old.Edition.Publisher
		retain("publisher")
	}
	if fresh.Edition.PublishedOn == "" {
		fresh.Edition.PublishedOn = old.Edition.PublishedOn
		retain("publishedOn")
	}
	return fresh
}
