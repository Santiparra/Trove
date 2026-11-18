package providers

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCoversUseValidatedLocalCacheWithoutCatalogCredentials(t *testing.T) {
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 4, 6))); err != nil {
		t.Fatal(err)
	}
	c := NewCovers(t.TempDir(), nil)
	calls := 0
	c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Header.Get("Authorization") != "" {
			t.Fatal("se envió clave al host de imágenes")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(data.Bytes())), Header: http.Header{}}, nil
	})
	first, err := c.Get(context.Background(), "https://assets.hardcover.app/cover.png")
	if err != nil || !strings.HasPrefix(first, "data:image/png;base64,") {
		t.Fatal("no se obtuvo portada local")
	}
	c.http.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) { calls++; return nil, errors.New("offline") })
	second, err := c.Get(context.Background(), "https://assets.hardcover.app/cover.png")
	if err != nil || second != first || calls != 1 {
		t.Fatal("no se conservó portada offline")
	}
	if _, err = c.Get(context.Background(), "http://127.0.0.1/private"); err == nil || calls != 1 {
		t.Fatal("se permitió un destino arbitrario")
	}
	if _, err = coverData([]byte(`<svg onload="alert(1)"></svg>`)); err == nil {
		t.Fatal("se aceptó una imagen ejecutable/no admitida")
	}
}
