package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
)

func readonly(path string) (*sql.DB, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > backup.MaxFileSize {
		return nil, errors.New("el respaldo no es un archivo regular o supera 256 MiB")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	uriPath := filepath.ToSlash(abs)
	if filepath.VolumeName(abs) != "" && !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}
	u := url.URL{Scheme: "file", Path: uriPath, RawQuery: "mode=ro"}
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err = db.Exec("PRAGMA query_only=ON; PRAGMA trusted_schema=OFF;"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func integrity(ctx context.Context, db *sql.DB) error {
	var result string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return errors.New("el respaldo SQLite no supera la comprobación de integridad")
	}
	rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return errors.New("el respaldo contiene referencias rotas")
	}
	return rows.Err()
}

func markNative(ctx context.Context, db *sql.DB, generation, created string) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS trove_backup (id INTEGER PRIMARY KEY CHECK(id=1),format TEXT NOT NULL,version INTEGER NOT NULL,generation TEXT NOT NULL,created_at TEXT NOT NULL);
DELETE FROM trove_backup;
INSERT INTO trove_backup(id,format,version,generation,created_at) VALUES(1,?,?,?,?)`, backup.Format, backup.Version, generation, created)
	return err
}

func nativeCopy(ctx context.Context, db *sql.DB, path, generation, created string) error {
	if _, err := db.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return err
	}
	copy, err := sql.Open("sqlite", path)
	if err != nil {
		return err
	}
	if err = markNative(ctx, copy, generation, created); err != nil {
		copy.Close()
		return err
	}
	return copy.Close()
}

func exportCatalog(ctx context.Context, db *sql.DB, doc backup.Document) (backup.Document, error) {
	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return doc, err
	}
	if version < 1 || version > 5 {
		return doc, errors.New("esquema SQLite no compatible con esta versión de Trove")
	}
	medium := "'books'"
	if version >= 3 {
		medium = "medium"
	}
	rows, err := db.QueryContext(ctx, "SELECT id,name,"+medium+" FROM series ORDER BY id")
	if err != nil {
		return doc, err
	}
	for rows.Next() {
		var v backup.Series
		if err = rows.Scan(&v.ID, &v.Name, &v.Medium); err != nil {
			rows.Close()
			return doc, err
		}
		if doc.Version == 1 {
			if v.Medium != books.MediumBooks {
				rows.Close()
				return doc, errors.New("el formato 1 no admite novelas ligeras")
			}
			v.Medium = ""
		}
		doc.Catalog.Series = append(doc.Catalog.Series, v)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return doc, err
	}
	favorite, owned := "favorite", "owned"
	if version == 1 {
		favorite = "0"
		owned = "0"
	}
	rows, err = db.QueryContext(ctx, "SELECT id,title,authors_json,notes,added_on,series_id,series_position,"+favorite+","+medium+" FROM books ORDER BY id")
	if err != nil {
		return doc, err
	}
	for rows.Next() {
		var v backup.Work
		var authors string
		var series sql.NullString
		var position sql.NullFloat64
		if err = rows.Scan(&v.ID, &v.Title, &authors, &v.Notes, &v.AddedOn, &series, &position, &v.Favorite, &v.Medium); err != nil {
			rows.Close()
			return doc, err
		}
		if err = json.Unmarshal([]byte(authors), &v.Authors); err != nil {
			rows.Close()
			return doc, err
		}
		v.SeriesID = series.String
		if position.Valid {
			v.Position = &position.Float64
		}
		doc.Catalog.Works = append(doc.Catalog.Works, v)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return doc, err
	}
	identifiers, metadata := "'[]'", "NULL"
	if version >= 4 {
		identifiers = "identifiers_json"
		metadata = "catalog_json"
	}
	rows, err = db.QueryContext(ctx, "SELECT id,book_id,language,format,identifier,"+owned+","+identifiers+","+metadata+" FROM editions ORDER BY id")
	if err != nil {
		return doc, err
	}
	for rows.Next() {
		var v backup.Edition
		var rawIDs string
		var rawMetadata sql.NullString
		if err = rows.Scan(&v.ID, &v.WorkID, &v.Language, &v.Format, &v.Identifier, &v.Owned, &rawIDs, &rawMetadata); err != nil {
			rows.Close()
			return doc, err
		}
		if err = json.Unmarshal([]byte(rawIDs), &v.Identifiers); err != nil {
			rows.Close()
			return doc, err
		}
		if rawMetadata.Valid {
			if err = json.Unmarshal([]byte(rawMetadata.String), &v.Catalog); err != nil {
				rows.Close()
				return doc, err
			}
		}
		if doc.Version < 3 && (len(v.Identifiers) > 0 || v.Catalog != nil) {
			rows.Close()
			return doc, errors.New("el respaldo usa un formato antiguo para datos extendidos")
		}
		if doc.Version >= 3 && len(v.Identifiers) == 0 {
			v.Identifiers = catalog.NormalizeIdentifier(v.Identifier, "manual")
		}
		doc.Catalog.Editions = append(doc.Catalog.Editions, v)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return doc, err
	}
	rows, err = db.QueryContext(ctx, "SELECT id,edition_id,status,started_on,finished_on FROM reading_cycles ORDER BY id")
	if err != nil {
		return doc, err
	}
	for rows.Next() {
		var v backup.Cycle
		if err = rows.Scan(&v.ID, &v.EditionID, &v.Status, &v.StartedOn, &v.FinishedOn); err != nil {
			rows.Close()
			return doc, err
		}
		doc.Catalog.Cycles = append(doc.Catalog.Cycles, v)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return doc, err
	}
	if version >= 5 {
		doc.Catalog.Novels, err = readNovels(ctx, db)
		if err != nil {
			return doc, err
		}
	}
	return doc, doc.Validate()
}

// Capture exports the JSON model from the completed SQLite copy, not from a later live read.
func (s *Store) Capture(ctx context.Context, path, generation, created string) (backup.Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if err := nativeCopy(ctx, s.db, path, generation, created); err != nil {
		return backup.Document{}, err
	}
	return s.ReadNative(ctx, path)
}

func (s *Store) ReadNative(ctx context.Context, path string) (backup.Document, error) {
	var doc backup.Document
	db, err := readonly(path)
	if err != nil {
		return doc, err
	}
	defer db.Close()
	if err = integrity(ctx, db); err != nil {
		return doc, err
	}
	if err = db.QueryRowContext(ctx, "SELECT format,version,generation,created_at FROM trove_backup WHERE id=1").Scan(&doc.Format, &doc.Version, &doc.Generation, &doc.CreatedAt); err != nil {
		return doc, errors.New("no es un respaldo SQLite identificado de Trove")
	}
	if doc.Format != backup.Format || !backup.SupportedVersion(doc.Version) {
		return doc, errors.New("versión de respaldo SQLite no admitida")
	}
	return exportCatalog(ctx, db, doc)
}

func (s *Store) Prepare(ctx context.Context, doc backup.Document, path string) error {
	if err := doc.Validate(); err != nil {
		return err
	}
	prepared, err := Open(path)
	if err != nil {
		return err
	}
	defer prepared.Close()
	tx, err := prepared.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, v := range doc.Catalog.Series {
		medium := v.Medium
		if doc.Version == 1 {
			medium = books.MediumBooks
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO series(id,name,name_key,medium) VALUES(?,?,?,?)", v.ID, v.Name, seriesKey(medium, v.Name), medium); err != nil {
			return err
		}
	}
	for _, v := range doc.Catalog.Works {
		authors, _ := json.Marshal(v.Authors)
		var series any
		if v.SeriesID != "" {
			series = v.SeriesID
		}
		var position any
		if v.Position != nil {
			position = *v.Position
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO books(id,title,authors_json,notes,added_on,series_id,series_position,favorite,medium) VALUES(?,?,?,?,?,?,?,?,?)", v.ID, v.Title, string(authors), v.Notes, v.AddedOn, series, position, v.Favorite, v.Medium); err != nil {
			return err
		}
	}
	for _, v := range doc.Catalog.Editions {
		ids, _ := json.Marshal(v.Identifiers)
		var metadata any
		if v.Catalog != nil {
			data, _ := json.Marshal(v.Catalog)
			metadata = string(data)
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO editions(id,book_id,language,format,identifier,owned,identifiers_json,catalog_json) VALUES(?,?,?,?,?,?,?,?)", v.ID, v.WorkID, v.Language, v.Format, v.Identifier, v.Owned, string(ids), metadata); err != nil {
			return err
		}
	}
	for _, v := range doc.Catalog.Cycles {
		if _, err = tx.ExecContext(ctx, "INSERT INTO reading_cycles(id,edition_id,status,started_on,finished_on) VALUES(?,?,?,?,?)", v.ID, v.EditionID, v.Status, v.StartedOn, v.FinishedOn); err != nil {
			return err
		}
	}
	for _, n := range doc.Catalog.Novels {
		if err = writeNovel(ctx, tx, n); err != nil {
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	if err = integrity(ctx, prepared.db); err != nil {
		return err
	}
	read, err := exportCatalog(ctx, prepared.db, backup.Document{Format: backup.Format, Version: doc.Version, Generation: doc.Generation, CreatedAt: doc.CreatedAt})
	if err != nil {
		return err
	}
	if read.ContentHash() != doc.ContentHash() {
		return errors.New("la base reconstruida no coincide con el respaldo")
	}
	return nil
}

// Install serializes all app access, retains a standalone safety backup and rolls back renames.
func (s *Store) Install(ctx context.Context, prepared string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := rand.Text()
	created := time.Now().UTC().Format(time.RFC3339Nano)
	safety := filepath.Join(filepath.Dir(s.path), "before-restore-"+id+".sqlite")
	if err := nativeCopy(ctx, s.db, safety, id, created); err != nil {
		return "", fmt.Errorf("no se pudo resguardar la biblioteca actual: %w", err)
	}
	if _, err := s.ReadNative(ctx, safety); err != nil {
		return safety, fmt.Errorf("la copia previa no pudo validarse; la biblioteca actual no se reemplazó: %w", err)
	}
	var mode string
	if err := s.db.QueryRowContext(ctx, "PRAGMA journal_mode=DELETE").Scan(&mode); err != nil {
		return safety, err
	}
	if mode != "delete" {
		return safety, errors.New("no se pudo finalizar el journal antes de restaurar")
	}
	if err := s.db.Close(); err != nil {
		previous, reopenErr := Open(s.path)
		if reopenErr == nil {
			s.db = previous.db
		}
		return safety, fmt.Errorf("no se pudo cerrar la base activa: %w; reapertura: %v", err, reopenErr)
	}
	old := filepath.Join(filepath.Dir(s.path), ".restore-old-"+id+".sqlite")
	reopen := func() error {
		v, err := Open(s.path)
		if err == nil {
			s.db = v.db
		}
		return err
	}
	if err := s.rename(s.path, old); err != nil {
		reopenErr := reopen()
		return safety, fmt.Errorf("no se pudo cerrar/reemplazar la base (cerrá otras instancias): %w; reapertura: %v", err, reopenErr)
	}
	if err := s.rename(prepared, s.path); err != nil {
		undo := s.rename(old, s.path)
		if undo != nil {
			return safety, fmt.Errorf("falló el reemplazo y la recuperación; copia disponible en %s: %w", safety, undo)
		}
		reopenErr := reopen()
		return safety, fmt.Errorf("restauración cancelada, biblioteca anterior recuperada: %w; reapertura: %v", err, reopenErr)
	}
	if err := reopen(); err != nil {
		failed := filepath.Join(filepath.Dir(s.path), ".restore-failed-"+id+".sqlite")
		if moveErr := s.rename(s.path, failed); moveErr != nil {
			return safety, fmt.Errorf("falló abrir la biblioteca restaurada; copia anterior en %s: %w", safety, moveErr)
		}
		if undo := s.rename(old, s.path); undo != nil {
			return safety, fmt.Errorf("falló la recuperación; copia anterior en %s: %w", safety, undo)
		}
		reopenErr := reopen()
		return safety, fmt.Errorf("base restaurada rechazada; biblioteca anterior recuperada: %w; reapertura: %v", err, reopenErr)
	}
	// The independently readable safety copy remains even after success.
	_ = os.Remove(old)
	return safety, nil
}
