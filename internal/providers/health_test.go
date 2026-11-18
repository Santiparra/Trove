package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func noWait(context.Context) error { return nil }

func TestCircuitPausesAfterFailuresAndAdmitsOnlyOneRecovery(t *testing.T) {
	m := newMonitor("test", "Test", true)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m.clock = func() time.Time { return now }
	ctx := context.Background()
	var calls atomic.Int32
	bad := func() error { calls.Add(1); return failure("network", "sintético") }
	for i := 0; i < 3; i++ {
		if err := m.run(ctx, noWait, bad); err == nil {
			t.Fatal("falta error")
		}
	}
	if state := m.snapshot(); state.State != "paused" || state.Failures != 3 || state.RetryAt == "" {
		t.Fatalf("estado incorrecto: %+v", state)
	}
	_ = m.run(ctx, noWait, bad)
	if calls.Load() != 3 {
		t.Fatal("consultó durante pausa")
	}
	now = now.Add(recoveryDelay)
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		done <- m.run(ctx, noWait, func() error { calls.Add(1); close(entered); <-release; return nil })
	}()
	<-entered
	if err := m.run(ctx, noWait, bad); err == nil {
		t.Error("admitió dos pruebas de recuperación")
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if state := m.snapshot(); state.State != "healthy" || state.Failures != 0 || state.RetryAt != "" || calls.Load() != 4 {
		t.Fatalf("no recuperó: %+v", state)
	}
}

func TestQueuedRequestsDoNotEscapeAnOpenedCircuit(t *testing.T) {
	m := newMonitor("test", "Test", true)
	ctx := context.Background()
	waiting, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	var called atomic.Bool
	go func() {
		done <- m.run(ctx, func(context.Context) error { close(waiting); <-release; return nil }, func() error { called.Store(true); return nil })
	}()
	<-waiting
	for i := 0; i < 3; i++ {
		_ = m.run(ctx, noWait, func() error { return failure("network", "sintético") })
	}
	close(release)
	if err := <-done; err == nil || called.Load() {
		t.Fatal("una solicitud en cola eludió el circuito")
	}
}

func TestCredentialResetIgnoresOldRequestsAndAccessErrorsBlockUntilReset(t *testing.T) {
	m := newMonitor("test", "Test", true)
	ctx := context.Background()
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		done <- m.run(ctx, noWait, func() error { close(entered); <-release; return failure("access", "secreto sintético") })
	}()
	<-entered
	m.reset(true)
	_ = m.run(ctx, noWait, func() error { return nil })
	close(release)
	<-done
	if m.snapshot().State != "healthy" {
		t.Fatal("respuesta de clave vieja contaminó el estado")
	}
	_ = m.run(ctx, noWait, func() error { return failure("access", "no publicar") })
	called := false
	_ = m.run(ctx, noWait, func() error { called = true; return nil })
	if called || m.snapshot().State != "access-denied" {
		t.Fatal("no pausó acceso rechazado")
	}
	m.reset(false)
	_ = m.run(ctx, noWait, func() error { called = true; return nil })
	if called || m.snapshot().State != "unconfigured" {
		t.Fatal("consultó sin clave")
	}
	m.reset(true)
	if err := m.run(ctx, noWait, func() error { return nil }); err != nil || m.snapshot().State != "healthy" {
		t.Fatal("nueva clave no habilitó consultas")
	}
}

func TestCancellationAndPacingDoNotBecomeProviderFailures(t *testing.T) {
	m := newMonitor("test", "Test", true)
	now := time.Now()
	m.clock = func() time.Time { return now }
	_ = m.run(context.Background(), func(context.Context) error { now = now.Add(time.Second); return nil }, func() error { now = now.Add(7 * time.Millisecond); return nil })
	if m.snapshot().LatencyMS != 7 {
		t.Fatal("latencia incluye espera local de ritmo")
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = m.run(ctx, noWait, func() error { cancel(); return failure("network", "cancelada") })
	if m.snapshot().State != "healthy" || m.snapshot().Failures != 0 {
		t.Fatal("cancelación se contó como caída")
	}
}

func TestHTTPRateLimitAndNotFoundAreDifferentHealthOutcomes(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Header.Get("Authorization") != "" {
			t.Error("Open Library recibió credencial")
		}
		if strings.Contains(r.URL.Path, "limited") {
			w.Header().Set("Retry-After", "120")
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()
	p := NewOpenLibrary()
	p.endpoint = server.URL
	p.http = server.Client()
	p.pace.interval = 0
	p.health = newMonitor("openlibrary", "Open Library", true)
	var data any
	_ = p.get(context.Background(), "/missing", &data)
	if p.health.snapshot().State != "healthy" {
		t.Fatal("404 normal se contó como caída")
	}
	_ = p.get(context.Background(), "/limited", &data)
	state := p.health.snapshot()
	if state.State != "paused" {
		t.Fatal("429 no pausó")
	}
	at, _ := time.Parse(time.RFC3339Nano, state.RetryAt)
	if time.Until(at) < 119*time.Second {
		t.Fatal("no respetó Retry-After")
	}
	_ = p.get(context.Background(), "/missing", &data)
	if requests.Load() != 2 {
		t.Fatal("consultó durante límite")
	}
}

func TestBackgroundChecksAreBoundedAndNeverSendKeysToSecondary(t *testing.T) {
	var primaryCalls, secondaryCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			primaryCalls.Add(1)
			if r.Header.Get("Authorization") != "Bearer synthetic-secret" {
				t.Error("falta acceso primario")
			}
			var query struct {
				Query string `json:"query"`
			}
			_ = json.NewDecoder(r.Body).Decode(&query)
			if !strings.Contains(query.Query, "TroveHealth") || !strings.Contains(query.Query, "limit:1") {
				t.Error("comprobación no acotada")
			}
			_, _ = w.Write([]byte(`{"data":{"books":[{"id":10}]}}`))
			return
		}
		secondaryCalls.Add(1)
		if r.Header.Get("Authorization") != "" || r.URL.Path != "/isbn/9780140328721.json" {
			t.Error("comprobación secundaria inválida")
		}
		_, _ = w.Write([]byte(`{"key":"/books/OL1M"}`))
	}))
	defer server.Close()
	hc := NewHardcover(func() (string, error) { return "synthetic-secret", nil })
	hc.endpoint = server.URL
	hc.http = server.Client()
	hc.pace.interval = 0
	ol := NewOpenLibrary()
	ol.endpoint = server.URL
	ol.http = server.Client()
	ol.pace.interval = 0
	h := NewHealth(hc, ol)
	defer h.Close()
	h.Check()
	h.Check()
	h.wg.Wait()
	state := h.State()
	if primaryCalls.Load() != 1 || secondaryCalls.Load() != 1 || state.Checking {
		t.Fatal("comprobaciones duplicadas")
	}
	for _, item := range state.Providers {
		if item.State != "healthy" || item.CheckedAt == "" {
			t.Fatalf("estado incorrecto: %+v", item)
		}
	}
	encoded, _ := json.Marshal(state)
	if strings.Contains(string(encoded), "synthetic-secret") {
		t.Fatal("se filtró clave")
	}
	h.Check()
	if primaryCalls.Load() != 1 || secondaryCalls.Load() != 1 {
		t.Fatal("ignoró espera entre comprobaciones")
	}
}

func TestMissingKeyAndInvalidProbePayloadAreNotReportedHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{}`)) }))
	defer server.Close()
	hc := NewHardcover(func() (string, error) { return "", nil })
	ol := NewOpenLibrary()
	ol.endpoint = server.URL
	ol.http = server.Client()
	ol.pace.interval = 0
	h := NewHealth(hc, ol)
	defer h.Close()
	h.Check()
	h.wg.Wait()
	state := h.State()
	if state.Providers[0].State != "unconfigured" || state.Providers[0].CheckedAt != "" || state.Providers[1].State != "degraded" {
		t.Fatalf("estado engañoso: %+v", state)
	}
}

func TestPublicProviderAccessRejectionRemainsRecoverable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(403) }))
	defer server.Close()
	p := NewOpenLibrary()
	p.endpoint = server.URL
	p.http = server.Client()
	p.pace.interval = 0
	p.health = newMonitor("openlibrary", "Open Library", true)
	var data any
	_ = p.get(context.Background(), "/test", &data)
	if p.health.snapshot().State != "degraded" {
		t.Fatal("API pública quedó bloqueada a la espera de una clave inexistente")
	}
}

func TestStartupCheckDoesNotWaitForNetworkAndShutdownCancelsIt(t *testing.T) {
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-r.Context().Done()
	}))
	defer server.Close()
	hc := NewHardcover(func() (string, error) { return "", nil })
	ol := NewOpenLibrary()
	ol.endpoint = server.URL
	ol.http = server.Client()
	ol.pace.interval = 0
	h := NewHealth(hc, ol)
	state := h.Check()
	<-entered
	if !state.Checking {
		t.Error("comprobación no se inició en segundo plano")
	}
	h.Close()
	if state = h.Check(); state.Checking || state.Providers[1].State != "unchecked" {
		t.Fatal("cierre contó cancelación como caída o reinició consultas")
	}
}
