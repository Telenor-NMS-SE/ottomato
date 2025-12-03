package manager

import (
	"testing"
	"time"
)

func TestAssign(t *testing.T) {
	state := NewMemoryStore()

	mgr := &Manager{
		state: state,
	}
	w := &MockWorker{id: "worker1"}
	wl := &MockWorkload{id: "workload1"}

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
			"worker1": &MockWorker{id: "worker1"},
			"worker2": &MockWorker{id: "worker2"},
		},
		workloads: map[string]Workload{
			"workload1": &MockWorkload{id: "workload1"},
			"workload2": &MockWorkload{id: "workload2"},
			"workload3": &MockWorkload{id: "workload3"},
			"workload4": &MockWorkload{id: "workload4"},
			"workload5": &MockWorkload{id: "workload5"},
			"workload6": &MockWorkload{id: "workload6"},
			"workload7": &MockWorkload{id: "workload7"},
		},
		associations: map[string]string{},
	}
	mgr := &Manager{
		state: state,
	}
	mgr.distributor()

	if exp, recv := len(state.workloads), len(state.associations); exp != recv {
		t.Errorf("expected length of distributed workloads to be %d got: %d", exp, recv)
	}

	counters := map[string]int{}
	for _, w := range state.GetAllWorkers() {
		counters[w.GetID()] = len(state.GetAssociations(w))
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
			"worker-0": &MockWorker{id: "worker0"},
			"worker-1": &MockWorker{id: "worker1"},
		},
		workloads: map[string]Workload{
			"workload0":  &MockWorkload{id: "workerload0"},
			"workload1":  &MockWorkload{id: "workerload1"},
			"workload2":  &MockWorkload{id: "workerload2"},
			"workload3":  &MockWorkload{id: "workerload3"},
			"workload4":  &MockWorkload{id: "workerload4"},
			"workload5":  &MockWorkload{id: "workerload5"},
			"workload6":  &MockWorkload{id: "workerload6"},
			"workload7":  &MockWorkload{id: "workerload7"},
			"workload8":  &MockWorkload{id: "workerload8"},
			"workload9":  &MockWorkload{id: "workerload9"},
			"workload10": &MockWorkload{id: "workerload10"},
			"workload11": &MockWorkload{id: "workerload11"},
			"workload12": &MockWorkload{id: "workerload12"},
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
		state: state,
	}
	mgr.rebalance()

	counters := map[string]int{}
	for _, w := range state.GetAllWorkers() {
		counters[w.GetID()] = len(state.GetAssociations(w))
	}

	_, _, delta := mgr.sort(counters)
	if delta > DELTA_MAX {
		t.Errorf("expected delta of rebalanced workloads to be no more than %d, got: %d", DELTA_MAX, delta)
	}
}

func TestDistributionCleanup(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"worker-0": &MockWorker{id: "worker-0"},
		},
		workloads: map[string]Workload{
			"workload-0": &MockWorkload{
				id: "workload-0",
				status: StatusDistributing,
				statusChange: time.Now().Add(-time.Hour),
			},
			"workload-1": &MockWorkload{
				id: "workload-1",
				status: StatusErr,
				statusChange: time.Now().Add(-time.Hour),
			},
		},
		associations: map[string]string{
			"workload-0": "worker-0",
		},
	}

	mgr := &Manager{
		state: state,
	}

	mgr.cleanup()
	if len(state.associations) != 0 {
		t.Errorf("expected distributed workloads to be empty, got: %d", len(state.associations))
	}

	if state.workloads["workload-1"].GetStatus() != StatusInit {
		t.Errorf("expected errornous workload to have ben set to %s, got: %s", StatusInit, state.workloads["workload-1"].GetStatus())
	}
}
