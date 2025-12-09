package manager

import (
	"context"
	"testing"
)

func TestStateGetAllWorkers(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{
			"worker0": &mockWorker{id: "worker0"},
			"worker1": &mockWorker{id: "worker1"},
			"worker2": &mockWorker{id: "worker2"},
		},
	}

	workers, err := state.GetAllWorkers(context.TODO())
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

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
			"worker0": &mockWorker{id: "worker0"},
			"worker1": &mockWorker{id: "worker1"},
			"worker2": &mockWorker{id: "worker2"},
		},
	}

	w, err := state.GetWorker(context.TODO(), "worker0")
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if exp, recv := "worker0", w.GetID(); exp != recv {
		t.Fatalf("expected to get '%s', but got: %s", exp, recv)
	}
}

func TestStateAddWorker(t *testing.T) {
	state := &MemoryStore{
		workers: map[string]Worker{},
	}

	if err := state.AddWorker(context.TODO(), &mockWorker{id: "worker0"}); err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

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
			"worker0": &mockWorker{id: "worker0"},
			"worker1": &mockWorker{id: "worker1"},
			"worker2": &mockWorker{id: "worker2"},
		},
	}

	if err := state.DeleteWorker(context.TODO(), &mockWorker{id: "worker0"}); err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

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
			"workload0": &mockWorkload{id: "workload0"},
			"workload1": &mockWorkload{id: "workload1"},
			"workload2": &mockWorkload{id: "workload2"},
		},
	}

	workloads, err := state.GetAllWorkloads(context.TODO())
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

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
			"workload0": &mockWorkload{id: "workload0"},
			"workload1": &mockWorkload{id: "workload1"},
			"workload2": &mockWorkload{id: "workload2"},
		},
	}

	wl, err := state.GetWorkload(context.TODO(), "workload0")
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if exp, recv := "workload0", wl.GetID(); exp != recv {
		t.Fatalf("expected to get '%s', but got: %s", exp, recv)
	}
}

func TestStateAddWorkload(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{},
	}

	if err := state.AddWorkload(context.TODO(), &mockWorkload{id: "workload0"}); err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

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
	wl := &mockWorkload{
		id:     "workload0",
		status: StatusErr,
	}

	state := &MemoryStore{
		workloads: map[string]Workload{
			wl.GetID(): wl,
		},
	}

	updated := &mockWorkload{
		id:     wl.GetID(),
		status: StatusRunning,
	}

	if err := state.UpdateWorkload(context.TODO(), updated); err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if state.workloads[wl.GetID()].GetStatus() != StatusRunning {
		t.Fatalf("expected workload to have an updated status to '%s', but got: %s", StatusRunning, state.workloads[wl.GetID()].GetStatus())
	}
}

func TestStateDeleteWorkload(t *testing.T) {
	state := &MemoryStore{
		workloads: map[string]Workload{
			"workload0": &mockWorkload{id: "workload0"},
			"workload1": &mockWorkload{id: "workload1"},
			"workload2": &mockWorkload{id: "workload2"},
		},
	}

	if err := state.DeleteWorkload(context.TODO(), &mockWorkload{id: "workload0"}); err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if exp, recv := 2, len(state.workloads); exp != recv {
		t.Fatalf("expected to see %d workload(s), but got: %d", exp, recv)
	}

	if _, ok := state.workloads["workload0"]; ok {
		t.Fatalf("expected 'workload0' to be removed from state, but it isn't")
	}
}

func TestGetAssociations(t *testing.T) {
	w := &mockWorker{id: "worker0"}
	wl := &mockWorkload{id: "workload0"}

	state := &MemoryStore{
		workers: map[string]Worker{
			w.GetID(): w,
		},
		workloads: map[string]Workload{
			"workload0": wl,
		},
		associations: map[string]string{
			"workload0": "worker0",
		},
	}

	assocs, err := state.GetAssociations(context.TODO(), w)
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if exp, recv := 1, len(assocs); exp != recv {
		t.Fatalf("expected to get %d association(s), but got: %d", exp, recv)
	}

	if exp, recv := wl.GetID(), assocs[0].GetID(); exp != recv {
		t.Fatalf("expected to see '%s' as the first association, but got: %s", exp, recv)
	}
}

func TestGetAssociation(t *testing.T) {
	w := &mockWorker{id: "worker0"}
	wl := &mockWorkload{id: "workload0"}

	state := &MemoryStore{
		workers: map[string]Worker{
			w.GetID(): w,
		},
		workloads: map[string]Workload{
			"workload0": wl,
		},
		associations: map[string]string{
			"workload0": "worker0",
		},
	}

	assoc, err := state.GetAssociation(context.TODO(), wl)
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if exp, recv := w.GetID(), assoc.GetID(); exp != recv {
		t.Fatalf("expected to get an association to '%s', but got: %s", exp, recv)
	}
}

func TestAssociate(t *testing.T) {
	w := &mockWorker{id: "worker0"}
	wl := &mockWorkload{id: "workload0"}

	state := &MemoryStore{
		workers: map[string]Worker{
			w.GetID(): w,
		},
		workloads: map[string]Workload{
			"workload0": wl,
		},
		associations: map[string]string{},
	}

	if err := state.Associate(context.TODO(), wl, w); err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if exp, recv := 1, len(state.associations); exp != recv {
		t.Fatalf("expected to see %d association(s), but got: %d", exp, recv)
	}

	workerId, ok := state.associations[wl.GetID()]
	if !ok {
		t.Fatalf("expected to find association for '%s', but didn't", wl.GetID())
	}

	if exp, recv := w.GetID(), workerId; exp != recv {
		t.Fatalf("expected to find an association towards '%s', but got: %s", exp, recv)
	}
}

func TestDisassociate(t *testing.T) {
	w := &mockWorker{id: "worker0"}
	wl := &mockWorkload{id: "workload0"}

	state := &MemoryStore{
		workers: map[string]Worker{
			w.GetID(): w,
		},
		workloads: map[string]Workload{
			"workload0": wl,
		},
		associations: map[string]string{
			wl.GetID(): w.GetID(),
		},
	}

	if err := state.Disassociate(context.TODO(), wl, w); err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}

	if exp, recv := 0, len(state.associations); exp != recv {
		t.Fatalf("expected to find %d association(s), but got: %d", exp, recv)
	}
}
