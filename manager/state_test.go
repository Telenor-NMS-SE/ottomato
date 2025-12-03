package manager

import "testing"

func TestStateGetAllWorkers(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"worker0": &MockWorker{id:"worker0"},
			"worker1": &MockWorker{id:"worker1"},
			"worker2": &MockWorker{id:"worker2"},
		},
	}

	workers := state.GetAllWorkers()

	if exp, recv := 3, len(workers); exp != recv {
		t.Fatalf("expected to receive %d workers, but got: %d", exp, recv)
	}

	found := map[string]bool{
		"worker0": false,
		"worker1": false,
		"worker2": false,
	}

	for _, w := range workers {
		f, ok := found[w.GetID()]
		if !ok {
			t.Errorf("unexpected worker '%s'", w.GetID())
			continue
		}

		if f {
			t.Errorf("worker '%s' has already been accounted for", w.GetID())
			continue
		}
		
		found[w.GetID()] = true
	}
}

func TestStateGetWorker(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"worker0": &MockWorker{id:"worker0"},
			"worker1": &MockWorker{id:"worker1"},
			"worker2": &MockWorker{id:"worker2"},
		},
	}

	w, ok := state.GetWorker("worker0")
	if !ok {
		t.Fatalf("expected to find worker 'worker0', but didn't")
	}

	if exp, recv := "worker0", w.GetID(); exp != recv {
		t.Fatalf("expected to get '%s', but got: %s", exp, recv)
	}
}

func TestStateAddWorker(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{},
	}

	state.AddWorker(&MockWorker{id:"worker0"})

	if exp, recv := 1, len(state.workers); exp != recv {
		t.Fatalf("expected to see %d worker(s), but found: %d", exp, recv)
	}

	w, ok := state.workers["worker0"]
	if !ok {
		t.Fatalf("expected to find 'worker0', but didn't")
	}

	if exp, recv := "worker0", w.GetID(); exp != recv {
		t.Fatalf("expected to get '%s', but got: %s", exp, recv)
	}
}

func TestStateDeleteWorker(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"worker0": &MockWorker{id:"worker0"},
			"worker1": &MockWorker{id:"worker1"},
			"worker2": &MockWorker{id:"worker2"},
		},
	}

	state.DeleteWorker(&MockWorker{id:"worker0"})

	if exp, recv := 2, len(state.workers); exp != recv {
		t.Fatalf("expected to see %d worker(s), but got: %d", exp, recv)
	}

	if _, ok := state.workers["worker0"]; ok {
		t.Fatalf("expected 'worker0' to be removed from state, but it isn't")
	}
}

func TestStateGetAllWorkloads(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{
			"workload0": &MockWorkload{id:"workload0"},
			"workload1": &MockWorkload{id:"workload1"},
			"workload2": &MockWorkload{id:"workload2"},
		},
	}

	workloads := state.GetAllWorkloads()

	if exp, recv := 3, len(workloads); exp != recv {
		t.Fatalf("expected to receive %d workers, but got: %d", exp, recv)
	}

	found := map[string]bool{
		"workload0": false,
		"workload1": false,
		"workload2": false,
	}

	for _, wl := range workloads {
		f, ok := found[wl.GetID()]
		if !ok {
			t.Errorf("unexpected workload '%s'", wl.GetID())
			continue
		}

		if f {
			t.Errorf("workload '%s' has already been accounted for", wl.GetID())
			continue
		}
		
		found[wl.GetID()] = true
	}
}

func TestStateGetWorkload(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{
			"workload0": &MockWorkload{id:"workload0"},
			"workload1": &MockWorkload{id:"workload1"},
			"workload2": &MockWorkload{id:"workload2"},
		},
	}

	wl, ok := state.GetWorkload("workload0")
	if !ok {
		t.Fatalf("expected to find workload 'workload0', but didn't")
	}

	if exp, recv := "workload0", wl.GetID(); exp != recv {
		t.Fatalf("expected to get '%s', but got: %s", exp, recv)
	}
}

func TestStateAddWorkload(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{},
	}

	state.AddWorkload(&MockWorkload{id:"workload0"})

	if exp, recv := 1, len(state.workloads); exp != recv {
		t.Fatalf("expected to see %d workload(s), but found: %d", exp, recv)
	}

	wl, ok := state.workloads["workload0"]
	if !ok {
		t.Fatalf("expected to find 'workload0', but didn't")
	}

	if exp, recv := "workload0", wl.GetID(); exp != recv {
		t.Fatalf("expected to get '%s', but got: %s", exp, recv)
	}
}

func TestStateUpdateWorkload(t *testing.T) {
	wl := &MockWorkload{
		id:"workload0",
		status: StatusErr,
	}

	state := &MemoryStore{
		workloads: map[string]Workload{
			wl.GetID(): wl,
		},
	}

	updated := &MockWorkload{
		id:     wl.GetID(),
		status: StatusRunning,
	}

	state.UpdateWorkload(updated)

	if state.workloads[wl.GetID()].GetStatus() != StatusRunning {
		t.Fatalf("expected workload to have an updated status to '%s', but got: %s", StatusRunning, state.workloads[wl.GetID()].GetStatus())
	}
}

func TestStateDeleteWorkload(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{
			"workload0": &MockWorkload{id:"workload0"},
			"workload1": &MockWorkload{id:"workload1"},
			"workload2": &MockWorkload{id:"workload2"},
		},
	}

	state.DeleteWorkload(&MockWorkload{id:"workload0"})

	if exp, recv := 2, len(state.workloads); exp != recv {
		t.Fatalf("expected to see %d workload(s), but got: %d", exp, recv)
	}

	if _, ok := state.workloads["workload0"]; ok {
		t.Fatalf("expected 'workload0' to be removed from state, but it isn't")
	}
}
