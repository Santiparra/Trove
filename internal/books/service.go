package books

import (
	"context"
	"crypto/rand"
	"errors"
	"time"
)

// Repository describes operations needed by the use cases, not a generic database API.
type Repository interface {
	List(context.Context, string) ([]Book, error)
	Get(context.Context, string) (Book, error)
	Save(context.Context, Book) (Book, error)
	ListSeries(context.Context) ([]Series, error)
	Delete(context.Context, string) error
	UpdateStatus(context.Context, string, string, string) error
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
	// A state change alone does not fabricate dates or replace any other field.
	if err = s.repository.UpdateStatus(ctx, b.ID, b.Cycle.ID, status); err != nil {
		return Book{}, err
	}
	return s.repository.Get(ctx, id)
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repository.Delete(ctx, id) }

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
	book.Edition.Language, book.Edition.Format, book.Edition.Identifier = in.Language, in.Format, in.Identifier
	book.Cycle.Status, book.Cycle.StartedOn, book.Cycle.FinishedOn = in.Status, in.StartedOn, in.FinishedOn
	book.Series = nil
	if in.SeriesName != "" {
		book.Series = &Series{Name: in.SeriesName, Position: in.SeriesPosition}
	}
	return s.repository.Save(ctx, book)
}
