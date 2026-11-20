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
			result = append(result, n)
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
	return s.repository.GetNovel(ctx, id)
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
	n.Canonicalize()
	if err := n.Validate(); err != nil {
		return Novel{}, err
	}
	return s.repository.SaveNovel(ctx, n)
}
func reconcile(n *Novel) error {
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
			v.Cycle.Status = "read"
		}
		for i := range n.Volumes {
			v := &n.Volumes[i]
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
	n.Title = strings.TrimSpace(in.Title)
	n.Authors = []string{}
	for _, author := range in.Authors {
		if value := strings.TrimSpace(author); value != "" {
			n.Authors = append(n.Authors, value)
		}
	}
	n.Status, n.Total, n.Completed, n.Notes, n.Favorite, n.AddedOn = in.Status, in.Total, in.Completed, in.Notes, in.Favorite, in.AddedOn
	if err := reconcile(&n); err != nil {
		return Novel{}, err
	}
	if n.Status == "reading" && n.NextNumber() == 0 {
		return Novel{}, errors.New("para empezar a leer indicá un progreso conocido con un volumen pendiente")
	}
	if n.Status == "reading" && previousStatus != "reading" && n.NextNumber() > 0 {
		v := volumeAt(&n, n.NextNumber())
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
	return s.persist(ctx, n)
}
func (s *Service) ChangeStatus(ctx context.Context, id, status string) (Novel, error) {
	n, err := s.Get(ctx, id)
	if err != nil {
		return Novel{}, err
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
	v.Cycle, err = v.Cycle.Transition("reading", s.today(), false, false)
	if err != nil {
		return Novel{}, err
	}
	n.Status = "reading"
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
	v.Cycle, err = v.Cycle.Transition("read", s.today(), false, false)
	if err != nil {
		return Novel{}, err
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
	if n.Completed != nil && in.Number <= *n.Completed {
		v.Cycle.Status = "read"
	}
	return s.persist(ctx, n)
}
