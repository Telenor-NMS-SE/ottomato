package manager

import (
	"errors"
	"time"
)

type Workload interface {
	GetID() string
	GetState() State
	SetState(State)
	LastStateChange() time.Time
}

var ErrWorkloadExists = errors.New("workload already exists")

func (m *Manager) Workloads() []Workload {
	m.workloadsMu.RLock()
	defer m.workloadsMu.RUnlock()

	workloads := make([]Workload, 0, len(m.workloads))
	for _, wl := range m.workloads {
		workloads = append(workloads, wl)
	}

	return workloads
}

func (m *Manager) AddWorkload(wl Workload) error {
	m.workloadsMu.Lock()
	defer m.workloadsMu.Unlock()

	if _, ok := m.workloads[wl.GetID()]; ok {
		return ErrWorkloadExists
	}

	m.workloads[wl.GetID()] = wl

	if m.eventCh != nil {
		m.eventCh <- NewWorkloadAddedEvent(m.id, wl)
	}

	return nil
}

func (m *Manager) DeleteWorkload(wl Workload) {
	m.workloadsMu.Lock()
	defer m.workloadsMu.Unlock()

	m.distributionsMu.Lock()
	defer m.distributionsMu.Unlock()

	delete(m.workloads, wl.GetID())
	delete(m.distributions, wl.GetID())

	if m.eventCh != nil {
		m.eventCh <- NewWorkloadDeletedEvent(m.id, wl)
	}
}
