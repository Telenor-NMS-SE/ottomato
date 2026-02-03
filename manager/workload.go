package manager

import (
	"context"
	"errors"
	"time"
)

type Workload interface {
	GetID() string
	GetStatus() Status
	SetStatus(Status)
	LastStatusChange() time.Time
}

type workload struct {
	id     string
	status Status
	change time.Time
}

func (wl *workload) GetID() string {
	return wl.id
}

func (wl *workload) GetStatus() Status {
	return wl.status
}

func (wl *workload) SetStatus(s Status) {
	wl.status = s
	wl.change = time.Now()
}

func (wl *workload) LastStatusChange() time.Time {
	return wl.change
}

var ErrWorkloadExists = errors.New("workload already exists")

func (m *Manager) Workloads(ctx context.Context) ([]Workload, error) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetAllWorkloads(ctx)
}

func (m *Manager) GetWorkload(ctx context.Context, id string) (Workload, error) {
	m.state.Lock()
	defer m.state.Unlock()

	return m.state.GetWorkload(ctx, id)
}

func (m *Manager) AddWorkload(ctx context.Context, wl Workload) error {
	m.state.Lock()
	defer m.state.Unlock()

	if err := m.state.AddWorkload(ctx, wl); err != nil {
		return err
	}

	m.signal.Event(NewWorkloadAddedEvent(m.id, wl))
	return nil
}

func (m *Manager) DeleteWorkload(ctx context.Context, wl Workload) error {
	m.state.Lock()
	defer m.state.Unlock()

	w, err := m.state.GetAssociation(ctx, wl)
	if err != nil {
		return err
	}

	if err := m.state.Disassociate(ctx, wl, w); err != nil {
		return err
	}

	if err := m.state.DeleteWorkload(ctx, wl); err != nil {
		return err
	}

	m.signal.Event(NewWorkloadDeletedEvent(m.id, wl))
	return nil
}
