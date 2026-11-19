// Package backup defines the portable library format and recovery use cases.
// It has no dependency on SQLite or Wails.
package backup

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/novels"
)

const Format = "trove-library"
const Version = 4

func SupportedVersion(version int) bool { return version >= 1 && version <= Version }

const MaxFileSize int64 = 256 << 20

type Series struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Medium string `json:"medium,omitempty"`
}
type Work struct {
	ID       string   `json:"id"`
	Medium   string   `json:"medium"`
	Title    string   `json:"title"`
	Authors  []string `json:"authors"`
	SeriesID string   `json:"seriesId,omitempty"`
	Position *float64 `json:"position,omitempty"`
	Notes    string   `json:"notes"`
	AddedOn  string   `json:"addedOn"`
	Favorite bool     `json:"favorite"`
}
type Edition struct {
	ID          string               `json:"id"`
	WorkID      string               `json:"workId"`
	Language    string               `json:"language"`
	Format      string               `json:"format"`
	Identifier  string               `json:"identifier"`
	Owned       bool                 `json:"owned"`
	Identifiers []catalog.Identifier `json:"identifiers,omitempty"`
	Catalog     *catalog.Snapshot    `json:"catalog,omitempty"`
}
type Cycle struct {
	ID         string `json:"id"`
	EditionID  string `json:"editionId"`
	Status     string `json:"status"`
	StartedOn  string `json:"startedOn"`
	FinishedOn string `json:"finishedOn"`
}
type Catalog struct {
	Novels   []novels.Novel `json:"novels,omitempty"`
	Series   []Series       `json:"series"`
	Works    []Work         `json:"works"`
	Editions []Edition      `json:"editions"`
	Cycles   []Cycle        `json:"cycles"`
}
type Document struct {
	Format     string  `json:"format"`
	Version    int     `json:"version"`
	Generation string  `json:"generation"`
	CreatedAt  string  `json:"createdAt"`
	Catalog    Catalog `json:"catalog"`
}

var generationPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,100}$`)

func (d *Document) Validate() error {
	if d.Format != Format || !SupportedVersion(d.Version) {
		return errors.New("no es un respaldo Trove compatible (formato o versión no admitidos)")
	}
	if !generationPattern.MatchString(d.Generation) {
		return errors.New("identificador de generación inválido")
	}
	if _, err := time.Parse(time.RFC3339Nano, d.CreatedAt); err != nil {
		return errors.New("fecha de generación inválida")
	}
	series := map[string]bool{}
	seriesMedia := map[string]string{}
	names := map[string]bool{}
	works := map[string]bool{}
	workMedia := map[string]string{}
	editions := map[string]string{}
	cycles := map[string]bool{}
	add := func(ids map[string]bool, id string) error {
		if strings.TrimSpace(id) == "" || ids[id] {
			return fmt.Errorf("identificador vacío o duplicado: %q", id)
		}
		ids[id] = true
		return nil
	}
	date := func(v string) bool {
		if v == "" {
			return true
		}
		_, err := time.Parse("2006-01-02", v)
		return err == nil
	}
	for _, v := range d.Catalog.Series {
		if err := add(series, v.ID); err != nil {
			return err
		}
		medium := v.Medium
		if d.Version == 1 {
			if medium != "" {
				return errors.New("el formato 1 no admite medios de serie explícitos")
			}
			medium = books.MediumBooks
		}
		if !books.ValidMedium(medium) {
			return errors.New("medio de serie no admitido")
		}
		seriesMedia[v.ID] = medium
		key := medium + ":" + strings.ToLower(strings.Join(strings.Fields(v.Name), " "))
		if key == "" || names[key] {
			return errors.New("serie sin nombre o con nombre duplicado")
		}
		names[key] = true
	}
	for _, v := range d.Catalog.Works {
		if err := add(works, v.ID); err != nil {
			return err
		}
		workMedia[v.ID] = v.Medium
		if !books.ValidMedium(v.Medium) || (d.Version == 1 && v.Medium != books.MediumBooks) {
			return errors.New("el respaldo contiene un medio que esta versión no admite")
		}
		if strings.TrimSpace(v.Title) == "" || !date(v.AddedOn) {
			return errors.New("obra sin título o con fecha de agregado inválida")
		}
		if v.SeriesID != "" && !series[v.SeriesID] {
			return errors.New("una obra referencia una serie inexistente")
		}
		if v.SeriesID != "" && seriesMedia[v.SeriesID] != v.Medium {
			return errors.New("una obra referencia una serie de otro medio")
		}
		if v.Position != nil && (v.SeriesID == "" || *v.Position < 0 || math.IsNaN(*v.Position) || math.IsInf(*v.Position, 0)) {
			return errors.New("posición de serie inválida")
		}
	}
	workEditions := map[string]int{}
	workCycles := map[string]int{}
	for _, v := range d.Catalog.Editions {
		if v.ID == "" || editions[v.ID] != "" || !works[v.WorkID] {
			return errors.New("edición duplicada, sin identidad o sin obra válida")
		}
		if !books.ValidFormat(v.Format) {
			return errors.New("formato de edición no admitido")
		}
		if d.Version < 3 && (len(v.Identifiers) > 0 || v.Catalog != nil) {
			return errors.New("el formato anterior no admite snapshots ni identificadores extendidos")
		}
		for _, id := range v.Identifiers {
			if !catalog.ValidIdentifier(id) {
				return errors.New("identificador de edición inválido")
			}
		}
		if v.Catalog != nil {
			if workMedia[v.WorkID] != books.MediumBooks {
				return errors.New("snapshot no admitido para este medio")
			}
			if err := v.Catalog.Validate(); err != nil {
				return err
			}
		}
		editions[v.ID] = v.WorkID
		workEditions[v.WorkID]++
	}
	for _, v := range d.Catalog.Cycles {
		if err := add(cycles, v.ID); err != nil {
			return err
		}
		work := editions[v.EditionID]
		if work == "" || !books.ValidStatus(v.Status) || !date(v.StartedOn) || !date(v.FinishedOn) {
			return errors.New("ciclo sin edición, estado o fechas válidos")
		}
		if v.StartedOn != "" && v.FinishedOn != "" && v.FinishedOn < v.StartedOn {
			return errors.New("finalización anterior al inicio en un ciclo")
		}
		workCycles[work]++
	}
	for id := range works {
		if workEditions[id] == 0 || workCycles[id] == 0 {
			return errors.New("obra sin edición o registro de lectura")
		}
	}
	if d.Version < 4 && len(d.Catalog.Novels) > 0 {
		return errors.New("el formato de respaldo no admite novelas agrupadas")
	}
	novelIDs, volumeIDs, novelEditionIDs, novelCycleIDs := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, n := range d.Catalog.Novels {
		if err := add(novelIDs, n.ID); err != nil {
			return err
		}
		if err := n.Validate(); err != nil {
			return err
		}
		for _, v := range n.Volumes {
			for _, entry := range []struct {
				ids map[string]bool
				id  string
			}{{volumeIDs, v.ID}, {novelEditionIDs, v.Edition.ID}, {novelCycleIDs, v.Cycle.ID}} {
				if err := add(entry.ids, entry.id); err != nil {
					return err
				}
			}
		}
	}
	d.Canonicalize()
	return nil
}

func (d *Document) Canonicalize() {
	for i := range d.Catalog.Novels {
		d.Catalog.Novels[i].Canonicalize()
	}
	sort.Slice(d.Catalog.Novels, func(i, j int) bool { return d.Catalog.Novels[i].ID < d.Catalog.Novels[j].ID })
	if d.Version >= 3 {
		for i := range d.Catalog.Editions {
			if len(d.Catalog.Editions[i].Identifiers) == 0 {
				d.Catalog.Editions[i].Identifiers = catalog.NormalizeIdentifier(d.Catalog.Editions[i].Identifier, "manual")
			}
		}
	}
	if d.Catalog.Series == nil {
		d.Catalog.Series = []Series{}
	}
	if d.Catalog.Works == nil {
		d.Catalog.Works = []Work{}
	}
	if d.Catalog.Editions == nil {
		d.Catalog.Editions = []Edition{}
	}
	if d.Catalog.Cycles == nil {
		d.Catalog.Cycles = []Cycle{}
	}
	for i := range d.Catalog.Works {
		if d.Catalog.Works[i].Authors == nil {
			d.Catalog.Works[i].Authors = []string{}
		}
	}
	sort.Slice(d.Catalog.Series, func(i, j int) bool { return d.Catalog.Series[i].ID < d.Catalog.Series[j].ID })
	sort.Slice(d.Catalog.Works, func(i, j int) bool { return d.Catalog.Works[i].ID < d.Catalog.Works[j].ID })
	sort.Slice(d.Catalog.Editions, func(i, j int) bool { return d.Catalog.Editions[i].ID < d.Catalog.Editions[j].ID })
	sort.Slice(d.Catalog.Cycles, func(i, j int) bool { return d.Catalog.Cycles[i].ID < d.Catalog.Cycles[j].ID })
}

func (d Document) ContentHash() string {
	d.Canonicalize()
	data, _ := json.Marshal(d.Catalog)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func ReadJSON(path string) (Document, error) {
	var d Document
	file, err := os.Open(path)
	if err != nil {
		return d, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return d, err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxFileSize {
		return d, errors.New("el respaldo no es un archivo regular o supera 256 MiB")
	}
	data, err := io.ReadAll(io.LimitReader(file, MaxFileSize+1))
	if err != nil {
		return d, err
	}
	if int64(len(data)) > MaxFileSize || !utf8.Valid(data) {
		return d, errors.New("respaldo demasiado grande o con texto inválido")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	// A missing catalog/collection is not the same as an explicitly empty library.
	envelope := struct {
		*Document
		Catalog *Catalog `json:"catalog"`
	}{Document: &d}
	if err = decoder.Decode(&envelope); err != nil {
		return d, fmt.Errorf("JSON inválido: %w", err)
	}
	if envelope.Catalog == nil || envelope.Catalog.Series == nil || envelope.Catalog.Works == nil || envelope.Catalog.Editions == nil || envelope.Catalog.Cycles == nil {
		return d, errors.New("el respaldo no contiene todas las colecciones de la biblioteca")
	}
	d.Catalog = *envelope.Catalog
	var trailing any
	if err = decoder.Decode(&trailing); err != io.EOF {
		return d, errors.New("el archivo contiene datos adicionales después del respaldo")
	}
	return d, d.Validate()
}

func WriteJSON(path string, value any) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(value)
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	return err
}
