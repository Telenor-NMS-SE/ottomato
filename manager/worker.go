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

func (m *Manager) Workers() []Worker {
	m.workersMu.RLock()
	defer m.workersMu.RUnlock()

	workers := make([]Worker, 0, len(m.workers))
	for _, w := range m.workers {
		workers = append(workers, w)
	}

	return workers
}

func (m *Manager) GetWorker(id string) (Worker, bool) {
	m.workersMu.RLock()
	defer m.workersMu.RUnlock()

	w, ok := m.workers[id]
	return w, ok
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

	m.workloadsMu.Lock()
	defer m.workloadsMu.Unlock()

	for workloadId, workerId := range m.distributions {
		if w.GetID() == workerId {
			delete(m.distributions, workloadId)

			if wl, ok := m.workloads[workloadId]; ok {
				wl.SetState(StateInit)
			}
		}
	}

	delete(m.workers, w.GetID())

	if m.eventCh != nil {
		m.eventCh <- NewWorkerDeletedEvent(m.id, w)
	}
}
