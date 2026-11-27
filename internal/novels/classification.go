package novels

import (
	"context"
	"github.com/Santiparra/Trove/internal/classification"
)

func withClassification(n Novel) Novel {
	var external *classification.Set
	if n.Catalog != nil {
		external = n.Catalog.Classification
	}
	n.EffectiveClassification = classification.Effective(n.Classification, external)
	for i := range n.Volumes {
		v := &n.Volumes[i]
		external = nil
		if v.Catalog != nil {
			external = v.Catalog.Classification
		}
		v.EffectiveClassification = classification.Effective(v.Classification, external)
	}
	return n
}

// An empty volumeID edits only the novel; otherwise the volume must belong to it.
func (s *Service) SetClassification(ctx context.Context, id, volumeID string, input classification.Input) (Novel, error) {
	in, err := classification.Normalize(input)
	if err != nil {
		return Novel{}, err
	}
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	if volumeID == "" {
		n.Classification = &in
		return s.persist(ctx, n)
	}
	for i := range n.Volumes {
		if n.Volumes[i].ID == volumeID {
			n.Volumes[i].Classification = &in
			return s.persist(ctx, n)
		}
	}
	return Novel{}, ErrNotFound
}
