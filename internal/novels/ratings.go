package novels

import (
	"context"
	"crypto/rand"
	"errors"

	"github.com/Santiparra/Trove/internal/books"
)

func validateRating(mode string, value *float64) error {
	if mode != "" && mode != "auto" && mode != "manual" {
		return errors.New("modo de puntaje inválido")
	}
	if mode == "" && value != nil {
		return errors.New("sin puntaje no puede guardar una nota")
	}
	if mode == "manual" && value == nil {
		return errors.New("indicá un puntaje manual")
	}
	return (books.Cycle{ID: "rating", Status: "read", Rating: value, RatingAutomatic: mode == "auto"}).Validate()
}

func (n *Novel) calculateRating() {
	if n.RatingMode == "" {
		n.Rating = nil
		return
	}
	if n.RatingMode != "auto" {
		return
	}
	var sum float64
	count := 0
	for _, v := range n.Volumes {
		value := v.Cycle.Rating
		if value == nil {
			for i := len(v.History) - 1; i >= 0; i-- {
				if v.History[i].Rating != nil {
					value = v.History[i].Rating
					break
				}
			}
		}
		if value != nil {
			sum += *value
			count++
		}
	}
	n.Rating = nil
	if count > 0 {
		value := sum / float64(count)
		n.Rating = &value
	}
}

func syncRating(r *ReadingRun, n Novel) {
	if r.RatingFrozen {
		return
	}
	r.RatingMode = n.RatingMode
	r.Cycle.Rating = n.Rating
	r.Cycle.RatingAutomatic = n.RatingMode == "auto"
	if n.Status == "read" || n.Status == "did-not-finish" {
		r.RatingFrozen = true
	}
}

func (s *Service) SetRating(ctx context.Context, id, mode string, value *float64) (Novel, error) {
	if mode == "auto" {
		value = nil
	}
	if err := validateRating(mode, value); err != nil {
		return Novel{}, err
	}
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	n.RatingMode, n.Rating = mode, value
	n.calculateRating()
	if len(n.Readings) == 0 {
		initialID := rand.Text()
		n.Readings = append(n.Readings, ReadingRun{ID: initialID, Cycle: books.Cycle{ID: initialID, Status: n.Status}, VolumeIDs: []string{}})
	}
	r := n.Run()
	if r == nil {
		r = &n.Readings[0]
	}
	// Changing the rating explicitly is a correction; later automatic recalculations
	// and provider refreshes must not rewrite the completed reading's snapshot.
	r.RatingFrozen = false
	syncRating(r, n)
	return s.persist(ctx, n)
}
