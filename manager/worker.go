package manager

import (
	"context"
	"errors"
)

type Worker interface {
	GetID() string
	Unload(Workload) error
	Load(Workload) error
}

var ErrWorkerExists = errors.New("worker already exists")

func (m *Manager) GetAssociation(ctx context.Context, wl Workload) (Worker, error) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetAssociation(ctx, wl)
}

func (m *Manager) GetAssosiactions(ctx context.Context, w Worker) ([]Workload, error) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetAssociations(ctx, w)
}

func (m *Manager) Associate(ctx context.Context, wl Workload, w Worker) error {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.Associate(ctx, wl, w)
}

func (m *Manager) Workers(ctx context.Context) ([]Worker, error) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetAllWorkers(ctx)
}

func (m *Manager) GetWorker(ctx context.Context, id string) (Worker, error) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetWorker(ctx, id)
}

func (m *Manager) AddWorker(ctx context.Context, w Worker) error {
	m.state.Lock()
	defer m.state.Unlock()

	if err := m.state.AddWorker(ctx, w); err != nil {
		return err
	}

	m.signal.Event(NewWorkerAddedEvent(m.id, w))
	return nil
}

func (m *Manager) DeleteWorker(ctx context.Context, w Worker) error {
	m.state.Lock()
	defer m.state.Unlock()

	assocs, err := m.state.GetAssociations(ctx, w)
	if err != nil {
		return err
	}

	for _, wl := range assocs {
		if err := m.state.Disassociate(ctx, wl, w); err != nil {
			return err
		}
		wl.SetStatus(StatusInit)
		if err := m.state.UpdateWorkload(ctx, wl); err != nil {
			return err
		}
	}

	if err := m.state.DeleteWorker(ctx, w); err != nil {
		return err
	}

	m.signal.Event(NewWorkerDeletedEvent(m.id, w))
	return nil
}
