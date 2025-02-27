package worker

import (
	"context"
	"testing"
	"time"
)

func TestWithWorkerID(t *testing.T) {
	w := Worker{}

	opt := WithWorkerID("test")
	opt(&w)

	if exp, recv := "test", w.config.id; exp != recv {
		t.Errorf("expected worker to have id %s, but got: %s", exp, recv)
	}
}

func TestWithSplay(t *testing.T) {
	w := Worker{}

	opt := WithPingSplay(time.Second, time.Millisecond)
	opt(&w)

	if exp, recv := time.Millisecond, w.config.splayLo; exp != recv {
		t.Errorf("expected worker to have splayLow set to %v, but got: %v", exp, recv)
	}

	if exp, recv := time.Second, w.config.splayHi; exp != recv {
		t.Errorf("expected worker to have splayHigh set to %v, but got: %v", exp, recv)
	}
}

func TestWithPingTimeout(t *testing.T) {
	w := Worker{}

	opt := WithPingTimeout(time.Second)
	opt(&w)

	if exp, recv := time.Second, w.config.pingTimeout; exp != recv {
		t.Errorf("expected worker to have pingTimeout of %v, but got: %v", exp, recv)
	}
}

func TestWithPingdownThreshold(t *testing.T) {
	w := Worker{}

	opt := WithPingdownThreshold(1337)
	opt(&w)

	if exp, recv := 1337, w.config.maxPingDown; exp != recv {
		t.Errorf("expected a pingdown threshold of %d, but got: %d", exp, recv)
	}
}

func TestWithEventCallback(t *testing.T) {
	w := Worker{}

	opt := WithEventCallback(func(ctx context.Context, e Event) {})
	opt(&w)

	if exp, recv := 1, len(w.config.eventCbs); exp != recv {
		t.Errorf("expected worker to %d event callbacks registered, but got: %d", exp, recv)
	}
}
