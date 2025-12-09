package manager

import (
	"context"
	"testing"
	"time"
)

type MockWorkload struct {
	id           string
	status       Status
	statusChange time.Time
}

func (wl *MockWorkload) GetID() string {
	return wl.id
}

func (wl *MockWorkload) SetStatus(s Status) {
	wl.statusChange = time.Now()
	wl.status = s
}

func (wl *MockWorkload) GetStatus() Status {
	return wl.status
}

func (wl *MockWorkload) LastStatusChange() time.Time {
	return wl.statusChange
}

func TestAddWorkload(t *testing.T) {
	state := NewMemoryStore()
	manager := Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &MockSignaller{},
	}
	workload := MockWorkload{id: "test"}

	manager.AddWorkload(context.TODO(), &workload)

	if len(state.workloads) != 1 {
		t.Errorf("expected exactly 1 workload, but got: %d", len(state.workloads))
	}

	if _, ok := state.workloads[workload.GetID()]; !ok {
		t.Errorf("expected to find workload '%s', but din't", workload.GetID())
	}
}

func TestGetWorkload(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{
			"test": &MockWorkload{id: "test"},
		},
	}
	manager := Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &MockSignaller{},
	}

	wl, err := manager.GetWorkload(context.TODO(), "test")
	if err != nil {
		t.Fatalf("unexpected error when getting workload: %v", err)
	}

	if wl == nil {
		t.Fatalf("unexpected nil pointer when getting workload")
	}

	if wl.GetID() != "test" {
		t.Errorf("expected to get workload 'test', but got: %s", wl.GetID())
	}
}

/*
func TestAddDuplicateWorkload(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{
			"test": &MockWorkload{id: "test"},
		},
	}
	manager := Manager{
		state: state,
		ctx:   context.TODO(),
		signal: &MockSignaller{},
	}

	err := manager.AddWorkload(&MockWorkload{id: "test"})
	if err == nil {
		t.Fatalf("expected an error when adding a duplicate workload, but got none")
	}

	if !errors.Is(err, ErrWorkloadExists) {
		t.Errorf("expected to get error '%v', but got: %v", ErrWorkloadExists, err)
	}
}
*/

func TestDeleteWorkload(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{
			"test": &MockWorkload{id: "test"},
		},
	}
	manager := Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &MockSignaller{},
	}

	manager.DeleteWorkload(context.TODO(), &MockWorkload{id: "test"})
	if len(state.workers) > 0 {
		t.Fatalf("expected workload count to be exactly 0, but got: %d", len(state.workers))
	}
}

func TestGetWorkloads(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{
			"test": &MockWorkload{id: "test"},
		},
	}
	manager := Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &MockSignaller{},
	}

	workloads, err := manager.Workloads(context.TODO())
	if err != nil {
		t.Fatalf("unexpected error when getting workloads: %v", err)
	}

	if len(workloads) != 1 {
		t.Fatalf("expected to get a slice of workers with a length of 1, but got: %d", len(workloads))
	}

	if workloads[0] == nil {
		t.Fatalf("expected the one worker to be a pointer, but got <nil>")
	}

	if workloads[0].GetID() != "test" {
		t.Fatalf("expected the one worker to be 'test', but got: %s", workloads[0].GetID())
	}
}
