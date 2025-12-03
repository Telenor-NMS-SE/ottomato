package store

import (
	"fmt"
	"log/slog"
	"sync"
)

type WorkerStore struct {
	mu sync.RWMutex
	kv map[string]struct{}
}

func New() *WorkerStore {
	return &WorkerStore{
		kv: map[string]struct{}{},
	}
}

func (s *WorkerStore) RegisterWorker(workerId string) {}

func (s *WorkerStore) RegisterWorkload(workloadName string, workerId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s.%s", workerId, workloadName)

	if _, exists := s.kv[key]; exists {
		slog.Error("unable to register workload", "reason", "workload already exists")
	}

	s.kv[key] = struct{}{}
}

func (s *WorkerStore) DeleteWorkload(workloadName string, workerId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s.%s", workerId, workloadName)
	delete(s.kv, key)

}

func (s *WorkerStore) UpdateWorkload(workloadName string, workerId string) {}
