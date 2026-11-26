package novels

import (
	"context"
	"errors"
	"github.com/Santiparra/Trove/internal/novelcatalog"
	"sort"
	"strconv"
)

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}
type PortalEntry struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Authors   []string `json:"authors"`
	Status    string   `json:"status"`
	Completed *int     `json:"completed"`
	Total     *int     `json:"total"`
	Number    int      `json:"number"`
	Label     string   `json:"label"`
	CoverURL  string   `json:"coverURL"`
}
type FinishedVolume struct {
	NovelID    string `json:"novelID"`
	NovelTitle string `json:"novelTitle"`
	Number     int    `json:"number"`
	Label      string `json:"label"`
	Title      string `json:"title"`
	CoverURL   string `json:"coverURL"`
	FinishedOn string `json:"finishedOn"`
}
type Dashboard struct {
	Total   int              `json:"total"`
	Counts  []StatusCount    `json:"counts"`
	Reading []PortalEntry    `json:"reading"`
	UpNext  []PortalEntry    `json:"upNext"`
	Recent  []FinishedVolume `json:"recent"`
}

func (n Novel) Label(number int) string {
	for _, v := range n.Volumes {
		if v.Number == number {
			if v.Label != "" {
				return v.Label
			}
			if v.Catalog != nil {
				if v.Catalog.Label != "" {
					return v.Catalog.Label
				}
				if label := novelcatalog.PrintedLabel(v.Catalog.Title, v.Catalog.OriginalTitle); label != "" {
					return label
				}
			}
		}
	}
	if n.Catalog != nil {
		if ref, ok := n.Catalog.Ref(number); ok {
			if ref.Label != "" {
				return ref.Label
			}
			if label := novelcatalog.PrintedLabel(ref.Title, ref.OriginalTitle); label != "" {
				return label
			}
			return "Lectura #" + strconv.Itoa(number)
		}
	}
	return strconv.Itoa(number)
}
func (n Novel) Cover(number int) string {
	for _, v := range n.Volumes {
		if v.Number == number && v.Catalog != nil && v.Catalog.CoverURL != "" {
			return v.Catalog.CoverURL
		}
	}
	if n.Catalog != nil {
		if ref, ok := n.Catalog.Ref(number); ok {
			return ref.CoverURL
		}
	}
	return ""
}
func (s *Service) Dashboard(ctx context.Context) (Dashboard, error) {
	items, err := s.repository.ListNovels(ctx, "")
	if err != nil {
		return Dashboard{}, err
	}
	d := Dashboard{Total: len(items), Counts: []StatusCount{}, Reading: []PortalEntry{}, UpNext: []PortalEntry{}, Recent: []FinishedVolume{}}
	counts := map[string]int{}
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.StatusChangedAt != b.StatusChangedAt {
			return a.StatusChangedAt > b.StatusChangedAt
		}
		if a.AddedOn != b.AddedOn {
			return a.AddedOn > b.AddedOn
		}
		return a.Title < b.Title
	})
	for _, n := range items {
		counts[n.Status]++
		number := n.NextNumber()
		if number == 0 && n.Completed != nil {
			number = *n.Completed
			if r := n.Run(); r != nil && r.Completed > 0 {
				for _, v := range n.Volumes {
					if v.ID == r.VolumeIDs[r.Completed-1] {
						number = v.Number
						break
					}
				}
			}
		}
		entry := PortalEntry{ID: n.ID, Title: n.Title, Authors: n.Authors, Status: n.Status, Completed: n.Completed, Total: n.ProgressTotal(), Number: number, Label: n.Label(number), CoverURL: n.Cover(number)}
		if n.Status == "reading" && len(d.Reading) < 4 {
			d.Reading = append(d.Reading, entry)
		}
		if n.Status == "up-next" && len(d.UpNext) < 5 {
			d.UpNext = append(d.UpNext, entry)
		}
		for _, v := range n.Volumes {
			if v.Cycle.Status == "read" && v.Cycle.FinishedOn != "" {
				d.Recent = append(d.Recent, FinishedVolume{NovelID: n.ID, NovelTitle: n.Title, Number: v.Number, Label: n.Label(v.Number), Title: v.Title, CoverURL: n.Cover(v.Number), FinishedOn: v.Cycle.FinishedOn})
			}
		}
	}
	for _, status := range []string{"to-read", "up-next", "reading", "read", "paused", "did-not-finish"} {
		d.Counts = append(d.Counts, StatusCount{Status: status, Count: counts[status]})
	}
	sort.Slice(d.Recent, func(i, j int) bool {
		a, b := d.Recent[i], d.Recent[j]
		if a.FinishedOn != b.FinishedOn {
			return a.FinishedOn > b.FinishedOn
		}
		if a.NovelID != b.NovelID {
			return a.NovelID < b.NovelID
		}
		return a.Number > b.Number
	})
	if len(d.Recent) > 5 {
		d.Recent = d.Recent[:5]
	}
	return d, nil
}

// SetCompleted is a historical correction: it never applies live reading dates.
func (s *Service) SetCompleted(ctx context.Context, id string, completed int, markRead bool) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	if n.ActiveVolumeID != "" {
		return Novel{}, errors.New("terminá la relectura individual antes de ajustar el progreso general")
	}
	n.Completed = &completed
	if n.Status == "read" && n.ProgressTotal() != nil && completed < *n.ProgressTotal() {
		n.Status = "up-next"
	}
	if markRead && n.ProgressTotal() != nil && completed == *n.ProgressTotal() {
		n.Status = "read"
	}
	if err = reconcile(&n); err != nil {
		return Novel{}, err
	}
	return s.persist(ctx, n)
}
