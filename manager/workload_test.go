package manager

import (
	"errors"
	"testing"
)

type MockWorkload struct {
	id    string
	state State
}

func (wl *MockWorkload) GetID() string {
	return wl.id
}

func (wl *MockWorkload) SetState(s State) {
	wl.state = s
}

func (wl *MockWorkload) GetState() State {
	return wl.state
}

func TestAddWorkload(t *testing.T) {
	manager := Manager{
		workloads: map[string]Workload{},
	}
	workload := MockWorkload{id: "test"}

	if err := manager.AddWorkload(&workload); err != nil {
		t.Fatalf("unexpected error when adding workload: %v", err)
	}

	if len(manager.workloads) != 1 {
		t.Errorf("expected exactly 1 workload, but got: %d", len(manager.workloads))
	}

	if _, ok := manager.workloads[workload.GetID()]; !ok {
		t.Errorf("expected to find workload '%s', but din't", workload.GetID())
	}
}

func TestAddDuplicateWorkload(t *testing.T) {
	manager := Manager{
		workloads: map[string]Workload{
			"test": &MockWorkload{id: "test"},
		},
	}

	err := manager.AddWorkload(&MockWorkload{id: "test"})
	if err == nil {
		t.Fatalf("expected an error when adding a duplicate workload, but got none")
	}

	if !errors.Is(err, ErrWorkloadExists) {
		t.Errorf("expected to get error '%v', but got: %v", ErrWorkloadExists, err)
	}
}

func TestDeleteWorkload(t *testing.T) {
	manager := Manager{
		workloads: map[string]Workload{
			"test": &MockWorkload{id: "test"},
		},
	}

	manager.DeleteWorkload(&MockWorkload{id: "test"})
	if len(manager.workers) > 0 {
		t.Fatalf("expected workload count to be exactly 0, but got: %d", len(manager.workers))
	}
}
