package providers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/novelcatalog"
)

type RanobeDB struct {
	http     *http.Client
	endpoint string
	pace     pace
	health   *monitor
}

func fmtID(id int64) string { return strconv.FormatInt(id, 10) }
func NewRanobeDB() *RanobeDB {
	return &RanobeDB{http: client(), endpoint: "https://ranobedb.org/api/v0", pace: pace{interval: 1100 * time.Millisecond}}
}
func (p *RanobeDB) get(ctx context.Context, path string, out any) error {
	call := func() error { return p.getRaw(ctx, path, out) }
	if p.health != nil {
		return p.health.run(ctx, p.pace.wait, call)
	}
	if err := p.pace.wait(ctx); err != nil {
		return err
	}
	return call()
}
func (p *RanobeDB) getRaw(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+path, nil)
	if err != nil {
		return errors.New("consulta RanobeDB inválida")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	response, err := p.http.Do(req)
	if err != nil {
		if isLocalOnly(err) {
			return err
		}
		return failure("network", "no se pudo conectar con RanobeDB")
	}
	return responseJSON(response, out)
}

type rnImage struct {
	Filename string `json:"filename"`
}

var imageName = regexp.MustCompile(`^[A-Za-z0-9_-]+\.(?:jpg|jpeg|png|webp)$`)

func rnCover(image *rnImage) string {
	if image == nil || !imageName.MatchString(image.Filename) {
		return ""
	}
	return "https://images.ranobedb.org/" + image.Filename
}

type rnBookRef struct {
	ID            int64    `json:"id"`
	Title         string   `json:"title"`
	OriginalTitle string   `json:"title_orig"`
	Type          string   `json:"book_type"`
	Order         float64  `json:"sort_order"`
	Image         *rnImage `json:"image"`
}
type rnSeries struct {
	ID            int64       `json:"id"`
	Title         string      `json:"title"`
	OriginalTitle string      `json:"title_orig"`
	Description   string      `json:"description"`
	Hidden        bool        `json:"hidden"`
	Books         []rnBookRef `json:"books"`
	Staff         []struct {
		ID     int64  `json:"staff_id"`
		Role   string `json:"role_type"`
		Name   string `json:"name"`
		Romaji string `json:"romaji"`
	} `json:"staff"`
	Book struct {
		Image *rnImage `json:"image"`
	} `json:"book"`
}

func rnDate(value int) string {
	if value <= 0 {
		return ""
	}
	raw := fmt.Sprintf("%08d", value)
	if len(raw) != 8 || raw[:4] == "9999" {
		return ""
	}
	if raw[4:6] == "99" {
		if _, err := time.Parse("2006", raw[:4]); err == nil {
			return raw[:4]
		}
		return ""
	}
	if raw[6:] == "99" {
		if _, err := time.Parse("2006-01", raw[:4]+"-"+raw[4:6]); err == nil {
			return raw[:4] + "-" + raw[4:6]
		}
		return ""
	}
	date := raw[:4] + "-" + raw[4:6] + "-" + raw[6:]
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return ""
	}
	return date
}

type rnRelease struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Lang        string   `json:"lang"`
	Format      string   `json:"format"`
	ISBN        string   `json:"isbn13"`
	Date        int      `json:"release_date"`
	Description string   `json:"description"`
	Pages       *int     `json:"pages"`
	Image       *rnImage `json:"image"`
	Hidden      bool     `json:"hidden"`
	Publishers  []struct {
		Name string `json:"name"`
	} `json:"publishers"`
	Books []struct {
		ID       int64  `json:"id"`
		Relation string `json:"rtype"`
	} `json:"books"`
}

func rnOption(r rnRelease) novelcatalog.Release {
	return novelcatalog.Release{ID: strconv.FormatInt(r.ID, 10), Title: r.Title, Language: r.Lang, Format: r.Format, ISBN: catalog.ISBN13(r.ISBN), PublishedOn: rnDate(r.Date)}
}
func (p *RanobeDB) searchSeries(ctx context.Context, values url.Values) (novelcatalog.SearchResult, error) {
	var data struct {
		Series []rnSeries      `json:"series"`
		Count  json.RawMessage `json:"count"`
		Page   int             `json:"currentPage"`
		Pages  int             `json:"totalPages"`
	}
	if err := p.get(ctx, "/series?"+values.Encode(), &data); err != nil {
		return novelcatalog.SearchResult{}, err
	}
	result := novelcatalog.SearchResult{Items: []novelcatalog.Candidate{}, Page: data.Page, Pages: data.Pages, Warnings: []string{}}
	var countText string
	_ = json.Unmarshal(data.Count, &countText)
	if countText == "" {
		countText = string(data.Count)
	}
	result.Total, _ = strconv.Atoi(countText)
	for _, s := range data.Series {
		if s.ID <= 0 || s.Hidden || strings.TrimSpace(s.Title) == "" || novelcatalog.ExplicitManga(s.Title) || novelcatalog.ExplicitManga(s.OriginalTitle) {
			continue
		}
		result.Items = append(result.Items, novelcatalog.Candidate{ID: strconv.FormatInt(s.ID, 10), Title: s.Title, OriginalTitle: s.OriginalTitle, CoverURL: rnCover(s.Book.Image)})
	}
	return result, nil
}
func (p *RanobeDB) Search(ctx context.Context, query, authorID string, page int) (novelcatalog.SearchResult, error) {
	query = strings.TrimSpace(query)
	if page < 1 || page > 1000 || (authorID != "" && !novelcatalog.ValidID(authorID)) || (authorID == "" && len([]rune(query)) < 2) {
		return novelcatalog.SearchResult{}, errors.New("indicá al menos dos caracteres para buscar novelas")
	}
	values := url.Values{"limit": {"12"}, "page": {strconv.Itoa(page)}}
	if authorID != "" {
		values.Set("staff", authorID)
		return p.searchSeries(ctx, values)
	}
	if isbn := catalog.ISBN13(query); isbn != "" {
		var list struct {
			Releases []rnRelease `json:"releases"`
		}
		if err := p.get(ctx, "/releases?"+url.Values{"q": {isbn}, "limit": {"3"}}.Encode(), &list); err != nil {
			return novelcatalog.SearchResult{}, err
		}
		result := novelcatalog.SearchResult{Items: []novelcatalog.Candidate{}, Page: 1, Pages: 1, Warnings: []string{}}
		seen := map[string]bool{}
		for _, r := range list.Releases {
			if r.Hidden || catalog.ISBN13(r.ISBN) != isbn {
				continue
			}
			var detail struct {
				Release rnRelease `json:"release"`
			}
			if err := p.get(ctx, "/release/"+strconv.FormatInt(r.ID, 10), &detail); err != nil {
				return result, err
			}
			if len(detail.Release.Books) != 1 || detail.Release.Books[0].Relation != "complete" {
				continue
			}
			var book struct {
				Book struct {
					Series struct {
						ID            int64  `json:"id"`
						Title         string `json:"title"`
						OriginalTitle string `json:"title_orig"`
					} `json:"series"`
				} `json:"book"`
			}
			if err := p.get(ctx, "/book/"+strconv.FormatInt(detail.Release.Books[0].ID, 10), &book); err != nil {
				return result, err
			}
			id := strconv.FormatInt(book.Book.Series.ID, 10)
			if !novelcatalog.ValidID(id) || seen[id] || novelcatalog.ExplicitManga(book.Book.Series.Title) {
				continue
			}
			seen[id] = true
			result.Items = append(result.Items, novelcatalog.Candidate{ID: id, Title: book.Book.Series.Title, OriginalTitle: book.Book.Series.OriginalTitle, MatchedReleaseID: strconv.FormatInt(r.ID, 10), MatchedBookID: strconv.FormatInt(detail.Release.Books[0].ID, 10)})
		}
		result.Total = len(result.Items)
		return result, nil
	}
	values.Set("q", query)
	result, err := p.searchSeries(ctx, values)
	if err != nil || len(result.Items) > 0 {
		return result, err
	}
	// Author navigation/free text uses the provider's staff filter, not a title suffix.
	var staff struct {
		Staff []struct {
			ID     int64  `json:"id"`
			Name   string `json:"name"`
			Romaji string `json:"romaji"`
		} `json:"staff"`
	}
	if err := p.get(ctx, "/staff?"+url.Values{"q": {query}, "limit": {"5"}}.Encode(), &staff); err != nil {
		return result, err
	}
	key := func(value string) string {
		fields := strings.Fields(strings.ToLower(value))
		sort.Strings(fields)
		return strings.Join(fields, " ")
	}
	for _, a := range staff.Staff {
		if a.ID > 0 && (key(a.Name) == key(query) || key(a.Romaji) == key(query)) {
			values.Del("q")
			values.Set("staff", strconv.FormatInt(a.ID, 10))
			return p.searchSeries(ctx, values)
		}
	}
	return result, nil
}
func (p *RanobeDB) Series(ctx context.Context, id string) (novelcatalog.Series, error) {
	if !novelcatalog.ValidID(id) {
		return novelcatalog.Series{}, errors.New("serie RanobeDB inválida")
	}
	var data struct {
		Series rnSeries `json:"series"`
	}
	if err := p.get(ctx, "/series/"+id, &data); err != nil {
		return novelcatalog.Series{}, err
	}
	raw := data.Series
	if raw.Hidden || strconv.FormatInt(raw.ID, 10) != id {
		return novelcatalog.Series{}, errors.New("RanobeDB no identificó la novela solicitada")
	}
	refs := []rnBookRef{}
	for _, b := range raw.Books {
		if b.Type == "main" {
			refs = append(refs, b)
		}
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].Order < refs[j].Order })
	s := novelcatalog.Series{ID: id, Title: raw.Title, OriginalTitle: raw.OriginalTitle, Description: plain(raw.Description), Authors: []novelcatalog.Author{}, Books: []novelcatalog.BookRef{}, FetchedAt: time.Now().UTC().Format(time.RFC3339Nano), Language: "en", Format: "print"}
	for i, b := range refs {
		s.Books = append(s.Books, novelcatalog.BookRef{Label: novelcatalog.PrintedLabel(b.Title, b.OriginalTitle), ID: strconv.FormatInt(b.ID, 10), Number: i + 1, Order: b.Order, Title: b.Title, OriginalTitle: b.OriginalTitle, CoverURL: rnCover(b.Image)})
	}
	if len(s.Books) > 0 {
		s.CoverURL = s.Books[0].CoverURL
	}
	seen := map[int64]bool{}
	for _, a := range raw.Staff {
		if a.Role != "author" || a.ID <= 0 || seen[a.ID] {
			continue
		}
		name := a.Romaji
		if name == "" {
			name = a.Name
		}
		if strings.TrimSpace(name) == "" {
			continue
		}
		seen[a.ID] = true
		s.Authors = append(s.Authors, novelcatalog.Author{ID: strconv.FormatInt(a.ID, 10), Name: name})
	}
	return s, s.Validate()
}
func (p *RanobeDB) Book(ctx context.Context, seriesID, bookID string) (novelcatalog.Book, error) {
	if !novelcatalog.ValidID(seriesID) || !novelcatalog.ValidID(bookID) {
		return novelcatalog.Book{}, errors.New("referencia de volumen inválida")
	}
	var data struct {
		Book struct {
			ID            int64       `json:"id"`
			Title         string      `json:"title"`
			OriginalTitle string      `json:"title_orig"`
			Description   string      `json:"description"`
			Hidden        bool        `json:"hidden"`
			Image         *rnImage    `json:"image"`
			Series        rnSeries    `json:"series"`
			Releases      []rnRelease `json:"releases"`
		} `json:"book"`
	}
	if err := p.get(ctx, "/book/"+bookID, &data); err != nil {
		return novelcatalog.Book{}, err
	}
	b := data.Book
	main := false
	for _, ref := range b.Series.Books {
		if strconv.FormatInt(ref.ID, 10) == bookID && (ref.Type == "main" || ref.Type == "") {
			main = true
		}
	}
	if b.Hidden || strconv.FormatInt(b.ID, 10) != bookID || strconv.FormatInt(b.Series.ID, 10) != seriesID || !main || novelcatalog.ExplicitManga(b.Title) || novelcatalog.ExplicitManga(b.OriginalTitle) {
		return novelcatalog.Book{}, errors.New("el resultado no es un volumen principal verificable de esta novela")
	}
	result := novelcatalog.Book{ID: bookID, SeriesID: seriesID, Title: b.Title, OriginalTitle: b.OriginalTitle, Description: plain(b.Description), CoverURL: rnCover(b.Image), Releases: []novelcatalog.Release{}}
	for _, r := range b.Releases {
		if r.ID > 0 && !r.Hidden && !novelcatalog.ExplicitManga(r.Title) {
			result.Releases = append(result.Releases, rnOption(r))
		}
	}
	sort.SliceStable(result.Releases, func(i, j int) bool {
		a, b := result.Releases[i], result.Releases[j]
		score := func(r novelcatalog.Release) int {
			n := 0
			if r.Language == "en" {
				n += 4
			}
			if r.Format == "print" {
				n += 2
			}
			if r.ISBN != "" {
				n++
			}
			return n
		}
		if score(a) == score(b) {
			return a.ID < b.ID
		}
		return score(a) > score(b)
	})
	return result, nil
}
func (p *RanobeDB) Release(ctx context.Context, bookID, releaseID string) (novelcatalog.Publication, error) {
	if !novelcatalog.ValidID(bookID) || !novelcatalog.ValidID(releaseID) {
		return novelcatalog.Publication{}, errors.New("publicación inválida")
	}
	var data struct {
		Release rnRelease `json:"release"`
	}
	if err := p.get(ctx, "/release/"+releaseID, &data); err != nil {
		return novelcatalog.Publication{}, err
	}
	r := data.Release
	if r.Hidden || strconv.FormatInt(r.ID, 10) != releaseID || len(r.Books) != 1 || strconv.FormatInt(r.Books[0].ID, 10) != bookID || r.Books[0].Relation != "complete" || novelcatalog.ExplicitManga(r.Title) {
		return novelcatalog.Publication{}, errors.New("la publicación es ambigua, parcial, una compilación o no corresponde a esta novela; no se mezcló metadata")
	}
	format := "other"
	switch r.Format {
	case "digital":
		format = "digital"
	case "audio":
		format = "audio"
	case "print":
	default:
		return novelcatalog.Publication{}, errors.New("formato de publicación RanobeDB no reconocido")
	}
	pub := novelcatalog.Publication{BookID: bookID, ReleaseID: releaseID, ReleaseTitle: r.Title, Language: r.Lang, ReleaseFormat: r.Format, Format: format, ISBN: catalog.ISBN13(r.ISBN), Description: plain(r.Description), CoverURL: rnCover(r.Image), Pages: r.Pages, PublishedOn: rnDate(r.Date)}
	names := []string{}
	for _, publisher := range r.Publishers {
		if publisher.Name != "" {
			names = append(names, publisher.Name)
		}
	}
	pub.Publisher = strings.Join(names, " · ")
	if pub.Pages != nil && *pub.Pages <= 0 {
		pub.Pages = nil
	}
	return pub, nil
}
func (p *RanobeDB) check(ctx context.Context) error {
	var data struct {
		Series []rnSeries `json:"series"`
	}
	call := func() error {
		if err := p.getRaw(ctx, "/series?q=Spice%20and%20Wolf&limit=1", &data); err != nil {
			return err
		}
		if len(data.Series) != 1 || data.Series[0].ID <= 0 || strings.TrimSpace(data.Series[0].Title) == "" {
			return failure("response", "RanobeDB no devolvió una respuesta reconocible")
		}
		return nil
	}
	if p.health != nil {
		return p.health.run(ctx, p.pace.wait, call)
	}
	if err := p.pace.wait(ctx); err != nil {
		return err
	}
	return call()
}
