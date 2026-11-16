package main

import (
	"embed"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/config"
	"github.com/Santiparra/Trove/internal/desktop"
	"github.com/Santiparra/Trove/internal/storage/sqlite"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

// Overridden by the test build, without assigning a public application version.
var defaultProfile = "development"
var buildID = "development"

func main() {
	profile := flag.String("profile", defaultProfile, "development, test o habitual")
	root := flag.String("data-dir", "", "raíz de datos; se añade siempre el nombre de perfil")
	flag.Parse()
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
	app := application.New(application.Options{
		Name: "Trove", Description: "Biblioteca personal multimedia local",
		Windows:  application.WindowsOptions{WebviewUserDataPath: filepath.Join(cfg.DataDir, "webview")},
		Services: []application.Service{application.NewService(desktop.NewLibrary(books.NewService(store), cfg.Profile, cfg.DataDir, buildID))},
		Assets:   application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Trove · " + cfg.Profile, Width: 1250, Height: 860, MinWidth: 960, MinHeight: 650,
		BackgroundColour: application.NewRGB(19, 23, 29), URL: "/",
	})
	if err = app.Run(); err != nil {
		log.Print(err)
	}
}
