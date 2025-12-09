package manager

import (
	"context"
	"errors"
	"sync"
)

type MemoryStore struct {
	mu sync.Mutex

	workers      map[string]Worker
	workloads    map[string]Workload
	associations map[string]string
}

var (
	ErrWorkerNotFound     = errors.New("no such worker")
	ErrWorkloadNotFound   = errors.New("no such workload")
	ErrMissingAssociation = errors.New("missing association")
)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		workers:      map[string]Worker{},
		workloads:    map[string]Workload{},
		associations: map[string]string{},
	}
}

func (s *MemoryStore) Lock() {
	s.mu.Lock()
}

func (s *MemoryStore) Unlock() {
	s.mu.Unlock()
}

func (s *MemoryStore) GetAllWorkers(_ context.Context) ([]Worker, error) {
	workers := make([]Worker, 0, len(s.workers))
	for _, worker := range s.workers {
		workers = append(workers, worker)
	}
	return workers, nil
}

func (s *MemoryStore) GetWorker(_ context.Context, id string) (Worker, error) {
	worker, ok := s.workers[id]
	if !ok {
		return nil, ErrWorkerNotFound
	}
	return worker, nil
}

func (s *MemoryStore) AddWorker(_ context.Context, w Worker) error {
	s.workers[w.GetID()] = w
	return nil
}

func (s *MemoryStore) DeleteWorker(_ context.Context, w Worker) error {
	delete(s.workers, w.GetID())
	return nil
}

func (s *MemoryStore) GetAllWorkloads(_ context.Context) ([]Workload, error) {
	workloads := make([]Workload, 0, len(s.workloads))
	for _, workload := range s.workloads {
		workloads = append(workloads, workload)
	}
	return workloads, nil
}

func (s *MemoryStore) GetWorkload(_ context.Context, id string) (Workload, error) {
	wl, ok := s.workloads[id]
	if !ok {
		return nil, ErrWorkloadNotFound
	}

	return wl, nil
}

func (s *MemoryStore) AddWorkload(_ context.Context, wl Workload) error {
	s.workloads[wl.GetID()] = wl
	return nil
}

func (s *MemoryStore) UpdateWorkload(_ context.Context, wl Workload) error {
	s.workloads[wl.GetID()] = wl
	return nil
}

func (s *MemoryStore) DeleteWorkload(_ context.Context, wl Workload) error {
	delete(s.workloads, wl.GetID())
	return nil
}

func (s *MemoryStore) GetAssociations(_ context.Context, w Worker) ([]Workload, error) {
	workloads := []Workload{}
	for workloadId, workerId := range s.associations {
		if workerId != w.GetID() {
			continue
		}

		if wl, ok := s.workloads[workloadId]; ok {
			workloads = append(workloads, wl)
		}
	}
	return workloads, nil
}

func (s *MemoryStore) GetAssociation(_ context.Context, wl Workload) (Worker, error) {
	workerId, ok := s.associations[wl.GetID()]
	if !ok {
		return nil, ErrMissingAssociation
	}

	w, ok := s.workers[workerId]
	if !ok {
		return nil, ErrWorkerNotFound
	}

	return w, nil
}

func (s *MemoryStore) Associate(_ context.Context, wl Workload, w Worker) error {
	s.associations[wl.GetID()] = w.GetID()
	return nil
}

func (s *MemoryStore) Disassociate(_ context.Context, wl Workload, w Worker) error {
	delete(s.associations, wl.GetID())
	return nil
}
