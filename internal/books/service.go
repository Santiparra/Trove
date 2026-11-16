package books

import (
	"context"
	"crypto/rand"
	"time"
)

// Repository describes operations needed by the use cases, not a generic database API.
type Repository interface {
	List(context.Context, string) ([]Book, error)
	Get(context.Context, string) (Book, error)
	Save(context.Context, Book) (Book, error)
	ListSeries(context.Context) ([]Series, error)
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
