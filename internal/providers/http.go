// Package providers translates remote catalogs; provider response types do not enter the domain.
package providers

import (
	"context"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
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
		return failure("access", "el proveedor rechazó el acceso; revisá la clave y sus permisos de catálogo")
	}
	if response.StatusCode == 429 {
		delay := recoveryDelay
		if seconds, err := strconv.Atoi(response.Header.Get("Retry-After")); err == nil && seconds > 0 {
			if seconds > 86400 {
				seconds = 86400
			}
			delay = time.Duration(seconds) * time.Second
		} else if date, err := http.ParseTime(response.Header.Get("Retry-After")); err == nil && time.Until(date) > 0 {
			delay = time.Until(date)
		}
		if delay > 24*time.Hour {
			delay = 24 * time.Hour
		}
		return &providerError{kind: "rate-limit", message: "el proveedor alcanzó su límite; esperá antes de reintentar", retry: delay}
	}
	if response.StatusCode == 404 {
		return failure("missing", "el proveedor no encontró este registro")
	}
	if response.StatusCode == 400 || response.StatusCode == 422 {
		return failure("request", "el proveedor no pudo resolver esta consulta")
	}
	if response.StatusCode != http.StatusOK {
		return failure("service", "el proveedor no está disponible; se conservan los datos locales")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 2<<20+1))
	if isLocalOnly(err) {
		return err
	}
	if err != nil || len(data) > 2<<20 {
		return failure("response", "respuesta de catálogo demasiado grande o incompleta")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err = decoder.Decode(output); err != nil {
		return failure("response", "respuesta de catálogo inválida")
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
