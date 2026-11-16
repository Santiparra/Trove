package desktop

import (
	"context"
	"errors"
	"time"

	"github.com/Santiparra/Trove/internal/backup"
)

// Backups adapts native file pickers and the library gate; the recovery rules live in backup.
type Backups struct {
	manager      *backup.Manager
	library      *Library
	folderPicker func() (string, error)
	filePicker   func() (string, error)
}

func NewBackups(manager *backup.Manager, library *Library, folder, file func() (string, error)) *Backups {
	return &Backups{manager: manager, library: library, folderPicker: folder, filePicker: file}
}
func (b *Backups) State() backup.Status { return b.manager.State() }
func (b *Backups) ChooseFolder() (backup.Status, error) {
	path, err := b.folderPicker()
	if err != nil || path == "" {
		return b.manager.State(), err
	}
	b.library.gate.RLock()
	defer b.library.gate.RUnlock()
	if b.library.closing {
		return b.manager.State(), errors.New("Trove se está cerrando")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return b.manager.Configure(ctx, path)
}
func (b *Backups) BackupNow() (backup.Status, error) {
	b.library.gate.RLock()
	defer b.library.gate.RUnlock()
	if b.library.closing {
		return b.manager.State(), errors.New("Trove se está cerrando")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := b.manager.Flush(ctx)
	return b.manager.State(), err
}
func (b *Backups) ChooseRestore() (*backup.Preview, error) {
	path, err := b.filePicker()
	if err != nil || path == "" {
		return nil, err
	}
	b.library.gate.RLock()
	defer b.library.gate.RUnlock()
	if b.library.closing {
		return nil, errors.New("Trove se está cerrando")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p, err := b.manager.PreviewRestore(ctx, path)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
func (b *Backups) CancelRestore(token string) { b.manager.CancelRestore(token) }
func (b *Backups) ConfirmRestore(token string) (backup.RestoreResult, error) {
	b.library.gate.Lock()
	defer b.library.gate.Unlock()
	if b.library.closing {
		return backup.RestoreResult{}, errors.New("Trove se está cerrando")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return b.manager.ConfirmRestore(ctx, token)
}

// Wails calls this lifecycle hook before the store is closed by main.
func (b *Backups) ServiceShutdown() error {
	b.library.gate.Lock()
	defer b.library.gate.Unlock()
	b.library.closing = true
	return b.manager.Close()
}
