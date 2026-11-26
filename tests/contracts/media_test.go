package contracts_test

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/desktop"
)

func TestAdmission_C33_ReadingMediaAreIsolatedAcrossAllOperations(t *testing.T) {
	h := start(t, t.TempDir(), "test")
	book, err := h.library.SaveBook(books.Input{Title: "Mismo título", SeriesName: "Misma serie", Status: "reading"})
	if err != nil {
		t.Fatal(err)
	}
	novel, err := h.library.SaveBook(books.Input{Medium: books.MediumLightNovels, Title: "Mismo título", SeriesName: "Misma serie", Status: "up-next"})
	if err != nil {
		t.Fatal(err)
	}
	if book.ID == novel.ID || book.Series.ID == novel.Series.ID {
		t.Fatal("se fusionaron identidades entre medios")
	}
	for _, medium := range []string{books.MediumBooks, books.MediumLightNovels} {
		items, err := h.library.ListBooks("Mismo", "", medium)
		if err != nil || len(items) != 1 || items[0].Medium != medium {
			t.Fatal("listado mezclado")
		}
		d, err := h.library.Dashboard(medium)
		if err != nil || d.Total != 1 || d.Medium != medium {
			t.Fatal("portal mezclado")
		}
		suggestions, err := h.library.SearchBooks("Mismo", medium)
		if err != nil || suggestions.Total != 1 || suggestions.Matches[0].Medium != medium {
			t.Fatal("búsqueda mezclada")
		}
		series, err := h.library.ListSeries(medium)
		if err != nil || len(series) != 1 || series[0].Medium != medium {
			t.Fatal("series mezcladas")
		}
	}
	wrong := books.MediumLightNovels
	if items, err := h.library.ListSeriesBooks(book.Series.ID, wrong); err != nil || len(items) != 0 {
		t.Fatal("la navegación de serie mostró otro medio")
	}
	if items, err := h.library.ListSeriesBooks(novel.Series.ID, wrong); err != nil || len(items) != 1 || items[0].ID != novel.ID {
		t.Fatal("la serie de novelas ligeras no recupera su volumen")
	}
	if _, err = h.library.GetBook(book.ID, wrong); err == nil {
		t.Fatal("leyó ficha de otro medio")
	}
	if _, err = h.library.ChangeBookStatus(book.ID, "read", wrong); err == nil {
		t.Fatal("cambió estado de otro medio")
	}
	if _, err = h.library.SaveBookNotes(book.ID, "No guardar", wrong); err == nil {
		t.Fatal("cambió notas de otro medio")
	}
	if _, err = h.library.SetBookFavorite(book.ID, true, wrong); err == nil {
		t.Fatal("cambió favorito de otro medio")
	}
	if _, err = h.library.SetBookOwned(book.ID, true, wrong); err == nil {
		t.Fatal("cambió propiedad de otro medio")
	}
	if err = h.library.DeleteBook(book.ID, wrong); err == nil {
		t.Fatal("eliminó otro medio")
	}
	if _, err = h.library.SaveBook(books.Input{ID: book.ID, Medium: wrong, Title: "No reclasificar", Status: "read"}); err == nil {
		t.Fatal("editar reclasificó un registro")
	}
	if _, err = h.library.Dashboard("future"); err == nil {
		t.Fatal("aceptó un medio desconocido")
	}
	untouched, err := h.library.GetBook(book.ID, books.MediumBooks)
	if err != nil || untouched.Title != book.Title || !reflect.DeepEqual(untouched.Cycle, book.Cycle) || untouched.Notes != book.Notes || untouched.Favorite {
		t.Fatal("se modificaron Libros")
	}
	if _, err = h.library.SaveBookNotes(novel.ID, "Nota del volumen", wrong); err != nil {
		t.Fatal(err)
	}
	if _, err = h.library.ChangeBookStatus(novel.ID, "reading", wrong); err != nil {
		t.Fatal(err)
	}
	if err = h.library.DeleteBook(novel.ID, wrong); err != nil {
		t.Fatal(err)
	}
	if _, err = h.library.GetBook(book.ID, books.MediumBooks); err != nil {
		t.Fatal("eliminar un volumen afectó Libros")
	}
}

func TestAdmission_C34_MixedLibraryRestoresFromEitherRepresentation(t *testing.T) {
	for _, name := range []string{"trove.json", "trove.sqlite"} {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			source := start(t, filepath.Join(root, "source"), "test")
			var records []books.Book
			for _, medium := range []string{books.MediumBooks, books.MediumLightNovels} {
				b, err := source.library.SaveBook(books.Input{Medium: medium, Title: "Obra sintética", SeriesName: "Serie compartida de nombre", Status: "read", Owned: true, Favorite: true})
				if err != nil {
					t.Fatal(err)
				}
				records = append(records, b)
			}
			state, err := source.manager.Configure(context.Background(), filepath.Join(root, "cloud"))
			if err != nil {
				t.Fatal(err)
			}
			target := start(t, filepath.Join(root, "target"), "test")
			file := filepath.Join(state.Folder, "generation-"+state.Generation, name)
			adapter := desktop.NewBackups(target.manager, target.library, nil, func() (string, error) { return file, nil })
			preview, err := adapter.ChooseRestore()
			if err != nil || preview == nil || preview.Works != 2 {
				t.Fatal("preparación perdió un medio")
			}
			if _, err = adapter.ConfirmRestore(preview.Token); err != nil {
				t.Fatal(err)
			}
			for _, record := range records {
				got, err := target.library.GetBook(record.ID, record.Medium)
				if err != nil || got.Series.ID != record.Series.ID || !got.Favorite || !got.Edition.Owned || got.Cycle.FinishedOn != "" {
					t.Fatal("recuperación de medio incompleta")
				}
			}
		})
	}
}

func TestAdmission_C35_LegacyJSONRestoresAsBooksWithoutInventingDates(t *testing.T) {
	root := t.TempDir()
	target := start(t, root, "test")
	novel, err := target.library.SaveBook(books.Input{Medium: books.MediumLightNovels, Title: "Copia previa", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "legacy.json")
	data := `{"format":"trove-library","version":1,"generation":"legacy-generation","createdAt":"2020-01-01T00:00:00Z","catalog":{"series":[{"id":"legacy-series","name":"Serie antigua"}],"works":[{"id":"legacy-work","medium":"books","title":"Libro antiguo","authors":[],"seriesId":"legacy-series","notes":"Conservar","addedOn":"","favorite":false}],"editions":[{"id":"legacy-edition","workId":"legacy-work","language":"","format":"","identifier":"","owned":true}],"cycles":[{"id":"legacy-cycle","editionId":"legacy-edition","status":"read","startedOn":"","finishedOn":""}]}}`
	if err = os.WriteFile(file, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	adapter := desktop.NewBackups(target.manager, target.library, nil, func() (string, error) { return file, nil })
	preview, err := adapter.ChooseRestore()
	if err != nil {
		t.Fatal(err)
	}
	result, err := adapter.ConfirmRestore(preview.Token)
	if err != nil {
		t.Fatal(err)
	}
	b, err := target.library.GetBook("legacy-work", books.MediumBooks)
	if err != nil || b.Series.ID != "legacy-series" || b.Series.Medium != books.MediumBooks || b.Cycle.StartedOn != "" || b.Cycle.FinishedOn != "" || b.AddedOn != "" {
		t.Fatal("se alteró el respaldo anterior")
	}
	volumes, err := target.library.ListBooks("", "", books.MediumLightNovels)
	if err != nil || len(volumes) != 0 {
		t.Fatal("una restauración completa mezcló bibliotecas")
	}
	safety, err := target.store.ReadNative(context.Background(), result.SafetyCopy)
	if err != nil || len(safety.Catalog.Works) != 1 || safety.Catalog.Works[0].ID != novel.ID {
		t.Fatal("la copia previa no conservó las novelas ligeras")
	}
}
