package manager

import (
	"errors"
	"fmt"
	"testing"
)

type MockWorker struct {
	id  string
	mgr *Manager
}

func (w *MockWorker) GetID() string {
	return w.id
}

func (w *MockWorker) Unload(amount uint32) ([]string, error) {
	unloaded := make([]string, 0, amount)

	if w.mgr == nil {
		for i := range amount {
			unloaded = append(unloaded, fmt.Sprintf("workload-%d", i))
		}

		return unloaded, nil
	}

	for workloadId, workerId := range w.mgr.distributions {
		if uint32(len(unloaded)) >= amount {
			break
		}

		if workerId == w.id {
			unloaded = append(unloaded, workloadId)
		}
	}

	return unloaded, nil
}

func (w *MockWorker) Load(wl Workload) error {
	return nil
}

func TestAddWorker(t *testing.T) {
	manager := Manager{
		workers: map[string]Worker{},
	}
	worker := MockWorker{id: "test"}

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
	worker := MockWorker{id: "test"}

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

	manager.DeleteWorker(&MockWorker{id: "test"})
	if len(manager.workers) > 0 {
		t.Fatalf("expected worker count to be exactly 0, but got: %d", len(manager.workers))
	}
}

func TestGetWorkers(t *testing.T) {
	manager := Manager{
		workers: map[string]Worker{
			"test": &MockWorker{id: "test"},
		},
	}

	workers := manager.Workers()
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
