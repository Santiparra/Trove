package catalog

import (
	"errors"
	"math"
	"time"
)

// Rating preserves the provider's aggregate, never a personal or blended score.
type Rating struct {
	Provider  string  `json:"provider"`
	Entity    string  `json:"entity"`
	ID        string  `json:"id"`
	Value     float64 `json:"value"`
	Scale     float64 `json:"scale"`
	Votes     *int    `json:"votes"`
	FetchedAt string  `json:"fetchedAt"`
}

func (r Rating) Validate() error {
	if !ValidID(r.Provider, r.ID) || (r.Entity != "work" && r.Entity != "series" && r.Entity != "book") || math.IsNaN(r.Value) || math.IsInf(r.Value, 0) || (r.Scale != 5 && r.Scale != 10) || r.Value < 0 || r.Value > r.Scale || (r.Votes != nil && *r.Votes < 1) {
		return errors.New("puntaje externo inválido")
	}
	if _, err := time.Parse(time.RFC3339Nano, r.FetchedAt); err != nil {
		return errors.New("fecha del puntaje externo inválida")
	}
	return nil
}
