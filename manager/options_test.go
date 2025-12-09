package manager

import (
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

func TestWithSignaller(t *testing.T) {
	mgr := &Manager{}
	signaller := &SlogSignaller{}

	WithSignaller(signaller)(mgr)

	if mgr.signal == nil {
		t.Fatalf("expected signaller to be a non-nil value")
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
