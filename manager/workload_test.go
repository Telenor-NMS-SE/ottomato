package manager

import (
	"testing"
	"time"
)

type MockWorkload struct {
	id          string
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
		state: state,
	}
	workload := MockWorkload{id: "test"}

	manager.AddWorkload(&workload)

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
		state: state,
	}

	wl, ok := manager.GetWorkload("test")
	if !ok {
		t.Fatalf("expected to get a workload, but didn't")
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
		state: state,
	}

	manager.DeleteWorkload(&MockWorkload{id: "test"})
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
		state: state,
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
