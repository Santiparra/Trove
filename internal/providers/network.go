package providers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// NetworkPolicy is shared by catalog clients and cover downloads, never by local cache reads.
type NetworkPolicy struct {
	mu        sync.Mutex
	path      string
	localOnly bool
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewNetworkPolicy(dir string) (*NetworkPolicy, error) {
	if dir == "" {
		return nil, errors.New("falta la carpeta del perfil")
	}
	p := &NetworkPolicy{path: filepath.Join(dir, "provider-network.json")}
	data, err := os.ReadFile(p.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("no se pudo leer el modo de red del perfil")
	}
	if err == nil {
		var settings struct {
			LocalOnly bool `json:"localOnly"`
		}
		if json.Unmarshal(data, &settings) != nil {
			return nil, errors.New("configuración de red del perfil inválida")
		}
		p.localOnly = settings.LocalOnly
	}
	p.ctx, p.cancel = context.WithCancel(context.Background())
	if p.localOnly {
		p.cancel()
	}
	return p, nil
}
func (p *NetworkPolicy) LocalOnly() bool { p.mu.Lock(); defer p.mu.Unlock(); return p.localOnly }
func (p *NetworkPolicy) SetLocalOnly(value bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if value == p.localOnly {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p.path), 0700); err != nil {
		return errors.New("no se pudo guardar el modo de red")
	}
	file, err := os.CreateTemp(filepath.Dir(p.path), ".provider-network-*")
	if err != nil {
		return errors.New("no se pudo guardar el modo de red")
	}
	name := file.Name()
	defer os.Remove(name)
	err = json.NewEncoder(file).Encode(struct {
		LocalOnly bool `json:"localOnly"`
	}{value})
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return errors.New("no se pudo guardar el modo de red")
	}
	if err = os.Rename(name, p.path); err != nil {
		return errors.New("no se pudo guardar el modo de red")
	}
	p.localOnly = value
	if value {
		p.cancel()
	} else {
		p.ctx, p.cancel = context.WithCancel(context.Background())
	}
	return nil
}
func (p *NetworkPolicy) protect(client *http.Client) {
	next := client.Transport
	if next == nil {
		next = http.DefaultTransport
	}
	client.Transport = &localTransport{policy: p, next: next}
}

type localTransport struct {
	policy *NetworkPolicy
	next   http.RoundTripper
}

func localOnlyError() error {
	return failure("local-only", "Modo solo local activo: las consultas y descargas de proveedores están desactivadas.")
}
func isLocalOnly(err error) bool {
	var problem *providerError
	return errors.As(err, &problem) && problem.kind == "local-only"
}
func (t *localTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.policy.mu.Lock()
	if t.policy.localOnly {
		t.policy.mu.Unlock()
		return nil, localOnlyError()
	}
	epoch := t.policy.ctx
	ctx, cancel := context.WithCancel(req.Context())
	stop := context.AfterFunc(epoch, cancel)
	t.policy.mu.Unlock()
	response, err := t.next.RoundTrip(req.WithContext(ctx))
	if err != nil {
		stop()
		cancel()
		if epoch.Err() != nil {
			return nil, localOnlyError()
		}
		return nil, err
	}
	response.Body = &localBody{ReadCloser: response.Body, epoch: epoch, done: func() { stop(); cancel() }}
	return response, nil
}

type localBody struct {
	io.ReadCloser
	epoch context.Context
	done  func()
}

func (b *localBody) Read(data []byte) (int, error) {
	n, err := b.ReadCloser.Read(data)
	if err != nil && b.epoch.Err() != nil {
		return n, localOnlyError()
	}
	return n, err
}
func (b *localBody) Close() error { err := b.ReadCloser.Close(); b.done(); return err }
