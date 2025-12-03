package manager

import (
	"errors"
	"time"
)

type Workload interface {
	GetID() string
	GetStatus() Status
	SetStatus(Status)
	LastStatusChange() time.Time
}

var ErrWorkloadExists = errors.New("workload already exists")

func (m *Manager) Workloads() []Workload {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetAllWorkloads()
}

func (m *Manager) GetWorkload(id string) (Workload, bool) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetWorkload(id)
}

func (m *Manager) AddWorkload(wl Workload) {
	m.state.Lock()
	defer m.state.Unlock()

	m.state.AddWorkload(wl)

	if m.eventCh != nil {
		m.eventCh <- NewWorkloadAddedEvent(m.id, wl)
	}
}

func (m *Manager) DeleteWorkload(wl Workload) {
	m.state.Lock()
	defer m.state.Unlock()

	if w, ok := m.state.GetAssociation(wl); ok {
		m.state.Disassociate(wl, w)

	}

	m.state.DeleteWorkload(wl)

	if m.eventCh != nil {
		m.eventCh <- NewWorkloadDeletedEvent(m.id, wl)
	}
}
