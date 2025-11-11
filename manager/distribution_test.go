package manager

import "testing"

func TestDistributor(t *testing.T) {
	mgr := &Manager{
		workers: map[string]Worker{
			"worker1": &MockWorker{},
			"worker2": &MockWorker{},
		},
		workloads: map[string]Workload{
			"workload1": &MockWorkload{},
			"workload2": &MockWorkload{},
			"workload3": &MockWorkload{},
			"workload4": &MockWorkload{},
			"workload5": &MockWorkload{},
			"workload6": &MockWorkload{},
			"workload7": &MockWorkload{},
		},
		distributions: map[string]string{},
	}
	mgr.distributor()

	if exp, recv := len(mgr.workloads), len(mgr.distributions); exp != recv {
		t.Errorf("expected length of distributed workloads to be %d got: %d", exp, recv)
	}

	_, _, delta := mgr.ChatGPTSortDelta()
	if delta != 1 {
		t.Errorf("expected delta of distributed workloads to be no more than 1, got: %d", delta)
	}

}

func TestRebalancer(t *testing.T) {

}

func TestChatGPTSortDelta(t *testing.T) {
	type TestCase struct {
		Input map[string]string
		Exp   struct {
			Hi    string
			Lo    string
			Delta uint32
		}
	}

	cases := []TestCase{
		{
			Input: map[string]string{
				"workload1": "worker1",
				"workload2": "worker1",
				"workload3": "worker1",
				"workload4": "worker1",
				"workload5": "worker2",
				"workload6": "worker2",
			},
			Exp: struct {
				Hi    string
				Lo    string
				Delta uint32
			}{
				Hi:    "worker1",
				Lo:    "worker2",
				Delta: uint32(2),
			},
		},
	}

	for _, tc := range cases {
		mgr := Manager{
			distributions: tc.Input,
		}

		hi, lo, delta := mgr.ChatGPTSortDelta()
		if hi != tc.Exp.Hi {
			t.Errorf("expected 'hi' to be '%s', but got: %s", tc.Exp.Hi, hi)
		}
		if lo != tc.Exp.Lo {
			t.Errorf("expected 'lo' to be '%s', but got: %s", tc.Exp.Lo, lo)
		}
		if delta != tc.Exp.Delta {
			t.Errorf("expected 'delta' to be '%d', but got: %d", tc.Exp.Delta, delta)
		}
	}
}
