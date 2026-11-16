package config

import (
	"path/filepath"
	"testing"
)

func TestProfilesCannotShareDataByDefaultOrWithCustomRoot(t *testing.T) {
	root := t.TempDir()
	paths := map[string]bool{}
	for _, profile := range []string{"development", "test", "habitual"} {
		cfg, err := Resolve(profile, root, root, filepath.Join(root, "Trove.exe"))
		if err != nil {
			t.Fatal(err)
		}
		if paths[cfg.DataDir] {
			t.Fatal("perfiles comparten directorio")
		}
		paths[cfg.DataDir] = true
		if filepath.Base(cfg.DataDir) != profile {
			t.Fatal("no aplica separación al root configurable")
		}
	}
	if _, err := Resolve("../habitual", root, root, root); err == nil {
		t.Fatal("perfil inválido aceptado")
	}
	testCfg, _ := Resolve("test", "", root, filepath.Join(root, "portable", "Trove.exe"))
	if testCfg.DataDir != filepath.Join(root, "portable", "data", "test") {
		t.Fatal("build de prueba no resuelve datos portables")
	}
}
