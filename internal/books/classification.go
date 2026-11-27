package books

import (
	"context"
	"github.com/Santiparra/Trove/internal/classification"
)

func withClassification(b Book) Book {
	var external *classification.Set
	if b.Edition.Catalog != nil {
		external = b.Edition.Catalog.Work.Classification
	}
	b.EffectiveClassification = classification.Effective(b.Classification, external)
	return b
}
func (s *Service) SetClassification(ctx context.Context, id string, input classification.Input) (Book, error) {
	in, err := classification.Normalize(input)
	if err != nil {
		return Book{}, err
	}
	b, err := s.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	b.Classification = &in
	b, err = s.repository.Save(ctx, b)
	return withClassification(b), err
}
