package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Santiparra/Trove/internal/novelcatalog"
)

func ranobeFixture(t *testing.T, releaseBody string) *RanobeDB {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/series":
			if r.URL.Query().Get("q") != "Classroom of the Elite" {
				t.Errorf("query no transparente: %q", r.URL.Query().Get("q"))
			}
			_, _ = w.Write([]byte(`{"series":[{"id":10,"title":"Classroom of the Elite","book":{"image":{"filename":"cover.jpg"}}},{"id":11,"title":"Classroom of the Elite (Manga)"}],"count":"2","currentPage":1,"totalPages":1}`))
		case "/series/10":
			_, _ = w.Write([]byte(`{"series":{"id":10,"title":"Classroom of the Elite","description":"An anime and manga tie-in.","books":[{"id":21,"title":"Part 2, Volume 1","book_type":"main","sort_order":2.5},{"id":99,"title":"Side story","book_type":"side","sort_order":2},{"id":20,"title":"Volume 1","book_type":"main","sort_order":1}],"staff":[{"staff_id":40,"role_type":"author","name":"Author","romaji":"Author A"}]}}`))
		case "/book/20":
			_, _ = w.Write([]byte(`{"book":{"id":20,"title":"Volume 1","description":"Novel synopsis","image":{"filename":"cover.jpg"},"series":{"id":10,"books":[{"id":20},{"id":21}]},"releases":[{"id":30,"title":"Volume 1 (Light Novel)","lang":"en","format":"print","isbn13":"9780547928227"}]}}`))
		case "/release/30":
			_, _ = w.Write([]byte(releaseBody))
		default:
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(server.Close)
	p := NewRanobeDB()
	p.endpoint = server.URL
	p.http = server.Client()
	p.pace.interval = 0
	return p
}

const validNovelRelease = `{"release":{"id":30,"title":"Volume 1 (Light Novel)","lang":"en","format":"print","isbn13":"9780547928227","release_date":20269999,"books":[{"id":20,"rtype":"complete"}],"publishers":[{"name":"Synthetic publisher"}]}}`

func TestRanobeSearchAndHierarchyUseProviderRulesWithoutMangaSuffixes(t *testing.T) {
	p := ranobeFixture(t, validNovelRelease)
	ctx := context.Background()
	result, err := p.Search(ctx, "Classroom of the Elite", "", 1)
	if err != nil || len(result.Items) != 1 || result.Items[0].ID != "10" {
		t.Fatalf("candidatos incorrectos: %+v %v", result, err)
	}
	series, err := p.Series(ctx, "10")
	if err != nil {
		t.Fatal(err)
	}
	if len(series.Books) != 2 || series.Books[1].Number != 2 || series.Books[1].Order != 2.5 || series.Books[1].Title != "Part 2, Volume 1" {
		t.Fatal("perdió orden original o incluyó extras")
	}
	pub, err := novelcatalog.New(p, nil).Resolve(ctx, series, 1, "30")
	if err != nil {
		t.Fatal(err)
	}
	if pub.SeriesID != "10" || pub.BookID != "20" || pub.ReleaseID != "30" || pub.Evidence != "ranobedb-main" || pub.ISBN != "9780547928227" || pub.PublishedOn != "2026" || pub.CoverURL != "https://images.ranobedb.org/cover.jpg" {
		t.Fatalf("publicación incorrecta: %+v", pub)
	}
}

func TestRanobeRejectsMangaPartialAndCompilationPublications(t *testing.T) {
	for _, body := range []string{
		`{"release":{"id":30,"title":"Series (Manga)","lang":"en","format":"print","books":[{"id":20,"rtype":"complete"}]}}`,
		`{"release":{"id":30,"title":"Partial","lang":"en","format":"print","books":[{"id":20,"rtype":"partial"}]}}`,
		`{"release":{"id":30,"title":"Omnibus","lang":"en","format":"print","books":[{"id":20,"rtype":"complete"},{"id":21,"rtype":"complete"}]}}`,
		`{"release":{"id":30,"title":"Wrong volume","lang":"en","format":"print","books":[{"id":21,"rtype":"complete"}]}}`,
	} {
		p := ranobeFixture(t, body)
		if _, err := p.Release(context.Background(), "20", "30"); err == nil {
			t.Fatal("aceptó una publicación no verificable")
		}
	}
	p := ranobeFixture(t, strings.Replace(validNovelRelease, `"lang":"en"`, `"lang":"ja"`, 1))
	s, err := p.Series(context.Background(), "10")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = novelcatalog.New(p, nil).Resolve(context.Background(), s, 1, "30"); err == nil {
		t.Fatal("cambió idioma detrás de la selección")
	}
}

func TestHardcoverNovelSupplementRequiresExactIdentityAndRejectsStructuredGraphicNovel(t *testing.T) {
	for _, test := range []struct {
		name, title, category, isbn10 string
		match                         bool
	}{
		{"light novel with misleading community tags", "Volume 1 (Light Novel)", "Light Novel", "054792822X", true},
		{"unannotated novel with exact ISBN", "Volume 1", "Book", "054792822X", true},
		{"explicit manga title", "Volume 1 (Manga)", "Book", "054792822X", false},
		{"graphic novel category without title marker", "Volume 1", "Graphic Novel", "054792822X", false},
		{"conflicting valid ISBN pair", "Volume 1 (Light Novel)", "Light Novel", "0131103628", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var query struct {
					Query     string            `json:"query"`
					Variables map[string]string `json:"variables"`
				}
				if json.NewDecoder(r.Body).Decode(&query) != nil {
					t.Error("consulta inválida")
				}
				if query.Variables["long"] != "9780547928227" || strings.Contains(query.Query, "search(query") || len(query.Variables) != 2 {
					t.Error("complementó por título o identificador incorrecto")
				}
				book := map[string]any{"id": 50, "title": test.title, "description": "Synopsis", "book_category_id": 10, "cached_tags": map[string]any{"Genre": []any{map[string]any{"tag": "Manga"}}}}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"book_categories": []any{map[string]any{"id": 10, "name": test.category}}, "editions": []any{map[string]any{"id": 60, "title": test.title, "isbn_13": "9780547928227", "isbn_10": test.isbn10, "pages": 240, "edition_format": "Paperback", "image": map[string]string{"url": "https://assets.hardcover.app/example.jpg"}, "language": map[string]string{"code2": "en"}, "book": book}}}})
			}))
			defer server.Close()
			p := NewHardcover(func() (string, error) { return "synthetic-secret", nil })
			p.http = server.Client()
			p.endpoint = server.URL
			p.pace.interval = 0
			extra, err := p.NovelPublication(context.Background(), "9780547928227", "en", "print")
			if err != nil {
				t.Fatal(err)
			}
			if (extra.EditionID != "") != test.match {
				t.Fatalf("desambiguación incorrecta: %+v", extra)
			}
			if test.match && (extra.Pages == nil || *extra.Pages != 240 || extra.WorkID != "50") {
				t.Fatal("no conservó metadata de edición")
			}
		})
	}
}

func TestRanobeDatesDoNotInventUnknownMonthOrDay(t *testing.T) {
	for input, want := range map[int]string{99999999: "", 20269999: "2026", 20260299: "2026-02", 20260230: "", 20260201: "2026-02-01"} {
		if got := rnDate(input); got != want {
			t.Fatalf("fecha %d: %s != %s", input, got, want)
		}
	}
}

func TestRanobeSharesLocalOnlyPolicyAndReportsHealthWithoutCredentials(t *testing.T) {
	policy, err := NewNetworkPolicy(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = policy.SetLocalOnly(true); err != nil {
		t.Fatal(err)
	}
	r := NewRanobeDB()
	r.pace.interval = 0
	h := NewHealth(NewHardcover(func() (string, error) { return "", nil }), NewOpenLibrary(), policy, r)
	defer h.Close()
	h.Check()
	if len(h.State().Providers) != 3 {
		t.Fatal("RanobeDB no aparece en salud")
	}
	if _, err = r.Search(context.Background(), "Classroom of the Elite", "", 1); !isLocalOnly(err) {
		t.Fatal("RanobeDB eludió modo local")
	}
	if r.health.snapshot().Failures != 0 {
		t.Fatal("modo local contó como caída")
	}
}

func TestHardcover979LookupDoesNotQueryEmptyISBN10(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var query struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&query)
		if len(query.Variables) != 1 || query.Variables["long"] != "9790000000001" || strings.Contains(query.Query, "$short") {
			t.Error("979 produjo búsqueda por ISBN-10 vacío")
		}
		_, _ = w.Write([]byte(`{"data":{"book_categories":[{"id":10,"name":"Light Novel"}],"editions":[{"id":60,"isbn_13":"9790000000001","edition_format":"Paperback","language":{"code2":"en"},"book":{"id":50,"title":"Novel","book_category_id":10}}]}}`))
	}))
	defer server.Close()
	p := NewHardcover(func() (string, error) { return "synthetic", nil })
	p.endpoint = server.URL
	p.http = server.Client()
	p.pace.interval = 0
	extra, err := p.NovelPublication(context.Background(), "9790000000001", "en", "print")
	if err != nil || extra.EditionID != "60" {
		t.Fatalf("no resolvió ISBN 979 seguro: %+v %v", extra, err)
	}
}
