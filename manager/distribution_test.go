package manager

import (
	"context"
	"testing"
	"time"
)

func TestAssign(t *testing.T) {
	state := NewMemoryStore()

	mgr := &Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &mockSignaller{},
	}
	w := &mockWorker{id: "worker1"}
	wl := &mockWorkload{id: "workload1"}

	mgr.Assign(w, wl)

	if len(state.workloads) != 1 {
		t.Fatalf("expected manager distributions to be exactly 1, but got: %d", len(state.associations))
	}

	if len(state.associations) != 1 {
		t.Fatalf("expected manager distributions to be exactly 1, but got: %d", len(state.associations))
	}

	if _, ok := state.workloads[wl.GetID()]; !ok {
		t.Fatalf("expected to find workload '%s', but didn't", wl.GetID())
	}

	workerId, ok := state.associations[wl.GetID()]
	if !ok {
		t.Fatalf("expected to find workloadId '%s' in distributions, but didn't", wl.GetID())
	}

	if workerId != w.GetID() {
		t.Errorf("expected to find workerId '%s', but got: %s", w.GetID(), workerId)
	}
}

func TestDistributor(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"worker1": &mockWorker{id: "worker1"},
			"worker2": &mockWorker{id: "worker2"},
		},
		workloads: map[string]Workload{
			"workload1": &mockWorkload{id: "workload1"},
			"workload2": &mockWorkload{id: "workload2"},
			"workload3": &mockWorkload{id: "workload3"},
			"workload4": &mockWorkload{id: "workload4"},
			"workload5": &mockWorkload{id: "workload5"},
			"workload6": &mockWorkload{id: "workload6"},
			"workload7": &mockWorkload{id: "workload7"},
		},
		associations: map[string]string{},
	}
	mgr := &Manager{
		state:               state,
		ctx:                 context.TODO(),
		signal:              &mockSignaller{},
		distributionTimeout: time.Minute,
	}
	mgr.distributor()

	if exp, recv := len(state.workloads), len(state.associations); exp != recv {
		t.Errorf("expected length of distributed workloads to be %d got: %d", exp, recv)
	}

	workers, err := state.GetAllWorkers(context.TODO())
	if err != nil {
		t.Fatalf("unexpected error when getting all workers: %v", err)
	}

	counters := map[string]int{}
	for _, w := range workers {
		assocs, err := state.GetAssociations(context.TODO(), w)
		if err != nil {
			t.Errorf("unexpected error when getting associations: %v", err)
			continue
		}

		counters[w.GetID()] = len(assocs)
	}

	_, _, delta := mgr.sort(counters)
	if delta != 1 {
		t.Errorf("expected delta of distributed workloads to be no more than 1, got: %d", delta)
	}

	for _, wl := range state.workloads {
		if exp, recv := StatusRunning, wl.GetStatus(); exp != recv {
			t.Errorf("expected state of '%s' to be '%s', but got: %s", wl.GetID(), exp.String(), recv.String())
		}
	}
}

func TestRebalancer(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"worker-0": &mockWorker{id: "worker0"},
			"worker-1": &mockWorker{id: "worker1"},
		},
		workloads: map[string]Workload{
			"workload0":  &mockWorkload{id: "workerload0"},
			"workload1":  &mockWorkload{id: "workerload1"},
			"workload2":  &mockWorkload{id: "workerload2"},
			"workload3":  &mockWorkload{id: "workerload3"},
			"workload4":  &mockWorkload{id: "workerload4"},
			"workload5":  &mockWorkload{id: "workerload5"},
			"workload6":  &mockWorkload{id: "workerload6"},
			"workload7":  &mockWorkload{id: "workerload7"},
			"workload8":  &mockWorkload{id: "workerload8"},
			"workload9":  &mockWorkload{id: "workerload9"},
			"workload10": &mockWorkload{id: "workerload10"},
			"workload11": &mockWorkload{id: "workerload11"},
			"workload12": &mockWorkload{id: "workerload12"},
		},
		associations: map[string]string{
			"workload0":  "worker1",
			"workload1":  "worker1",
			"workload2":  "worker1",
			"workload3":  "worker1",
			"workload4":  "worker1",
			"workload5":  "worker1",
			"workload6":  "worker1",
			"workload7":  "worker1",
			"workload8":  "worker0",
			"workload9":  "worker0",
			"workload10": "worker0",
			"workload11": "worker0",
			"workload12": "worker0",
		},
	}

	mgr := &Manager{
		state:    state,
		ctx:      context.TODO(),
		signal:   &mockSignaller{},
		maxDelta: 5,
	}
	mgr.rebalance()

	workers, err := state.GetAllWorkers(context.TODO())
	if err != nil {
		t.Fatalf("unexpected error when getting all workers: %v", err)
	}

	counters := map[string]int{}
	for _, w := range workers {
		assocs, err := state.GetAssociations(context.TODO(), w)
		if err != nil {
			t.Errorf("unexpected error when getting associations: %v", err)
			continue
		}

		counters[w.GetID()] = len(assocs)
	}

	_, _, delta := mgr.sort(counters)
	if delta > mgr.maxDelta {
		t.Errorf("expected delta of rebalanced workloads to be no more than %d, got: %d", mgr.maxDelta, delta)
	}
}

func TestDistributionCleanup(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"worker-0": &mockWorker{id: "worker-0"},
		},
		workloads: map[string]Workload{
			"workload-0": &mockWorkload{
				id:           "workload-0",
				status:       StatusDistributing,
				statusChange: time.Now().Add(-time.Hour),
			},
			"workload-1": &mockWorkload{
				id:           "workload-1",
				status:       StatusErr,
				statusChange: time.Now().Add(-time.Hour),
			},
		},
		associations: map[string]string{
			"workload-0": "worker-0",
		},
	}

	mgr := &Manager{
		state:  state,
		ctx:    context.TODO(),
		signal: &mockSignaller{},
	}

	mgr.cleanup()
	if len(state.associations) != 0 {
		t.Errorf("expected distributed workloads to be empty, got: %d", len(state.associations))
	}

	if state.workloads["workload-1"].GetStatus() != StatusInit {
		t.Errorf("expected errornous workload to have ben set to %s, got: %s", StatusInit, state.workloads["workload-1"].GetStatus())
	}
}
