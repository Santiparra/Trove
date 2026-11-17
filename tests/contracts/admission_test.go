// These are application admission contracts through the same Go adapters bound by Wails.
// They deliberately use no window, provider, personal export or repository discovery.
package contracts_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/config"
	"github.com/Santiparra/Trove/internal/desktop"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
)

type host struct {
	store   *sqlite.Store
	manager *backup.Manager
	library *desktop.Library
	dir     string
}

func start(t *testing.T, root, profile string) host {
	t.Helper()
	working := filepath.Join(root, "empty-working-directory")
	if err := os.MkdirAll(working, 0700); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Resolve(profile, filepath.Join(root, "data"), working, filepath.Join(root, "portable", "Trove.exe"))
	if err != nil {
		t.Fatal(err)
	}
	store, err := sqlite.Open(filepath.Join(cfg.DataDir, "trove.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	manager := backup.New(store, cfg.DataDir, profile)
	store.SetOnChange(manager.MarkDirty)
	t.Cleanup(func() {
		if err := manager.Close(); err != nil {
			t.Error(err)
		}
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	library := desktop.NewLibrary(books.NewService(store), profile, cfg.DataDir, "contract", desktop.BuildInfo{Version: "1.2.3", Revision: strings.Repeat("a", 40)})
	return host{store, manager, library, cfg.DataDir}
}

func TestAdmission_C01_EmptyDirectoryStartsOffline(t *testing.T) {
	h := start(t, t.TempDir(), "test")
	items, err := h.library.ListBooks("", "", books.MediumBooks)
	if err != nil || len(items) != 0 {
		t.Fatal("una biblioteca nueva debe iniciar vacía sin proveedor")
	}
	if h.manager.State().Phase != "unconfigured" {
		t.Fatal("no debe elegir una salida de respaldo por su cuenta")
	}
	if _, err = os.Stat(filepath.Join(h.dir, "trove.sqlite")); err != nil {
		t.Fatal(err)
	}
}

func TestAdmission_C02_ProfilesDoNotShareLibraryOrBackupOutput(t *testing.T) {
	root := t.TempDir()
	a := start(t, root, "test")
	b := start(t, root, "habitual")
	if _, err := a.library.SaveBook(books.Input{Title: "Solo prueba", Status: "to-read"}); err != nil {
		t.Fatal(err)
	}
	items, err := b.library.ListBooks("", "", books.MediumBooks)
	if err != nil || len(items) != 0 {
		t.Fatal("se filtraron datos entre perfiles")
	}
	cloud := filepath.Join(root, "backup-output")
	sa, err := a.manager.Configure(context.Background(), cloud)
	if err != nil {
		t.Fatal(err)
	}
	sb, err := b.manager.Configure(context.Background(), cloud)
	if err != nil {
		t.Fatal(err)
	}
	if sa.Folder == sb.Folder {
		t.Fatal("perfiles comparten salida")
	}
}

func TestAdmission_C03_WireDTOsPreserveMeaning(t *testing.T) {
	h := start(t, t.TempDir(), "test")
	var input books.Input
	if err := json.Unmarshal([]byte(`{"title":"Contrato sintético","authors":["Autora ficticia"],"status":"read","favorite":true,"owned":true,"notes":"Conservar","seriesName":"Serie ficticia","seriesPosition":1.5}`), &input); err != nil {
		t.Fatal(err)
	}
	book, err := h.library.SaveBook(input)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(book)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		ID     string `json:"id"`
		Medium string `json:"medium"`
		Series struct {
			Medium string `json:"medium"`
		} `json:"series"`
		Favorite bool `json:"favorite"`
		Edition  struct {
			Owned bool `json:"owned"`
		} `json:"edition"`
		Cycle struct {
			StartedOn  string `json:"startedOn"`
			FinishedOn string `json:"finishedOn"`
		} `json:"cycle"`
	}
	if err = json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	if wire.ID == "" || wire.Medium != books.MediumBooks || wire.Series.Medium != books.MediumBooks || !wire.Favorite || !wire.Edition.Owned || wire.Cycle.StartedOn != "" || wire.Cycle.FinishedOn != "" {
		t.Fatal("el contrato de transporte pierde datos o fabrica fechas")
	}
	info := h.library.Info()
	if info.Version != "1.2.3" || info.Revision != strings.Repeat("a", 40) || info.Profile != "test" {
		t.Fatal("identidad no corresponde al paquete")
	}
	got, err := h.library.GetBook(book.ID, books.MediumBooks)
	if err != nil || got.Series == nil || got.Series.ID == "" || *got.Series.Position != 1.5 {
		t.Fatal("vínculo/posición no recuperable desde el adaptador")
	}
}

func TestAdmission_C04_InvalidWriteDoesNotCreatePartialRecords(t *testing.T) {
	h := start(t, t.TempDir(), "test")
	if _, err := h.library.SaveBook(books.Input{Title: "", Status: "reading", SeriesName: "No crear"}); err == nil {
		t.Fatal("se admitió escritura inválida")
	}
	if _, err := h.library.SaveBook(books.Input{Title: "Fechas inválidas", Status: "read", StartedOn: "2026-02-01", FinishedOn: "2026-01-01"}); err == nil {
		t.Fatal("se admitió cronología inválida")
	}
	items, err := h.library.ListBooks("", "", books.MediumBooks)
	if err != nil || len(items) != 0 {
		t.Fatal("hubo persistencia parcial")
	}
	series, err := h.library.ListSeries(books.MediumBooks)
	if err != nil || len(series) != 0 {
		t.Fatal("se creó una serie de una escritura rechazada")
	}
}

func recoverRepresentation(t *testing.T, name string) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	source := start(t, filepath.Join(root, "source"), "test")
	b, err := source.library.SaveBook(books.Input{Title: "Rescate sintético", Status: "read", Favorite: true, Owned: true, Notes: "Historial desconocido"})
	if err != nil {
		t.Fatal(err)
	}
	state, err := source.manager.Configure(ctx, filepath.Join(root, "source-cloud"))
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(state.Folder, "generation-"+state.Generation, name)
	data, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	isolated := filepath.Join(root, "standalone")
	if err = os.MkdirAll(isolated, 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(isolated, name)
	if err = os.WriteFile(file, data, 0600); err != nil {
		t.Fatal(err)
	}
	target := start(t, filepath.Join(root, "target"), "habitual")
	previous, err := target.library.SaveBook(books.Input{Title: "Biblioteca anterior", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	adapter := desktop.NewBackups(target.manager, target.library, nil, func() (string, error) { return file, nil })
	p, err := adapter.ChooseRestore()
	if err != nil {
		t.Fatal(err)
	}
	if p == nil || p.Works != 1 || p.Profile != "habitual" {
		t.Fatal("la preparación no describe el destino")
	}
	if _, err = target.library.GetBook(previous.ID, books.MediumBooks); err != nil {
		t.Fatal("preparar ya sustituyó la biblioteca")
	}
	result, err := adapter.ConfirmRestore(p.Token)
	if err != nil {
		t.Fatal(err)
	}
	if result.SafetyCopy == "" {
		t.Fatal("falta copia anterior")
	}
	got, err := target.library.GetBook(b.ID, books.MediumBooks)
	if err != nil || !got.Favorite || !got.Edition.Owned || got.Cycle.FinishedOn != "" || got.Notes != b.Notes {
		t.Fatal("recuperación no equivalente")
	}
	if _, err = target.library.GetBook(previous.ID, books.MediumBooks); err == nil {
		t.Fatal("se mezclaron bibliotecas")
	}
	if target.manager.State().Folder != "" {
		t.Fatal("la recuperación importó rutas de otro perfil")
	}
}

func TestAdmission_C05_SQLiteRecoveryIsStandalone(t *testing.T) {
	recoverRepresentation(t, "trove.sqlite")
}
func TestAdmission_C06_JSONRecoveryIsStandalone(t *testing.T) { recoverRepresentation(t, "trove.json") }

func TestAdmission_C07_ExternalFileIsNotARecoveryBackup(t *testing.T) {
	root := t.TempDir()
	h := start(t, root, "test")
	b, err := h.library.SaveBook(books.Input{Title: "Conservar", Status: "to-read"})
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "external.csv")
	if err = os.WriteFile(file, []byte("Title,Authors,Read Status\nEjemplo,Autor,read\n"), 0600); err != nil {
		t.Fatal(err)
	}
	adapter := desktop.NewBackups(h.manager, h.library, nil, func() (string, error) { return file, nil })
	if _, err = adapter.ChooseRestore(); err == nil {
		t.Fatal("se aceptó una exportación externa como respaldo")
	}
	if _, err = adapter.ConfirmRestore("not-a-token"); err == nil {
		t.Fatal("se restauró sin preparación confirmable")
	}
	if _, err = h.library.GetBook(b.ID, books.MediumBooks); err != nil {
		t.Fatal("el rechazo alteró la biblioteca")
	}
}

func TestAdmission_C08_BackupConfigurationFailureDoesNotRejectLocalWork(t *testing.T) {
	root := t.TempDir()
	h := start(t, root, "test")
	blocked := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocked, []byte("block"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := h.manager.Configure(context.Background(), blocked); err == nil {
		t.Fatal("se aceptó un archivo como directorio")
	}
	book, err := h.library.SaveBook(books.Input{Title: "Guardado local", Status: "up-next"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = h.library.GetBook(book.ID, books.MediumBooks); err != nil {
		t.Fatal("el fallo de salida bloqueó la biblioteca")
	}
}

func TestAdmission_C09_InvalidProfileOrDataRootHasNoSilentFallback(t *testing.T) {
	root := t.TempDir()
	if _, err := config.Resolve("../habitual", root, root, filepath.Join(root, "Trove.exe")); err == nil {
		t.Fatal("perfil inválido admitido")
	}
	blocked := filepath.Join(root, "blocked")
	if err := os.WriteFile(blocked, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if s, err := sqlite.Open(filepath.Join(blocked, "trove.sqlite")); err == nil {
		s.Close()
		t.Fatal("se eligió otra ruta silenciosamente")
	}
	value, err := os.ReadFile(blocked)
	if err != nil || string(value) != "unchanged" {
		t.Fatal("se modificó el obstáculo o eligió otro destino")
	}
}
