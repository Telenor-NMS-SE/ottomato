package manager

import (
	"testing"
	"time"
)

func TestAssign(t *testing.T) {
	mgr := &Manager{
		workloads:     map[string]Workload{},
		distributions: map[string]string{},
	}
	w := &MockWorker{id: "worker1"}
	wl := &MockWorkload{id: "workload1"}

	mgr.Assign(w, wl)

	if len(mgr.workloads) != 1 {
		t.Fatalf("expected manager distributions to be exactly 1, but got: %d", len(mgr.distributions))
	}

	if len(mgr.distributions) != 1 {
		t.Fatalf("expected manager distributions to be exactly 1, but got: %d", len(mgr.distributions))
	}

	if _, ok := mgr.workloads[wl.GetID()]; !ok {
		t.Fatalf("expected to find workload '%s', but didn't", wl.GetID())
	}

	workerId, ok := mgr.distributions[wl.GetID()]
	if !ok {
		t.Fatalf("expected to find workloadId '%s' in distributions, but didn't", wl.GetID())
	}

	if workerId != w.GetID() {
		t.Errorf("expected to find workerId '%s', but got: %s", w.GetID(), workerId)
	}
}

func TestDistributor(t *testing.T) {
	mgr := &Manager{
		workers: map[string]Worker{
			"worker1": &MockWorker{id: "worker1"},
			"worker2": &MockWorker{id: "worker2"},
		},
		workloads: map[string]Workload{
			"workload1": &MockWorkload{id: "worker1"},
			"workload2": &MockWorkload{id: "worker2"},
			"workload3": &MockWorkload{id: "worker3"},
			"workload4": &MockWorkload{id: "worker4"},
			"workload5": &MockWorkload{id: "worker5"},
			"workload6": &MockWorkload{id: "worker6"},
			"workload7": &MockWorkload{id: "worker7"},
		},
		distributions: map[string]string{},
	}
	mgr.distributor()

	if exp, recv := len(mgr.workloads), len(mgr.distributions); exp != recv {
		t.Errorf("expected length of distributed workloads to be %d got: %d", exp, recv)
	}

	_, _, delta := mgr.sort()
	if delta != 1 {
		t.Errorf("expected delta of distributed workloads to be no more than 1, got: %d", delta)
	}

}

func TestRebalancer(t *testing.T) {
	mgr := &Manager{
		workers: map[string]Worker{
			"worker-0": &MockWorker{id: "worker-0"},
			"worker-1": &MockWorker{id: "worker-1"},
		},
		workloads: map[string]Workload{
			"workload-0":  &MockWorkload{id: "workerload-0"},
			"workload-1":  &MockWorkload{id: "workerload-1"},
			"workload-2":  &MockWorkload{id: "workerload-2"},
			"workload-3":  &MockWorkload{id: "workerload-3"},
			"workload-4":  &MockWorkload{id: "workerload-4"},
			"workload-5":  &MockWorkload{id: "workerload-5"},
			"workload-6":  &MockWorkload{id: "workerload-6"},
			"workload-7":  &MockWorkload{id: "workerload-7"},
			"workload-8":  &MockWorkload{id: "workerload-8"},
			"workload-9":  &MockWorkload{id: "workerload-9"},
			"workload-10": &MockWorkload{id: "workerload-10"},
			"workload-11": &MockWorkload{id: "workerload-11"},
			"workload-12": &MockWorkload{id: "workerload-12"},
		},
		distributions: map[string]string{
			"workload-0":  "worker-1",
			"workload-1":  "worker-1",
			"workload-2":  "worker-1",
			"workload-3":  "worker-1",
			"workload-4":  "worker-1",
			"workload-5":  "worker-1",
			"workload-6":  "worker-1",
			"workload-7":  "worker-1",
			"workload-8":  "worker-0",
			"workload-9":  "worker-0",
			"workload-10": "worker-0",
			"workload-11": "worker-0",
			"workload-12": "worker-0",
		},
	}
	mgr.rebalance()

	_, _, delta := mgr.sort()
	if delta > DELTA_MAX {
		t.Errorf("expected delta of rebalanced workloads to be no more than %d, got: %d", DELTA_MAX, delta)
	}
}

func TestDistributionCleanup(t *testing.T) {
	mgr := &Manager{
		workers: map[string]Worker{
			"worker-0": &MockWorker{id: "worker-0"},
		},
		workloads: map[string]Workload{
			"workload-0": &MockWorkload{id: "workload-0", state: StateDistributing, stateChange: time.Now().Add(-time.Hour)},
			"workload-1": &MockWorkload{id: "workload-1", state: StateErr, stateChange: time.Now().Add(-time.Hour)},
		},
		distributions: map[string]string{
			"workload-0": "worker-0",
		},
	}

	mgr.cleanup()
	if len(mgr.distributions) != 0 {
		t.Errorf("expected distributed workloads to be empty, got: %d", len(mgr.distributions))
	}
	if mgr.workloads["workload-1"].GetState() != StateInit {
		t.Errorf("expected errornous workload to have ben set to %s, got: %s", StateInit, mgr.workloads["workload-1"].GetState())
	}
}
