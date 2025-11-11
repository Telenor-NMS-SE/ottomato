package manager

import "testing"

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

}
