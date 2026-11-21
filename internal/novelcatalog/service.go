package novelcatalog

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Santiparra/Trove/internal/catalog"
)

type Primary interface {
	Search(context.Context, string, string, int) (SearchResult, error)
	Series(context.Context, string) (Series, error)
	Book(context.Context, string, string) (Book, error)
	Release(context.Context, string, string) (Publication, error)
}
type Secondary interface {
	NovelPublication(context.Context, string, string, string) (Supplement, error)
}
type Service struct {
	mu        sync.Mutex
	books     map[string]*bookEntry
	primary   Primary
	secondary Secondary
}

func New(primary Primary, secondary Secondary) *Service {
	return &Service{primary: primary, secondary: secondary, books: map[string]*bookEntry{}}
}
func (s *Service) Search(ctx context.Context, q, authorID string, page int) (SearchResult, error) {
	return s.primary.Search(ctx, q, authorID, page)
}
func (s *Service) Series(ctx context.Context, id string) (Series, error) {
	return s.primary.Series(ctx, id)
}
func (s *Service) Releases(ctx context.Context, series Series, number int) ([]Release, error) {
	ref, ok := series.Ref(number)
	if !ok {
		return nil, errors.New("no hay un volumen principal confirmado para ese número")
	}
	book, err := s.book(ctx, series.ID, ref.ID)
	if err != nil {
		return nil, err
	}
	return book.Releases, nil
}
func (s *Service) Resolve(ctx context.Context, series Series, number int, releaseID string) (Publication, error) {
	explicitRelease := releaseID != ""
	if err := series.Validate(); err != nil {
		return Publication{}, err
	}
	ref, ok := series.Ref(number)
	if !ok {
		return Publication{}, errors.New("RanobeDB todavía no confirma ese volumen principal; el progreso local se conserva")
	}
	book, err := s.book(ctx, series.ID, ref.ID)
	if err != nil {
		return Publication{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	base := Publication{SeriesID: series.ID, BookID: ref.ID, Number: number, Order: ref.Order, Title: book.Title, OriginalTitle: book.OriginalTitle, Description: book.Description, CoverURL: book.CoverURL, Evidence: "ranobedb-main", FetchedAt: now, Sources: map[string]Source{}, Warnings: []string{}}
	bookSource := Source{Provider: RanobeDB, Entity: "book", ID: book.ID, FetchedAt: now}
	if book.FetchedAt != "" {
		bookSource.FetchedAt = book.FetchedAt
	}
	if base.Description != "" {
		base.Sources["description"] = bookSource
	}
	if base.CoverURL != "" {
		base.Sources["cover"] = bookSource
	}
	if releaseID == "" {
		base.Label = ref.Label
		if base.Label == "" {
			base.Label = PrintedLabel(book.Title, book.OriginalTitle)
		}
		lang := series.Language
		if lang == "" {
			lang = "en"
		}
		format := series.Format
		if format == "" {
			format = "print"
		}
		for _, r := range book.Releases {
			if r.Language == lang && r.Format == format {
				releaseID = r.ID
				break
			}
		}
	}
	if releaseID != "" {
		listed := false
		for _, r := range book.Releases {
			if r.ID == releaseID {
				listed = true
				break
			}
		}
		if !listed {
			return Publication{}, errors.New("la publicación no pertenece al volumen seleccionado")
		}
		publication, err := s.primary.Release(ctx, book.ID, releaseID)
		if err != nil {
			if !explicitRelease {
				base.Warnings = append(base.Warnings, "No se pudo verificar la publicación preferida. Se conserva solo metadata del volumen; elegí otra publicación.")
				return base, base.Validate()
			}
			return Publication{}, err
		}
		for _, listed := range book.Releases {
			if listed.ID == releaseID && (listed.Language != publication.Language || listed.Format != publication.ReleaseFormat || (listed.ISBN != "" && listed.ISBN != publication.ISBN)) {
				if explicitRelease {
					return Publication{}, errors.New("la identidad, idioma o formato de la publicación cambió; volvé a seleccionarla")
				}
				base.Warnings = append(base.Warnings, "La publicación preferida cambió; no se asumió otra edición.")
				return base, base.Validate()
			}
		}
		publication.SeriesID, publication.BookID, publication.Number, publication.Order = series.ID, book.ID, number, ref.Order
		publication.Title, publication.OriginalTitle = book.Title, book.OriginalTitle
		publication.Label = ref.Label
		if publication.Label == "" {
			publication.Label = PrintedLabel(book.Title, book.OriginalTitle)
		}
		publication.FetchedAt, publication.Evidence = now, "ranobedb-main"
		publication.Sources = map[string]Source{}
		publication.Warnings = []string{}
		rSource := Source{Provider: RanobeDB, Entity: "release", ID: releaseID, FetchedAt: now}
		for _, field := range []string{"language", "format", "isbn", "publisher", "publishedOn", "pages", "cover", "description"} {
			publication.Sources[field] = rSource
		}
		if publication.Description == "" {
			publication.Description = base.Description
			publication.Sources["description"] = bookSource
		}
		// An edition-specific cover is preferred to the volume-level fallback.
		if publication.ISBN != "" && s.secondary != nil && (publication.CoverURL == "" || publication.Description == "" || publication.Pages == nil || publication.Publisher == "" || publication.PublishedOn == "") {
			extra, e := s.secondary.NovelPublication(ctx, publication.ISBN, publication.Language, publication.ReleaseFormat)
			if e != nil {
				publication.Warnings = append(publication.Warnings, "Hardcover no pudo complementar esta publicación; se conservó RanobeDB.")
			} else {
				publication.Warnings = append(publication.Warnings, extra.Warnings...)
				if extra.EditionID != "" {
					publication.HardcoverCategory = extra.Category
					if extra.Category != "" {
						publication.Sources["classification"] = Source{Provider: catalog.Hardcover, Entity: "work", ID: extra.WorkID, FetchedAt: now}
					}
					hc := Source{Provider: catalog.Hardcover, Entity: "edition", ID: extra.EditionID, FetchedAt: now}
					if publication.CoverURL == "" && extra.CoverURL != "" {
						publication.CoverURL = extra.CoverURL
						publication.Sources["cover"] = hc
					}
					if publication.Pages == nil && extra.Pages != nil {
						publication.Pages = extra.Pages
						publication.Sources["pages"] = hc
					}
					if publication.Publisher == "" && extra.Publisher != "" {
						publication.Publisher = extra.Publisher
						publication.Sources["publisher"] = hc
					}
					if publication.PublishedOn == "" && extra.PublishedOn != "" {
						publication.PublishedOn = extra.PublishedOn
						publication.Sources["publishedOn"] = hc
					}
					if publication.Description == "" && extra.Description != "" {
						publication.Description = extra.Description
						publication.Sources["description"] = Source{Provider: catalog.Hardcover, Entity: "work", ID: extra.WorkID, FetchedAt: now}
					}
					if publication.ReleaseFormat == "print" && extra.Format != "" {
						publication.Format = extra.Format
						publication.Sources["format"] = hc
					}
				}
			}
		}
		if publication.CoverURL == "" {
			publication.CoverURL = base.CoverURL
			publication.Sources["cover"] = bookSource
		}
		base = publication
	} else {
		language, format := series.Language, series.Format
		if language == "" {
			language = "en"
		}
		if format == "" {
			format = "print"
		}
		base.Warnings = append(base.Warnings, fmt.Sprintf("Sin publicación coincidente para %s / %s. Se muestra información del volumen; podés elegir otra publicación. La ausencia de ISBN solo limita el complemento Hardcover.", language, format))
	}
	if err := base.Validate(); err != nil {
		return Publication{}, err
	}
	return base, nil
}
func KeepMissing(old, fresh Publication) Publication {
	if old.BookID != fresh.BookID || old.ReleaseID != fresh.ReleaseID {
		return fresh
	}
	keep := func(field string) {
		if source, ok := old.Sources[field]; ok {
			fresh.Sources[field] = source
		}
	}
	if fresh.Description == "" {
		fresh.Description = old.Description
		keep("description")
	}
	if fresh.CoverURL == "" {
		fresh.CoverURL = old.CoverURL
		keep("cover")
	}
	if fresh.Pages == nil {
		fresh.Pages = old.Pages
		keep("pages")
	}
	if fresh.Publisher == "" {
		fresh.Publisher = old.Publisher
		keep("publisher")
	}
	if fresh.PublishedOn == "" {
		fresh.PublishedOn = old.PublishedOn
		keep("publishedOn")
	}
	return fresh
}
