package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/novels"
)

func readNovels(ctx context.Context, db *sql.DB) ([]novels.Novel, error) {
	items := []novels.Novel{}
	rows, err := db.QueryContext(ctx, "SELECT id,data_json FROM novels ORDER BY id")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id, raw string
		var n novels.Novel
		if err = rows.Scan(&id, &raw); err == nil {
			err = json.Unmarshal([]byte(raw), &n)
		}
		if err != nil {
			rows.Close()
			return nil, err
		}
		if id != n.ID {
			rows.Close()
			return nil, errors.New("identidad de novela inconsistente")
		}
		n.Volumes = []novels.Volume{}
		items = append(items, n)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	for i := range items {
		rows, err = db.QueryContext(ctx, "SELECT id,number,data_json FROM novel_volumes WHERE novel_id=? ORDER BY number", items[i].ID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id, raw string
			var number int
			var v novels.Volume
			if err = rows.Scan(&id, &number, &raw); err == nil {
				err = json.Unmarshal([]byte(raw), &v)
			}
			if err != nil {
				rows.Close()
				return nil, err
			}
			if id != v.ID || number != v.Number {
				rows.Close()
				return nil, errors.New("identidad de volumen inconsistente")
			}
			items[i].Volumes = append(items[i].Volumes, v)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		items[i].Canonicalize()
		if err = items[i].Validate(); err != nil {
			return nil, err
		}
	}
	return items, nil
}
func writeNovel(ctx context.Context, tx *sql.Tx, n novels.Novel) error {
	volumes := n.Volumes
	n.Volumes = nil
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "INSERT INTO novels(id,data_json) VALUES(?,?) ON CONFLICT(id) DO UPDATE SET data_json=excluded.data_json", n.ID, string(data)); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, "DELETE FROM novel_volumes WHERE novel_id=?", n.ID); err != nil {
		return err
	}
	for _, v := range volumes {
		data, err = json.Marshal(v)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO novel_volumes(id,novel_id,number,data_json) VALUES(?,?,?,?)", v.ID, n.ID, v.Number, string(data)); err != nil {
			return err
		}
	}
	return nil
}
func (s *Store) ListNovels(ctx context.Context, query string) ([]novels.Novel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items, err := readNovels(ctx, s.db)
	if err != nil {
		return nil, err
	}
	result := []novels.Novel{}
	query = strings.ToLower(strings.TrimSpace(query))
	queryISBN := catalog.ISBN13(query)
	for _, n := range items {
		text := n.Title + " " + strings.Join(n.Authors, " ")
		isbnMatch := false
		if n.Catalog != nil {
			text += " " + n.Catalog.Title + " " + n.Catalog.OriginalTitle
			for _, a := range n.Catalog.Authors {
				text += " " + a.Name
			}
		}
		for _, v := range n.Volumes {
			text += " " + v.Title + " " + v.Edition.Identifier
			if queryISBN != "" && catalog.ISBN13(v.Edition.Identifier) == queryISBN {
				isbnMatch = true
			}
			if v.Catalog != nil {
				text += " " + v.Catalog.Title + " " + v.Catalog.OriginalTitle + " " + v.Catalog.ISBN
				if queryISBN != "" && v.Catalog.ISBN == queryISBN {
					isbnMatch = true
				}
			}
		}
		if query == "" || strings.Contains(strings.ToLower(text), query) || isbnMatch {
			result = append(result, n)
		}
	}
	return result, nil
}
func (s *Store) GetNovel(ctx context.Context, id string) (novels.Novel, error) {
	items, err := s.ListNovels(ctx, "")
	if err != nil {
		return novels.Novel{}, err
	}
	for _, n := range items {
		if n.ID == id {
			return n, nil
		}
	}
	return novels.Novel{}, novels.ErrNotFound
}
func (s *Store) SaveNovel(ctx context.Context, n novels.Novel) (novels.Novel, error) {
	if err := n.Validate(); err != nil {
		return novels.Novel{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return novels.Novel{}, err
	}
	defer tx.Rollback()
	if err = writeNovel(ctx, tx, n); err != nil {
		return novels.Novel{}, err
	}
	if err = tx.Commit(); err != nil {
		return novels.Novel{}, err
	}
	s.changed()
	return n, nil
}
func (s *Store) DeleteNovel(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	result, err := s.db.ExecContext(ctx, "DELETE FROM novels WHERE id=?", id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return novels.ErrNotFound
	}
	s.changed()
	return nil
}
