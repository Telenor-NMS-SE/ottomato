package worker

import (
	"context"
	"testing"
	"time"
)

type MockState struct {
	kv map[string]any //nolint:all
}

func (k *MockState) RegisterWorker(id string) {}

func (k *MockState) RegisterWorkload(name string, id string) {}

func (k *MockState) DeleteWorkload(name string, id string) {}

func (k *MockState) UpdateWorkload(name string, id string) {}

type MockWorkload struct {
	name string
}

func (mo *MockWorkload) Init(context.Context) error {
	return nil
}

func (mo *MockWorkload) Name() string {
	return mo.name
}

func (mo *MockWorkload) Stop() error {
	return nil
}

func (mo *MockWorkload) Ping(ctx context.Context) error {
	return nil
}

func (mo *MockWorkload) Info() map[string]any {
	return map[string]any{}
}

func (mo *MockWorkload) RunTask(ctx context.Context, task *Task) (Result, error) {
	return Result{JobID: "test", Return: "test"}, nil
}

func TestNewWorker(t *testing.T) {
	kv := &MockState{}

	mgr, err := New(context.Background(), kv)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	if exp, recv := 0, len(mgr.Workloads()); exp != recv {
		t.Errorf("expected length of managed objects to be %d, but recieved %d", exp, recv)
	}
}

func TestAddWorkload(t *testing.T) {
	kv := &MockState{}

	w, err := New(context.Background(), kv)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	obj := MockWorkload{name: "test"}

	wl, err := w.AddWorkload(context.Background(), &obj)
	if err != nil {
		t.Fatalf("failed to add workload: %v", err)
	}

	name, ok := wl["name"]
	if !ok {
		t.Errorf("expected to find a workload name in the returned metadata, but found none")
	}

	if exp, recv := obj.name, name; exp != recv {
		t.Errorf("expected workload name to be '%s', but got: %s", exp, recv)
	}

	if exp, recv := 1, len(w.Workloads()); exp != recv {
		t.Errorf("expected length of managed objects to be %d, but recieved %d", exp, recv)
	}
}

func TestRemoveManagedObject(t *testing.T) {
	kv := &MockState{}

	w, err := New(context.Background(), kv)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	obj := MockWorkload{name: "test"}

	if _, err := w.AddWorkload(context.Background(), &obj); err != nil {
		t.Fatalf("failed to add workload: %v", err)
	}

	if err := w.DeleteWorkload(obj.Name()); err != nil {
		t.Errorf("failed to delete workload: %v", err)
	}

	if exp, recv := 0, len(w.Workloads()); exp != recv {
		t.Errorf("expected length of managed objects to be %d, but recieved %d", exp, recv)
	}
}

func TestRunTask(t *testing.T) {
	kv := &MockState{}

	w, err := New(context.Background(), kv)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	obj := MockWorkload{name: "test"}

	if _, err := w.AddWorkload(context.Background(), &obj); err != nil {
		t.Fatalf("failed to add workload: %v", err)
	}

	job, err := w.RunTask(context.Background(), "test", &Task{Command: "test"})
	if err != nil {
		t.Fatalf("coult not run task: %s", err.Error())
	}

	if exp, recv := "test", job.JobID; exp != recv {
		t.Errorf("expected job ID %s, but got: %s", exp, recv)
	}
}

func TestEventCallback(t *testing.T) {
	kv := &MockState{}

	var hit bool
	opts := []Option{
		WithEventCallback(func(ctx context.Context, e Event) {
			hit = true
		}),
	}

	w, err := New(context.Background(), kv, opts...)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	obj := MockWorkload{name: "test"}
	if _, err := w.AddWorkload(context.Background(), &obj); err != nil {
		t.Fatalf("Could not add new workload; %s", err.Error())
	}

	// wait, because events are async
	time.Sleep(10 * time.Millisecond)

	if exp, recv := true, hit; exp != recv {
		t.Errorf("expected callback to be executed, but it wasn't")
	}
}
