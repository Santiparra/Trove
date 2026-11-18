package providers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/Santiparra/Trove/internal/catalog"
)

type OpenLibrary struct {
	http     *http.Client
	endpoint string
	pace     pace
	health   *monitor
}

func NewOpenLibrary() *OpenLibrary {
	httpClient := client()
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || req.URL.Scheme != "https" || req.URL.Host != "openlibrary.org" || req.URL.User != nil {
			return errors.New("redirección de Open Library no permitida")
		}
		return nil
	}
	return &OpenLibrary{http: httpClient, endpoint: "https://openlibrary.org", pace: pace{interval: time.Second}}
}
func (p *OpenLibrary) get(ctx context.Context, path string, out any) error {
	if p.health != nil {
		return p.health.run(ctx, p.pace.wait, func() error { return p.getRaw(ctx, path, out) })
	}
	if err := p.pace.wait(ctx); err != nil {
		return errors.New("consulta secundaria cancelada")
	}
	return p.getRaw(ctx, path, out)
}
func (p *OpenLibrary) getRaw(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+path, nil)
	if err != nil {
		return errors.New("consulta secundaria inválida")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	response, err := p.http.Do(req)
	if err != nil {
		return failure("network", "Open Library no está disponible")
	}
	err = responseJSON(response, out)
	var problem *providerError
	// This public API has no user key to replace; an access rejection must remain recoverable.
	if errors.As(err, &problem) && problem.kind == "access" {
		return failure("service", "Open Library rechazó temporalmente la consulta")
	}
	return err
}

var editionKey = regexp.MustCompile(`^/books/(OL[0-9]+M)$`)
var workKey = regexp.MustCompile(`^/works/(OL[0-9]+W)$`)

func (p *OpenLibrary) Lookup(ctx context.Context, isbn string) (catalog.Supplement, error) {
	if !catalog.ValidISBN13(isbn) {
		return catalog.Supplement{}, errors.New("se requiere ISBN-13 válido")
	}
	// Query an exact identifier; inspect the edition before trusting its cover or work.
	var item struct {
		Key    string   `json:"key"`
		ISBN10 []string `json:"isbn_10"`
		ISBN13 []string `json:"isbn_13"`
		Covers []int64  `json:"covers"`
		Works  []struct {
			Key string `json:"key"`
		} `json:"works"`
	}
	if err := p.get(ctx, "/isbn/"+url.PathEscape(isbn)+".json", &item); err != nil {
		return catalog.Supplement{}, err
	}
	match := editionKey.FindStringSubmatch(item.Key)
	if match == nil {
		return catalog.Supplement{}, errors.New("el secundario no identificó una edición")
	}
	confirmed, conflicting := false, false
	for _, v := range append(item.ISBN13, item.ISBN10...) {
		if value := catalog.ISBN13(v); value != "" {
			if value == isbn {
				confirmed = true
			} else {
				conflicting = true
			}
		}
	}
	if !confirmed {
		return catalog.Supplement{Warnings: []string{"Open Library devolvió una edición sin ISBN equivalente verificable; no se mezcló metadata."}}, nil
	}
	if conflicting {
		return catalog.Supplement{Warnings: []string{"Open Library agrupa ISBN no equivalentes; se requiere revisión y no se mezclaron campos."}}, nil
	}
	result := catalog.Supplement{EditionID: match[1], Warnings: []string{}}
	for _, id := range item.Covers {
		if id > 0 {
			result.CoverURL = "https://covers.openlibrary.org/b/id/" + strconv.FormatInt(id, 10) + "-M.jpg"
			break
		}
	}
	if len(item.Works) > 0 {
		key := workKey.FindStringSubmatch(item.Works[0].Key)
		if key != nil {
			var work struct {
				Description json.RawMessage `json:"description"`
			}
			if err := p.get(ctx, item.Works[0].Key+".json", &work); err != nil {
				return result, nil
			}
			var description string
			if json.Unmarshal(work.Description, &description) != nil {
				var wrapped struct {
					Value string `json:"value"`
				}
				_ = json.Unmarshal(work.Description, &wrapped)
				description = wrapped.Value
			}
			result.WorkID = key[1]
			result.Description = plain(description)
		}
	}
	return result, nil
}
