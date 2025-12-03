package manager

import (
	"errors"
)

type Worker interface {
	GetID() string
	Unload(Workload) error
	Load(Workload) error
}

var ErrWorkerExists = errors.New("worker already exists")

func (m *Manager) GetAssociation(wl Workload) (Worker, bool) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetAssociation(wl)
}

func (m *Manager) GetAssosiactions(w Worker) []Workload {
	m.state.Lock()
	defer m.state.Unlock()

	return nil
}

func (m *Manager) Associate(wl Workload, w Worker) {
	m.state.Lock()
	defer m.state.Unlock()

	m.state.Associate(wl, w)
}

func (m *Manager) Workers() []Worker {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetAllWorkers()
}

func (m *Manager) GetWorker(id string) (Worker, bool) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetWorker(id)
}

func (m *Manager) AddWorker(w Worker) {
	m.state.Lock()
	defer m.state.Unlock()

	m.state.AddWorker(w)

	if m.eventCh != nil {
		m.eventCh <- NewWorkerAddedEvent(m.id, w)
	}
}

func (m *Manager) DeleteWorker(w Worker) {
	m.state.Lock()
	defer m.state.Unlock()

	for _, wl := range m.state.GetAssociations(w) {
		m.state.Disassociate(wl, w)
		wl.SetStatus(StatusInit)
		m.state.UpdateWorkload(wl)
	}

	m.state.DeleteWorker(w)

	if m.eventCh != nil {
		m.eventCh <- NewWorkerDeletedEvent(m.id, w)
	}
}

