package novelcatalog

import (
	"context"
	"time"
)

type bookEntry struct {
	book  Book
	err   error
	done  chan struct{}
	until time.Time
}
type freshBookKey struct{}

// Coalesce only selected-book details. Search results and personal library data are never cached here.
func (s *Service) book(ctx context.Context, seriesID, bookID string) (Book, error) {
	key := seriesID + ":" + bookID
	fresh, _ := ctx.Value(freshBookKey{}).(bool)
	s.mu.Lock()
	if existing, ok := s.books[key]; ok {
		select {
		case <-existing.done:
			if !fresh && time.Now().Before(existing.until) {
				s.mu.Unlock()
				return existing.book, existing.err
			}
		default:
			s.mu.Unlock()
			select {
			case <-ctx.Done():
				return Book{}, ctx.Err()
			case <-existing.done:
				return existing.book, existing.err
			}
		}
	}
	if len(s.books) >= 32 {
		for id, entry := range s.books {
			select {
			case <-entry.done:
				delete(s.books, id)
			default:
			}
		}
	}
	if len(s.books) >= 32 {
		s.mu.Unlock()
		return s.primary.Book(ctx, seriesID, bookID)
	}
	entry := &bookEntry{done: make(chan struct{})}
	s.books[key] = entry
	s.mu.Unlock()
	book, err := s.primary.Book(ctx, seriesID, bookID)
	s.mu.Lock()
	entry.book, entry.err = book, err
	entry.until = time.Now().Add(2 * time.Minute)
	if err != nil {
		delete(s.books, key)
	}
	close(entry.done)
	s.mu.Unlock()
	return book, err
}
func (s *Service) ResolveFresh(ctx context.Context, series Series, number int, releaseID string) (Publication, error) {
	return s.Resolve(context.WithValue(ctx, freshBookKey{}, true), series, number, releaseID)
}
