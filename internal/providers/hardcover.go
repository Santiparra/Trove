package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Santiparra/Trove/internal/catalog"
)

type Hardcover struct {
	token    func() (string, error)
	http     *http.Client
	endpoint string
	pace     pace
	health   *monitor
}

func NewHardcover(token func() (string, error)) *Hardcover {
	return &Hardcover{token: token, http: client(), endpoint: "https://api.hardcover.app/v1/graphql", pace: pace{interval: time.Second}}
}
func (p *Hardcover) query(ctx context.Context, query string, variables any, out any) error {
	if p.health != nil {
		return p.health.run(ctx, p.pace.wait, func() error { return p.queryRaw(ctx, query, variables, out) })
	}
	if err := p.pace.wait(ctx); err != nil {
		return errors.New("consulta de catálogo cancelada")
	}
	return p.queryRaw(ctx, query, variables, out)
}
func (p *Hardcover) queryRaw(ctx context.Context, query string, variables any, out any) error {
	token, err := p.token()
	if err != nil || token == "" {
		return failure("unconfigured", "configurá una clave de lectura Hardcover para consultar el catálogo")
	}
	body, err := json.Marshal(map[string]any{"query": query, "variables": variables})
	if err != nil {
		return errors.New("no se pudo preparar la consulta")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("consulta de catálogo inválida")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	response, err := p.http.Do(req)
	if err != nil {
		if isLocalOnly(err) {
			return err
		}
		return failure("network", "no se pudo conectar con Hardcover; los datos locales siguen disponibles")
	}
	var envelope struct {
		Data   json.RawMessage   `json:"data"`
		Errors []json.RawMessage `json:"errors"`
	}
	if err = responseJSON(response, &envelope); err != nil {
		return err
	}
	if len(envelope.Errors) > 0 || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		for _, item := range envelope.Errors {
			var problem struct {
				Extensions struct {
					Code string `json:"code"`
				} `json:"extensions"`
			}
			_ = json.Unmarshal(item, &problem)
			switch strings.ToLower(problem.Extensions.Code) {
			case "invalid-jwt", "permission-denied", "access-denied", "unauthenticated", "forbidden":
				return failure("access", "Hardcover rechazó el acceso; revisá la clave y sus permisos")
			}
		}
		return failure("response", "Hardcover no pudo resolver la consulta; revisá el acceso o reintentá más tarde")
	}
	if err = json.Unmarshal(envelope.Data, out); err != nil {
		return failure("response", "Hardcover devolvió datos incompatibles")
	}
	return nil
}

type remoteImage struct {
	URL string `json:"url"`
}

func cover(v *remoteImage) string {
	if v != nil && catalog.ValidCoverURL(v.URL) {
		return v.URL
	}
	return ""
}

type hcEdition struct {
	ID       int64        `json:"id"`
	Title    string       `json:"title"`
	ISBN10   string       `json:"isbn_10"`
	ISBN13   string       `json:"isbn_13"`
	ASIN     string       `json:"asin"`
	Pages    *int         `json:"pages"`
	Format   string       `json:"edition_format"`
	Date     string       `json:"release_date"`
	Image    *remoteImage `json:"image"`
	Language *struct {
		Code string `json:"code2"`
	} `json:"language"`
	Publisher *struct {
		Name string `json:"name"`
	} `json:"publisher"`
	Book struct {
		ID        int64        `json:"id"`
		Title     string       `json:"title"`
		Canonical *int64       `json:"canonical_id"`
		Image     *remoteImage `json:"image"`
	} `json:"book"`
}

func edition(v hcEdition) catalog.Edition {
	format := strings.ToLower(v.Format)
	switch {
	case strings.Contains(format, "paperback"):
		format = "paperback"
	case strings.Contains(format, "hard"):
		format = "hardcover"
	case strings.Contains(format, "audio"):
		format = "audio"
	case strings.Contains(format, "ebook") || strings.Contains(format, "digital") || strings.Contains(format, "kindle"):
		format = "digital"
	case format == "":
	default:
		format = "other"
	}
	declared := []catalog.Identifier{}
	for _, id := range []struct{ scheme, value string }{{"isbn10", v.ISBN10}, {"isbn13", v.ISBN13}, {"asin", v.ASIN}} {
		if id.value != "" {
			declared = append(declared, catalog.Identifier{Scheme: id.scheme, Value: id.value, Source: catalog.Hardcover})
		}
	}
	title := strings.TrimSpace(v.Title)
	if title == "" {
		title = v.Book.Title
	}
	out := catalog.Edition{ID: strconv.FormatInt(v.ID, 10), WorkID: strconv.FormatInt(v.Book.ID, 10), Title: title, Format: format, Pages: v.Pages, PublishedOn: v.Date, CoverURL: cover(v.Image), Declared: declared}
	if v.Language != nil {
		out.Language = v.Language.Code
	}
	if v.Publisher != nil {
		out.Publisher = v.Publisher.Name
	}
	return out
}

const editionFields = `id title isbn_10 isbn_13 asin pages edition_format release_date image { url } language { code2 } publisher { name } book { id title canonical_id image { url } }`

func (p *Hardcover) Search(ctx context.Context, text string) ([]catalog.Work, error) {
	if isbn := catalog.ISBN13(text); isbn != "" {
		var data struct {
			Editions []hcEdition `json:"editions"`
		}
		query := `query TroveISBN($long:String!,$short:String!){editions(limit:20,where:{_or:[{isbn_13:{_eq:$long}},{isbn_10:{_eq:$short}}]}){` + editionFields + `}}`
		if err := p.query(ctx, query, map[string]string{"long": isbn, "short": catalog.ISBN10(isbn)}, &data); err != nil {
			return nil, err
		}
		result := []catalog.Work{}
		seen := map[string]bool{}
		for _, v := range data.Editions {
			if v.Book.Canonical != nil {
				continue
			}
			id := strconv.FormatInt(v.Book.ID, 10)
			if !seen[id] {
				seen[id] = true
				result = append(result, catalog.Work{ID: id, Title: v.Book.Title, Authors: []string{}, CoverURL: cover(v.Image), MatchedEditionID: strconv.FormatInt(v.ID, 10)})
			}
		}
		return result, nil
	}
	var data struct {
		Search struct {
			Results json.RawMessage `json:"results"`
		} `json:"search"`
	}
	if err := p.query(ctx, `query TroveSearch($q:String!){search(query:$q,query_type:"Book",per_page:10,page:1){results}}`, map[string]string{"q": text}, &data); err != nil {
		return nil, err
	}
	raw := data.Search.Results
	if len(raw) > 0 && raw[0] == '"' {
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			return nil, errors.New("búsqueda de catálogo inválida")
		}
		raw = []byte(text)
	}
	var parsed struct {
		Hits []struct {
			Document map[string]json.RawMessage `json:"document"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, errors.New("formato de búsqueda Hardcover no reconocido")
	}
	result := []catalog.Work{}
	seen := map[string]bool{}
	for _, hit := range parsed.Hits {
		doc := hit.Document
		var id any
		decoder := json.NewDecoder(bytes.NewReader(doc["id"]))
		decoder.UseNumber()
		if decoder.Decode(&id) != nil {
			continue
		}
		key := fmt.Sprint(id)
		if !catalog.ValidID(catalog.Hardcover, key) || seen[key] {
			continue
		}
		seen[key] = true
		var title string
		_ = json.Unmarshal(doc["title"], &title)
		if title == "" {
			continue
		}
		var authors []string
		_ = json.Unmarshal(doc["author_names"], &authors)
		if authors == nil {
			authors = []string{}
		}
		var img remoteImage
		_ = json.Unmarshal(doc["image"], &img)
		if img.URL == "" {
			_ = json.Unmarshal(doc["cached_image"], &img)
		}
		result = append(result, catalog.Work{ID: key, Title: title, Authors: authors, CoverURL: cover(&img)})
	}
	return result, nil
}

func (p *Hardcover) Editions(ctx context.Context, id string) ([]catalog.Edition, error) {
	number, _ := strconv.ParseInt(id, 10, 64)
	var data struct {
		Editions []hcEdition `json:"editions"`
	}
	if err := p.query(ctx, `query TroveEditions($id:Int!){editions(where:{book_id:{_eq:$id}},order_by:{users_count:desc},limit:20){`+editionFields+`}}`, map[string]int64{"id": number}, &data); err != nil {
		return nil, err
	}
	result := []catalog.Edition{}
	for _, v := range data.Editions {
		result = append(result, edition(v))
	}
	return result, nil
}

func (p *Hardcover) work(ctx context.Context, id int64) (catalog.Work, error) {
	var data struct {
		Books []struct {
			ID            int64        `json:"id"`
			Title         string       `json:"title"`
			Description   string       `json:"description"`
			Canonical     *int64       `json:"canonical_id"`
			Image         *remoteImage `json:"image"`
			Contributions []struct {
				Author struct {
					Name string `json:"name"`
				} `json:"author"`
			} `json:"contributions"`
			Series []struct {
				Position    *float64 `json:"position"`
				Compilation bool     `json:"compilation"`
				Series      struct {
					ID   int64  `json:"id"`
					Name string `json:"name"`
				} `json:"series"`
			} `json:"book_series"`
		} `json:"books"`
	}
	seen := map[int64]bool{}
	for attempt := 0; attempt < 4; attempt++ {
		if seen[id] {
			return catalog.Work{}, errors.New("Hardcover devolvió una redirección circular")
		}
		seen[id] = true
		query := `query TroveWork($id:Int!){books(where:{id:{_eq:$id}},limit:1){id title canonical_id description image{url} contributions(where:{contributor_role:{slug:{_eq:"author"}},contributable_type:{_eq:"Book"}}){author{name}} book_series(limit:12){position compilation series{id name}}}}`
		if err := p.query(ctx, query, map[string]int64{"id": id}, &data); err != nil {
			return catalog.Work{}, err
		}
		if len(data.Books) != 1 {
			return catalog.Work{}, errors.New("la obra ya no está disponible en Hardcover")
		}
		v := data.Books[0]
		if v.Canonical != nil {
			id = *v.Canonical
			continue
		}
		out := catalog.Work{ID: strconv.FormatInt(v.ID, 10), Title: v.Title, Description: plain(v.Description), CoverURL: cover(v.Image), Authors: []string{}, Series: []catalog.Series{}}
		for _, author := range v.Contributions {
			if author.Author.Name != "" {
				out.Authors = append(out.Authors, author.Author.Name)
			}
		}
		for _, series := range v.Series {
			if !series.Compilation {
				out.Series = append(out.Series, catalog.Series{ID: strconv.FormatInt(series.Series.ID, 10), Name: series.Series.Name, Position: series.Position})
			}
		}
		return out, nil
	}
	return catalog.Work{}, errors.New("demasiadas redirecciones de obra")
}

func (p *Hardcover) Resolve(ctx context.Context, id string) (catalog.Work, catalog.Edition, error) {
	number, _ := strconv.ParseInt(id, 10, 64)
	var data struct {
		Editions []hcEdition `json:"editions"`
	}
	if err := p.query(ctx, `query TroveEdition($id:Int!){editions(where:{id:{_eq:$id}},limit:1){`+editionFields+`}}`, map[string]int64{"id": number}, &data); err != nil {
		return catalog.Work{}, catalog.Edition{}, err
	}
	if len(data.Editions) != 1 {
		return catalog.Work{}, catalog.Edition{}, errors.New("la edición ya no está disponible")
	}
	v := data.Editions[0]
	work, err := p.work(ctx, v.Book.ID)
	if err != nil {
		return catalog.Work{}, catalog.Edition{}, err
	}
	out := edition(v)
	out.WorkID = work.ID
	return work, out, nil
}
