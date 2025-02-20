package worker

import (
	"context"
	"testing"
	"time"
)

type MockKvStore struct {
	kv map[string]any
}

func (k *MockKvStore) RegisterWorker(id string) {}

func (k *MockKvStore) RegisterWorkload(name string, id string) {}

func (k *MockKvStore) DeleteWorkload(name string, id string) {}

func (k *MockKvStore) UpdateWorkload(name string, id string) {}

type MockManagedObject struct {
	name string
}

func (mo *MockManagedObject) Init(context.Context) error {
	return nil
}

func (mo *MockManagedObject) Name() string {
	return mo.name
}

func (mo *MockManagedObject) Stop() error {
	return nil
}

func (mo *MockManagedObject) Ping(ctx context.Context) error {
	return nil
}

func (mo *MockManagedObject) Info() map[string]any {
	return map[string]any{}
}

func (mo *MockManagedObject) RunTask(ctx context.Context, target string, task *Task) (Result, error) {
	return Result{JobID: "test", Return: "test"}, nil
}

func TestNewWorker(t *testing.T) {
	kv := &MockKvStore{}

	mgr, err := New(context.Background(), kv)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	if exp, recv := 0, len(mgr.Workloads()); exp != recv {
		t.Errorf("expected length of managed objects to be %d, but recieved %d", exp, recv)
	}
}

func TestAddWorkload(t *testing.T) {
	kv := &MockKvStore{}

	mgr, err := New(context.Background(), kv)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	obj := MockManagedObject{name: "test"}

	if err := mgr.AddWorkload(context.Background(), &obj); err != nil {
		t.Fatalf("failed to add workload: %v", err)
	}

	if exp, recv := 1, len(mgr.Workloads()); exp != recv {
		t.Errorf("expected length of managed objects to be %d, but recieved %d", exp, recv)
	}
}

func TestRemoveManagedObject(t *testing.T) {
	kv := &MockKvStore{}

	mgr, err := New(context.Background(), kv)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	obj := MockManagedObject{name: "test"}

	if err := mgr.AddWorkload(context.Background(), &obj); err != nil {
		t.Fatalf("failed to add workload: %v", err)
	}

	if err := mgr.DeleteWorkload(obj.Name()); err != nil {
		t.Errorf("failed to delete workload: %v", err)
	}

	if exp, recv := 0, len(mgr.Workloads()); exp != recv {
		t.Errorf("expected length of managed objects to be %d, but recieved %d", exp, recv)
	}
}

func TestRunTask(t *testing.T) {
	kv := &MockKvStore{}

	mgr, err := New(context.Background(), kv)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	obj := MockManagedObject{name: "test"}

	if err := mgr.AddWorkload(context.Background(), &obj); err != nil {
		t.Fatalf("failed to add workload: %v", err)
	}

	job, err := mgr.RunTask(context.Background(), "test", &Task{Command: "test"})
	if err != nil {
		t.Fatalf("coult not run task: %s", err.Error())
	}

	if exp, recv := "test", job.JobID; exp != recv {
		t.Errorf("expected job ID %s, but got: %s", exp, recv)
	}
}

func TestEventCallback(t *testing.T) {
	kv := &MockKvStore{}

	var hit bool
	opts := []Option{
		WithEventCallback(func(ctx context.Context, e Event) {
			hit = true
		}),
	}

	mgr, err := New(context.Background(), kv, opts...)
	if err != nil {
		t.Fatalf("coult not create new manager: %s", err.Error())
	}

	obj := MockManagedObject{name: "test"}
	mgr.AddWorkload(context.Background(), &obj)

	// wait, because events are async
	time.Sleep(10 * time.Millisecond)

	if exp, recv := true, hit; exp != recv {
		t.Errorf("expected callback to be executed, but it wasn't")
	}
}