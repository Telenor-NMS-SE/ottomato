package manager

import (
	"errors"
	"fmt"
	"testing"
)

type MockWorker struct {
	id string
}

func (w *MockWorker) GetID() string {
	return w.id
}

func (w *MockWorker) Unload(amount uint32) ([]string, error) {
	unloaded := make([]string, 0, amount)

	for i := range amount {
		unloaded = append(unloaded, fmt.Sprintf("workload-%d", i))
	}

	return unloaded, nil
}

func TestAddWorker(t *testing.T) {
	manager := Manager{
		workers:  map[string]Worker{},
	}
	worker := MockWorker{"test"}

	if err := manager.AddWorker(&worker); err != nil {
		t.Fatalf("unexpected error when adding a worker: %v", err)
	}

	if len(manager.workers) != 1 {
		t.Errorf("expected manager to have exactly 1 worker, but got: %d", len(manager.workers))
	}

	if _, ok := manager.workers[worker.GetID()]; !ok {
		t.Errorf("expected manager to have stored worker '%s', but it wasn't found.", worker.GetID())
	}
}

func TestAddDuplicateWorker(t *testing.T) {
	manager := Manager{
		workers: map[string]Worker{
			"test": &MockWorker{id: "test"},
		},
	}
	worker := MockWorker{"test"}

	err := manager.AddWorker(&worker)
	if err == nil {
		t.Fatalf("expected an error when adding a duplicate worker, but got none")
	}

	if !errors.Is(err, ErrWorkerExists) {
		t.Fatalf("expected to get '%v' error, but got: %v", ErrWorkerExists, err)
	}
}

func TestDeleteWorker(t *testing.T) {
	manager := Manager{
		workers: map[string]Worker{
			"test": &MockWorker{id: "test"},
		},
	}

	manager.DeleteWorker(&MockWorker{"test"})
	if len(manager.workers) > 0 {
		t.Fatalf("expected worker count to be exactly 0, but got: %d", len(manager.workers))
	}
}