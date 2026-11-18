package books

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/Santiparra/Trove/internal/catalog"
	"sort"
	"time"
)

// Repository exposes concrete reading operations; every list has an explicit medium.
type Repository interface {
	List(context.Context, string, string) ([]Book, error)
	Get(context.Context, string) (Book, error)
	Save(context.Context, Book) (Book, error)
	ListSeries(context.Context, string) ([]Series, error)
	Delete(context.Context, string) error
	UpdateCycle(context.Context, string, Cycle) error
	UpdateNotes(context.Context, string, string) error
	UpdateFavorite(context.Context, string, bool) error
	UpdateOwned(context.Context, string, string, bool) error
	UpdateCatalog(context.Context, string, string, *catalog.Snapshot, []catalog.Identifier) error
}

type Service struct {
	repository Repository
	today      func() string
	newID      func() string
	medium     string
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, today: func() string { return time.Now().Format("2006-01-02") }, newID: rand.Text, medium: MediumBooks}
}

// A scope is immutable. Switching a tab never changes another request's service context.
func (s *Service) ForMedium(medium string) (*Service, error) {
	if !ValidMedium(medium) {
		return nil, errors.New("el medio de lectura no es válido")
	}
	copy := *s
	copy.medium = medium
	return &copy, nil
}

func (s *Service) List(ctx context.Context, query string) ([]Book, error) {
	return s.repository.List(ctx, query, s.medium)
}
func (s *Service) Get(ctx context.Context, id string) (Book, error) {
	b, err := s.repository.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	if b.Medium != s.medium {
		return Book{}, ErrNotFound
	}
	return b, nil
}
func (s *Service) ListSeries(ctx context.Context) ([]Series, error) {
	return s.repository.ListSeries(ctx, s.medium)
}

func (s *Service) ListByStatus(ctx context.Context, query, status string) ([]Book, error) {
	if status != "" && !ValidStatus(status) {
		return nil, errors.New("el estado de lectura no es válido")
	}
	items, err := s.List(ctx, query)
	if err != nil {
		return nil, err
	}
	if status == "" {
		return items, nil
	}
	result := make([]Book, 0)
	for _, b := range items {
		if b.Cycle.Status == status {
			result = append(result, b)
		}
	}
	return result, nil
}

func (s *Service) ChangeStatus(ctx context.Context, id, status string) (Book, error) {
	if !ValidStatus(status) {
		return Book{}, errors.New("el estado de lectura no es válido")
	}
	b, err := s.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	cycle, err := b.Cycle.Transition(status, s.today(), false, false)
	if err != nil {
		return Book{}, err
	}
	if err = s.repository.UpdateCycle(ctx, b.ID, cycle); err != nil {
		return Book{}, err
	}
	return s.Get(ctx, id)
}
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	return s.repository.Delete(ctx, id)
}
func (s *Service) SaveNotes(ctx context.Context, id, notes string) (Book, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return Book{}, err
	}
	if err := s.repository.UpdateNotes(ctx, id, notes); err != nil {
		return Book{}, err
	}
	return s.Get(ctx, id)
}
func (s *Service) SetFavorite(ctx context.Context, id string, value bool) (Book, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return Book{}, err
	}
	if err := s.repository.UpdateFavorite(ctx, id, value); err != nil {
		return Book{}, err
	}
	return s.Get(ctx, id)
}
func (s *Service) SetOwned(ctx context.Context, id string, value bool) (Book, error) {
	b, err := s.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	if err = s.repository.UpdateOwned(ctx, id, b.Edition.ID, value); err != nil {
		return Book{}, err
	}
	return s.Get(ctx, id)
}
func (s *Service) SeriesBooks(ctx context.Context, id string) ([]Book, error) {
	items, err := s.List(ctx, "")
	if err != nil {
		return nil, err
	}
	result := make([]Book, 0)
	for _, b := range items {
		if b.Series != nil && b.Series.ID == id {
			result = append(result, b)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		a, b := result[i].Series.Position, result[j].Series.Position
		if a == nil {
			return false
		}
		if b == nil {
			return true
		}
		return *a < *b
	})
	return result, nil
}

type SearchResults struct {
	Matches []Book `json:"matches"`
	Total   int    `json:"total"`
}

func (s *Service) Search(ctx context.Context, query string) (SearchResults, error) {
	items, err := s.List(ctx, query)
	if err != nil {
		return SearchResults{}, err
	}
	result := SearchResults{Matches: items, Total: len(items)}
	if len(result.Matches) > 6 {
		result.Matches = result.Matches[:6]
	}
	return result, nil
}

func (s *Service) Save(ctx context.Context, input Input) (Book, error) {
	return s.save(ctx, input, nil)
}

func (s *Service) save(ctx context.Context, input Input, snapshot *catalog.Snapshot) (Book, error) {
	in, err := input.Normalize(s.today())
	if err != nil {
		return Book{}, err
	}
	if in.Medium != s.medium {
		return Book{}, errors.New("el registro no corresponde a este medio")
	}
	book := Book{ID: s.newID(), Medium: s.medium, Edition: Edition{ID: s.newID()}, Cycle: Cycle{ID: s.newID()}}
	if in.ID != "" {
		book, err = s.Get(ctx, in.ID)
		if err != nil {
			return Book{}, err
		}
	}
	book.Title, book.Authors, book.AddedOn, book.Notes = in.Title, in.Authors, in.AddedOn, in.Notes
	book.Favorite, book.Edition.Owned = in.Favorite, in.Owned
	book.Edition.Language, book.Edition.Format, book.Edition.Identifier = in.Language, in.Format, in.Identifier
	retained := []catalog.Identifier{}
	for _, v := range book.Edition.Identifiers {
		if v.Source != "manual" {
			retained = append(retained, v)
		}
	}
	book.Edition.Identifiers = catalog.MergeIdentifiers(catalog.NormalizeIdentifier(in.Identifier, "manual"), retained)
	if snapshot != nil {
		book.Edition.Catalog = snapshot
		book.Edition.Identifiers = catalog.MergeIdentifiers(book.Edition.Identifiers, snapshot.Identifiers)
	}
	book.Cycle.StartedOn, book.Cycle.FinishedOn = in.StartedOn, in.FinishedOn
	book.Cycle, err = book.Cycle.Transition(in.Status, s.today(), in.StartedOnEdited, in.FinishedOnEdited)
	if err != nil {
		return Book{}, err
	}
	book.Series = nil
	if in.SeriesName != "" {
		book.Series = &Series{Name: in.SeriesName, Position: in.SeriesPosition, Medium: s.medium}
	}
	return s.repository.Save(ctx, book)
}
