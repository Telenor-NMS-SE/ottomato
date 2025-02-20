package store

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Store struct {
	mu sync.RWMutex
	kv map[string]struct{}
}

func New(ctx context.Context) *Store {
	return &Store{
		kv: map[string]struct{}{},
	}
}

func (s *Store) RegisterWorker(workerId string) {}

func (s *Store) RegisterWorkload(workloadName string, workerId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s.%s", workerId, workloadName)

	if _, exists := s.kv[key]; exists {
		slog.Error("unable to register workload", "reason", "workload already exists")
	}

	s.kv[key] = struct{}{}
}

func (s *Store) DeleteWorkload(workloadName string, workerId string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s.%s", workerId, workloadName)

	if _, exists := s.kv[key]; exists {
		delete(s.kv, key)
	}
}

func (s *Store) UpdateWorkload(workloadName string, workerId string) {}
