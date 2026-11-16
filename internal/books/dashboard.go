package books

import (
	"context"
	"sort"
)

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type Dashboard struct {
	Total            int           `json:"total"`
	Counts           []StatusCount `json:"counts"`
	CurrentlyReading []Book        `json:"currentlyReading"`
	RecentlyFinished []Book        `json:"recentlyFinished"`
	ReadWithoutDate  int           `json:"readWithoutDate"`
}

func (s *Service) Dashboard(ctx context.Context) (Dashboard, error) {
	items, err := s.repository.List(ctx, "")
	if err != nil {
		return Dashboard{}, err
	}
	d := Dashboard{Total: len(items), Counts: make([]StatusCount, 0), CurrentlyReading: make([]Book, 0), RecentlyFinished: make([]Book, 0)}
	counts := map[string]int{}
	for _, b := range items {
		counts[b.Cycle.Status]++
		if b.Cycle.Status == "reading" {
			d.CurrentlyReading = append(d.CurrentlyReading, b)
		}
		if b.Cycle.Status == "read" {
			if b.Cycle.FinishedOn == "" {
				d.ReadWithoutDate++
			} else {
				d.RecentlyFinished = append(d.RecentlyFinished, b)
			}
		}
	}
	for _, status := range []string{"reading", "up-next", "read", "paused", "did-not-finish", "to-read"} {
		d.Counts = append(d.Counts, StatusCount{status, counts[status]})
	}
	sort.Slice(d.RecentlyFinished, func(i, j int) bool {
		a, b := d.RecentlyFinished[i], d.RecentlyFinished[j]
		if a.Cycle.FinishedOn == b.Cycle.FinishedOn {
			return a.ID < b.ID
		}
		return a.Cycle.FinishedOn > b.Cycle.FinishedOn
	})
	if len(d.RecentlyFinished) > 5 {
		d.RecentlyFinished = d.RecentlyFinished[:5]
	}
	if len(d.CurrentlyReading) > 6 {
		d.CurrentlyReading = d.CurrentlyReading[:6]
	}
	return d, nil
}
