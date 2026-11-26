package desktop

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Santiparra/Trove/internal/books"
	"github.com/Santiparra/Trove/internal/novelcatalog"
	"github.com/Santiparra/Trove/internal/novels"
	"github.com/Santiparra/Trove/internal/providers"
)

// Novels serializes aggregate writes with restore and shutdown through the shared gate.
type Novels struct {
	service      *novels.Service
	library      *Library
	catalog      *novelcatalog.Service
	network      *providers.NetworkPolicy
	mu           sync.Mutex
	previews     map[string]novelSelection
	searchCancel map[string]context.CancelFunc
}

func NewNovels(service *novels.Service, library *Library, catalog *novelcatalog.Service, network *providers.NetworkPolicy) *Novels {
	return &Novels{service: service, library: library, catalog: catalog, network: network, previews: map[string]novelSelection{}, searchCancel: map[string]context.CancelFunc{}}
}

type novelSelection struct {
	series      novelcatalog.Series
	publication novelcatalog.Publication
	at          time.Time
}

func (n *Novels) List(query, status string) ([]novels.Novel, error) {
	n.library.gate.RLock()
	defer n.library.gate.RUnlock()
	return n.service.List(context.Background(), query, status)
}
func (n *Novels) Get(id string) (novels.Novel, error) {
	n.library.gate.RLock()
	defer n.library.gate.RUnlock()
	return n.service.Get(context.Background(), id)
}
func (n *Novels) Dashboard() (novels.Dashboard, error) {
	n.library.gate.RLock()
	defer n.library.gate.RUnlock()
	return n.service.Dashboard(context.Background())
}
func (n *Novels) SetCompleted(id string, completed int, markRead bool) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) {
		return n.service.SetCompleted(context.Background(), id, completed, markRead)
	})
}
func (n *Novels) mutate(action func() (novels.Novel, error)) (novels.Novel, error) {
	n.library.gate.Lock()
	defer n.library.gate.Unlock()
	if n.library.closing {
		return novels.Novel{}, errors.New("Trove se está cerrando")
	}
	return action()
}
func (n *Novels) Save(input novels.Input) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.Save(context.Background(), input) })
}
func (n *Novels) ChangeStatus(id, status string) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.ChangeStatus(context.Background(), id, status) })
}
func (n *Novels) StartVolume(id string) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.Start(context.Background(), id) })
}
func (n *Novels) FinishVolume(id string) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.Finish(context.Background(), id) })
}
func (n *Novels) SaveVolume(id string, input novels.VolumeInput) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.SaveVolume(context.Background(), id, input) })
}
func (n *Novels) RereadNovel(id string, volumeIDs []string) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.RereadNovel(context.Background(), id, volumeIDs) })
}
func (n *Novels) RereadVolume(id, volumeID string) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.RereadVolume(context.Background(), id, volumeID) })
}
func (n *Novels) SaveReading(id, volumeID string, input books.ReadingInput) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.SaveReading(context.Background(), id, volumeID, input) })
}
func (n *Novels) SaveRun(id string, input books.ReadingInput) (novels.Novel, error) {
	return n.mutate(func() (novels.Novel, error) { return n.service.SaveRun(context.Background(), id, input) })
}
func (n *Novels) Delete(id string) error {
	_, err := n.mutate(func() (novels.Novel, error) { return novels.Novel{}, n.service.Delete(context.Background(), id) })
	return err
}
