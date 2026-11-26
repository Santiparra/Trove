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

func TestHardcoverWorkRatingUsesOriginalFivePointAggregate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !strings.Contains(body.Query, "rating ratings_count") {
			t.Error("rating aggregate not requested")
		}
		_, _ = w.Write([]byte(`{"data":{"books":[{"id":10,"title":"Synthetic book","rating":4.135,"ratings_count":12}]}}`))
	}))
	defer server.Close()
	p := NewHardcover(func() (string, error) { return "synthetic-secret", nil })
	p.endpoint = server.URL
	p.pace.interval = 0
	work, err := p.work(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if work.Rating == nil || work.Rating.Value != 4.135 || work.Rating.Scale != 5 || *work.Rating.Votes != 12 || work.Rating.ID != "10" || work.Rating.Entity != "work" {
		t.Fatal("aggregate metadata changed")
	}
}

func TestRanobeSeriesAndBookRatingsAreNotReleaseOrHardcoverScores(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/series/10":
			_, _ = w.Write([]byte(`{"series":{"id":10,"title":"Synthetic novel","rating":{"score":8.65,"count":10},"books":[{"id":20,"title":"Volume 1","sort_order":1,"book_type":"main"}]}}`))
		case "/book/20":
			_, _ = w.Write([]byte(`{"book":{"id":20,"title":"Volume 1","rating":{"score":7.333,"count":3},"series":{"id":10,"books":[{"id":20}]},"releases":[{"id":30,"title":"Volume 1","lang":"en","format":"print"}]}}`))
		case "/release/30":
			_, _ = w.Write([]byte(`{"release":{"id":30,"title":"Volume 1","lang":"en","format":"print","books":[{"id":20,"rtype":"complete"}]}}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	p := NewRanobeDB()
	p.endpoint = server.URL
	p.pace.interval = 0
	series, err := p.Series(context.Background(), "10")
	if err != nil {
		t.Fatal(err)
	}
	if series.Rating == nil || series.Rating.Value != 8.65 || series.Rating.Scale != 10 || *series.Rating.Votes != 10 {
		t.Fatal("series aggregate lost")
	}
	publication, err := novelcatalog.New(p, nil).Resolve(context.Background(), series, 1, "30")
	if err != nil {
		t.Fatal(err)
	}
	if publication.Rating == nil || publication.Rating.Value != 7.333 || publication.Rating.Entity != "book" || publication.Rating.ID != "20" || *publication.Rating.Votes != 3 {
		t.Fatal("book aggregate was lost or confused with series/release")
	}
	if ranobeRating(nil, "book", "20", series.FetchedAt) != nil || ranobeRating(&rnRating{Score: 0, Count: 0}, "book", "20", series.FetchedAt) != nil {
		t.Fatal("no votes became zero score")
	}
}
