package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/Santiparra/Trove/internal/backup"
	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/catalog"
	"github.com/Santiparra/Trove/internal/config"
	"github.com/Santiparra/Trove/internal/credentials"
	"github.com/Santiparra/Trove/internal/desktop"
	"github.com/Santiparra/Trove/internal/providers"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

// Overridden by the test build, without assigning a public application version.
var defaultProfile = "development"
var buildID = "development"
var version = "development"
var revision = "unknown"

func main() {
	profile := flag.String("profile", defaultProfile, "development, test o habitual")
	root := flag.String("data-dir", "", "raíz de datos; se añade siempre el nombre de perfil")
	showVersion := flag.Bool("version", false, "mostrar identidad de la compilación sin abrir la biblioteca")
	flag.Parse()
	if *showVersion {
		fmt.Printf("Trove %s; revision=%s; profile=%s; platform=%s/%s; build=%s\n", version, revision, *profile, runtime.GOOS, runtime.GOARCH, buildID)
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Resolve(*profile, *root, cwd, exe)
	if err != nil {
		log.Fatal(err)
	}
	store, err := sqlite.Open(filepath.Join(cfg.DataDir, "trove.sqlite"))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()
	backups := backup.New(store, cfg.DataDir, cfg.Profile)
	defer backups.Close()
	store.SetOnChange(backups.MarkDirty)
	library := desktop.NewLibrary(books.NewService(store), cfg.Profile, cfg.DataDir, buildID, desktop.BuildInfo{Version: version, Revision: revision})
	vault := credentials.New(cfg.DataDir)
	hardcover, openlibrary := providers.NewHardcover(vault.Get), providers.NewOpenLibrary()
	health := providers.NewHealth(hardcover, openlibrary)
	defer health.Close()
	health.Check()
	catalogService := desktop.NewCatalog(catalog.New(hardcover, openlibrary), vault, providers.NewCovers(filepath.Join(cfg.DataDir, "catalog-covers")), library, health)
	app := application.New(application.Options{
		Name: "Trove", Description: "Biblioteca personal multimedia local",
		Windows:  application.WindowsOptions{WebviewUserDataPath: filepath.Join(cfg.DataDir, "webview")},
		Services: []application.Service{application.NewService(library), application.NewService(catalogService)},
		Assets:   application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
	})
	app.RegisterService(application.NewService(desktop.NewBackups(backups, library,
		func() (string, error) {
			return app.Dialog.OpenFile().SetTitle("Elegir carpeta raíz de respaldos de Trove").CanChooseDirectories(true).CanChooseFiles(false).CanCreateDirectories(true).PromptForSingleSelection()
		},
		func() (string, error) {
			dialog := app.Dialog.OpenFile().SetTitle("Restaurar un respaldo de Trove").CanChooseFiles(true).CanChooseDirectories(false).AddFilter("Respaldos Trove", "*.sqlite;*.json").AddFilter("Respaldo SQLite", "*.sqlite").AddFilter("Respaldo JSON", "*.json")
			state := backups.State()
			if state.Folder != "" {
				dir := state.Folder
				if state.Generation != "" {
					dir = filepath.Join(dir, "generation-"+state.Generation)
				}
				dialog.SetDirectory(dir)
			}
			return dialog.PromptForSingleSelection()
		},
	)))
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Trove · " + cfg.Profile, Width: 1250, Height: 860, MinWidth: 960, MinHeight: 650,
		BackgroundColour: application.NewRGB(19, 23, 29), URL: "/",
	})
	if err = app.Run(); err != nil {
		log.Print(err)
	}
}
