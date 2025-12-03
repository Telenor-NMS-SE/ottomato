package manager

import (
	"sync"
)

type MemoryStore struct {
	mu sync.Mutex

	workers      map[string]Worker
	workloads    map[string]Workload
	associations map[string]string
}

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

func (s *MemoryStore) GetAllWorkers() []Worker {
	workers := make([]Worker, 0, len(s.workers))
	for _, worker := range s.workers {
		workers = append(workers, worker)
	}
	return workers
}

func (s *MemoryStore) GetWorker(id string) (Worker, bool) {
	worker, ok := s.workers[id]
	return worker, ok
}

func (s *MemoryStore) AddWorker(w Worker) {
	s.workers[w.GetID()] = w
}

func (s *MemoryStore) DeleteWorker(w Worker) {
	delete(s.workers, w.GetID())
}

func (s *MemoryStore) GetAllWorkloads() []Workload {
	workloads := make([]Workload, 0, len(s.workloads))
	for _, workload := range s.workloads {
		workloads = append(workloads, workload)
	}
	return workloads
}

func (s *MemoryStore) GetWorkload(id string) (Workload, bool) {
	wl, ok := s.workloads[id]
	return wl, ok
}

func (s *MemoryStore) AddWorkload(wl Workload) {
	s.workloads[wl.GetID()] = wl
}

func (s *MemoryStore) UpdateWorkload(wl Workload) {
	s.workloads[wl.GetID()] = wl
}

func (s *MemoryStore) DeleteWorkload(wl Workload) {
	delete(s.workloads, wl.GetID())
}

func (s *MemoryStore) GetAssociations(w Worker) []Workload {
	workloads := []Workload{}
	for workloadId, workerId := range s.associations {
		if workerId != w.GetID() {
			continue
		}

		if wl, ok := s.workloads[workloadId]; ok {
			workloads = append(workloads, wl)
		}
	}
	return workloads
}

func (s *MemoryStore) GetAssociation(wl Workload) (Worker, bool) {
	workerId, ok := s.associations[wl.GetID()]
	if !ok {
		return nil, false
	}

	w, ok := s.workers[workerId]
	return w, ok
}

func (s *MemoryStore) Associate(wl Workload, w Worker) {
	s.associations[wl.GetID()] = w.GetID()
}

func (s *MemoryStore) Disassociate(wl Workload, w Worker) {
	delete(s.associations, wl.GetID())
}
