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

func TestHardcoverClassificationMapsGenresAndMoodsWithoutPrivateTaggings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !strings.Contains(body.Query, "taggable_counts(") || strings.Contains(body.Query, "user{") {
			t.Error("classification must use aggregate public tags")
		}
		_, _ = w.Write([]byte(`{"data":{"books":[{"id":10,"title":"Synthetic book","taggable_counts":[{"tag":{"id":4,"tag":"Fantasy","tag_category":{"slug":"genre"}}},{"tag":{"id":5,"tag":"Adventurous","tag_category":{"slug":"mood"}}},{"tag":{"id":6,"tag":"Warning","tag_category":{"slug":"content_warning"}}},{"tag":null}]}]}}`))
	}))
	defer server.Close()
	p := NewHardcover(func() (string, error) { return "synthetic-secret", nil })
	p.endpoint = server.URL
	p.pace.interval = 0
	w, err := p.work(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if w.Classification == nil || len(w.Classification.Genres) != 1 || len(w.Classification.Tags) != 1 || w.Classification.Tags[0].Origins[0].Category != "mood" || w.Classification.Genres[0].Origins[0].ID != "4" {
		t.Fatal("classification lost taxonomy, identity or mixed warnings into moods")
	}
}

func TestRanobeClassificationBelongsToSeriesNotNestedBookSeries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/series/10":
			_, _ = w.Write([]byte(`{"series":{"id":10,"title":"Synthetic novel","tags":[{"id":7,"name":"fantasy","ttype":"genre"},{"id":26,"name":"anime tie-in","ttype":"tag"}],"books":[{"id":20,"title":"Volume 1","book_type":"main","sort_order":1}]}}`))
		case "/book/20":
			_, _ = w.Write([]byte(`{"book":{"id":20,"title":"Volume 1","series":{"id":10,"books":[{"id":20}],"tags":[{"id":7,"name":"fantasy","ttype":"genre"}]},"releases":[]}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	p := NewRanobeDB()
	p.endpoint = server.URL
	p.pace.interval = 0
	s, err := p.Series(context.Background(), "10")
	if err != nil {
		t.Fatal(err)
	}
	if s.Classification == nil || s.Classification.Genres[0].Name != "fantasy" || s.Classification.Tags[0].Origins[0].ID != "26" {
		t.Fatal("series classification missing")
	}
	pub, err := novelcatalog.New(p, nil).Resolve(context.Background(), s, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	if pub.Classification != nil {
		t.Fatal("inherited nested series tags into the book")
	}
}
