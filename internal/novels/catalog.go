package novels

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"

	"github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/novelcatalog"
)

type CatalogMatch struct {
	Kind    string `json:"kind"`
	Novel   *Novel `json:"novel"`
	Message string `json:"message"`
}

func (s *Service) MatchCatalog(ctx context.Context, series novelcatalog.Series, pub novelcatalog.Publication) (CatalogMatch, error) {
	items, err := s.repository.ListNovels(ctx, "")
	if err != nil {
		return CatalogMatch{}, err
	}
	for _, n := range items {
		if n.Catalog != nil && n.Catalog.ID == series.ID {
			return CatalogMatch{Kind: "existing", Novel: &n, Message: "Esta novela ya está en tu biblioteca."}, nil
		}
	}
	if pub.ISBN != "" {
		for _, n := range items {
			for _, v := range n.Volumes {
				if catalog.ISBN13(v.Edition.Identifier) == pub.ISBN || catalog.PreferredISBN(v.Edition.Identifiers) == pub.ISBN {
					return CatalogMatch{Kind: "review", Novel: &n, Message: "Hay un volumen local con el mismo ISBN; revisá la vinculación antes de agregar."}, nil
				}
			}
		}
	}
	return CatalogMatch{Kind: "new"}, nil
}
func applyPublication(n *Novel, series novelcatalog.Series, pub novelcatalog.Publication, changePreference bool) error {
	if err := series.Validate(); err != nil {
		return err
	}
	if err := pub.Validate(); err != nil {
		return err
	}
	ref, ok := series.Ref(pub.Number)
	if !ok || ref.ID != pub.BookID || pub.SeriesID != series.ID {
		return errors.New("la publicación no pertenece a esta novela")
	}
	if n.Catalog != nil && !novelcatalog.SameOrder(*n.Catalog, series) {
		return errors.New("la serie o su orden cambiaron; revisá el vínculo antes de reemplazar identidades")
	}
	if changePreference && pub.ReleaseID != "" {
		series.Language, series.Format = pub.Language, pub.ReleaseFormat
	}
	v := volumeAt(n, pub.Number)
	old := v.Catalog
	personalISBN := catalog.ISBN13(v.Edition.Identifier)
	ownedISBN := old != nil && old.ISBN != "" && personalISBN == old.ISBN
	if old != nil && old.ReleaseID != "" && pub.ReleaseID == "" {
		return errors.New("no hay una publicación nueva verificable; se conserva la anterior hasta elegir otra o desvincular")
	}
	if personalISBN != "" && pub.ISBN != "" && personalISBN != pub.ISBN && !ownedISBN {
		return errors.New("el ISBN personal no coincide con la publicación seleccionada; revisalo antes de vincular")
	}
	if old != nil {
		pub = novelcatalog.KeepMissing(*old, pub)
	}
	autoIdentifier := v.Edition.Identifier == "" || ownedISBN
	if autoIdentifier {
		v.Edition.Identifier = pub.ISBN
	}
	source := "manual"
	if autoIdentifier {
		source = novelcatalog.RanobeDB
	}
	v.Edition.Identifiers = catalog.MergeIdentifiers(catalog.NormalizeIdentifier(v.Edition.Identifier, source), catalog.NormalizeIdentifier(pub.ISBN, novelcatalog.RanobeDB))
	if v.Edition.Language == "" || (old != nil && v.Edition.Language == old.Language) {
		v.Edition.Language = pub.Language
	}
	if v.Edition.Format == "" || (old != nil && v.Edition.Format == old.Format) {
		v.Edition.Format = pub.Format
	}
	if n.Completed != nil && v.Number <= *n.Completed {
		v.Cycle.Status = "read"
	}
	v.Catalog = &pub
	n.Catalog = &series
	return nil
}
func (s *Service) AddCatalog(ctx context.Context, series novelcatalog.Series, pub novelcatalog.Publication, status string, completed *int) (CatalogMatch, error) {
	match, err := s.MatchCatalog(ctx, series, pub)
	if err != nil || match.Kind != "new" {
		return match, err
	}
	n := Novel{ID: rand.Text(), Title: series.Title, Authors: []string{}, Status: status, Completed: completed, AddedOn: s.today(), Volumes: []Volume{}}
	for _, author := range series.Authors {
		n.Authors = append(n.Authors, author.Name)
	}
	if err = reconcile(&n); err != nil {
		return CatalogMatch{}, err
	}
	if err = applyPublication(&n, series, pub, true); err != nil {
		return CatalogMatch{}, err
	}
	if n.Status == "reading" && n.NextNumber() > 0 {
		v := volumeAt(&n, n.NextNumber())
		v.Cycle, err = v.Cycle.Transition("reading", s.today(), false, false)
		if err != nil {
			return CatalogMatch{}, err
		}
	}
	stored, err := s.persist(ctx, n)
	if err != nil {
		return CatalogMatch{}, err
	}
	return CatalogMatch{Kind: "created", Novel: &stored, Message: "Novela agregada a tu biblioteca."}, nil
}
func (s *Service) LinkCatalog(ctx context.Context, id string, series novelcatalog.Series, pub novelcatalog.Publication) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	items, err := s.repository.ListNovels(ctx, "")
	if err != nil {
		return Novel{}, err
	}
	for _, other := range items {
		if other.ID != n.ID && other.Catalog != nil && other.Catalog.ID == series.ID {
			return Novel{}, errors.New("esa novela de catálogo ya está vinculada en otro registro local")
		}
	}
	if err = applyPublication(&n, series, pub, true); err != nil {
		return Novel{}, err
	}
	return s.persist(ctx, n)
}
func (s *Service) RefreshPublication(ctx context.Context, id string, expected novelcatalog.Series, fresh novelcatalog.Series, pub novelcatalog.Publication) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	if n.Catalog == nil || n.Catalog.ID != expected.ID || n.Catalog.Language != expected.Language || n.Catalog.Format != expected.Format {
		return Novel{}, errors.New("el vínculo cambió durante la consulta; no se reemplazó metadata")
	}
	fresh.Language, fresh.Format = n.Catalog.Language, n.Catalog.Format
	for _, v := range n.Volumes {
		if v.Number == pub.Number && v.Catalog != nil && v.Catalog.ReleaseID == pub.ReleaseID && v.Catalog.ISBN != "" && pub.ISBN != "" && v.Catalog.ISBN != pub.ISBN {
			return Novel{}, errors.New("el ISBN de la publicación cambió; revisá el vínculo antes de reemplazar metadata")
		}
	}
	if err = applyPublication(&n, fresh, pub, false); err != nil {
		return Novel{}, err
	}
	return s.persist(ctx, n)
}
func (s *Service) UnlinkCatalog(ctx context.Context, id string) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	n.Catalog = nil
	for i := range n.Volumes {
		n.Volumes[i].Catalog = nil
	}
	return s.persist(ctx, n)
}
func (s *Service) FindCatalogSeries(ctx context.Context, id string) (*Novel, error) {
	items, err := s.repository.ListNovels(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, n := range items {
		if n.Catalog != nil && strings.TrimSpace(n.Catalog.ID) == id {
			return &n, nil
		}
	}
	return nil, nil
}
