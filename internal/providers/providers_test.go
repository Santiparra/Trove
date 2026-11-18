package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHardcoverNormalizesISBNRequestAndDeduplicatesSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer synthetic-secret" || r.URL.RawQuery != "" {
			t.Error("credencial/URL no corresponde")
		}
		var body struct {
			Query     string            `json:"query"`
			Variables map[string]string `json:"variables"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil {
			t.Fatal("consulta inválida")
		}
		if strings.Contains(body.Query, "TroveISBN") {
			if body.Variables["long"] != "9780547928227" || body.Variables["short"] != "054792822X" {
				t.Error("no se consultaron equivalencias")
			}
			_, _ = w.Write([]byte(`{"data":{"editions":[{"id":30,"title":"Edición","book":{"id":10,"title":"Obra"}},{"id":31,"title":"Otra","book":{"id":10,"title":"Obra"}}]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"data":{"search":{"results":{"hits":[{"document":{"id":"10","title":"Obra","author_names":["Autor"]}},{"document":{"id":"10","title":"Duplicado"}}]}}}}`))
	}))
	defer server.Close()
	p := NewHardcover(func() (string, error) { return "synthetic-secret", nil })
	p.endpoint = server.URL
	p.pace.interval = 0
	items, err := p.Search(context.Background(), "0-547-92822-X")
	if err != nil || len(items) != 1 || items[0].MatchedEditionID != "30" {
		t.Fatal("lookup ISBN incorrecto")
	}
	items, err = p.Search(context.Background(), "Autor")
	if err != nil || len(items) != 1 || len(items[0].Authors) != 1 {
		t.Fatal("búsqueda textual incorrecta")
	}
}

func TestProviderErrorsAndRedirectsNeverExposeOrForwardKey(t *testing.T) {
	for _, status := range []int{401, 429, 200} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{"errors":[{"message":"synthetic-secret"}]}`))
			}))
			defer server.Close()
			p := NewHardcover(func() (string, error) { return "synthetic-secret", nil })
			p.endpoint = server.URL
			p.pace.interval = 0
			_, err := p.Search(context.Background(), "Title")
			if err == nil || strings.Contains(err.Error(), "synthetic-secret") {
				t.Fatal("error ocultado o secreto expuesto")
			}
		})
	}
	called := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true }))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, http.StatusFound) }))
	defer redirect.Close()
	p := NewHardcover(func() (string, error) { return "synthetic-secret", nil })
	p.endpoint = redirect.URL
	p.pace.interval = 0
	if _, err := p.Search(context.Background(), "Title"); err == nil || called {
		t.Fatal("se siguió redirección de API")
	}
}

func TestOpenLibraryConfirmsEditionAndDoesNotReceivePrimaryAuthorization(t *testing.T) {
	conflict := false
	workCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Error("se envió autorización principal al secundario")
		}
		if strings.HasPrefix(r.URL.Path, "/isbn/") {
			if conflict {
				_, _ = w.Write([]byte(`{"key":"/books/OL1M","isbn_10":["0131103628","0131103709"],"isbn_13":["9780131103627"],"covers":[1]}`))
			} else {
				_, _ = w.Write([]byte(`{"key":"/books/OL1M","isbn_13":["9780131103627"],"covers":[1],"works":[{"key":"/works/OL2W"}]}`))
			}
			return
		}
		workCalls++
		_, _ = w.Write([]byte(`{"description":{"value":"<p>Texto público</p>"}}`))
	}))
	defer server.Close()
	p := NewOpenLibrary()
	p.endpoint = server.URL
	p.pace.interval = 0
	item, err := p.Lookup(context.Background(), "9780131103627")
	if err != nil || item.Description != "Texto público" || item.EditionID != "OL1M" || item.WorkID != "OL2W" || item.CoverURL == "" {
		t.Fatal("no se obtuvo metadata de edición confirmada")
	}
	conflict = true
	item, err = p.Lookup(context.Background(), "9780131103627")
	if err != nil || item.CoverURL != "" || len(item.Warnings) == 0 || workCalls != 1 {
		t.Fatal("se mezclaron ISBN contradictorios")
	}
}
