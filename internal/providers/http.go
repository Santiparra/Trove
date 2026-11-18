// Package providers translates remote catalogs; provider response types do not enter the domain.
package providers

import (
	"context"
	"encoding/json"
	"errors"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const userAgent = "Trove/1.0 (https://github.com/Santiparra/Trove)"

func client() *http.Client {
	return &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

type pace struct {
	mu       sync.Mutex
	next     time.Time
	interval time.Duration
}

func (p *pace) wait(ctx context.Context) error {
	p.mu.Lock()
	at := time.Now()
	if p.next.After(at) {
		at = p.next
	}
	p.next = at.Add(p.interval)
	p.mu.Unlock()
	timer := time.NewTimer(time.Until(at))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
func responseJSON(response *http.Response, output any) error {
	defer response.Body.Close()
	if response.StatusCode == 401 || response.StatusCode == 403 {
		return errors.New("el proveedor rechazó el acceso; revisá la clave y sus permisos de catálogo")
	}
	if response.StatusCode == 429 {
		return errors.New("el proveedor alcanzó su límite; esperá antes de reintentar")
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("el proveedor no está disponible; se conservan los datos locales")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 2<<20+1))
	if err != nil || len(data) > 2<<20 {
		return errors.New("respuesta de catálogo demasiado grande o incompleta")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err = decoder.Decode(output); err != nil {
		return errors.New("respuesta de catálogo inválida")
	}
	return nil
}

var tags = regexp.MustCompile(`(?s)<[^>]*>`)

func plain(value string) string {
	value = strings.ReplaceAll(value, "<br>", "\n")
	value = strings.ReplaceAll(value, "<br/>", "\n")
	value = strings.ReplaceAll(value, "</p>", "\n\n")
	return strings.TrimSpace(html.UnescapeString(tags.ReplaceAllString(value, "")))
}
