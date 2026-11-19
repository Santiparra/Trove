package desktop

import (
	"context"
	"errors"

	"github.com/Santiparra/Trove/internal/novels"
)

// Novels serializes aggregate writes with restore and shutdown through the shared gate.
type Novels struct {
	service *novels.Service
	library *Library
}

func NewNovels(service *novels.Service, library *Library) *Novels {
	return &Novels{service: service, library: library}
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
func (n *Novels) Delete(id string) error {
	_, err := n.mutate(func() (novels.Novel, error) { return novels.Novel{}, n.service.Delete(context.Background(), id) })
	return err
}
