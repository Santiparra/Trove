package novelcatalog

import (
	"context"
	"sync/atomic"
	"testing"
)

type cachePrimary struct {
	calls            atomic.Int32
	entered, release chan struct{}
}

func (p *cachePrimary) Search(context.Context, string, string, int) (SearchResult, error) {
	return SearchResult{}, nil
}
func (p *cachePrimary) Series(context.Context, string) (Series, error) { return Series{}, nil }
func (p *cachePrimary) Book(context.Context, string, string) (Book, error) {
	if p.calls.Add(1) == 1 && p.entered != nil {
		close(p.entered)
		<-p.release
	}
	return Book{ID: "20", SeriesID: "10", Releases: []Release{{ID: "30"}}}, nil
}
func (p *cachePrimary) Release(context.Context, string, string) (Publication, error) {
	return Publication{}, nil
}
func TestSelectedBookDetailsAreCoalescedAndReused(t *testing.T) {
	p := &cachePrimary{entered: make(chan struct{}), release: make(chan struct{})}
	s := New(p, nil)
	ctx := context.Background()
	series := Series{ID: "10", Books: []BookRef{{ID: "20", Number: 1}}}
	done := make(chan error, 2)
	go func() { _, err := s.Releases(ctx, series, 1); done <- err }()
	<-p.entered
	go func() { _, err := s.Releases(ctx, series, 1); done <- err }()
	close(p.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := s.Releases(ctx, series, 1); err != nil {
		t.Fatal(err)
	}
	if p.calls.Load() != 1 {
		t.Fatal("repitió detalles del mismo libro")
	}
	if _, err := s.book(context.WithValue(ctx, freshBookKey{}, true), "10", "20"); err != nil {
		t.Fatal(err)
	}
	if p.calls.Load() != 2 {
		t.Fatal("refresco explícito no saltó la caché")
	}
}
