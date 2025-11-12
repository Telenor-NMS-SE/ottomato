package manager

import (
	"context"
	"testing"
	"time"
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

func TestWithDistributionInterval(t *testing.T) {
	mgr := &Manager{}
	WithDistributorInterval(time.Hour)(mgr)

	if exp, recv := time.Hour, mgr.distributionInterval; exp != recv {
		t.Errorf("expected distribution interval to be '%s', but got '%s'", exp, recv)
	}
}

func TestWithRebalanceInterval(t *testing.T) {
	mgr := &Manager{}
	WithRebalanceInterval(time.Minute)(mgr)

	if exp, recv := time.Minute, mgr.rebalanceInterval; exp != recv {
		t.Errorf("expected rebalance interval to be '%s', but got '%s'", exp, recv)
	}
}
func TestWithCleanupInterval(t *testing.T) {
	mgr := &Manager{}
	WithCleanupInterval(time.Minute)(mgr)

	if exp, recv := time.Minute, mgr.cleanupInterval; exp != recv {
		t.Errorf("expected rebalance interval to be '%s', but got '%s'", exp, recv)
	}
}

func TestWithCleanupMaxTime(t *testing.T) {
	mgr := &Manager{}
	WithCleanupMaxTime(time.Minute)(mgr)

	if exp, recv := time.Minute, mgr.cleanupMaxTime; exp != recv {
		t.Errorf("expected rebalance interval to be '%s', but got '%s'", exp, recv)
	}
}
