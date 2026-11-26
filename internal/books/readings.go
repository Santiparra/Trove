package books

import (
	"context"
	"errors"
)

func (s *Service) Reread(ctx context.Context, id string) (Book, error) {
	b, err := s.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	completedBefore := false
	for _, c := range b.History {
		if c.Status == "read" {
			completedBefore = true
		}
	}
	if b.Cycle.Status != "read" && !(b.Cycle.Status == "did-not-finish" && completedBefore) {
		return Book{}, errors.New("terminá la lectura actual antes de releer")
	}
	if b.Cycle.Edition == nil {
		b.Cycle.Edition = b.Edition.ReadingEdition()
	}
	b.History = append(b.History, b.Cycle)
	b.Cycle = Cycle{ID: s.newID(), Status: "reading", StartedOn: s.today(), Edition: b.Edition.ReadingEdition()}
	return s.repository.Save(ctx, b)
}

func (s *Service) SaveReading(ctx context.Context, id string, in ReadingInput) (Book, error) {
	b, err := s.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	cycleID := in.ID
	if cycleID == "" {
		cycleID = s.newID()
	}
	c, err := in.Cycle(cycleID)
	if err != nil {
		return Book{}, err
	}
	if c.Edition == nil && in.ID != "" {
		if in.ID == b.Cycle.ID {
			c.Edition = b.Cycle.Edition
		}
		for _, old := range b.History {
			if old.ID == in.ID {
				c.Edition = old.Edition
				break
			}
		}
	}
	if in.ID == "" {
		if c.Status != "read" && c.Status != "did-not-finish" {
			return Book{}, errors.New("una lectura pasada debe estar terminada o sin terminar")
		}
		b.History = append(b.History, c)
	} else if in.ID == b.Cycle.ID {
		b.Cycle = c
	} else {
		found := false
		for i := range b.History {
			if b.History[i].ID == in.ID {
				b.History[i] = c
				found = true
				break
			}
		}
		if !found {
			return Book{}, ErrNotFound
		}
	}
	return s.repository.Save(ctx, b)
}
