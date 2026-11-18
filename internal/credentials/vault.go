// Package credentials keeps catalog keys outside library data and portable backups.
package credentials

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Vault struct {
	path string
	mu   sync.RWMutex
}

func New(dir string) *Vault { return &Vault{path: filepath.Join(dir, "catalog-credentials.bin")} }
func normalize(value string) (string, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		value = strings.TrimSpace(value[7:])
	}
	if len(value) < 8 || len(value) > 12000 || strings.ContainsAny(value, " \r\n\t\x00") {
		return "", errors.New("la clave tiene un formato inválido")
	}
	return value, nil
}
func (v *Vault) Get() (string, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	data, err := os.ReadFile(v.path)
	if err != nil {
		return "", errors.New("no hay una clave de catálogo disponible")
	}
	plain, err := unprotect(data)
	if err != nil {
		return "", errors.New("no se pudo desbloquear la clave en esta cuenta Windows; volvé a configurarla")
	}
	defer func() {
		for i := range plain {
			plain[i] = 0
		}
	}()
	return normalize(string(plain))
}
func (v *Vault) Save(value string) error {
	value, err := normalize(value)
	if err != nil {
		return err
	}
	data, err := protect([]byte(value))
	if err != nil {
		return errors.New("Windows no pudo proteger la clave")
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if err = os.MkdirAll(filepath.Dir(v.path), 0700); err != nil {
		return errors.New("no se pudo preparar la configuración de catálogo")
	}
	f, err := os.CreateTemp(filepath.Dir(v.path), ".catalog-key-")
	if err != nil {
		return errors.New("no se pudo preparar la clave")
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	if err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(name, v.path)
	}
	if err != nil {
		return errors.New("no se pudo guardar la clave protegida")
	}
	return nil
}
func (v *Vault) Clear() error {
	v.mu.Lock()
	defer v.mu.Unlock()
	err := os.Remove(v.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return errors.New("no se pudo quitar la clave")
	}
	return nil
}
