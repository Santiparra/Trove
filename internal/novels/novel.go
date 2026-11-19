// Package novels owns novel-level progress and independent volume records.
package novels

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
)

const MaxVolumes = 10000

var ErrNotFound = errors.New("la novela no existe; actualizá la biblioteca")

type Volume struct {
	ID          string        `json:"id"`
	Number      int           `json:"number"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Edition     books.Edition `json:"edition"`
	Cycle       books.Cycle   `json:"cycle"`
	Notes       string        `json:"notes"`
	Favorite    bool          `json:"favorite"`
}
type Novel struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Authors   []string `json:"authors"`
	Status    string   `json:"status"`
	Total     *int     `json:"total"`
	Completed *int     `json:"completed"`
	Notes     string   `json:"notes"`
	Favorite  bool     `json:"favorite"`
	AddedOn   string   `json:"addedOn"`
	Volumes   []Volume `json:"volumes"`
}
type Input struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Authors   []string `json:"authors"`
	Status    string   `json:"status"`
	Total     *int     `json:"total"`
	Completed *int     `json:"completed"`
	Notes     string   `json:"notes"`
	Favorite  bool     `json:"favorite"`
	AddedOn   string   `json:"addedOn"`
}
type VolumeInput struct {
	ID          string `json:"id"`
	Number      int    `json:"number"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Format      string `json:"format"`
	Identifier  string `json:"identifier"`
	Owned       bool   `json:"owned"`
	Favorite    bool   `json:"favorite"`
	Notes       string `json:"notes"`
	StartedOn   string `json:"startedOn"`
	FinishedOn  string `json:"finishedOn"`
}

func validDate(value string) bool {
	if value == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}
func (n *Novel) Canonicalize() {
	if n.Authors == nil {
		n.Authors = []string{}
	}
	if n.Volumes == nil {
		n.Volumes = []Volume{}
	}
	for i := range n.Volumes {
		if n.Volumes[i].Edition.Identifiers == nil {
			n.Volumes[i].Edition.Identifiers = []catalog.Identifier{}
		}
	}
	sort.Slice(n.Volumes, func(i, j int) bool { return n.Volumes[i].Number < n.Volumes[j].Number })
}
func (n Novel) Validate() error {
	if n.Volumes == nil {
		return errors.New("falta la colección de volúmenes de la novela")
	}
	if strings.TrimSpace(n.ID) == "" || strings.TrimSpace(n.Title) == "" || !books.ValidStatus(n.Status) || !validDate(n.AddedOn) {
		return errors.New("identidad, título, estado o fecha de novela inválidos")
	}
	if n.Total != nil && (*n.Total < 1 || *n.Total > MaxVolumes) {
		return errors.New("el total debe estar entre 1 y 10000, o ser desconocido")
	}
	if n.Completed != nil && (*n.Completed < 0 || *n.Completed > MaxVolumes || (n.Total != nil && *n.Completed > *n.Total)) {
		return errors.New("el progreso no puede superar el total ni ser negativo")
	}
	if n.Status == "read" && n.Total != nil && (n.Completed == nil || *n.Completed != *n.Total) {
		return errors.New("una novela leída con total conocido debe tener progreso completo")
	}
	if n.Status == "reading" && n.NextNumber() == 0 {
		return errors.New("la lectura activa requiere progreso conocido y un volumen pendiente")
	}
	ids, numbers, editionIDs, cycleIDs := map[string]bool{}, map[int]bool{}, map[string]bool{}, map[string]bool{}
	for _, v := range n.Volumes {
		if v.ID == "" || ids[v.ID] || v.Number < 1 || v.Number > MaxVolumes || numbers[v.Number] || (n.Total != nil && v.Number > *n.Total) {
			return errors.New("identidad o número de volumen inválido o duplicado")
		}
		ids[v.ID], numbers[v.Number] = true, true
		if editionIDs[v.Edition.ID] || cycleIDs[v.Cycle.ID] {
			return errors.New("identidades de edición o ciclo duplicadas")
		}
		editionIDs[v.Edition.ID], cycleIDs[v.Cycle.ID] = true, true
		if v.Edition.ID == "" || v.Cycle.ID == "" || !books.ValidFormat(v.Edition.Format) || !books.ValidStatus(v.Cycle.Status) || !validDate(v.Cycle.StartedOn) || !validDate(v.Cycle.FinishedOn) || (v.Cycle.StartedOn != "" && v.Cycle.FinishedOn != "" && v.Cycle.FinishedOn < v.Cycle.StartedOn) {
			return errors.New("edición o fechas del volumen inválidas")
		}
		if v.Edition.Catalog != nil {
			return errors.New("el proveedor de novelas todavía no está habilitado")
		}
		for _, id := range v.Edition.Identifiers {
			if !catalog.ValidIdentifier(id) {
				return errors.New("identificador de volumen inválido")
			}
		}
		if n.Completed != nil && ((v.Number <= *n.Completed) != (v.Cycle.Status == "read")) {
			return errors.New("el historial contradice el progreso completado")
		}
	}
	if n.Completed != nil {
		for number := 1; number <= *n.Completed; number++ {
			if !numbers[number] {
				return errors.New("faltan los registros del progreso histórico conocido")
			}
		}
	}
	return nil
}
func (n Novel) NextNumber() int {
	if n.Completed == nil || n.Status == "read" || (n.Total != nil && *n.Completed >= *n.Total) {
		return 0
	}
	return *n.Completed + 1
}
