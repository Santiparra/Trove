package providers

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestLocalOnlyPersistsBlocksBothCatalogsAndKeepsCachedCovers(t *testing.T) {
	dir := t.TempDir()
	policy, err := NewNetworkPolicy(dir)
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	var imageData bytes.Buffer
	_ = png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 3)))
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(imageData.Bytes())), Header: http.Header{}}, nil
	})
	covers := NewCovers(t.TempDir(), nil)
	covers.http.Transport = transport
	policy.protect(covers.http)
	url := "https://assets.hardcover.app/cached.png"
	first, err := covers.Get(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	hc := NewHardcover(func() (string, error) { return "synthetic", nil })
	hc.http.Transport = transport
	hc.pace.interval = 0
	ol := NewOpenLibrary()
	ol.http.Transport = transport
	ol.pace.interval = 0
	health := NewHealth(hc, ol, policy)
	defer health.Close()
	if _, err = health.SetLocalOnly(true); err != nil {
		t.Fatal(err)
	}
	health.Check()
	health.wg.Wait()
	var output any
	if err = hc.query(context.Background(), "test", nil, &output); !isLocalOnly(err) {
		t.Fatal("Hardcover no bloqueado")
	}
	if err = ol.get(context.Background(), "/test", &output); !isLocalOnly(err) {
		t.Fatal("Open Library no bloqueado")
	}
	if _, err = covers.Get(context.Background(), "https://assets.hardcover.app/missing.png"); err == nil {
		t.Fatal("descargó portada sin caché")
	}
	second, err := covers.Get(context.Background(), url)
	if err != nil || second != first || calls.Load() != 1 {
		t.Fatal("modo local perdió caché o consultó la red")
	}
	for _, state := range health.State().Providers {
		if state.Failures != 0 || state.CheckedAt != "" {
			t.Fatal("modo local se contó como caída")
		}
	}
	reloaded, err := NewNetworkPolicy(dir)
	if err != nil || !reloaded.LocalOnly() {
		t.Fatal("modo no persistido")
	}
	other, err := NewNetworkPolicy(t.TempDir())
	if err != nil || other.LocalOnly() {
		t.Fatal("modo contaminó otro perfil")
	}
	if err = policy.SetLocalOnly(false); err != nil {
		t.Fatal(err)
	}
	if _, err = covers.Get(context.Background(), "https://assets.hardcover.app/new.png"); err != nil || calls.Load() != 2 {
		t.Fatal("no reactivó red")
	}
	reloaded, err = NewNetworkPolicy(dir)
	if err != nil || reloaded.LocalOnly() {
		t.Fatal("reactivación no persistida")
	}
}

func TestLocalOnlyCancelsResponseBodiesAlreadyInFlight(t *testing.T) {
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":`))
		w.(http.Flusher).Flush()
		close(entered)
		<-r.Context().Done()
	}))
	defer server.Close()
	policy, err := NewNetworkPolicy(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	hc := NewHardcover(func() (string, error) { return "synthetic", nil })
	hc.endpoint = server.URL
	hc.http = server.Client()
	hc.pace.interval = 0
	ol := NewOpenLibrary()
	health := NewHealth(hc, ol, policy)
	defer health.Close()
	done := make(chan error, 1)
	go func() { var out any; done <- hc.query(context.Background(), "test", nil, &out) }()
	<-entered
	if _, err = health.SetLocalOnly(true); err != nil {
		t.Fatal(err)
	}
	if err = <-done; !isLocalOnly(err) {
		t.Fatalf("no canceló cuerpo remoto: %v", err)
	}
	state := health.State().Providers[0]
	if state.Failures != 0 || state.CheckedAt != "" {
		t.Fatal("cancelación manual marcó caída")
	}
}
