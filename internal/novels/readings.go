package novels

import (
	"context"
	"crypto/rand"
	"errors"
	"sort"

	"github.com/Santiparra/Trove/internal/books"
)

type ReadingRun struct {
	ID        string      `json:"id"`
	Cycle     books.Cycle `json:"cycle"`
	VolumeIDs []string    `json:"volumeIDs"`
	Completed int         `json:"completed"`
}

func (v Volume) HasCompleted() bool {
	if v.Cycle.Status == "read" {
		return true
	}
	for _, c := range v.History {
		if c.Status == "read" {
			return true
		}
	}
	return false
}
func (n Novel) Run() *ReadingRun {
	for i := range n.Readings {
		if n.Readings[i].ID == n.ActiveRunID {
			return &n.Readings[i]
		}
	}
	return nil
}
func (n Novel) ProgressTotal() *int {
	if r := n.Run(); r != nil {
		total := len(r.VolumeIDs)
		return &total
	}
	return n.Total
}
func (n Novel) validateReadings() error {
	ids := map[string]bool{}
	volumes := map[string]Volume{}
	for _, v := range n.Volumes {
		volumes[v.ID] = v
	}
	for _, r := range n.Readings {
		if r.ID == "" || ids[r.ID] || r.Cycle.ID != r.ID || r.Completed < 0 || r.Completed > len(r.VolumeIDs) {
			return errors.New("recorrido de novela inválido")
		}
		ids[r.ID] = true
		if err := r.Cycle.Validate(); err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, id := range r.VolumeIDs {
			if _, ok := volumes[id]; !ok || seen[id] {
				return errors.New("volumen inválido en el recorrido")
			}
			seen[id] = true
		}
	}
	if n.ActiveRunID != "" && !ids[n.ActiveRunID] {
		return errors.New("recorrido actual inexistente")
	}
	if n.ActiveVolumeID != "" {
		v, ok := volumes[n.ActiveVolumeID]
		if !ok || !books.ValidStatus(n.PreviousStatus) || !books.ValidStatus(v.Cycle.Status) || v.Cycle.Status == "read" || n.Status != v.Cycle.Status {
			return errors.New("relectura individual inválida")
		}
	}
	if r := n.Run(); r != nil && n.ActiveVolumeID == "" {
		if n.Completed == nil || *n.Completed != r.Completed || n.Status != r.Cycle.Status || (n.Status == "read" && r.Completed != len(r.VolumeIDs)) {
			return errors.New("el recorrido contradice el progreso")
		}
		for i, id := range r.VolumeIDs {
			v := volumes[id]
			completed := false
			for _, c := range append(append([]books.Cycle{}, v.History...), v.Cycle) {
				if c.RunID == r.ID && c.Status == "read" {
					completed = true
				}
			}
			if i < r.Completed && !completed {
				return errors.New("falta una lectura completada del recorrido")
			}
		}
	}
	for _, v := range n.Volumes {
		for _, c := range append(append([]books.Cycle{}, v.History...), v.Cycle) {
			if c.RunID != "" && !ids[c.RunID] {
				return errors.New("la lectura referencia un recorrido inexistente")
			}
		}
	}
	return nil
}

func (n Novel) completedBefore(number int) (int, error) {
	if r := n.Run(); r != nil {
		for i, id := range r.VolumeIDs {
			for _, v := range n.Volumes {
				if v.ID == id && v.Number == number {
					return i, nil
				}
			}
		}
		return 0, errors.New("el volumen no forma parte del recorrido actual")
	}
	return number - 1, nil
}
func archiveVolume(v *Volume) {
	if v.Cycle.Edition == nil {
		v.Cycle.Edition = v.Edition.ReadingEdition()
	}
	if v.Cycle.HasReading() {
		v.History = append(v.History, v.Cycle)
	}
	v.Cycle = books.Cycle{ID: rand.Text(), Status: "to-read", Edition: v.Edition.ReadingEdition()}
}
func prepareRunVolume(n *Novel, v *Volume) {
	if r := n.Run(); r != nil && n.ActiveVolumeID == "" && v.Cycle.RunID != r.ID {
		archiveVolume(v)
		v.Cycle.RunID = r.ID
	}
}

func (s *Service) RereadNovel(ctx context.Context, id string, volumeIDs []string) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	n, err = beginRun(n, volumeIDs)
	if err != nil {
		return Novel{}, err
	}
	return s.persist(ctx, n)
}

func beginRun(n Novel, volumeIDs []string) (Novel, error) {
	if n.ActiveVolumeID != "" || (n.Run() != nil && n.Status != "read" && n.Status != "did-not-finish") {
		return Novel{}, errors.New("terminá el recorrido actual antes de iniciar otro")
	}
	if n.Status == "reading" {
		return Novel{}, errors.New("terminá o pausá la lectura actual antes de iniciar otro recorrido")
	}
	selected := map[string]bool{}
	for _, id := range volumeIDs {
		if id == "" || selected[id] {
			return Novel{}, errors.New("selección de volúmenes inválida")
		}
		selected[id] = true
	}
	chosen := []Volume{}
	for _, v := range n.Volumes {
		if (len(volumeIDs) == 0 && v.HasCompleted()) || selected[v.ID] {
			if !v.HasCompleted() {
				return Novel{}, errors.New("solo se pueden releer volúmenes ya leídos")
			}
			chosen = append(chosen, v)
			delete(selected, v.ID)
		}
	}
	if len(chosen) == 0 || len(selected) > 0 {
		return Novel{}, errors.New("elegí volúmenes leídos de esta novela")
	}
	sort.Slice(chosen, func(i, j int) bool { return chosen[i].Number < chosen[j].Number })
	if len(n.Readings) == 0 {
		initialID := rand.Text()
		initial := ReadingRun{ID: initialID, Cycle: books.Cycle{ID: initialID, Status: n.Status}, VolumeIDs: []string{}}
		for i := range n.Volumes {
			v := &n.Volumes[i]
			if v.Cycle.HasReading() {
				initial.VolumeIDs = append(initial.VolumeIDs, v.ID)
				v.Cycle.RunID = initialID
				if v.HasCompleted() {
					initial.Completed++
				}
			}
		}
		n.Readings = append(n.Readings, initial)
	}
	runID := rand.Text()
	run := ReadingRun{ID: runID, Cycle: books.Cycle{ID: runID, Status: "up-next"}, VolumeIDs: []string{}}
	for _, v := range chosen {
		run.VolumeIDs = append(run.VolumeIDs, v.ID)
	}
	n.Readings = append(n.Readings, run)
	n.ActiveRunID = runID
	zero := 0
	n.Completed = &zero
	n.Status = "up-next"
	// Selecting the journey does not start or complete any volume.
	return n, nil
}
func (s *Service) RereadVolume(ctx context.Context, id, volumeID string) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	if n.ActiveVolumeID != "" || n.Status == "reading" || (n.Run() != nil && n.Status != "read" && n.Status != "did-not-finish") {
		return Novel{}, errors.New("terminá la lectura actual antes de releer otro volumen")
	}
	for i := range n.Volumes {
		v := &n.Volumes[i]
		if v.ID == volumeID {
			if !v.HasCompleted() {
				return Novel{}, errors.New("el volumen todavía no tiene una lectura completada")
			}
			archiveVolume(v)
			v.Cycle.Status = "reading"
			v.Cycle.StartedOn = s.today()
			n.PreviousStatus = n.Status
			n.ActiveVolumeID = v.ID
			n.Status = "reading"
			return s.persist(ctx, n)
		}
	}
	return Novel{}, ErrNotFound
}

func (s *Service) SaveReading(ctx context.Context, id, volumeID string, in books.ReadingInput) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	for i := range n.Volumes {
		v := &n.Volumes[i]
		if v.ID != volumeID {
			continue
		}
		cycleID := in.ID
		if cycleID == "" {
			cycleID = rand.Text()
		}
		c, err := in.Cycle(cycleID)
		if err != nil {
			return Novel{}, err
		}
		if c.Edition == nil && in.ID != "" {
			if in.ID == v.Cycle.ID {
				c.Edition = v.Cycle.Edition
			}
			for _, old := range v.History {
				if old.ID == in.ID {
					c.Edition = old.Edition
					break
				}
			}
		}
		if in.ID == "" {
			if c.Status != "read" && c.Status != "did-not-finish" {
				return Novel{}, errors.New("una lectura pasada debe estar terminada o sin terminar")
			}
			v.History = append(v.History, c)
		} else if in.ID == v.Cycle.ID {
			c.RunID = v.Cycle.RunID
			v.Cycle = c
			if n.ActiveVolumeID == v.ID {
				if c.Status == "read" || c.Status == "did-not-finish" {
					n.ActiveVolumeID = ""
					n.Status = n.PreviousStatus
					n.PreviousStatus = ""
				} else {
					n.Status = c.Status
				}
			}
		} else {
			found := false
			for j := range v.History {
				if v.History[j].ID == in.ID {
					c.RunID = v.History[j].RunID
					v.History[j] = c
					found = true
					break
				}
			}
			if !found {
				return Novel{}, ErrNotFound
			}
		}
		return s.persist(ctx, n)
	}
	return Novel{}, ErrNotFound
}

func (s *Service) SaveRun(ctx context.Context, id string, in books.ReadingInput) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	for i := range n.Readings {
		r := &n.Readings[i]
		if r.ID != in.ID {
			continue
		}
		c, err := in.Cycle(r.ID)
		if err != nil {
			return Novel{}, err
		}
		if r.ID == n.ActiveRunID && c.Status != r.Cycle.Status {
			return Novel{}, errors.New("el estado del recorrido actual se cambia desde Seguimiento")
		}
		r.Cycle = c
		return s.persist(ctx, n)
	}
	return Novel{}, ErrNotFound
}
