package desktop

import (
	"context"
	"crypto/rand"
	"errors"
	"sync"
	"time"

	"github.com/Santiparra/Trove/internal/books"
	cat "github.com/Santiparra/Trove/internal/catalog"
)

type CatalogKey interface {
	Get() (string, error)
	Save(string) error
	Clear() error
}
type CatalogCovers interface {
	Get(context.Context, string) (string, error)
}
type CatalogState struct {
	Provider   string `json:"provider"`
	Configured bool   `json:"configured"`
}
type CatalogPreview struct {
	Token    string               `json:"token"`
	Snapshot cat.Snapshot         `json:"snapshot"`
	Match    *books.CatalogResult `json:"match"`
}
type previewEntry struct {
	snapshot cat.Snapshot
	at       time.Time
}
type Catalog struct {
	service  *cat.Service
	key      CatalogKey
	covers   CatalogCovers
	library  *Library
	mu       sync.Mutex
	previews map[string]previewEntry
}

func NewCatalog(service *cat.Service, key CatalogKey, covers CatalogCovers, library *Library) *Catalog {
	return &Catalog{service: service, key: key, covers: covers, library: library, previews: map[string]previewEntry{}}
}
func (c *Catalog) State() CatalogState {
	token, err := c.key.Get()
	return CatalogState{Provider: cat.Hardcover, Configured: err == nil && token != ""}
}
func (c *Catalog) SaveKey(value string) (CatalogState, error) {
	c.library.gate.RLock()
	defer c.library.gate.RUnlock()
	if c.library.closing {
		return CatalogState{}, errors.New("Trove se está cerrando")
	}
	if err := c.key.Save(value); err != nil {
		return c.State(), err
	}
	return c.State(), nil
}
func (c *Catalog) ClearKey() (CatalogState, error) {
	if err := c.key.Clear(); err != nil {
		return c.State(), err
	}
	return c.State(), nil
}
func (c *Catalog) Search(query string) (cat.SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return c.service.Search(ctx, query)
}
func (c *Catalog) Editions(workID string) ([]cat.Edition, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return c.service.Editions(ctx, workID)
}
func (c *Catalog) Preview(editionID string) (CatalogPreview, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	snapshot, err := c.service.Resolve(ctx, editionID)
	if err != nil {
		return CatalogPreview{}, err
	}
	return c.prepare(snapshot)
}

// OpenWork opens an owned edition first, otherwise the most informative returned edition.
func (c *Catalog) OpenWork(workID string) (CatalogPreview, error) {
	if !cat.ValidID(cat.Hardcover, workID) {
		return CatalogPreview{}, errors.New("obra de catálogo inválida")
	}
	items, err := c.library.ListBooks("", "", books.MediumBooks)
	if err != nil {
		return CatalogPreview{}, err
	}
	for _, item := range items {
		if snapshot := item.Edition.Catalog; snapshot != nil && snapshot.Provider == cat.Hardcover && snapshot.Work.ID == workID {
			return c.prepare(*snapshot)
		}
	}
	editions, err := c.Editions(workID)
	if err != nil {
		return CatalogPreview{}, err
	}
	if len(editions) == 0 {
		return CatalogPreview{}, errors.New("no hay ediciones disponibles para esta obra")
	}
	best, score := editions[0], -1
	for _, edition := range editions {
		n := 0
		for _, value := range []string{edition.Language, edition.Format, edition.Publisher, edition.PublishedOn, edition.CoverURL} {
			if value != "" {
				n++
			}
		}
		if edition.Pages != nil {
			n++
		}
		if len(edition.Declared) > 0 {
			n++
		}
		if n > score {
			best, score = edition, n
		}
	}
	return c.Preview(best.ID)
}

func (c *Catalog) prepare(snapshot cat.Snapshot) (CatalogPreview, error) {
	c.library.gate.RLock()
	defer c.library.gate.RUnlock()
	s, _ := c.library.service.ForMedium(books.MediumBooks)
	match, err := s.MatchCatalog(context.Background(), snapshot)
	if err != nil {
		return CatalogPreview{}, err
	}
	token := rand.Text()
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, entry := range c.previews {
		if now.Sub(entry.at) > 10*time.Minute {
			delete(c.previews, id)
		}
	}
	if len(c.previews) >= 8 {
		var oldest string
		var at time.Time
		for id, v := range c.previews {
			if oldest == "" || v.at.Before(at) {
				oldest, at = id, v.at
			}
		}
		delete(c.previews, oldest)
	}
	c.previews[token] = previewEntry{snapshot, now}
	return CatalogPreview{Token: token, Snapshot: snapshot, Match: match}, nil
}
func (c *Catalog) prepared(token string) (cat.Snapshot, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.previews[token]
	if !ok || time.Since(v.at) > 10*time.Minute {
		return cat.Snapshot{}, errors.New("la selección de catálogo expiró; revisá de nuevo la edición")
	}
	return v.snapshot, nil
}
func (c *Catalog) Add(token, seriesID, status string) (books.CatalogResult, error) {
	snapshot, err := c.prepared(token)
	if err != nil {
		return books.CatalogResult{}, err
	}
	c.library.gate.Lock()
	defer c.library.gate.Unlock()
	if c.library.closing {
		return books.CatalogResult{}, errors.New("Trove se está cerrando")
	}
	s, err := c.library.service.ForMedium(books.MediumBooks)
	if err != nil {
		return books.CatalogResult{}, err
	}
	return s.AddCatalog(context.Background(), snapshot, seriesID, status)
}
func (c *Catalog) Link(bookID, token string) (books.Book, error) {
	snapshot, err := c.prepared(token)
	if err != nil {
		return books.Book{}, err
	}
	c.library.gate.Lock()
	defer c.library.gate.Unlock()
	if c.library.closing {
		return books.Book{}, errors.New("Trove se está cerrando")
	}
	s, err := c.library.service.ForMedium(books.MediumBooks)
	if err != nil {
		return books.Book{}, err
	}
	return s.LinkCatalog(context.Background(), bookID, snapshot)
}
func (c *Catalog) Refresh(bookID string) (books.Book, error) {
	before, err := c.library.GetBook(bookID, books.MediumBooks)
	if err != nil {
		return books.Book{}, err
	}
	old := before.Edition.Catalog
	if old == nil {
		return books.Book{}, errors.New("esta edición no está vinculada")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	fresh, err := c.service.Resolve(ctx, old.Edition.ID)
	if err != nil {
		return books.Book{}, err
	}
	c.library.gate.Lock()
	defer c.library.gate.Unlock()
	if c.library.closing {
		return books.Book{}, errors.New("Trove se está cerrando")
	}
	s, _ := c.library.service.ForMedium(books.MediumBooks)
	current, err := s.Get(ctx, bookID)
	if err != nil {
		return books.Book{}, err
	}
	if current.Edition.ID != before.Edition.ID || current.Edition.Catalog == nil || current.Edition.Catalog.Edition.ID != old.Edition.ID {
		return books.Book{}, errors.New("la vinculación cambió durante el refresco; revisá la ficha")
	}
	fresh = cat.KeepMissing(*current.Edition.Catalog, fresh)
	return s.LinkCatalog(ctx, bookID, fresh)
}
func (c *Catalog) Unlink(bookID string) (books.Book, error) {
	c.library.gate.Lock()
	defer c.library.gate.Unlock()
	if c.library.closing {
		return books.Book{}, errors.New("Trove se está cerrando")
	}
	s, _ := c.library.service.ForMedium(books.MediumBooks)
	return s.UnlinkCatalog(context.Background(), bookID)
}
func (c *Catalog) CoverDataURL(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return c.covers.Get(ctx, url)
}
