package desktop

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"github.com/Santiparra/Trove/internal/novelcatalog"
	"github.com/Santiparra/Trove/internal/novels"
)

type NovelPreview struct {
	Token       string                   `json:"token"`
	Series      novelcatalog.Series      `json:"series"`
	Publication novelcatalog.Publication `json:"publication"`
	Match       novels.CatalogMatch      `json:"match"`
}

// BrowseCatalog loads a series once; publication details are deliberately resolved afterwards.
func (n *Novels) BrowseCatalog(seriesID, bookID, releaseID string) (NovelPreview, error) {
	if n.catalog == nil {
		return NovelPreview{}, errors.New("catálogo de novelas no disponible")
	}
	n.library.gate.RLock()
	existing, err := n.service.FindCatalogSeries(context.Background(), seriesID)
	n.library.gate.RUnlock()
	if err != nil {
		return NovelPreview{}, err
	}
	if existing != nil {
		if releaseID == "" {
			return NovelPreview{Series: *existing.Catalog, Match: novels.CatalogMatch{Kind: "existing", Novel: existing}}, nil
		}
		for _, v := range existing.Volumes {
			if v.Catalog != nil && v.Catalog.BookID == bookID && v.Catalog.ReleaseID == releaseID {
				return NovelPreview{Series: *existing.Catalog, Match: novels.CatalogMatch{Kind: "existing", Novel: existing}}, nil
			}
		}
		return n.prepareCatalog(*existing.Catalog, novelcatalog.Publication{})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	series, err := n.catalog.Series(ctx, seriesID)
	if err != nil {
		return NovelPreview{}, err
	}
	return n.prepareCatalog(series, novelcatalog.Publication{})
}

func (n *Novels) SearchCatalog(query, authorID string, page int, scope string) (novelcatalog.SearchResult, error) {
	if n.catalog == nil {
		return novelcatalog.SearchResult{}, errors.New("catálogo de novelas no disponible")
	}
	n.mu.Lock()
	if scope != "list" {
		scope = "popover"
	}
	if previous := n.searchCancel[scope]; previous != nil {
		previous()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	n.searchCancel[scope] = cancel
	n.mu.Unlock()
	defer cancel()
	return n.catalog.Search(ctx, query, authorID, page)
}
func (n *Novels) prepareCatalog(series novelcatalog.Series, pub novelcatalog.Publication) (NovelPreview, error) {
	n.library.gate.RLock()
	defer n.library.gate.RUnlock()
	match, err := n.service.MatchCatalog(context.Background(), series, pub)
	if err != nil {
		return NovelPreview{}, err
	}
	token := rand.Text()
	n.mu.Lock()
	defer n.mu.Unlock()
	for id, value := range n.previews {
		if time.Since(value.at) > 10*time.Minute {
			delete(n.previews, id)
		}
	}
	if len(n.previews) >= 12 {
		var oldest string
		var at time.Time
		for id, value := range n.previews {
			if oldest == "" || value.at.Before(at) {
				oldest, at = id, value.at
			}
		}
		delete(n.previews, oldest)
	}
	n.previews[token] = novelSelection{series: series, publication: pub, at: time.Now()}
	return NovelPreview{Token: token, Series: series, Publication: pub, Match: match}, nil
}
func (n *Novels) selection(token string) (novelSelection, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	value, ok := n.previews[token]
	if !ok || time.Since(value.at) > 10*time.Minute {
		return novelSelection{}, errors.New("la selección venció; volvé a revisar la publicación")
	}
	return value, nil
}
func (n *Novels) OpenCatalog(seriesID string, number int, bookID, releaseID string) (NovelPreview, error) {
	if n.catalog == nil {
		return NovelPreview{}, errors.New("catálogo de novelas no disponible")
	}
	n.library.gate.RLock()
	existing, err := n.service.FindCatalogSeries(context.Background(), seriesID)
	n.library.gate.RUnlock()
	if err != nil {
		return NovelPreview{}, err
	}
	if existing != nil && releaseID == "" {
		return NovelPreview{Series: *existing.Catalog, Match: novels.CatalogMatch{Kind: "existing", Novel: existing, Message: "Esta novela ya está en tu biblioteca."}}, nil
	}
	if existing != nil && releaseID != "" {
		for _, v := range existing.Volumes {
			if v.Catalog != nil && v.Catalog.BookID == bookID && v.Catalog.ReleaseID == releaseID {
				return NovelPreview{Series: *existing.Catalog, Match: novels.CatalogMatch{Kind: "existing", Novel: existing, Message: "Esta publicación ya está en tu biblioteca."}}, nil
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	series, err := n.catalog.Series(ctx, seriesID)
	if err != nil {
		return NovelPreview{}, err
	}
	if number < 1 || number > len(series.Books) {
		number = 1
	}
	if bookID != "" {
		found := false
		for _, ref := range series.Books {
			if ref.ID == bookID {
				number, found = ref.Number, true
				break
			}
		}
		if !found {
			return NovelPreview{}, errors.New("el ISBN apunta a un libro que no es principal de esta novela")
		}
	}
	pub, err := n.catalog.Resolve(ctx, series, number, releaseID)
	if err != nil {
		return NovelPreview{}, err
	}
	return n.prepareCatalog(series, pub)
}
func (n *Novels) PrepareCatalog(id string, number int) (NovelPreview, error) {
	if n.catalog == nil {
		return NovelPreview{}, errors.New("catálogo de novelas no disponible")
	}
	novel, err := n.Get(id)
	if err != nil {
		return NovelPreview{}, err
	}
	if novel.Catalog == nil {
		return NovelPreview{}, errors.New("esta novela todavía no está vinculada")
	}
	if number < 1 {
		number = novel.NextNumber()
		if number == 0 && novel.Completed != nil {
			number = *novel.Completed
		}
		if number < 1 {
			number = 1
		}
	}
	for _, v := range novel.Volumes {
		if v.Number == number && v.Catalog != nil {
			return n.prepareCatalog(*novel.Catalog, *v.Catalog)
		}
	}
	return n.prepareCatalog(*novel.Catalog, novelcatalog.Publication{Number: number})
}
func (n *Novels) CatalogReleases(token string, number int) ([]novelcatalog.Release, error) {
	if n.catalog == nil {
		return nil, errors.New("catálogo de novelas no disponible")
	}
	value, err := n.selection(token)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return n.catalog.Releases(ctx, value.series, number)
}
func (n *Novels) PreviewPublication(token string, number int, releaseID string) (NovelPreview, error) {
	if n.catalog == nil {
		return NovelPreview{}, errors.New("catálogo de novelas no disponible")
	}
	value, err := n.selection(token)
	if err != nil {
		return NovelPreview{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	pub, err := n.catalog.Resolve(ctx, value.series, number, releaseID)
	if err != nil {
		return NovelPreview{}, err
	}
	return n.prepareCatalog(value.series, pub)
}
func (n *Novels) AddCatalog(token, status string, completed *int) (novels.CatalogMatch, error) {
	value, err := n.selection(token)
	if err != nil {
		return novels.CatalogMatch{}, err
	}
	n.library.gate.Lock()
	defer n.library.gate.Unlock()
	if n.library.closing {
		return novels.CatalogMatch{}, errors.New("Trove se está cerrando")
	}
	return n.service.AddCatalog(context.Background(), value.series, value.publication, status, completed)
}
func (n *Novels) LinkCatalog(id, token string) (novels.Novel, error) {
	value, err := n.selection(token)
	if err != nil {
		return novels.Novel{}, err
	}
	return n.mutate(func() (novels.Novel, error) {
		return n.service.LinkCatalog(context.Background(), id, value.series, value.publication)
	})
}
func (n *Novels) UnlinkCatalog(id string) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.UnlinkCatalog(context.Background(), id) })
}
func (n *Novels) RefreshCatalog(id string, number int, refreshSeries bool) (novels.Novel, error) {
	if n.catalog == nil {
		return novels.Novel{}, errors.New("catálogo de novelas no disponible")
	}
	n.library.gate.RLock()
	before, err := n.service.Get(context.Background(), id)
	epoch := n.library.dataEpoch
	n.library.gate.RUnlock()
	if err != nil {
		return novels.Novel{}, err
	}
	if before.Catalog == nil {
		return novels.Novel{}, errors.New("novela sin vínculo de catálogo")
	}
	if number < 1 {
		number = before.NextNumber()
		if number == 0 && before.Completed != nil {
			number = *before.Completed
		}
		if number < 1 {
			number = 1
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	fresh := *before.Catalog
	if refreshSeries {
		fresh, err = n.catalog.Series(ctx, fresh.ID)
		if err != nil {
			return novels.Novel{}, err
		}
		fresh.Language, fresh.Format = before.Catalog.Language, before.Catalog.Format
	}
	releaseID := ""
	for _, v := range before.Volumes {
		if v.Number == number && v.Catalog != nil {
			releaseID = v.Catalog.ReleaseID
		}
	}
	var pub novelcatalog.Publication
	if refreshSeries {
		pub, err = n.catalog.ResolveFresh(ctx, fresh, number, releaseID)
	} else {
		pub, err = n.catalog.Resolve(ctx, fresh, number, releaseID)
	}
	if err != nil {
		return novels.Novel{}, err
	}
	return n.mutate(func() (novels.Novel, error) {
		if n.library.dataEpoch != epoch {
			return novels.Novel{}, errors.New("se restauró la biblioteca durante la consulta; no se reemplazó metadata")
		}
		if n.network != nil && n.network.LocalOnly() {
			return novels.Novel{}, errors.New("modo solo local activo; no se aplicó el refresco remoto")
		}
		return n.service.RefreshPublication(context.Background(), id, *before.Catalog, fresh, pub)
	})
}
