package manager

import (
	"errors"
	"testing"
	"time"
)

type MockWorkload struct {
	id          string
	state       State
	stateChange time.Time
}

func (wl *MockWorkload) GetID() string {
	return wl.id
}

func (wl *MockWorkload) SetState(s State) {
	wl.stateChange = time.Now()
	wl.state = s
}

func (wl *MockWorkload) GetState() State {
	return wl.state
}

func (wl *MockWorkload) LastStateChange() time.Time {
	return wl.stateChange
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

func TestGetWorkloads(t *testing.T) {
	manager := Manager{
		workloads: map[string]Workload{
			"test": &MockWorkload{id: "test"},
		},
	}

	workloads := manager.Workloads()
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
