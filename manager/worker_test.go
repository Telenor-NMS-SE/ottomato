package manager

import (
	"context"
	"testing"
)

type MockWorker struct {
	id  string
	mgr *Manager
}

func (w *MockWorker) GetID() string {
	return w.id
}

func (w *MockWorker) Unload(wl Workload) error {
	return nil
}

func (w *MockWorker) Load(wl Workload) error {
	return nil
}

func TestAddWorker(t *testing.T) {
	state := NewMemoryStore()
	manager := Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &MockSignaller{},
	}
	worker := MockWorker{id: "test"}

	if err := manager.AddWorker(context.TODO(), &worker); err != nil {
		t.Fatalf("unexpected error when adding worker: %v", err)
	}

	if len(state.workers) != 1 {
		t.Errorf("expected manager to have exactly 1 worker, but got: %d", len(state.workers))
	}

	if _, ok := state.workers[worker.GetID()]; !ok {
		t.Errorf("expected manager to have stored worker '%s', but it wasn't found.", worker.GetID())
	}
}

func TestGetWorker(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"test": &MockWorker{id: "test"},
		},
	}
	manager := Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &MockSignaller{},
	}

	w, err := manager.GetWorker(context.TODO(), "test")
	if err != nil {
		t.Fatalf("unexpected error when getting worker: %v", err)
	}

	if w == nil {
		t.Fatalf("unexpected nil pointer when getting worker")
	}

	if w.GetID() != "test" {
		t.Errorf("expected to get worker 'test', but got: %s", w.GetID())
	}
}

/*
func TestAddDuplicateWorker(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"test": &MockWorker{id: "test"},
		},
	}
	manager := Manager{
	}
	worker := MockWorker{id: "test"}

	manager.AddWorker(&worker)

	if !errors.Is(err, ErrWorkerExists) {
		t.Fatalf("expected to get '%v' error, but got: %v", ErrWorkerExists, err)
	}
}
*/

func TestDeleteWorker(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"test": &MockWorker{id: "test"},
		},
		workloads: map[string]Workload{
			"test": &MockWorkload{id: "test"},
		},
		associations: map[string]string{
			"test": "test",
		},
	}
	manager := Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &MockSignaller{},
	}

	if err := manager.DeleteWorker(context.TODO(), &MockWorker{id: "test"}); err != nil {
		t.Fatalf("unexpected error when deleting worker: %v", err)
	}

	if len(state.workers) > 0 {
		t.Fatalf("expected worker count to be exactly 0, but got: %d", len(state.workers))
	}

	if len(state.associations) > 0 {
		t.Fatalf("expected workload to be removed from the worker in distributions, but found %d entries", len(state.associations))
	}
}

func TestGetWorkers(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"test": &MockWorker{id: "test"},
		},
	}
	manager := Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &MockSignaller{},
	}

	workers, err := manager.Workers(context.TODO())
	if err != nil {
		t.Fatalf("unexpected error when getting workers: %v", err)
	}

	if len(workers) != 1 {
		t.Fatalf("expected to get a slice of workers with a length of 1, but got: %d", len(workers))
	}

	if workers[0] == nil {
		t.Fatalf("expected the one worker to be a pointer, but got <nil>")
	}

	if workers[0].GetID() != "test" {
		t.Fatalf("expected the one worker to be 'test', but got: %s", workers[0].GetID())
	}
}
