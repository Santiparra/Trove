package backup

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Backend interface {
	Capture(context.Context, string, string, string) (Document, error)
	ReadNative(context.Context, string) (Document, error)
	Prepare(context.Context, Document, string) error
	Install(context.Context, string) (string, error)
}

type Status struct {
	Profile     string `json:"profile"`
	Folder      string `json:"folder"`
	Phase       string `json:"phase"`
	Pending     bool   `json:"pending"`
	LastSavedAt string `json:"lastSavedAt"`
	Generation  string `json:"generation"`
	LastError   string `json:"lastError"`
}
type Preview struct {
	Token      string `json:"token"`
	Source     string `json:"source"`
	Kind       string `json:"kind"`
	CreatedAt  string `json:"createdAt"`
	Generation string `json:"generation"`
	Profile    string `json:"profile"`
	DataDir    string `json:"dataDir"`
	Works      int    `json:"works"`
	Editions   int    `json:"editions"`
	Cycles     int    `json:"cycles"`
}
type RestoreResult struct {
	SafetyCopy string `json:"safetyCopy"`
	Works      int    `json:"works"`
}
type RestoreChoice struct {
	Generation string `json:"generation"`
	Label      string `json:"label"`
	CreatedAt  string `json:"createdAt"`
}

func (m *Manager) restoreChoices() ([]generation, error) {
	folder := m.State().Folder
	if folder == "" {
		return []generation{}, nil
	}
	items, err := completeGenerations(folder)
	if errors.Is(err, os.ErrNotExist) {
		return []generation{}, nil
	}
	if err != nil {
		return nil, errors.New("no se pudieron leer las generaciones de respaldo")
	}
	if len(items) > 2 {
		items = items[:2]
	}
	return items, nil
}
func (m *Manager) RestoreChoices() ([]RestoreChoice, error) {
	m.op.Lock()
	defer m.op.Unlock()
	items, err := m.restoreChoices()
	if err != nil {
		return nil, err
	}
	choices := make([]RestoreChoice, 0, len(items))
	for i, item := range items {
		label := "Actual"
		if i == 1 {
			label = "Anterior"
		}
		choices = append(choices, RestoreChoice{Generation: item.Generation, Label: label, CreatedAt: item.CreatedAt})
	}
	return choices, nil
}
func (m *Manager) PreviewGeneration(ctx context.Context, id string) (Preview, error) {
	m.op.Lock()
	defer m.op.Unlock()
	items, err := m.restoreChoices()
	if err != nil {
		return Preview{}, err
	}
	for _, item := range items {
		if item.Generation == id {
			return m.previewRestore(ctx, filepath.Join(item.path, "trove.sqlite"), &item.manifest)
		}
	}
	return Preview{}, errors.New("el respaldo seleccionado ya no está disponible o no pasó la validación; actualizá la selección")
}

type manifest struct {
	Format      string `json:"format"`
	Version     int    `json:"version"`
	Generation  string `json:"generation"`
	CreatedAt   string `json:"createdAt"`
	ContentHash string `json:"contentHash"`
	SQLiteHash  string `json:"sqliteHash"`
	JSONHash    string `json:"jsonHash"`
}
type generation struct {
	manifest
	path    string
	created time.Time
}
type restoreJob struct {
	preview         Preview
	dir, path, hash string
	expires         time.Time
}
type settings struct {
	Version int    `json:"version"`
	Root    string `json:"root"`
}

type Manager struct {
	backend          Backend
	dir, profile     string
	mu               sync.Mutex
	op               sync.Mutex
	status           Status
	root             string
	epoch            uint64
	wake, stop, done chan struct{}
	retry            chan struct{}
	closeOnce        sync.Once
	closeErr         error
	prepared         *restoreJob
	delay            time.Duration
}

func New(backend Backend, dir, profile string) *Manager {
	m := &Manager{backend: backend, dir: dir, profile: profile, status: Status{Profile: profile, Phase: "unconfigured"}, wake: make(chan struct{}, 1), retry: make(chan struct{}, 1), stop: make(chan struct{}), done: make(chan struct{}), delay: 2 * time.Second}
	data, err := os.ReadFile(filepath.Join(dir, "backup-settings.json"))
	if err == nil {
		var config settings
		if err = json.Unmarshal(data, &config); err == nil && config.Version == 1 && config.Root != "" {
			m.root = config.Root
			m.status.Folder, err = m.resolveFolder(config.Root)
			if err != nil {
				if abs, e := filepath.Abs(config.Root); e == nil {
					m.status.Folder = filepath.Join(abs, profile)
				}
			}
		} else {
			err = errors.New("configuración de respaldo inválida; elegí de nuevo la carpeta")
		}
		if err != nil {
			m.status.Phase = "error"
			m.status.LastError = err.Error()
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		m.status.Phase = "error"
		m.status.LastError = err.Error()
	}
	go m.run()
	if m.status.Folder != "" {
		m.MarkDirty()
	}
	return m
}

func (m *Manager) State() Status { m.mu.Lock(); defer m.mu.Unlock(); return m.status }
func (m *Manager) MarkDirty() {
	m.mu.Lock()
	m.epoch++
	m.status.Pending = true
	if m.status.Folder != "" && m.status.Phase != "saving" {
		m.status.Phase = "pending"
	}
	m.mu.Unlock()
	select {
	case m.wake <- struct{}{}:
	default:
	}
}

func (m *Manager) run() {
	defer close(m.done)
	var timer *time.Timer
	var tick <-chan time.Time
	reset := func(delay time.Duration) {
		if timer == nil {
			timer = time.NewTimer(delay)
		} else {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(delay)
		}
		tick = timer.C
	}
	for {
		select {
		case <-m.wake:
			reset(m.delay)
		case <-m.retry:
			reset(30 * time.Second)
		case <-tick:
			tick = nil
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := m.Flush(ctx)
			cancel()
			if err != nil {
				reset(30 * time.Second)
			}
		case <-m.stop:
			if timer != nil {
				timer.Stop()
			}
			state := m.State()
			if state.Folder != "" && state.Pending {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				m.closeErr = m.Flush(ctx)
				cancel()
			}
			m.op.Lock()
			m.clearPrepared()
			m.op.Unlock()
			return
		}
	}
}
func (m *Manager) Close() error {
	m.closeOnce.Do(func() { close(m.stop) })
	<-m.done
	return m.closeErr
}

func containsPath(root, path string) bool {
	rel, err := filepath.Rel(strings.ToLower(root), strings.ToLower(path))
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func (m *Manager) resolveFolder(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("elegí una carpeta de respaldo")
	}
	if m.profile != "development" && m.profile != "test" && m.profile != "habitual" {
		return "", errors.New("perfil de respaldo inválido")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	dest := filepath.Join(abs, m.profile)
	active, err := filepath.Abs(m.dir)
	if err != nil {
		return "", err
	}
	if containsPath(active, dest) || containsPath(dest, active) {
		return "", errors.New("elegí una carpeta de respaldo separada de los datos activos")
	}
	if err = os.MkdirAll(dest, 0700); err != nil {
		return "", err
	}
	dest, err = filepath.EvalSymlinks(dest)
	if err != nil {
		return "", err
	}
	active, err = filepath.EvalSymlinks(active)
	if err != nil {
		return "", err
	}
	if containsPath(active, dest) || containsPath(dest, active) {
		return "", errors.New("la carpeta de respaldo enlaza con los datos activos")
	}
	return dest, nil
}

func (m *Manager) Configure(ctx context.Context, root string) (Status, error) {
	m.op.Lock()
	dest, err := m.resolveFolder(root)
	if err == nil {
		tmp := filepath.Join(m.dir, ".backup-settings-"+rand.Text()+".json")
		err = WriteJSON(tmp, settings{Version: 1, Root: root})
		if err == nil {
			err = os.Rename(tmp, filepath.Join(m.dir, "backup-settings.json"))
		}
		_ = os.Remove(tmp)
	}
	if err == nil {
		m.mu.Lock()
		m.status.Folder = dest
		m.root = root
		m.status.LastError = ""
		m.status.LastSavedAt = ""
		m.status.Generation = ""
		m.mu.Unlock()
		m.MarkDirty()
	}
	m.op.Unlock()
	if err == nil {
		err = m.Flush(ctx)
	}
	return m.State(), err
}

func fileHash(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() > MaxFileSize {
		return "", errors.New("archivo de respaldo inválido o demasiado grande")
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, io.LimitReader(f, MaxFileSize+1)); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func completeGenerations(folder string) ([]generation, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}
	result := make([]generation, 0)
	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 || !strings.HasPrefix(entry.Name(), "generation-") {
			continue
		}
		path := filepath.Join(folder, entry.Name())
		info, err := os.Stat(filepath.Join(path, "complete.json"))
		if err != nil || info.Size() > 65536 {
			continue
		}
		data, err := os.ReadFile(filepath.Join(path, "complete.json"))
		if err != nil {
			continue
		}
		var v manifest
		if json.Unmarshal(data, &v) != nil || v.Format != "trove-generation" || v.Version != 1 || entry.Name() != "generation-"+v.Generation || !generationPattern.MatchString(v.Generation) {
			continue
		}
		created, err := time.Parse(time.RFC3339Nano, v.CreatedAt)
		if err != nil {
			continue
		}
		native, err := fileHash(filepath.Join(path, "trove.sqlite"))
		if err != nil || native != v.SQLiteHash {
			continue
		}
		logical, err := fileHash(filepath.Join(path, "trove.json"))
		if err != nil || logical != v.JSONHash {
			continue
		}
		doc, err := ReadJSON(filepath.Join(path, "trove.json"))
		if err != nil || doc.Generation != v.Generation || doc.CreatedAt != v.CreatedAt || doc.ContentHash() != v.ContentHash {
			continue
		}
		result = append(result, generation{manifest: v, path: path, created: created})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].created.After(result[j].created) })
	return result, nil
}

func (m *Manager) Flush(ctx context.Context) (err error) {
	m.op.Lock()
	defer m.op.Unlock()
	m.mu.Lock()
	root, epoch := m.root, m.epoch
	if root == "" {
		m.mu.Unlock()
		return nil
	}
	m.status.Phase = "saving"
	m.mu.Unlock()
	defer func() {
		if err != nil {
			m.mu.Lock()
			m.status.Phase = "error"
			m.status.Pending = true
			m.status.LastError = err.Error()
			m.mu.Unlock()
			select {
			case m.retry <- struct{}{}:
			default:
			}
		}
	}()
	folder, err := m.resolveFolder(root)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.status.Folder = folder
	m.mu.Unlock()
	previous, err := completeGenerations(folder)
	if err != nil {
		return err
	}
	id := rand.Text()
	created := time.Now().UTC().Format(time.RFC3339Nano)
	staging, err := os.MkdirTemp(folder, ".pending-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	doc, err := m.backend.Capture(ctx, filepath.Join(staging, "trove.sqlite"), id, created)
	if err != nil {
		return err
	}
	if err = doc.Validate(); err != nil {
		return err
	}
	if doc.Generation != id || doc.CreatedAt != created {
		return errors.New("las representaciones no identifican la misma generación")
	}
	hash := doc.ContentHash()
	var last manifest
	if len(previous) > 0 && previous[0].ContentHash == hash {
		last = previous[0].manifest
		if len(previous) > 2 {
			for _, old := range previous[2:] {
				if err = os.RemoveAll(old.path); err != nil {
					return fmt.Errorf("no se pudo retirar una generación anterior: %w", err)
				}
			}
		}
	} else {
		if err = WriteJSON(filepath.Join(staging, "trove.json"), doc); err != nil {
			return err
		}
		native, err := fileHash(filepath.Join(staging, "trove.sqlite"))
		if err != nil {
			return err
		}
		logical, err := fileHash(filepath.Join(staging, "trove.json"))
		if err != nil {
			return err
		}
		last = manifest{Format: "trove-generation", Version: 1, Generation: id, CreatedAt: created, ContentHash: hash, SQLiteHash: native, JSONHash: logical}
		if err = WriteJSON(filepath.Join(staging, "complete.json"), last); err != nil {
			return err
		}
		if err = os.Rename(staging, filepath.Join(folder, "generation-"+id)); err != nil {
			return err
		}
		// Remove only known, complete older generations, never unrelated/incomplete directories.
		if len(previous) > 1 {
			for _, old := range previous[1:] {
				if err = os.RemoveAll(old.path); err != nil {
					return fmt.Errorf("respaldo creado, pero no se pudo retirar una generación anterior: %w", err)
				}
			}
		}
	}
	m.mu.Lock()
	m.status.Pending = m.epoch != epoch
	m.status.Phase = "current"
	if m.status.Pending {
		m.status.Phase = "pending"
	}
	m.status.LastSavedAt = last.CreatedAt
	m.status.Generation = last.Generation
	m.status.LastError = ""
	m.mu.Unlock()
	return nil
}

func (m *Manager) clearPrepared() {
	if m.prepared != nil {
		_ = os.RemoveAll(m.prepared.dir)
		m.prepared = nil
	}
}

func (m *Manager) PreviewRestore(ctx context.Context, source string) (Preview, error) {
	m.op.Lock()
	defer m.op.Unlock()
	return m.previewRestore(ctx, source, nil)
}
func (m *Manager) previewRestore(ctx context.Context, source string, expected *manifest) (Preview, error) {
	m.clearPrepared()
	dir, err := os.MkdirTemp(m.dir, ".restore-")
	if err != nil {
		return Preview{}, err
	}
	ok := false
	defer func() {
		if !ok {
			os.RemoveAll(dir)
		}
	}()
	f, err := os.Open(source)
	if err != nil {
		return Preview{}, err
	}
	header := make([]byte, 16)
	_, _ = io.ReadFull(f, header)
	f.Close()
	var doc Document
	kind := "JSON"
	if string(header) == "SQLite format 3\x00" {
		kind = "SQLite"
		doc, err = m.backend.ReadNative(ctx, source)
	} else {
		doc, err = ReadJSON(source)
	}
	if err != nil {
		return Preview{}, err
	}
	if err = doc.Validate(); err != nil {
		return Preview{}, err
	}
	if expected != nil && (doc.Generation != expected.Generation || doc.CreatedAt != expected.CreatedAt || doc.ContentHash() != expected.ContentHash) {
		return Preview{}, errors.New("el respaldo seleccionado cambió; actualizá la selección")
	}
	path := filepath.Join(dir, "prepared.sqlite")
	if err = m.backend.Prepare(ctx, doc, path); err != nil {
		return Preview{}, err
	}
	hash, err := fileHash(path)
	if err != nil {
		return Preview{}, err
	}
	p := Preview{Token: rand.Text(), Source: source, Kind: kind, CreatedAt: doc.CreatedAt, Generation: doc.Generation, Profile: m.profile, DataDir: m.dir, Works: len(doc.Catalog.Works) + len(doc.Catalog.Novels), Editions: len(doc.Catalog.Editions), Cycles: len(doc.Catalog.Cycles)}
	for _, n := range doc.Catalog.Novels {
		p.Editions += len(n.Volumes)
		p.Cycles += len(n.Volumes)
	}
	m.prepared = &restoreJob{preview: p, dir: dir, path: path, hash: hash, expires: time.Now().Add(10 * time.Minute)}
	ok = true
	return p, nil
}

func (m *Manager) CancelRestore(token string) {
	m.op.Lock()
	defer m.op.Unlock()
	if m.prepared != nil && m.prepared.preview.Token == token {
		m.clearPrepared()
	}
}

func (m *Manager) ConfirmRestore(ctx context.Context, token string) (RestoreResult, error) {
	m.op.Lock()
	defer m.op.Unlock()
	job := m.prepared
	if job == nil || job.preview.Token != token {
		return RestoreResult{}, errors.New("no hay una restauración validada para confirmar")
	}
	defer m.clearPrepared()
	if time.Now().After(job.expires) {
		return RestoreResult{}, errors.New("la confirmación expiró; seleccioná de nuevo el archivo")
	}
	hash, err := fileHash(job.path)
	if err != nil || hash != job.hash {
		return RestoreResult{}, errors.New("la copia preparada cambió; seleccioná de nuevo el archivo")
	}
	safety, err := m.backend.Install(ctx, job.path)
	if err != nil {
		return RestoreResult{}, err
	}
	m.MarkDirty()
	return RestoreResult{SafetyCopy: safety, Works: job.preview.Works}, nil
}
