// Package sqlite implements persistence; SQL never crosses into the book service.
package sqlite

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct {
	db       *sql.DB
	mu       sync.RWMutex
	path     string
	onChange func()
	rename   func(string, string) error
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db, path: path, rename: os.Rename}
	if err = store.initialize(path); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error          { s.mu.Lock(); defer s.mu.Unlock(); return s.db.Close() }
func (s *Store) SetOnChange(fn func()) { s.mu.Lock(); defer s.mu.Unlock(); s.onChange = fn }
func (s *Store) changed() {
	if s.onChange != nil {
		s.onChange()
	}
}

func (s *Store) initialize(path string) error {
	if _, err := s.db.Exec("PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;"); err != nil {
		return err
	}
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	files := []string{"migrations/001_books.sql", "migrations/002_personal_flags.sql", "migrations/003_reading_media.sql", "migrations/004_catalog.sql", "migrations/005_novels.sql", "migrations/006_novel_catalog.sql", "migrations/007_reading_history.sql", "migrations/008_classification.sql"}
	if version > len(files) {
		return errors.New("la base pertenece a una versión más nueva de Trove")
	}
	if version == len(files) {
		return nil
	}
	if version > 0 {
		backup := filepath.Join(filepath.Dir(path), fmt.Sprintf("before-schema-%d-%s.sqlite", len(files), time.Now().Format("20060102-150405.000000000")))
		if _, err := s.db.Exec("VACUUM INTO ?", backup); err != nil {
			return fmt.Errorf("no se pudo crear la copia previa a la migración: %w", err)
		}
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for i := version; i < len(files); i++ {
		ddl, err := migrations.ReadFile(files[i])
		if err != nil {
			return err
		}
		if _, err = tx.Exec(string(ddl)); err != nil {
			return err
		}
		if i == 3 {
			if err = normalizeLegacyIdentifiers(tx); err != nil {
				return err
			}
		}
		if _, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

const selectBooks = `SELECT b.id, b.title, b.authors_json, b.notes, b.added_on,
 s.id, s.name, b.series_position, b.favorite,
 e.id, e.language, e.format, e.identifier, e.owned, c.id, c.status, c.started_on, c.finished_on, b.medium, s.medium, e.identifiers_json, e.catalog_json, c.details_json, b.classification_json
 FROM books b
 LEFT JOIN series s ON s.id=b.series_id
 JOIN editions e ON e.book_id=b.id
 JOIN reading_cycles c ON c.edition_id=e.id AND c.is_current=1`

type scanner interface{ Scan(...any) error }

func scanBook(row scanner) (books.Book, error) {
	var b books.Book
	var authors string
	var sid, name, seriesMedium sql.NullString
	var pos sql.NullFloat64
	var ids string
	var metadata sql.NullString
	var details string
	var rawClassification sql.NullString
	err := row.Scan(&b.ID, &b.Title, &authors, &b.Notes, &b.AddedOn, &sid, &name, &pos, &b.Favorite,
		&b.Edition.ID, &b.Edition.Language, &b.Edition.Format, &b.Edition.Identifier, &b.Edition.Owned,
		&b.Cycle.ID, &b.Cycle.Status, &b.Cycle.StartedOn, &b.Cycle.FinishedOn, &b.Medium, &seriesMedium, &ids, &metadata, &details, &rawClassification)
	if errors.Is(err, sql.ErrNoRows) {
		return b, books.ErrNotFound
	}
	if err != nil {
		return b, err
	}
	if err = readCycleDetails(details, &b.Cycle); err != nil {
		return b, err
	}
	if err = json.Unmarshal([]byte(authors), &b.Authors); err != nil {
		return b, err
	}
	if err = json.Unmarshal([]byte(ids), &b.Edition.Identifiers); err != nil {
		return b, err
	}
	if len(b.Edition.Identifiers) == 0 {
		b.Edition.Identifiers = catalog.NormalizeIdentifier(b.Edition.Identifier, "manual")
	}
	if metadata.Valid {
		if err = json.Unmarshal([]byte(metadata.String), &b.Edition.Catalog); err != nil {
			return b, err
		}
	}
	if rawClassification.Valid {
		if err = json.Unmarshal([]byte(rawClassification.String), &b.Classification); err != nil {
			return b, err
		}
	}
	if sid.Valid {
		b.Series = &books.Series{ID: sid.String, Name: name.String, Medium: seriesMedium.String}
		if pos.Valid {
			b.Series.Position = &pos.Float64
		}
	}
	return b, nil
}

func (s *Store) Get(ctx context.Context, id string) (books.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, err := scanBook(s.db.QueryRowContext(ctx, selectBooks+" WHERE b.id=?", id))
	if err == nil {
		b.History, err = readHistory(ctx, s.db, b.Edition.ID)
	}
	return b, err
}

func (s *Store) List(ctx context.Context, query, medium string) ([]books.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !books.ValidMedium(medium) {
		return nil, errors.New("medio no válido")
	}
	rows, err := s.db.QueryContext(ctx, selectBooks+" WHERE b.medium=? ORDER BY b.title COLLATE NOCASE, b.id", medium)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]books.Book, 0)
	query = strings.ToLower(strings.TrimSpace(query))
	queryISBN := catalog.ISBN13(query)
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		// Go case folding avoids SQLite's ASCII-only LOWER for non-English text.
		text := b.Title + " " + strings.Join(b.Authors, " ") + " " + b.Edition.Identifier
		if b.Series != nil {
			text += " " + b.Series.Name
		}
		isbnMatch := false
		for _, id := range b.Edition.Identifiers {
			text += " " + id.Value
			if queryISBN != "" && catalog.ISBN13(id.Value) == queryISBN {
				isbnMatch = true
			}
		}
		if query == "" || strings.Contains(strings.ToLower(text), query) || isbnMatch {
			result = append(result, b)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()
	for i := range result {
		result[i].History, err = readHistory(ctx, s.db, result[i].Edition.ID)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *Store) ListSeries(ctx context.Context, medium string) ([]books.Series, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !books.ValidMedium(medium) {
		return nil, errors.New("medio no válido")
	}
	rows, err := s.db.QueryContext(ctx, "SELECT id,name,medium FROM series WHERE medium=? AND id IN (SELECT series_id FROM books WHERE medium=?) ORDER BY name COLLATE NOCASE", medium, medium)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]books.Series, 0)
	for rows.Next() {
		var v books.Series
		if err := rows.Scan(&v.ID, &v.Name, &v.Medium); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (s *Store) Save(ctx context.Context, b books.Book) (books.Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !books.ValidMedium(b.Medium) {
		return b, errors.New("medio no válido")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return b, err
	}
	defer tx.Rollback()
	var seriesID any
	if b.Series != nil {
		key := seriesKey(b.Medium, b.Series.Name)
		id := rand.Text()
		if _, err = tx.ExecContext(ctx, "INSERT INTO series(id,name,name_key,medium) VALUES(?,?,?,?) ON CONFLICT(name_key) DO NOTHING", id, b.Series.Name, key, b.Medium); err != nil {
			return b, err
		}
		if err = tx.QueryRowContext(ctx, "SELECT id,name,medium FROM series WHERE name_key=? AND medium=?", key, b.Medium).Scan(&b.Series.ID, &b.Series.Name, &b.Series.Medium); err != nil {
			return b, err
		}
		seriesID = b.Series.ID
	}
	authors, err := json.Marshal(b.Authors)
	if err != nil {
		return b, err
	}
	var position any
	if b.Series != nil && b.Series.Position != nil {
		position = *b.Series.Position
	}
	var personalClassification any
	if b.Classification != nil {
		raw, e := json.Marshal(b.Classification)
		if e != nil {
			return b, e
		}
		personalClassification = string(raw)
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO books(id,title,authors_json,notes,added_on,series_id,series_position,favorite,medium,classification_json) VALUES(?,?,?,?,?,?,?,?,?,?)
 ON CONFLICT(id) DO UPDATE SET title=excluded.title,authors_json=excluded.authors_json,notes=excluded.notes,added_on=excluded.added_on,series_id=excluded.series_id,series_position=excluded.series_position,favorite=excluded.favorite,classification_json=excluded.classification_json WHERE books.medium=excluded.medium`,
		b.ID, b.Title, string(authors), b.Notes, b.AddedOn, seriesID, position, b.Favorite, b.Medium, personalClassification)
	if err != nil {
		return b, fmt.Errorf("guardar libro: %w", err)
	}
	if err = requireUpdated(result, nil); err != nil {
		return b, err
	}
	ids, err := json.Marshal(b.Edition.Identifiers)
	if err != nil {
		return b, err
	}
	var metadata any
	if b.Edition.Catalog != nil {
		data, err := json.Marshal(b.Edition.Catalog)
		if err != nil {
			return b, err
		}
		metadata = string(data)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO editions(id,book_id,language,format,identifier,owned,identifiers_json,catalog_json) VALUES(?,?,?,?,?,?,?,?)
 ON CONFLICT(id) DO UPDATE SET language=excluded.language,format=excluded.format,identifier=excluded.identifier,owned=excluded.owned,identifiers_json=excluded.identifiers_json,catalog_json=excluded.catalog_json`, b.Edition.ID, b.ID, b.Edition.Language, b.Edition.Format, b.Edition.Identifier, b.Edition.Owned, string(ids), metadata)
	if err != nil {
		return b, err
	}
	if err = writeBookReadings(ctx, tx, b); err != nil {
		return b, err
	}
	if err = tx.Commit(); err != nil {
		return b, err
	}
	s.changed()
	return b, nil
}

func seriesKey(medium, name string) string {
	return medium + ":" + strings.ToLower(strings.Join(strings.Fields(name), " "))
}

func normalizeLegacyIdentifiers(tx *sql.Tx) error {
	rows, err := tx.Query("SELECT id,identifier FROM editions")
	if err != nil {
		return err
	}
	type entry struct{ id, value string }
	items := []entry{}
	for rows.Next() {
		var v entry
		if err = rows.Scan(&v.id, &v.value); err != nil {
			rows.Close()
			return err
		}
		items = append(items, v)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, v := range items {
		data, _ := json.Marshal(catalog.NormalizeIdentifier(v.value, "manual"))
		if _, err = tx.Exec("UPDATE editions SET identifiers_json=? WHERE id=?", string(data), v.id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) UpdateCatalog(ctx context.Context, id, editionID string, snapshot *catalog.Snapshot, identifiers []catalog.Identifier) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids, err := json.Marshal(identifiers)
	if err != nil {
		return err
	}
	var metadata any
	if snapshot != nil {
		data, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		metadata = string(data)
	}
	err = requireUpdated(s.db.ExecContext(ctx, "UPDATE editions SET catalog_json=?,identifiers_json=? WHERE id=? AND book_id=?", metadata, string(ids), editionID, id))
	if err == nil {
		s.changed()
	}
	return err
}

func (s *Store) UpdateCycle(ctx context.Context, bookID string, cycle books.Cycle) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, err := s.db.ExecContext(ctx, `UPDATE reading_cycles SET status=?,started_on=?,finished_on=? WHERE id=? AND edition_id IN (SELECT id FROM editions WHERE book_id=?)`, cycle.Status, cycle.StartedOn, cycle.FinishedOn, cycle.ID, bookID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return books.ErrNotFound
	}
	s.changed()
	return nil
}

func requireUpdated(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return books.ErrNotFound
	}
	return nil
}
func (s *Store) UpdateNotes(ctx context.Context, id, notes string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	err := requireUpdated(s.db.ExecContext(ctx, "UPDATE books SET notes=? WHERE id=?", notes, id))
	if err == nil {
		s.changed()
	}
	return err
}
func (s *Store) UpdateFavorite(ctx context.Context, id string, value bool) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	err := requireUpdated(s.db.ExecContext(ctx, "UPDATE books SET favorite=? WHERE id=?", value, id))
	if err == nil {
		s.changed()
	}
	return err
}
func (s *Store) UpdateOwned(ctx context.Context, bookID, editionID string, value bool) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	err := requireUpdated(s.db.ExecContext(ctx, "UPDATE editions SET owned=? WHERE id=? AND book_id=?", value, editionID, bookID))
	if err == nil {
		s.changed()
	}
	return err
}

func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM reading_cycles WHERE edition_id IN (SELECT id FROM editions WHERE book_id=?)`, id); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM editions WHERE book_id=?", id); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, "DELETE FROM books WHERE id=?", id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return books.ErrNotFound
	}
	if err = tx.Commit(); err == nil {
		s.changed()
	}
	return err
}
