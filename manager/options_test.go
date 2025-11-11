package manager

import (
	"context"
	"testing"
)

func TestWithManagerID(t *testing.T) {
	id := "test-id"
	mgr := &Manager{}

	WithManagerID(id)(mgr)

	if exp, recv := id, mgr.id; exp != recv {
		t.Errorf("expected manager to have id '%s', but got: %s", exp, recv)
	}
}

func TestWithEventCallback(t *testing.T) {
	mgr := &Manager{
		eventCbs: []func(context.Context, *Event){},
	}

	WithEventCallback(func(ctx context.Context, e *Event) {})(mgr)

	if exp, recv := 1, len(mgr.eventCbs); exp != recv {
		t.Fatalf("expected to have %d callbacks registered, but got: %d", exp, recv)
	}
}
