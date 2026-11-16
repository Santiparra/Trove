// Package desktop is a thin Wails-facing adapter. It owns no business rules or SQL.
package desktop

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/Santiparra/Trove/internal/books"
)

type Choice struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
type Info struct {
	Profile  string   `json:"profile"`
	DataDir  string   `json:"dataDir"`
	BuildID  string   `json:"buildID"`
	Statuses []Choice `json:"statuses"`
	Formats  []Choice `json:"formats"`
}

type Library struct {
	gate    sync.RWMutex
	closing bool
	service *books.Service
	info    Info
}

func NewLibrary(service *books.Service, profile, dir, buildID string) *Library {
	return &Library{service: service, info: Info{
		Profile: profile, DataDir: dir, BuildID: buildID,
		Statuses: []Choice{{"to-read", "Por leer"}, {"up-next", "A continuación"}, {"reading", "Leyendo"}, {"read", "Leído"}, {"paused", "Pausado"}, {"did-not-finish", "Sin terminar"}},
		Formats:  []Choice{{"", "Sin especificar"}, {"paperback", "Tapa blanda"}, {"hardcover", "Tapa dura"}, {"digital", "Digital"}, {"audio", "Audiolibro"}, {"other", "Otro"}},
	}}
}

func (l *Library) Info() Info { return l.info }
func (l *Library) ListBooks(query, status string) ([]books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	result, err := l.service.ListByStatus(context.Background(), query, status)
	if err != nil {
		log.Printf("listar libros: %v", err)
		return nil, errors.New("no se pudo leer la biblioteca; intentá actualizar el listado")
	}
	return result, nil
}

func (l *Library) Dashboard() (books.Dashboard, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	return l.service.Dashboard(context.Background())
}
func (l *Library) ChangeBookStatus(id, status string) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	if l.closing {
		return books.Book{}, errors.New("Trove se está cerrando")
	}
	return l.service.ChangeStatus(context.Background(), id, status)
}
func (l *Library) DeleteBook(id string) error {
	l.gate.RLock()
	defer l.gate.RUnlock()
	if l.closing {
		return errors.New("Trove se está cerrando")
	}
	return l.service.Delete(context.Background(), id)
}
func (l *Library) GetBook(id string) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	return l.service.Get(context.Background(), id)
}
func (l *Library) ListSeries() ([]books.Series, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	return l.service.ListSeries(context.Background())
}
func (l *Library) SaveBook(input books.Input) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	if l.closing {
		return books.Book{}, errors.New("Trove se está cerrando")
	}
	return l.service.Save(context.Background(), input)
}

func (l *Library) SaveBookNotes(id, notes string) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	if l.closing {
		return books.Book{}, errors.New("Trove se está cerrando")
	}
	return l.service.SaveNotes(context.Background(), id, notes)
}
func (l *Library) SetBookFavorite(id string, value bool) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	if l.closing {
		return books.Book{}, errors.New("Trove se está cerrando")
	}
	return l.service.SetFavorite(context.Background(), id, value)
}
func (l *Library) SetBookOwned(id string, value bool) (books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	if l.closing {
		return books.Book{}, errors.New("Trove se está cerrando")
	}
	return l.service.SetOwned(context.Background(), id, value)
}
func (l *Library) ListSeriesBooks(id string) ([]books.Book, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	return l.service.SeriesBooks(context.Background(), id)
}
func (l *Library) SearchBooks(query string) (books.SearchResults, error) {
	l.gate.RLock()
	defer l.gate.RUnlock()
	return l.service.Search(context.Background(), query)
}
