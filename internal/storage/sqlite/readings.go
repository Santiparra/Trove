package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/Santiparra/Trove/internal/books"
)

// Core dates/status stay queryable; reading-specific personal data is independent of the current edition.
type cycleDetails struct {
	Rating  *float64              `json:"rating,omitempty"`
	Edition *books.ReadingEdition `json:"edition,omitempty"`
	RunID   string                `json:"runID,omitempty"`
}

func cycleJSON(c books.Cycle) (string, error) {
	data, err := json.Marshal(cycleDetails{c.Rating, c.Edition, c.RunID})
	return string(data), err
}
func readCycleDetails(raw string, c *books.Cycle) error {
	var d cycleDetails
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return err
	}
	c.Rating, c.Edition, c.RunID = d.Rating, d.Edition, d.RunID
	return nil
}
func readHistory(ctx context.Context, db *sql.DB, editionID string) ([]books.Cycle, error) {
	rows, err := db.QueryContext(ctx, "SELECT id,status,started_on,finished_on,details_json FROM reading_cycles WHERE edition_id=? AND is_current=0 ORDER BY position,id", editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []books.Cycle{}
	for rows.Next() {
		var c books.Cycle
		var raw string
		if err = rows.Scan(&c.ID, &c.Status, &c.StartedOn, &c.FinishedOn, &raw); err != nil {
			return nil, err
		}
		if err = readCycleDetails(raw, &c); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
func writeBookReadings(ctx context.Context, tx *sql.Tx, b books.Book) error {
	seen := map[string]bool{}
	for _, c := range append(append([]books.Cycle{}, b.History...), b.Cycle) {
		if seen[c.ID] {
			return errors.New("identidad de lectura duplicada")
		}
		seen[c.ID] = true
		if err := c.Validate(); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM reading_cycles WHERE edition_id=?", b.Edition.ID); err != nil {
		return err
	}
	for i, c := range append(append([]books.Cycle{}, b.History...), b.Cycle) {
		raw, err := cycleJSON(c)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO reading_cycles(id,edition_id,status,started_on,finished_on,is_current,position,details_json) VALUES(?,?,?,?,?,?,?,?)`, c.ID, b.Edition.ID, c.Status, c.StartedOn, c.FinishedOn, i == len(b.History), i, raw); err != nil {
			return err
		}
	}
	return nil
}
