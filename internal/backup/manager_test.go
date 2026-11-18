package backup_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
)

type failingCapture struct {
	backup.Backend
	fail atomic.Bool
}

func (f *failingCapture) Capture(ctx context.Context, p, g, c string) (backup.Document, error) {
	if f.fail.Load() {
		return backup.Document{}, errors.New("fallo sintético de disco")
	}
	return f.Backend.Capture(ctx, p, g, c)
}

func generations(t *testing.T, folder string) []string {
	t.Helper()
	entries, err := os.ReadDir(folder)
	if err != nil {
		t.Fatal(err)
	}
	result := []string{}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "generation-") {
			result = append(result, filepath.Join(folder, e.Name()))
		}
	}
	return result
}

func TestRestoreChoicesSelectExactGenerationAndRevalidateBeforePreparation(t *testing.T) {
	ctx := context.Background()
	base := t.TempDir()
	store, err := sqlite.Open(filepath.Join(base, "profile", "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	service := books.NewService(store)
	b, err := service.Save(ctx, books.Input{Title: "Sintético", Status: "to-read", Notes: "Anterior"})
	if err != nil {
		t.Fatal(err)
	}
	m := backup.New(store, filepath.Join(base, "profile"), "test")
	defer m.Close()
	state, err := m.Configure(ctx, filepath.Join(base, "backups"))
	if err != nil {
		t.Fatal(err)
	}
	previous := state.Generation
	if _, err = service.SaveNotes(ctx, b.ID, "Actual"); err != nil {
		t.Fatal(err)
	}
	m.MarkDirty()
	if err = m.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	current := m.State().Generation
	choices, err := m.RestoreChoices()
	if err != nil || len(choices) != 2 || choices[0].Label != "Actual" || choices[0].Generation != current || choices[1].Label != "Anterior" || choices[1].Generation != previous || choices[0].CreatedAt == "" || choices[1].CreatedAt == "" {
		t.Fatalf("opciones incorrectas: %+v %v", choices, err)
	}
	p, err := m.PreviewGeneration(ctx, previous)
	if err != nil || p.Generation != previous {
		t.Fatal("no preparó generación elegida")
	}
	stored, _ := service.Get(ctx, b.ID)
	if stored.Notes != "Actual" {
		t.Fatal("vista previa modificó biblioteca")
	}
	result, err := m.ConfirmRestore(ctx, p.Token)
	if err != nil || result.SafetyCopy == "" {
		t.Fatal("restauración sin copia previa")
	}
	stored, _ = service.Get(ctx, b.ID)
	if stored.Notes != "Anterior" {
		t.Fatal("restauró generación incorrecta")
	}
	if _, err = m.PreviewGeneration(ctx, "../unknown"); err == nil {
		t.Fatal("aceptó generación ajena")
	}
	if err = os.WriteFile(filepath.Join(state.Folder, "generation-"+current, "trove.sqlite"), []byte("corrupt"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.PreviewGeneration(ctx, current); err == nil {
		t.Fatal("no revalidó generación alterada")
	}
	stored, _ = service.Get(ctx, b.ID)
	if stored.Notes != "Anterior" {
		t.Fatal("fallo modificó biblioteca")
	}
}

func TestAutomaticGenerationsRetentionFailureAndRestoreBothFormats(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	data := filepath.Join(root, "data", "test")
	s, err := sqlite.Open(filepath.Join(data, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	backend := &failingCapture{Backend: s}
	m := backup.New(backend, data, "test")
	defer m.Close()
	s.SetOnChange(m.MarkDirty)
	if m.State().Phase != "unconfigured" {
		t.Fatal("debe pedir carpeta inicialmente")
	}
	service := books.NewService(s)
	b, err := service.Save(ctx, books.Input{Title: "Ejemplo", Status: "read", Favorite: true, Owned: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.Configure(ctx, filepath.Join(root, "data")); err == nil {
		t.Fatal("permitió respaldar sobre los datos activos")
	}
	state, err := m.Configure(ctx, filepath.Join(root, "cloud"))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(state.Folder) != "test" || state.Pending || state.Phase != "current" {
		t.Fatal("configuración/perfil incorrectos")
	}
	first := state.Generation
	if err = m.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if m.State().Generation != first || len(generations(t, state.Folder)) != 1 {
		t.Fatal("regeneró sin cambios")
	}
	for _, note := range []string{"Segunda", "Tercera"} {
		if _, err = service.SaveNotes(ctx, b.ID, note); err != nil {
			t.Fatal(err)
		}
		if err = m.Flush(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if len(generations(t, state.Folder)) != 2 {
		t.Fatal("no conserva exactamente dos generaciones completas")
	}
	valid := m.State().Generation
	backend.fail.Store(true)
	if _, err = service.SaveNotes(ctx, b.ID, "Local sin respaldo"); err != nil {
		t.Fatal(err)
	}
	if err = m.Flush(ctx); err == nil {
		t.Fatal("se esperaba fallo")
	}
	if !m.State().Pending || m.State().Phase != "error" || m.State().Generation != valid {
		t.Fatal("ocultó fallo o perdió última generación")
	}
	if current, err := service.Get(ctx, b.ID); err != nil || current.Notes != "Local sin respaldo" {
		t.Fatal("fallo de respaldo afectó guardado local")
	}
	backend.fail.Store(false)
	latest := filepath.Join(state.Folder, "generation-"+valid)
	for _, name := range []string{"trove.json", "trove.sqlite"} {
		preview, err := m.PreviewRestore(ctx, filepath.Join(latest, name))
		if err != nil {
			t.Fatal(err)
		}
		if preview.Works != 1 || preview.Profile != "test" {
			t.Fatal("vista previa incorrecta")
		}
		result, err := m.ConfirmRestore(ctx, preview.Token)
		if err != nil {
			t.Fatal(err)
		}
		if result.SafetyCopy == "" {
			t.Fatal("no creó resguardo antes de reemplazar")
		}
		current, err := service.Get(ctx, b.ID)
		if err != nil || current.Notes != "Tercera" || !current.Favorite || !current.Edition.Owned || current.Cycle.FinishedOn != "" {
			t.Fatal("restauración alteró datos")
		}
	}
}

func TestInvalidOrModifiedRestoreNeverTouchesActiveData(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	data := filepath.Join(root, "test")
	s, err := sqlite.Open(filepath.Join(data, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	m := backup.New(s, data, "test")
	defer m.Close()
	s.SetOnChange(m.MarkDirty)
	b, err := books.NewService(s).Save(ctx, books.Input{Title: "Original", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	state, err := m.Configure(ctx, filepath.Join(root, "cloud"))
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(state.Folder, "generation-"+state.Generation, "trove.json")
	preview, err := m.PreviewRestore(ctx, source)
	if err != nil {
		t.Fatal(err)
	}
	m.CancelRestore(preview.Token)
	if _, err = m.ConfirmRestore(ctx, preview.Token); err == nil {
		t.Fatal("confirmó una restauración cancelada")
	}
	invalid := filepath.Join(root, "invalid.json")
	if err = os.WriteFile(invalid, []byte(`{"format":"trove-library","version":999}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.PreviewRestore(ctx, invalid); err == nil {
		t.Fatal("aceptó versión desconocida")
	}
	preview, err = m.PreviewRestore(ctx, source)
	if err != nil {
		t.Fatal(err)
	}
	// Changing the chosen source after preview cannot change the frozen prepared copy.
	if err = os.WriteFile(source, []byte(`{"broken":true}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.ConfirmRestore(ctx, preview.Token); err != nil {
		t.Fatal(err)
	}
	if current, err := books.NewService(s).Get(ctx, b.ID); err != nil || current.Title != "Original" {
		t.Fatal("el archivo cambiado afectó la copia validada")
	}
}

func TestDebounceAndShutdownSavePendingChangesWithoutManualExport(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	data := filepath.Join(root, "data", "test")
	s, err := sqlite.Open(filepath.Join(data, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	m := backup.New(s, data, "test")
	defer m.Close()
	s.SetOnChange(m.MarkDirty)
	state, err := m.Configure(ctx, filepath.Join(root, "cloud"))
	if err != nil {
		t.Fatal(err)
	}
	initial := state.Generation
	b, err := books.NewService(s).Save(ctx, books.Input{Title: "Automático", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		state = m.State()
		if state.Phase == "current" && !state.Pending && state.Generation != initial {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if state.Phase != "current" || state.Pending || state.Generation == initial {
		t.Fatal("no se generó respaldo después de la espera de agrupación")
	}
	if _, err = books.NewService(s).SaveNotes(ctx, b.ID, "Al cerrar"); err != nil {
		t.Fatal(err)
	}
	if err = m.Close(); err != nil {
		t.Fatal(err)
	}
	state = m.State()
	doc, err := backup.ReadJSON(filepath.Join(state.Folder, "generation-"+state.Generation, "trove.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Catalog.Works) != 1 || doc.Catalog.Works[0].Notes != "Al cerrar" {
		t.Fatal("el cierre no completó cambios pendientes")
	}
	// Startup checks content, including changes that happened without a running scheduler.
	if _, err = books.NewService(s).SaveNotes(ctx, b.ID, "Cambio anterior al arranque"); err != nil {
		t.Fatal(err)
	}
	next := backup.New(s, data, "test")
	defer next.Close()
	s.SetOnChange(next.MarkDirty)
	if err = next.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if next.State().Generation == state.Generation {
		t.Fatal("no detectó cambios pendientes al reiniciar")
	}
}

func TestIncompleteGenerationsAreIgnoredAndPreparedTamperingIsRejected(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	dir := filepath.Join(root, "data", "test")
	s, err := sqlite.Open(filepath.Join(dir, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	m := backup.New(s, dir, "test")
	defer m.Close()
	s.SetOnChange(m.MarkDirty)
	b, err := books.NewService(s).Save(ctx, books.Input{Title: "Conservar", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	state, err := m.Configure(ctx, filepath.Join(root, "cloud"))
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(state.Folder, "generation-"+state.Generation, "trove.json")
	incomplete := filepath.Join(state.Folder, "generation-incomplete")
	if err = os.Mkdir(incomplete, 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(incomplete, "trove.json"), []byte("partial"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = books.NewService(s).SaveNotes(ctx, b.ID, "Cambio"); err != nil {
		t.Fatal(err)
	}
	if err = m.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = backup.ReadJSON(original); err != nil {
		t.Fatal("una carpeta incompleta sustituyó la copia anterior")
	}
	if _, err = os.Stat(incomplete); err != nil {
		t.Fatal("se borró una carpeta no reconocida como generación completa")
	}
	p, err := m.PreviewRestore(ctx, original)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := filepath.Glob(filepath.Join(dir, ".restore-*", "prepared.sqlite"))
	if err != nil || len(paths) != 1 {
		t.Fatal("falta copia preparada")
	}
	if err = os.WriteFile(paths[0], []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = m.ConfirmRestore(ctx, p.Token); err == nil {
		t.Fatal("instaló una copia preparada modificada")
	}
	if current, err := books.NewService(s).Get(ctx, b.ID); err != nil || current.Notes != "Cambio" {
		t.Fatal("el rechazo alteró la biblioteca activa")
	}
}

func TestRestoreKeepsTargetProfileBackupConfiguration(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	cloud := filepath.Join(root, "cloud")
	sourceDir := filepath.Join(root, "data", "test")
	targetDir := filepath.Join(root, "data", "habitual")
	a, err := sqlite.Open(filepath.Join(sourceDir, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := sqlite.Open(filepath.Join(targetDir, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	ma := backup.New(a, sourceDir, "test")
	defer ma.Close()
	a.SetOnChange(ma.MarkDirty)
	mb := backup.New(b, targetDir, "habitual")
	defer mb.Close()
	b.SetOnChange(mb.MarkDirty)
	if _, err = books.NewService(a).Save(ctx, books.Input{Title: "Copia explícita", Status: "to-read"}); err != nil {
		t.Fatal(err)
	}
	sa, err := ma.Configure(ctx, cloud)
	if err != nil {
		t.Fatal(err)
	}
	sb, err := mb.Configure(ctx, cloud)
	if err != nil {
		t.Fatal(err)
	}
	if sa.Folder == sb.Folder {
		t.Fatal("perfiles comparten salida de respaldo")
	}
	file := filepath.Join(sa.Folder, "generation-"+sa.Generation, "trove.json")
	p, err := mb.PreviewRestore(ctx, file)
	if err != nil {
		t.Fatal(err)
	}
	if p.Profile != "habitual" {
		t.Fatal("vista previa identifica perfil de origen como destino")
	}
	if _, err = mb.ConfirmRestore(ctx, p.Token); err != nil {
		t.Fatal(err)
	}
	if err = mb.Flush(ctx); err != nil {
		t.Fatal(err)
	}
	if mb.State().Folder != sb.Folder {
		t.Fatal("restaurar reemplazó la configuración local del perfil")
	}
	if _, err = backup.ReadJSON(file); err != nil {
		t.Fatal("respaldo del otro perfil alterado")
	}
}
