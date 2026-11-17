// Package desktop adapts Wails requests; business rules and medium isolation live in Go.
package desktop

import (
	"context"
	"errors"
	"sync"

	"github.com/Santiparra/Trove/internal/books"
)

type Choice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
type Info struct {
	Version  string   `json:"version"`
	Revision string   `json:"revision"`
	Profile  string   `json:"profile"`
	DataDir  string   `json:"dataDir"`
	BuildID  string   `json:"buildID"`
	Statuses []Choice `json:"statuses"`
	Formats  []Choice `json:"formats"`
}
type BuildInfo struct {
	Version  string
	Revision string
}
type Library struct {
	gate    sync.RWMutex
	closing bool
	service *books.Service
	info    Info
}

func NewLibrary(service *books.Service, profile, dir, buildID string, identity ...BuildInfo) *Library {
	meta := BuildInfo{Version: "development", Revision: "unknown"}
	if len(identity) > 0 {
		meta = identity[0]
	}
	return &Library{service: service, info: Info{Version: meta.Version, Revision: meta.Revision, Profile: profile, DataDir: dir, BuildID: buildID,
		Statuses: []Choice{{"to-read", "Por leer"}, {"up-next", "A continuación"}, {"reading", "Leyendo"}, {"read", "Leído"}, {"paused", "Pausado"}, {"did-not-finish", "Sin terminar"}},
		Formats:  []Choice{{"", "Sin especificar"}, {"paperback", "Tapa blanda"}, {"hardcover", "Tapa dura"}, {"digital", "Digital"}, {"audio", "Audiolibro"}, {"other", "Otro"}},
	}}
}

// Called under the library gate. Scopes are request-local, never a mutable selected tab.
func (l *Library) scoped(medium string, write bool) (*books.Service, error) {
	if write && l.closing {
		return nil, errors.New("Trove se está cerrando")
	}
	return l.service.ForMedium(medium)
}
func (l *Library) Info() Info { return l.info }
func (l *Library) ListBooks(query, status, medium string) ([]books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, false)
	if err != nil {
		return nil, err
	}
	return s.ListByStatus(context.Background(), query, status)
}
func (l *Library) Dashboard(medium string) (books.Dashboard, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, false)
	if err != nil {
		return books.Dashboard{}, err
	}
	return s.Dashboard(context.Background())
}
func (l *Library) GetBook(id, medium string) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, false)
	if err != nil {
		return books.Book{}, err
	}
	return s.Get(context.Background(), id)
}
func (l *Library) ListSeries(medium string) ([]books.Series, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, false)
	if err != nil {
		return nil, err
	}
	return s.ListSeries(context.Background())
}
func (l *Library) SaveBook(input books.Input) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	medium := input.Medium
	if medium == "" {
		medium = books.MediumBooks
	}
	s, err := l.scoped(medium, true)
	if err != nil {
		return books.Book{}, err
	}
	return s.Save(context.Background(), input)
}
func (l *Library) ChangeBookStatus(id, status, medium string) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, true)
	if err != nil {
		return books.Book{}, err
	}
	return s.ChangeStatus(context.Background(), id, status)
}
func (l *Library) DeleteBook(id, medium string) error {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, true)
	if err != nil {
		return err
	}
	return s.Delete(context.Background(), id)
}
func (l *Library) SaveBookNotes(id, notes, medium string) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, true)
	if err != nil {
		return books.Book{}, err
	}
	return s.SaveNotes(context.Background(), id, notes)
}
func (l *Library) SetBookFavorite(id string, value bool, medium string) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, true)
	if err != nil {
		return books.Book{}, err
	}
	return s.SetFavorite(context.Background(), id, value)
}
func (l *Library) SetBookOwned(id string, value bool, medium string) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, true)
	if err != nil {
		return books.Book{}, err
	}
	return s.SetOwned(context.Background(), id, value)
}
func (l *Library) ListSeriesBooks(id, medium string) ([]books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, false)
	if err != nil {
		return nil, err
	}
	return s.SeriesBooks(context.Background(), id)
}
func (l *Library) SearchBooks(query, medium string) (books.SearchResults, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	s, err := l.scoped(medium, false)
	if err != nil {
		return books.SearchResults{}, err
	}
	return s.Search(context.Background(), query)
}
