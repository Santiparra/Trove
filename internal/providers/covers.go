package providers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	img "image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/Santiparra/Trove/internal/catalog"
	_ "golang.org/x/image/webp"
)

type coverFlight struct {
	done  chan struct{}
	value string
	err   error
}
type Covers struct {
	dir     string
	http    *http.Client
	mu      sync.Mutex
	flights map[string]*coverFlight
}

func NewCovers(dir string) *Covers {
	h := client()
	h.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 || !catalog.ValidCoverURL(req.URL.String()) {
			return errors.New("redirección de portada no permitida")
		}
		return nil
	}
	return &Covers{dir: dir, http: h, flights: map[string]*coverFlight{}}
}

const maxCover = 6 << 20

func coverData(data []byte) (string, error) {
	if len(data) > maxCover {
		return "", errors.New("portada demasiado grande")
	}
	config, format, err := img.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width <= 0 || config.Height <= 0 || config.Width > 8000 || config.Height > 8000 {
		return "", errors.New("portada inválida")
	}
	mime := map[string]string{"jpeg": "image/jpeg", "png": "image/png", "gif": "image/gif", "webp": "image/webp"}[format]
	if mime == "" {
		return "", errors.New("formato de portada no admitido")
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}
func (c *Covers) Get(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", nil
	}
	if !catalog.ValidCoverURL(url) {
		return "", errors.New("destino de portada no permitido")
	}
	sum := sha256.Sum256([]byte(url))
	path := filepath.Join(c.dir, hex.EncodeToString(sum[:])+".cover")
	if data, err := os.ReadFile(path); err == nil {
		if result, err := coverData(data); err == nil {
			return result, nil
		}
	}
	c.mu.Lock()
	if f := c.flights[url]; f != nil {
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return "", errors.New("portada cancelada")
		case <-f.done:
			return f.value, f.err
		}
	}
	f := &coverFlight{done: make(chan struct{})}
	c.flights[url] = f
	c.mu.Unlock()
	f.value, f.err = c.load(ctx, url, path)
	c.mu.Lock()
	delete(c.flights, url)
	close(f.done)
	c.mu.Unlock()
	return f.value, f.err
}
func (c *Covers) load(ctx context.Context, url, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", errors.New("portada inválida")
	}
	req.Header.Set("User-Agent", userAgent)
	response, err := c.http.Do(req)
	if err != nil {
		return "", errors.New("no se pudo recuperar la portada")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return "", errors.New("portada no disponible")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxCover+1))
	if err != nil {
		return "", errors.New("portada incompleta")
	}
	result, err := coverData(data)
	if err != nil {
		return "", err
	}
	if os.MkdirAll(c.dir, 0700) == nil {
		if file, err := os.CreateTemp(c.dir, ".cover-"); err == nil {
			name := file.Name()
			_, writeErr := file.Write(data)
			closeErr := file.Close()
			if writeErr == nil && closeErr == nil {
				_ = os.Rename(name, path)
			}
			_ = os.Remove(name)
		}
	}
	return result, nil
}
