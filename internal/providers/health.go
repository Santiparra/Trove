package providers

import (
	"context"
	"errors"
	"sync"
	"time"
)

const failureLimit = 3
const recoveryDelay = time.Minute
const checkInterval = 15 * time.Second

// ProviderStatus describes catalog API calls, not the provider website or image CDN.
type ProviderStatus struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	State     string `json:"state"`
	Detail    string `json:"detail"`
	CheckedAt string `json:"checkedAt"`
	LatencyMS int64  `json:"latencyMS"`
	Failures  int    `json:"failures"`
	RetryAt   string `json:"retryAt"`
	Checking  bool   `json:"checking"`
}

type HealthReport struct {
	LocalOnly   bool             `json:"localOnly"`
	Providers   []ProviderStatus `json:"providers"`
	Checking    bool             `json:"checking"`
	NextCheckAt string           `json:"nextCheckAt"`
}

type providerError struct {
	kind, message string
	retry         time.Duration
}

func (e *providerError) Error() string   { return e.message }
func failure(kind, message string) error { return &providerError{kind: kind, message: message} }

type monitor struct {
	mu        sync.Mutex
	status    ProviderStatus
	clock     func() time.Time
	until     time.Time
	epoch     uint64
	sequence  uint64
	completed uint64
	inflight  int
	trial     bool
}

func newMonitor(id, name string, configured bool) *monitor {
	m := &monitor{status: ProviderStatus{ID: id, Name: name}, clock: time.Now}
	m.reset(configured)
	return m
}
func (m *monitor) reset(configured bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.epoch++
	m.until = time.Time{}
	m.trial = false
	m.inflight = 0
	m.completed = m.sequence
	m.status = ProviderStatus{ID: m.status.ID, Name: m.status.Name, State: "unchecked"}
	if !configured {
		m.status.State = "unconfigured"
		m.status.Detail = "Falta configurar la clave de catálogo."
	}
}
func (m *monitor) snapshot() ProviderStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	status := m.status
	status.Checking = m.inflight > 0
	return status
}
func (m *monitor) run(ctx context.Context, wait func(context.Context) error, call func() error) error {
	m.mu.Lock()
	if m.status.State == "unconfigured" {
		m.mu.Unlock()
		return failure("unconfigured", "Configurá la clave de catálogo para consultar este proveedor.")
	}
	if m.status.State == "access-denied" {
		m.mu.Unlock()
		return failure("access", "El proveedor rechazó el acceso; revisá o reemplazá la clave en Claves.")
	}
	if !m.until.IsZero() {
		if m.clock().Before(m.until) || m.trial || m.inflight > 0 {
			m.mu.Unlock()
			return failure("paused", "Consultas a este proveedor en pausa temporal; los datos locales siguen disponibles.")
		}
		m.trial = true
	}
	m.sequence++
	sequence, epoch := m.sequence, m.epoch
	recovering := m.trial
	m.inflight++
	m.mu.Unlock()
	if err := wait(ctx); err != nil {
		m.finish(epoch, sequence, 0, err, false)
		return errors.New("consulta de catálogo cancelada")
	}
	if err := ctx.Err(); err != nil {
		m.finish(epoch, sequence, 0, err, false)
		return err
	}
	// A queued request may have been admitted before another request opened the circuit.
	m.mu.Lock()
	if epoch != m.epoch {
		m.mu.Unlock()
		return errors.New("el acceso al proveedor cambió; reintentá la consulta")
	}
	if m.status.State == "access-denied" || m.status.State == "unconfigured" || (!m.until.IsZero() && !recovering) {
		m.inflight--
		m.mu.Unlock()
		return failure("paused", "Consultas a este proveedor en pausa; los datos locales siguen disponibles.")
	}
	m.mu.Unlock()
	started := m.clock()
	err := call()
	m.finish(epoch, sequence, m.clock().Sub(started), err, ctx.Err() != context.Canceled)
	return err
}
func (m *monitor) finish(epoch, sequence uint64, duration time.Duration, err error, observe bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if epoch != m.epoch {
		return
	}
	m.inflight--
	if !observe {
		m.trial = false
		return
	}
	if isLocalOnly(err) {
		m.trial = false
		return
	}
	// Requests that started before an opened circuit cannot reopen it on completion.
	if sequence < m.completed || (!m.until.IsZero() && !m.trial) || m.status.State == "access-denied" || m.status.State == "unconfigured" {
		return
	}
	m.completed = sequence
	m.trial = false
	now := m.clock()
	m.status.CheckedAt = now.UTC().Format(time.RFC3339Nano)
	m.status.LatencyMS = duration.Milliseconds()
	var problem *providerError
	if err == nil || (errors.As(err, &problem) && (problem.kind == "missing" || problem.kind == "request")) {
		m.status.State, m.status.Detail = "healthy", "La API respondió correctamente."
		m.status.Failures = 0
		m.status.RetryAt = ""
		m.until = time.Time{}
		return
	}
	if problem != nil && problem.kind == "unconfigured" {
		m.status.State, m.status.Detail = "unconfigured", "Falta configurar la clave de catálogo."
		return
	}
	if problem != nil && problem.kind == "access" {
		m.status.State, m.status.Detail = "access-denied", "Acceso rechazado. Revisá la clave y sus permisos."
		m.status.RetryAt = ""
		m.until = time.Time{}
		return
	}
	m.status.Failures++
	m.status.State, m.status.Detail = "degraded", "La consulta falló; los datos locales se conservan."
	if problem != nil && problem.kind == "response" {
		m.status.Detail = "La API devolvió una respuesta incompatible."
	}
	if problem != nil && problem.kind == "network" {
		m.status.Detail = "No se pudo conectar o se agotó el tiempo de espera."
	}
	if m.status.Failures >= failureLimit || !m.until.IsZero() || (problem != nil && problem.kind == "rate-limit") {
		delay := recoveryDelay
		if problem != nil && problem.retry > delay {
			delay = problem.retry
		}
		m.until = now.Add(delay)
		m.status.State = "paused"
		m.status.RetryAt = m.until.UTC().Format(time.RFC3339Nano)
		if problem != nil && problem.kind == "rate-limit" {
			m.status.Detail = "Límite de solicitudes alcanzado. Se respetará la espera del proveedor."
		}
	}
}

// Health owns only bounded, explicit checks; it never periodically polls providers.
type Health struct {
	network     *NetworkPolicy
	hardcover   *Hardcover
	openlibrary *OpenLibrary
	mu          sync.Mutex
	checking    bool
	next        time.Time
	closed      bool
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

func NewHealth(hardcover *Hardcover, openlibrary *OpenLibrary, network *NetworkPolicy) *Health {
	ctx, cancel := context.WithCancel(context.Background())
	h := &Health{hardcover: hardcover, openlibrary: openlibrary, network: network, ctx: ctx, cancel: cancel}
	if network != nil {
		network.protect(hardcover.http)
		network.protect(openlibrary.http)
	}
	token, err := hardcover.token()
	hardcover.health = newMonitor("hardcover", "Hardcover", err == nil && token != "")
	openlibrary.health = newMonitor("openlibrary", "Open Library", true)
	return h
}
func (h *Health) State() HealthReport {
	h.mu.Lock()
	defer h.mu.Unlock()
	report := HealthReport{Providers: []ProviderStatus{h.hardcover.health.snapshot(), h.openlibrary.health.snapshot()}, Checking: h.checking}
	if h.network != nil {
		report.LocalOnly = h.network.LocalOnly()
	}
	if !h.next.IsZero() {
		report.NextCheckAt = h.next.UTC().Format(time.RFC3339Nano)
	}
	return report
}

// Check starts at most one pair of lightweight probes; cooldowns also apply to checks.
func (h *Health) Check() HealthReport {
	h.mu.Lock()
	if h.closed || h.checking || time.Now().Before(h.next) || (h.network != nil && h.network.LocalOnly()) {
		h.mu.Unlock()
		return h.State()
	}
	h.checking = true
	h.next = time.Now().Add(checkInterval)
	h.wg.Add(1)
	h.mu.Unlock()
	go func() {
		defer h.wg.Done()
		ctx, cancel := context.WithTimeout(h.ctx, 8*time.Second)
		defer cancel()
		var probes sync.WaitGroup
		for _, check := range []func(context.Context) error{h.hardcover.check, h.openlibrary.check} {
			probes.Add(1)
			go func(probe func(context.Context) error) { defer probes.Done(); _ = probe(ctx) }(check)
		}
		probes.Wait()
		h.mu.Lock()
		h.checking = false
		h.mu.Unlock()
	}()
	return h.State()
}
func (h *Health) CredentialsChanged() {
	token, err := h.hardcover.token()
	h.hardcover.health.reset(err == nil && token != "")
}
func (h *Health) SetLocalOnly(value bool) (HealthReport, error) {
	if h.network == nil {
		return h.State(), errors.New("modo de red no disponible")
	}
	if err := h.network.SetLocalOnly(value); err != nil {
		return h.State(), err
	}
	return h.State(), nil
}
func (h *Health) Close() {
	h.mu.Lock()
	h.closed = true
	h.cancel()
	h.mu.Unlock()
	h.wg.Wait()
}

func (p *Hardcover) check(ctx context.Context) error {
	var data struct {
		Books []struct {
			ID int64 `json:"id"`
		} `json:"books"`
	}
	return p.health.run(ctx, p.pace.wait, func() error {
		if err := p.queryRaw(ctx, `query TroveHealth { books(limit:1) { id } }`, nil, &data); err != nil {
			return err
		}
		if len(data.Books) != 1 || data.Books[0].ID <= 0 {
			return failure("response", "Hardcover no devolvió una respuesta de catálogo reconocible")
		}
		return nil
	})
}
func (p *OpenLibrary) check(ctx context.Context) error {
	var data struct {
		Key string `json:"key"`
	}
	return p.health.run(ctx, p.pace.wait, func() error {
		if err := p.getRaw(ctx, "/isbn/9780140328721.json", &data); err != nil {
			var problem *providerError
			if errors.As(err, &problem) && problem.kind == "missing" {
				return failure("response", "Open Library no devolvió el registro de comprobación")
			}
			return err
		}
		if !editionKey.MatchString(data.Key) {
			return failure("response", "Open Library no devolvió una respuesta de catálogo reconocible")
		}
		return nil
	})
}
