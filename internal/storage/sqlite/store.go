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
	"time"

	"github.com/Santiparra/Trove/internal/books"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if err = store.initialize(path); err != nil {
		db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) initialize(path string) error {
	if _, err := s.db.Exec("PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;"); err != nil {
		return err
	}
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	files := []string{"migrations/001_books.sql", "migrations/002_personal_flags.sql"}
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
		if _, err = tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", i+1)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

const selectBooks = `SELECT b.id, b.title, b.authors_json, b.notes, b.added_on,
 s.id, s.name, b.series_position, b.favorite,
 e.id, e.language, e.format, e.identifier, e.owned, c.id, c.status, c.started_on, c.finished_on
 FROM books b
 LEFT JOIN series s ON s.id=b.series_id
 JOIN editions e ON e.book_id=b.id
 JOIN reading_cycles c ON c.edition_id=e.id`

type scanner interface{ Scan(...any) error }

func scanBook(row scanner) (books.Book, error) {
	var b books.Book
	var authors string
	var sid, name sql.NullString
	var pos sql.NullFloat64
	err := row.Scan(&b.ID, &b.Title, &authors, &b.Notes, &b.AddedOn, &sid, &name, &pos, &b.Favorite,
		&b.Edition.ID, &b.Edition.Language, &b.Edition.Format, &b.Edition.Identifier, &b.Edition.Owned,
		&b.Cycle.ID, &b.Cycle.Status, &b.Cycle.StartedOn, &b.Cycle.FinishedOn)
	if errors.Is(err, sql.ErrNoRows) {
		return b, books.ErrNotFound
	}
	if err != nil {
		return b, err
	}
	if err = json.Unmarshal([]byte(authors), &b.Authors); err != nil {
		return b, err
	}
	if sid.Valid {
		b.Series = &books.Series{ID: sid.String, Name: name.String}
		if pos.Valid {
			b.Series.Position = &pos.Float64
		}
	}
	return b, nil
}

func (s *Store) Get(ctx context.Context, id string) (books.Book, error) {
	return scanBook(s.db.QueryRowContext(ctx, selectBooks+" WHERE b.id=?", id))
}

func (s *Store) List(ctx context.Context, query string) ([]books.Book, error) {
	rows, err := s.db.QueryContext(ctx, selectBooks+" ORDER BY b.title COLLATE NOCASE, b.id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]books.Book, 0)
	query = strings.ToLower(strings.TrimSpace(query))
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		// Go case folding avoids SQLite's ASCII-only LOWER for non-English titles.
		if query == "" || strings.Contains(strings.ToLower(b.Title), query) {
			result = append(result, b)
		}
	}
	return result, rows.Err()
}

func (s *Store) ListSeries(ctx context.Context) ([]books.Series, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id,name FROM series WHERE id IN (SELECT series_id FROM books) ORDER BY name COLLATE NOCASE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]books.Series, 0)
	for rows.Next() {
		var v books.Series
		if err := rows.Scan(&v.ID, &v.Name); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, rows.Err()
}

func (s *Store) Save(ctx context.Context, b books.Book) (books.Book, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return b, err
	}
	defer tx.Rollback()
	var seriesID any
	if b.Series != nil {
		key := strings.ToLower(strings.Join(strings.Fields(b.Series.Name), " "))
		id := rand.Text()
		if _, err = tx.ExecContext(ctx, "INSERT INTO series(id,name,name_key) VALUES(?,?,?) ON CONFLICT(name_key) DO NOTHING", id, b.Series.Name, key); err != nil {
			return b, err
		}
		if err = tx.QueryRowContext(ctx, "SELECT id,name FROM series WHERE name_key=?", key).Scan(&b.Series.ID, &b.Series.Name); err != nil {
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
	_, err = tx.ExecContext(ctx, `INSERT INTO books(id,title,authors_json,notes,added_on,series_id,series_position,favorite) VALUES(?,?,?,?,?,?,?,?)
 ON CONFLICT(id) DO UPDATE SET title=excluded.title,authors_json=excluded.authors_json,notes=excluded.notes,added_on=excluded.added_on,series_id=excluded.series_id,series_position=excluded.series_position,favorite=excluded.favorite`,
		b.ID, b.Title, string(authors), b.Notes, b.AddedOn, seriesID, position, b.Favorite)
	if err != nil {
		return b, fmt.Errorf("guardar libro: %w", err)
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO editions(id,book_id,language,format,identifier,owned) VALUES(?,?,?,?,?,?)
 ON CONFLICT(id) DO UPDATE SET language=excluded.language,format=excluded.format,identifier=excluded.identifier,owned=excluded.owned`, b.Edition.ID, b.ID, b.Edition.Language, b.Edition.Format, b.Edition.Identifier, b.Edition.Owned)
	if err != nil {
		return b, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO reading_cycles(id,edition_id,status,started_on,finished_on) VALUES(?,?,?,?,?)
 ON CONFLICT(id) DO UPDATE SET status=excluded.status,started_on=excluded.started_on,finished_on=excluded.finished_on`, b.Cycle.ID, b.Edition.ID, b.Cycle.Status, b.Cycle.StartedOn, b.Cycle.FinishedOn)
	if err != nil {
		return b, err
	}
	if err = tx.Commit(); err != nil {
		return b, err
	}
	return b, nil
}

func (s *Store) UpdateCycle(ctx context.Context, bookID string, cycle books.Cycle) error {
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
	return requireUpdated(s.db.ExecContext(ctx, "UPDATE books SET notes=? WHERE id=?", notes, id))
}
func (s *Store) UpdateFavorite(ctx context.Context, id string, value bool) error {
	return requireUpdated(s.db.ExecContext(ctx, "UPDATE books SET favorite=? WHERE id=?", value, id))
}
func (s *Store) UpdateOwned(ctx context.Context, bookID, editionID string, value bool) error {
	return requireUpdated(s.db.ExecContext(ctx, "UPDATE editions SET owned=? WHERE id=? AND book_id=?", value, editionID, bookID))
}

func (s *Store) Delete(ctx context.Context, id string) error {
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
	return tx.Commit()
}
