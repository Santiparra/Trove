package novels

import (
	"context"
	"crypto/rand"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
)

type Repository interface {
	ListNovels(context.Context, string) ([]Novel, error)
	GetNovel(context.Context, string) (Novel, error)
	SaveNovel(context.Context, Novel) (Novel, error)
	DeleteNovel(context.Context, string) error
}
type Service struct {
	repository Repository
	today      func() string
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, today: func() string { return time.Now().Format("2006-01-02") }}
}
func (s *Service) List(ctx context.Context, query, status string) ([]Novel, error) {
	if status != "" && !books.ValidStatus(status) {
		return nil, errors.New("estado de novela inválido")
	}
	items, err := s.repository.ListNovels(ctx, query)
	if err != nil {
		return nil, err
	}
	result := make([]Novel, 0, len(items))
	for _, n := range items {
		if status == "" || n.Status == status {
			result = append(result, withClassification(n))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		a, b := strings.ToLower(result[i].Title), strings.ToLower(result[j].Title)
		if a == b {
			return result[i].ID < result[j].ID
		}
		return a < b
	})
	return result, nil
}
func (s *Service) Get(ctx context.Context, id string) (Novel, error) {
	n, err := s.repository.GetNovel(ctx, id)
	return withClassification(n), err
}
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repository.DeleteNovel(ctx, id)
}
func volumeAt(n *Novel, number int) *Volume {
	for i := range n.Volumes {
		if n.Volumes[i].Number == number {
			return &n.Volumes[i]
		}
	}
	n.Volumes = append(n.Volumes, Volume{ID: rand.Text(), Number: number, Edition: books.Edition{ID: rand.Text()}, Cycle: books.Cycle{ID: rand.Text(), Status: "to-read"}})
	return &n.Volumes[len(n.Volumes)-1]
}
func (s *Service) persist(ctx context.Context, n Novel) (Novel, error) {
	n.calculateRating()
	if r := n.Run(); r != nil && n.ActiveVolumeID == "" {
		syncRating(r, n)
		r.Cycle.Status = n.Status
		if n.Completed != nil {
			r.Completed = *n.Completed
		}
		if n.Status == "read" && r.ReturnTo != nil {
			resume := r.ReturnTo
			n.Status, n.Completed, n.ActiveRunID = resume.Status, resume.Completed, resume.ActiveRunID
			n.RatingMode, n.Rating = resume.RatingMode, resume.Rating
			n.calculateRating()
			for _, id := range resume.PendingCycleIDs {
				for i := range n.Volumes {
					v := &n.Volumes[i]
					for j, c := range v.History {
						if c.ID == id {
							v.History = append(v.History[:j], v.History[j+1:]...)
							v.History = append(v.History, v.Cycle)
							v.Cycle = c
							break
						}
					}
				}
			}
		}
	}
	if n.ActiveRunID == "" && n.ActiveVolumeID == "" && len(n.Readings) > 0 {
		initial := &n.Readings[0]
		syncRating(initial, n)
		initial.Cycle.Status = n.Status
		initial.VolumeIDs = []string{}
		initial.Completed = 0
		if n.Completed != nil {
			initial.Completed = *n.Completed
		}
		for i := range n.Volumes {
			v := &n.Volumes[i]
			if v.Cycle.HasReading() {
				initial.VolumeIDs = append(initial.VolumeIDs, v.ID)
				if v.Cycle.RunID == "" {
					v.Cycle.RunID = initial.ID
				}
			}
		}
	}
	previous, err := s.repository.GetNovel(ctx, n.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Novel{}, err
	}
	if err != nil || previous.Status != n.Status {
		n.StatusChangedAt = time.Now().UTC().Format(time.RFC3339Nano)
	} else {
		n.StatusChangedAt = previous.StatusChangedAt
	}
	n.Canonicalize()
	if err := n.Validate(); err != nil {
		return Novel{}, err
	}
	n, err = s.repository.SaveNovel(ctx, n)
	return withClassification(n), err
}
func reconcile(n *Novel) error {
	if n.ActiveVolumeID != "" {
		return nil
	}
	if r := n.Run(); r != nil {
		if n.Completed == nil || *n.Completed < 0 || *n.Completed > len(r.VolumeIDs) {
			return errors.New("progreso fuera del recorrido seleccionado")
		}
		if n.Status == "read" {
			count := len(r.VolumeIDs)
			n.Completed = &count
		}
		for i, id := range r.VolumeIDs {
			for j := range n.Volumes {
				v := &n.Volumes[j]
				if v.ID != id {
					continue
				}
				if i < *n.Completed {
					prepareRunVolume(n, v)
					v.Cycle.Status = "read"
				} else if v.Cycle.RunID == r.ID && v.Cycle.Status == "read" {
					v.Cycle.Status = "to-read"
				}
			}
		}
		return nil
	}
	if n.Total != nil && (*n.Total < 1 || *n.Total > MaxVolumes) {
		return errors.New("total de volúmenes inválido")
	}
	if n.Completed != nil && (*n.Completed < 0 || *n.Completed > MaxVolumes || (n.Total != nil && *n.Completed > *n.Total)) {
		return errors.New("progreso de volúmenes inválido")
	}
	if n.Status == "read" && n.Total != nil {
		count := *n.Total
		n.Completed = &count
	}
	if n.Completed != nil {
		for number := 1; number <= *n.Completed; number++ {
			v := volumeAt(n, number)
			if len(n.Readings) > 0 && v.Cycle.RunID != "" && v.Cycle.RunID != n.Readings[0].ID {
				continue
			}
			v.Cycle.Status = "read"
		}
		for i := range n.Volumes {
			v := &n.Volumes[i]
			if len(n.Readings) > 0 && v.Cycle.RunID != "" && v.Cycle.RunID != n.Readings[0].ID {
				continue
			}
			if v.Number > *n.Completed && v.Cycle.Status == "read" {
				v.Cycle.Status = "to-read"
			}
		}
	} else if n.Status == "read" {
		for i := range n.Volumes {
			n.Volumes[i].Cycle.Status = "read"
		}
	}
	return nil
}
func (s *Service) Save(ctx context.Context, in Input) (Novel, error) {
	n := Novel{ID: in.ID}
	if in.ID != "" {
		var err error
		n, err = s.Get(ctx, in.ID)
		if err != nil {
			return Novel{}, err
		}
	} else {
		n.ID = rand.Text()
		if in.AddedOn == "" {
			in.AddedOn = s.today()
		}
	}
	previousStatus := n.Status
	previousCompleted := n.Completed
	if in.ID != "" && previousStatus == "read" && in.Status == "reading" {
		var err error
		n, err = beginRun(n, nil)
		if err != nil {
			return Novel{}, err
		}
		// Keep the new journey's progress, not the completed previous journey.
		in.Completed = n.Completed
	}
	if in.ProgressVolumeID != "" {
		found := false
		for _, v := range n.Volumes {
			if v.ID == in.ProgressVolumeID {
				completed, err := n.completedBefore(v.Number)
				if err != nil {
					return Novel{}, err
				}
				in.Completed = &completed
				found = true
				break
			}
		}
		if !found {
			return Novel{}, errors.New("el volumen no pertenece a esta novela")
		}
	}
	if in.ProgressBookID != "" {
		if n.Catalog == nil {
			return Novel{}, errors.New("no hay catálogo para identificar ese volumen")
		}
		found := false
		for _, ref := range n.Catalog.Books {
			if ref.ID == in.ProgressBookID {
				completed, err := n.completedBefore(ref.Number)
				if err != nil {
					return Novel{}, err
				}
				in.Completed = &completed
				found = true
				break
			}
		}
		if !found {
			return Novel{}, errors.New("el volumen elegido no pertenece a esta novela")
		}
	}
	n.Title = strings.TrimSpace(in.Title)
	n.Authors = []string{}
	for _, author := range in.Authors {
		if value := strings.TrimSpace(author); value != "" {
			n.Authors = append(n.Authors, value)
		}
	}
	n.Status, n.Total, n.Completed, n.Notes, n.Favorite, n.AddedOn = in.Status, in.Total, in.Completed, in.Notes, in.Favorite, in.AddedOn
	if n.ActiveVolumeID != "" {
		if (previousCompleted == nil) != (n.Completed == nil) || (previousCompleted != nil && *previousCompleted != *n.Completed) {
			return Novel{}, errors.New("terminá la relectura individual antes de ajustar el progreso general")
		}
		for i := range n.Volumes {
			v := &n.Volumes[i]
			if v.ID != n.ActiveVolumeID {
				continue
			}
			var err error
			v.Cycle, err = v.Cycle.Transition(in.Status, s.today(), false, false)
			if err != nil {
				return Novel{}, err
			}
			if in.Status == "read" || in.Status == "did-not-finish" {
				n.ActiveVolumeID = ""
				n.Status = n.PreviousStatus
				n.PreviousStatus = ""
			}
			return s.persist(ctx, n)
		}
	}
	if err := reconcile(&n); err != nil {
		return Novel{}, err
	}
	if n.Status == "reading" && n.NextNumber() == 0 {
		return Novel{}, errors.New("para empezar a leer indicá un progreso conocido con un volumen pendiente")
	}
	if n.Status == "reading" && previousStatus != "reading" && n.NextNumber() > 0 {
		v := volumeAt(&n, n.NextNumber())
		prepareRunVolume(&n, v)
		var err error
		v.Cycle, err = v.Cycle.Transition("reading", s.today(), false, false)
		if err != nil {
			return Novel{}, err
		}
	}
	if n.Status != "reading" && n.Status != "read" {
		for i := range n.Volumes {
			v := &n.Volumes[i]
			if v.Number == n.NextNumber() {
				v.Cycle.Status = n.Status
			}
		}
	}
	if r := n.Run(); r != nil && n.ActiveVolumeID == "" {
		var err error
		r.Cycle, err = r.Cycle.Transition(n.Status, s.today(), false, false)
		if err != nil {
			return Novel{}, err
		}
	}
	return s.persist(ctx, n)
}
func (s *Service) ChangeStatus(ctx context.Context, id, status string) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	if n.ActiveVolumeID != "" {
		if status == "read" {
			return s.Finish(ctx, id)
		}
		for i := range n.Volumes {
			v := &n.Volumes[i]
			if v.ID != n.ActiveVolumeID {
				continue
			}
			v.Cycle, err = v.Cycle.Transition(status, s.today(), false, false)
			if err != nil {
				return Novel{}, err
			}
			n.Status = status
			if status == "did-not-finish" {
				n.ActiveVolumeID = ""
				n.Status = n.PreviousStatus
				n.PreviousStatus = ""
			}
			return s.persist(ctx, n)
		}
	}
	return s.Save(ctx, Input{ID: n.ID, Title: n.Title, Authors: n.Authors, Status: status, Total: n.Total, Completed: n.Completed, Notes: n.Notes, Favorite: n.Favorite, AddedOn: n.AddedOn})
}
func (s *Service) Start(ctx context.Context, id string) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	number := n.NextNumber()
	if number == 0 {
		return Novel{}, errors.New("indicá progreso conocido y un volumen pendiente antes de empezar")
	}
	v := volumeAt(&n, number)
	prepareRunVolume(&n, v)
	v.Cycle, err = v.Cycle.Transition("reading", s.today(), false, false)
	if err != nil {
		return Novel{}, err
	}
	n.Status = "reading"
	if r := n.Run(); r != nil && n.ActiveVolumeID == "" && r.Cycle.StartedOn == "" {
		r.Cycle.StartedOn = s.today()
	}
	return s.persist(ctx, n)
}
func (s *Service) Finish(ctx context.Context, id string) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	number := n.NextNumber()
	if number == 0 {
		return Novel{}, errors.New("no hay un volumen pendiente con número conocido")
	}
	v := volumeAt(&n, number)
	prepareRunVolume(&n, v)
	v.Cycle, err = v.Cycle.Transition("read", s.today(), false, false)
	if err != nil {
		return Novel{}, err
	}
	if n.ActiveVolumeID != "" {
		n.ActiveVolumeID = ""
		n.Status = n.PreviousStatus
		n.PreviousStatus = ""
		return s.persist(ctx, n)
	}
	if r := n.Run(); r != nil {
		count := *n.Completed + 1
		n.Completed = &count
		n.Status = "up-next"
		if count == len(r.VolumeIDs) {
			n.Status = "read"
			if r.Cycle.StartedOn != "" {
				r.Cycle.FinishedOn = s.today()
			}
		}
		return s.persist(ctx, n)
	}
	n.Completed = &number
	n.Status = "up-next"
	if n.Total != nil && number == *n.Total {
		n.Status = "read"
	}
	return s.persist(ctx, n)
}
func (s *Service) SaveVolume(ctx context.Context, id string, in VolumeInput) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
	}
	if in.Number < 1 || in.Number > MaxVolumes || (n.Total != nil && in.Number > *n.Total) {
		return Novel{}, errors.New("número de volumen fuera del total conocido")
	}
	if in.ID != "" {
		found := false
		for _, v := range n.Volumes {
			if v.ID == in.ID {
				found = true
				if v.Number != in.Number {
					return Novel{}, errors.New("el número del volumen existente no puede cambiarse")
				}
			}
		}
		if !found {
			return Novel{}, errors.New("el volumen no pertenece a esta novela")
		}
	}
	for _, v := range n.Volumes {
		if v.Number == in.Number && in.ID == "" {
			return Novel{}, errors.New("ese volumen ya está registrado; editá su ficha")
		}
	}
	v := volumeAt(&n, in.Number)
	if v.Cycle.Edition == nil {
		v.Cycle.Edition = v.Edition.ReadingEdition()
	}
	v.Label = strings.TrimSpace(in.Label)
	if isbn := catalog.ISBN13(in.Identifier); v.Catalog != nil && isbn != "" && v.Catalog.ISBN != "" && isbn != v.Catalog.ISBN {
		return Novel{}, errors.New("el ISBN no coincide con la publicación vinculada; elegí otra publicación antes de cambiarlo")
	}
	v.Title, v.Description, v.Notes, v.Favorite = strings.TrimSpace(in.Title), in.Description, in.Notes, in.Favorite
	v.Edition.Language, v.Edition.Format, v.Edition.Identifier, v.Edition.Owned = strings.TrimSpace(in.Language), in.Format, strings.TrimSpace(in.Identifier), in.Owned
	v.Edition.Identifiers = catalog.NormalizeIdentifier(v.Edition.Identifier, "manual")
	if v.Catalog != nil {
		v.Edition.Identifiers = catalog.MergeIdentifiers(v.Edition.Identifiers, catalog.NormalizeIdentifier(v.Catalog.ISBN, "ranobedb"))
	}
	v.Cycle.StartedOn, v.Cycle.FinishedOn = in.StartedOn, in.FinishedOn
	if n.Run() == nil && n.Completed != nil && in.Number <= *n.Completed && n.ActiveVolumeID != v.ID {
		v.Cycle.Status = "read"
	}
	return s.persist(ctx, n)
}
