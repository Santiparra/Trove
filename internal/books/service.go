package books

import (
	"context"
	"crypto/rand"
	"errors"
	"sort"
	"time"
)

// Repository describes operations needed by the use cases, not a generic database API.
type Repository interface {
	List(context.Context, string) ([]Book, error)
	Get(context.Context, string) (Book, error)
	Save(context.Context, Book) (Book, error)
	ListSeries(context.Context) ([]Series, error)
	Delete(context.Context, string) error
	UpdateCycle(context.Context, string, Cycle) error
	UpdateNotes(context.Context, string, string) error
	UpdateFavorite(context.Context, string, bool) error
	UpdateOwned(context.Context, string, string, bool) error
}

type Service struct {
	repository Repository
	today      func() string
	newID      func() string
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, today: func() string { return time.Now().Format("2006-01-02") }, newID: rand.Text}
}

func (s *Service) List(ctx context.Context, query string) ([]Book, error) {
	return s.repository.List(ctx, query)
}
func (s *Service) Get(ctx context.Context, id string) (Book, error) { return s.repository.Get(ctx, id) }
func (s *Service) ListSeries(ctx context.Context) ([]Series, error) {
	return s.repository.ListSeries(ctx)
}

func (s *Service) ListByStatus(ctx context.Context, query, status string) ([]Book, error) {
	if status != "" && !ValidStatus(status) {
		return nil, errors.New("el estado de lectura no es válido")
	}
	items, err := s.repository.List(ctx, query)
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
	b, err := s.repository.Get(ctx, id)
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
	return s.repository.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repository.Delete(ctx, id) }

func (s *Service) SaveNotes(ctx context.Context, id, notes string) (Book, error) {
	if err := s.repository.UpdateNotes(ctx, id, notes); err != nil {
		return Book{}, err
	}
	return s.repository.Get(ctx, id)
}

func (s *Service) SetFavorite(ctx context.Context, id string, value bool) (Book, error) {
	if err := s.repository.UpdateFavorite(ctx, id, value); err != nil {
		return Book{}, err
	}
	return s.repository.Get(ctx, id)
}

func (s *Service) SetOwned(ctx context.Context, id string, value bool) (Book, error) {
	b, err := s.repository.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	if err = s.repository.UpdateOwned(ctx, id, b.Edition.ID, value); err != nil {
		return Book{}, err
	}
	return s.repository.Get(ctx, id)
}

func (s *Service) SeriesBooks(ctx context.Context, id string) ([]Book, error) {
	items, err := s.repository.List(ctx, "")
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
	items, err := s.repository.List(ctx, query)
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
	in, err := input.Normalize(s.today())
	if err != nil {
		return Book{}, err
	}
	book := Book{ID: s.newID(), Edition: Edition{ID: s.newID()}, Cycle: Cycle{ID: s.newID()}}
	if in.ID != "" {
		book, err = s.repository.Get(ctx, in.ID)
		if err != nil {
			return Book{}, err
		}
	}
	book.Title, book.Authors, book.AddedOn, book.Notes = in.Title, in.Authors, in.AddedOn, in.Notes
	book.Favorite, book.Edition.Owned = in.Favorite, in.Owned
	book.Edition.Language, book.Edition.Format, book.Edition.Identifier = in.Language, in.Format, in.Identifier
	book.Cycle.StartedOn, book.Cycle.FinishedOn = in.StartedOn, in.FinishedOn
	book.Cycle, err = book.Cycle.Transition(in.Status, s.today(), in.StartedOnEdited, in.FinishedOnEdited)
	if err != nil {
		return Book{}, err
	}
	book.Series = nil
	if in.SeriesName != "" {
		book.Series = &Series{Name: in.SeriesName, Position: in.SeriesPosition}
	}
	return s.repository.Save(ctx, book)
}
