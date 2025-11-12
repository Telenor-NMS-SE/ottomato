package manager

import (
	"errors"
	"fmt"
)

type Worker interface {
	GetID() string
	Unload(uint32) ([]string, error)
	Load(Workload) error
}

var ErrWorkerExists = errors.New("worker already exists")

func (m *Manager) Workers() []Worker {
	m.workersMu.RLock()
	defer m.workersMu.RUnlock()

	workers := make([]Worker, 0, len(m.workers))
	for k, w := range m.workers {
		fmt.Printf("key: %s, value: %s\n", k, w.GetID())

		workers = append(workers, w)
	}

	return workers
}

func (m *Manager) AddWorker(w Worker) error {
	m.workersMu.Lock()
	defer m.workersMu.Unlock()

	if _, ok := m.workers[w.GetID()]; ok {
		return ErrWorkerExists
	}

	m.workers[w.GetID()] = w

	if m.eventCh != nil {
		m.eventCh <- NewWorkerAddedEvent(m.id, w)
	}

	return nil
}

func (m *Manager) DeleteWorker(w Worker) {
	m.workersMu.Lock()
	defer m.workersMu.Unlock()

	m.distributionsMu.Lock()
	defer m.distributionsMu.Unlock()

	delete(m.workers, w.GetID())

	for wl, wo := range m.distributions {
		if w.GetID() == wo {
			delete(m.distributions, wl)
		}
	}

	if m.eventCh != nil {
		m.eventCh <- NewWorkerDeletedEvent(m.id, w)
	}
}
