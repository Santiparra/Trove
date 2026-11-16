package config

import (
	"errors"
	"path/filepath"
)

type Config struct {
	Profile string
	DataDir string
}

// Resolve always namespaces a custom root by profile, including restored test copies.
func Resolve(profile, root, workingDir, executable string) (Config, error) {
	switch profile {
	case "development", "test", "habitual":
	default:
		return Config{}, errors.New("perfil inválido: usar development, test o habitual")
	}
	if root == "" {
		if profile == "development" {
			root = filepath.Join(workingDir, ".local")
		} else {
			root = filepath.Join(filepath.Dir(executable), "data")
		}
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return Config{}, err
	}
	return Config{Profile: profile, DataDir: filepath.Join(root, profile)}, nil
}
