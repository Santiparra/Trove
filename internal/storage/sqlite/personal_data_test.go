package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Santiparra/Trove/internal/books"
)

func TestPersonalActionsAndSeriesOrdering(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	service := books.NewService(s)
	var id string
	for _, pos := range []*float64{nil, pointer(3), pointer(1.5), pointer(1)} {
		b, err := service.Save(ctx, books.Input{Title: "Título sintético", Status: "up-next", SeriesName: "Serie de ejemplo", SeriesPosition: pos, Notes: "Original"})
		if err != nil {
			t.Fatal(err)
		}
		if pos == nil {
			id = b.ID
		}
	}
	b, err := service.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.SetFavorite(ctx, id, true); err != nil {
		t.Fatal(err)
	}
	if _, err = service.SetOwned(ctx, id, true); err != nil {
		t.Fatal(err)
	}
	if _, err = service.SaveNotes(ctx, id, "Nueva nota\ncon contenido"); err != nil {
		t.Fatal(err)
	}
	updated, err := service.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Favorite || !updated.Edition.Owned || updated.Notes != "Nueva nota\ncon contenido" || updated.Cycle != b.Cycle || updated.Series.ID != b.Series.ID {
		t.Fatal("acción personal alteró otros campos")
	}
	items, err := service.SeriesBooks(ctx, b.Series.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 4 || *items[0].Series.Position != 1 || *items[1].Series.Position != 1.5 || *items[2].Series.Position != 3 || items[3].Series.Position != nil {
		t.Fatal("serie no ordenada por posición")
	}
	if _, err = service.SetFavorite(ctx, id, false); err != nil {
		t.Fatal(err)
	}
	if _, err = service.SetOwned(ctx, id, false); err != nil {
		t.Fatal(err)
	}
	updated, err = service.Get(ctx, id)
	if err != nil || updated.Favorite || updated.Edition.Owned {
		t.Fatal("no se desmarcaron propiedades")
	}
	if _, err = service.SaveNotes(ctx, "missing", "nota"); !errors.Is(err, books.ErrNotFound) {
		t.Fatal("ID inexistente aceptado")
	}
}

func pointer(n float64) *float64 { return &n }

func TestSaveAppliesLiveDatesButHonorsManualEdits(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	service := books.NewService(s)
	b, err := service.Save(ctx, books.Input{Title: "Lectura nueva", Status: "reading"})
	if err != nil {
		t.Fatal(err)
	}
	if b.Cycle.StartedOn != time.Now().Format("2006-01-02") {
		t.Fatal("crear leyendo no registró inicio")
	}
	b, err = service.Save(ctx, books.Input{ID: b.ID, Title: b.Title, Status: "read", StartedOn: b.Cycle.StartedOn, FinishedOnEdited: true})
	if err != nil {
		t.Fatal(err)
	}
	if b.Cycle.FinishedOn != "" {
		t.Fatal("no respetó finalización vacía explícita")
	}
	b, err = service.Save(ctx, books.Input{ID: b.ID, Title: b.Title, Status: "read", StartedOn: "2020-01-01", FinishedOn: "2020-01-10", StartedOnEdited: true, FinishedOnEdited: true})
	if err != nil {
		t.Fatal(err)
	}
	if b.Cycle.StartedOn != "2020-01-01" || b.Cycle.FinishedOn != "2020-01-10" {
		t.Fatal("no respetó fechas manuales")
	}
	historical, err := service.Save(ctx, books.Input{Title: "Lectura antigua", Status: "read"})
	if err != nil {
		t.Fatal(err)
	}
	if historical.Cycle.StartedOn != "" || historical.Cycle.FinishedOn != "" {
		t.Fatal("inventó fechas de un libro antiguo")
	}
}

func TestVersionOneMigratesAndHasReadableSafetyCopy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "trove.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	ddl, err := migrations.ReadFile("migrations/001_books.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(string(ddl)); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO books(id,title,authors_json,notes,added_on) VALUES('book','Ejemplo','[]','Conservar','2020-01-01');
INSERT INTO editions(id,book_id) VALUES('edition','book');
INSERT INTO reading_cycles(id,edition_id,status) VALUES('cycle','edition','read'); PRAGMA user_version=1;`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	b, err := s.Get(context.Background(), "book")
	if err != nil {
		t.Fatal(err)
	}
	if b.Notes != "Conservar" || b.Favorite || b.Edition.Owned || b.Cycle.FinishedOn != "" {
		t.Fatal("migración alteró datos previos")
	}
	var version int
	if err = s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 2 {
		t.Fatal("versión de esquema incorrecta")
	}
	paths, err := filepath.Glob(filepath.Join(dir, "before-schema-2-*.sqlite"))
	if err != nil || len(paths) != 1 {
		t.Fatal("falta copia previa")
	}
	copy, err := sql.Open("sqlite", paths[0])
	if err != nil {
		t.Fatal(err)
	}
	defer copy.Close()
	if err = copy.QueryRow("PRAGMA user_version").Scan(&version); err != nil || version != 1 {
		t.Fatal("la copia no conserva esquema anterior")
	}
	var note string
	if err = copy.QueryRow("SELECT notes FROM books WHERE id='book'").Scan(&note); err != nil || note != "Conservar" {
		t.Fatal("copia previa incompleta")
	}
}
