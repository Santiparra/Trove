package books

import (
	"context"
	"errors"

	"github.com/Santiparra/Trove/internal/catalog"
)

type CatalogResult struct {
	Book    Book   `json:"book"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

func (s *Service) AddCatalog(ctx context.Context, snapshot catalog.Snapshot, seriesID string) (CatalogResult, error) {
	if s.medium != MediumBooks {
		return CatalogResult{}, errors.New("el catálogo remoto de esta entrega es solo para Libros")
	}
	if err := snapshot.Validate(); err != nil {
		return CatalogResult{}, err
	}
	items, err := s.List(ctx, "")
	if err != nil {
		return CatalogResult{}, err
	}
	isbn := catalog.PreferredISBN(snapshot.Identifiers)
	for _, item := range items {
		if old := item.Edition.Catalog; old != nil && old.Provider == snapshot.Provider {
			if old.Edition.ID == snapshot.Edition.ID {
				return CatalogResult{item, "existing", "Esta edición ya está en tu biblioteca; no se creó un duplicado."}, nil
			}
			if old.Work.ID == snapshot.Work.ID {
				return CatalogResult{item, "review", "Esta obra ya existe con otra edición. Revisá su ficha antes de cambiar la vinculación; no se fusionaron datos."}, nil
			}
		}
		if isbn != "" && catalog.PreferredISBN(item.Edition.Identifiers) == isbn {
			return CatalogResult{item, "review", "Hay una edición local con el mismo ISBN. Podés vincularla explícitamente sin perder su seguimiento."}, nil
		}
	}
	input := Input{Medium: MediumBooks, Title: snapshot.Edition.Title, Authors: snapshot.Work.Authors, Language: snapshot.Edition.Language, Format: snapshot.Edition.Format, Identifier: isbn, Status: "to-read"}
	if seriesID != "" {
		found := false
		for _, v := range snapshot.Work.Series {
			if v.ID == seriesID {
				input.SeriesName = v.Name
				input.SeriesPosition = v.Position
				found = true
				break
			}
		}
		if !found {
			return CatalogResult{}, errors.New("la serie elegida no pertenece al candidato")
		}
	}
	book, err := s.save(ctx, input, &snapshot)
	if err != nil {
		return CatalogResult{}, err
	}
	return CatalogResult{book, "created", "Edición agregada con identidad local y snapshot."}, nil
}

func (s *Service) LinkCatalog(ctx context.Context, id string, snapshot catalog.Snapshot) (Book, error) {
	if s.medium != MediumBooks {
		return Book{}, errors.New("catálogo no disponible para este medio")
	}
	if err := snapshot.Validate(); err != nil {
		return Book{}, err
	}
	book, err := s.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	localISBN := catalog.ISBN13(book.Edition.Identifier)
	remoteISBN := catalog.PreferredISBN(snapshot.Identifiers)
	if localISBN != "" && remoteISBN != "" && localISBN != remoteISBN {
		return Book{}, errors.New("el ISBN local y el de la edición seleccionada no coinciden; revisá la edición antes de vincular")
	}
	items, err := s.List(ctx, "")
	if err != nil {
		return Book{}, err
	}
	for _, other := range items {
		if other.ID != id && other.Edition.Catalog != nil && other.Edition.Catalog.Provider == snapshot.Provider && other.Edition.Catalog.Edition.ID == snapshot.Edition.ID {
			return Book{}, errors.New("esa edición ya está vinculada a otro registro local")
		}
	}
	retained := []catalog.Identifier{}
	for _, v := range book.Edition.Identifiers {
		if v.Source != catalog.Hardcover {
			retained = append(retained, v)
		}
	}
	ids := catalog.MergeIdentifiers(retained, snapshot.Identifiers)
	if err = s.repository.UpdateCatalog(ctx, id, book.Edition.ID, &snapshot, ids); err != nil {
		return Book{}, err
	}
	return s.Get(ctx, id)
}

func (s *Service) UnlinkCatalog(ctx context.Context, id string) (Book, error) {
	if s.medium != MediumBooks {
		return Book{}, errors.New("catálogo no disponible para este medio")
	}
	book, err := s.Get(ctx, id)
	if err != nil {
		return Book{}, err
	}
	if err = s.repository.UpdateCatalog(ctx, id, book.Edition.ID, nil, book.Edition.Identifiers); err != nil {
		return Book{}, err
	}
	return s.Get(ctx, id)
}
