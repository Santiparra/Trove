// Package desktop is a thin Wails-facing adapter. It owns no business rules or SQL.
package desktop

import (
	"context"
	"errors"
	"log"

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
	service *books.Service
	info    Info
}

func NewLibrary(service *books.Service, profile, dir, buildID string) *Library {
	return &Library{service: service, info: Info{
		Profile: profile, DataDir: dir, BuildID: buildID,
		Statuses: []Choice{{"to-read", "Por leer"}, {"reading", "Leyendo"}, {"read", "Leído"}, {"paused", "Pausado"}, {"did-not-finish", "Sin terminar"}},
		Formats:  []Choice{{"", "Sin especificar"}, {"paperback", "Tapa blanda"}, {"hardcover", "Tapa dura"}, {"digital", "Digital"}, {"audio", "Audiolibro"}, {"other", "Otro"}},
	}}
}

func (l *Library) Info() Info { return l.info }
func (l *Library) ListBooks(query string) ([]books.Book, error) {
	result, err := l.service.List(context.Background(), query)
	if err != nil {
		log.Printf("listar libros: %v", err)
		return nil, errors.New("no se pudo leer la biblioteca; intentá actualizar el listado")
	}
	return result, nil
}
func (l *Library) GetBook(id string) (books.Book, error) {
	return l.service.Get(context.Background(), id)
}
func (l *Library) ListSeries() ([]books.Series, error) {
	return l.service.ListSeries(context.Background())
}
func (l *Library) SaveBook(input books.Input) (books.Book, error) {
	return l.service.Save(context.Background(), input)
}
