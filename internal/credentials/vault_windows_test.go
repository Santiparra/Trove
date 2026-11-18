//go:build windows

package credentials

import (
	"bytes"
	"os"
	"testing"
)

func TestVaultProtectsKeyAndAllowsReconfiguration(t *testing.T) {
	v := New(t.TempDir())
	secret := "synthetic-catalog-key-not-a-real-token"
	if err := v.Save("Bearer " + secret); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(v.path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte(secret)) {
		t.Fatal("clave guardada en claro")
	}
	if got, err := v.Get(); err != nil || got != secret {
		t.Fatal("la clave protegida no se recuperó")
	}
	if err = v.Save("invalid token with spaces"); err == nil {
		t.Fatal("formato inválido aceptado")
	}
	if got, err := v.Get(); err != nil || got != secret {
		t.Fatal("una clave inválida sustituyó la anterior")
	}
	if err = v.Clear(); err != nil {
		t.Fatal(err)
	}
	if _, err = v.Get(); err == nil {
		t.Fatal("no se quitó la clave")
	}
}
