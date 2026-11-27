// Package classification keeps local genres and tags independent of provider taxonomies.
package classification

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Origin struct {
	Provider string `json:"provider"`
	ID       string `json:"id"`
	Category string `json:"category"`
}
type Term struct {
	Name    string   `json:"name"`
	Origins []Origin `json:"origins"`
}
type Set struct {
	Genres []Term `json:"genres"`
	Tags   []Term `json:"tags"`
}
type Hidden struct {
	Kind     string `json:"kind"`
	Provider string `json:"provider"`
	ID       string `json:"id"`
}
type Input struct {
	Genres []string `json:"genres"`
	Tags   []string `json:"tags"`
	Hidden []Hidden `json:"hidden"`
}

func key(name string) string { return strings.ToLower(strings.Join(strings.Fields(name), " ")) }
func name(value string) (string, error) {
	value = strings.Join(strings.Fields(value), " ")
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > 80 {
		return "", errors.New("cada género o etiqueta debe tener entre 1 y 80 caracteres")
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", errors.New("el término contiene caracteres inválidos")
		}
	}
	return value, nil
}
func validOrigin(o Origin) bool {
	if o.Provider != "hardcover" && o.Provider != "ranobedb" {
		return false
	}
	id, err := strconv.ParseUint(o.ID, 10, 64)
	return err == nil && id > 0 && o.Category != "" && len(o.Category) <= 40
}
func Normalize(in Input) (Input, error) {
	if len(in.Genres) > 100 || len(in.Tags) > 100 || len(in.Hidden) > 400 {
		return Input{}, errors.New("demasiados términos de clasificación")
	}
	out := Input{Genres: []string{}, Tags: []string{}, Hidden: []Hidden{}}
	for i, list := range [][]string{in.Genres, in.Tags} {
		seen := map[string]bool{}
		result := []string{}
		for _, raw := range list {
			n, err := name(raw)
			if err != nil {
				return Input{}, err
			}
			if !seen[key(n)] {
				seen[key(n)] = true
				result = append(result, n)
			}
		}
		sort.Slice(result, func(i, j int) bool { return key(result[i]) < key(result[j]) })
		if i == 0 {
			out.Genres = result
		} else {
			out.Tags = result
		}
	}
	seen := map[Hidden]bool{}
	for _, h := range in.Hidden {
		if (h.Kind != "genre" && h.Kind != "tag") || !validOrigin(Origin{h.Provider, h.ID, h.Kind}) {
			return Input{}, errors.New("referencia de clasificación oculta inválida")
		}
		if !seen[h] {
			seen[h] = true
			out.Hidden = append(out.Hidden, h)
		}
	}
	sort.Slice(out.Hidden, func(i, j int) bool {
		a, b := out.Hidden[i], out.Hidden[j]
		return a.Kind+"/"+a.Provider+"/"+a.ID < b.Kind+"/"+b.Provider+"/"+b.ID
	})
	return out, nil
}
func (s Set) Validate() error {
	if len(s.Genres) > 200 || len(s.Tags) > 200 {
		return errors.New("demasiados términos externos")
	}
	for _, list := range [][]Term{s.Genres, s.Tags} {
		for _, t := range list {
			if _, err := name(t.Name); err != nil {
				return err
			}
			if len(t.Origins) == 0 || len(t.Origins) > 8 {
				return errors.New("procedencia de clasificación inválida")
			}
			for _, o := range t.Origins {
				if !validOrigin(o) {
					return errors.New("identidad de clasificación externa inválida")
				}
			}
		}
	}
	return nil
}
func (s Set) ValidateFor(provider string) error {
	if err := s.Validate(); err != nil {
		return err
	}
	for _, list := range [][]Term{s.Genres, s.Tags} {
		for _, t := range list {
			for _, o := range t.Origins {
				if o.Provider != provider {
					return errors.New("la clasificación no pertenece al proveedor del snapshot")
				}
			}
		}
	}
	return nil
}

type ProviderTerm struct{ Name, ID, Category string }

func ProviderSet(provider string, terms []ProviderTerm) *Set {
	s := &Set{Genres: []Term{}, Tags: []Term{}}
	for _, raw := range terms {
		category := strings.ToLower(strings.TrimSpace(raw.Category))
		kind := ""
		switch category {
		case "genre", "genres":
			kind = "genre"
		case "tag", "tags", "mood", "moods", "theme", "themes":
			kind = "tag"
		default:
			continue
		}
		n, err := name(raw.Name)
		o := Origin{provider, raw.ID, raw.Category}
		if err != nil || !validOrigin(o) {
			continue
		}
		term := Term{Name: n, Origins: []Origin{o}}
		if kind == "genre" {
			s.Genres = append(s.Genres, term)
		} else {
			s.Tags = append(s.Tags, term)
		}
	}
	if len(s.Genres) == 0 && len(s.Tags) == 0 {
		return nil
	}
	return Effective(nil, s)
}

// Effective combines exact normalized labels for presentation while preserving every identity.
// Manual data and suppression references remain independent of refreshed provider snapshots.
func Effective(in *Input, external *Set) *Set {
	if in == nil && external == nil {
		return nil
	}
	out := &Set{Genres: []Term{}, Tags: []Term{}}
	for i, kind := range []string{"genre", "tag"} {
		manual := []string{}
		imported := []Term{}
		if in != nil {
			if i == 0 {
				manual = in.Genres
			} else {
				manual = in.Tags
			}
		}
		if external != nil {
			if i == 0 {
				imported = external.Genres
			} else {
				imported = external.Tags
			}
		}
		terms := map[string]Term{}
		for _, label := range manual {
			terms[key(label)] = Term{Name: label, Origins: []Origin{{Provider: "manual", ID: key(label), Category: kind}}}
		}
		for _, t := range imported {
			for _, o := range t.Origins {
				hidden := false
				if in != nil {
					for _, h := range in.Hidden {
						if h.Kind == kind && h.Provider == o.Provider && h.ID == o.ID {
							hidden = true
							break
						}
					}
				}
				if hidden {
					continue
				}
				k := key(t.Name)
				v, ok := terms[k]
				if !ok {
					v = Term{Name: t.Name, Origins: []Origin{}}
				}
				duplicate := false
				for _, existing := range v.Origins {
					if existing == o {
						duplicate = true
						break
					}
				}
				if !duplicate {
					v.Origins = append(v.Origins, o)
				}
				terms[k] = v
			}
		}
		list := []Term{}
		for _, t := range terms {
			list = append(list, t)
		}
		sort.Slice(list, func(i, j int) bool { return key(list[i].Name) < key(list[j].Name) })
		if i == 0 {
			out.Genres = list
		} else {
			out.Tags = list
		}
	}
	return out
}
