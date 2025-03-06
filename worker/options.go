package worker

import (
	"context"
	"time"
)

type Option func(*Worker)

// Set an explicit Worker ID, default: UUID
func WithWorkerID(id string) Option {
	return func(w *Worker) {
		w.config.id = id
	}
}

// Set a splay interval, default: lo 8, hi 10
func WithPingSplay(hi, lo time.Duration) Option {
	return func(w *Worker) {
		w.config.splayLo = lo
		w.config.splayHi = hi
	}
}

// Set a ping timeout, default: 10 seconds
func WithPingTimeout(t time.Duration) Option {
	return func(w *Worker) {
		w.config.pingTimeout = t
	}
}

// Set pingdown threshold, default: 2
func WithPingdownThreshold(n int) Option {
	return func(w *Worker) {
		w.config.maxPingDown = n
	}
}

// Add a callback executes when the worker encounters an error
func WithErrorCallback(fn func(error)) Option {
	return func(w *Worker) {
		w.config.errCb = fn
	}
}

// Add a callback that is executed when an event is created
func WithEventCallback(fn func(context.Context, Event)) Option {
	return func(w *Worker) {
		w.config.eventCbs = append(w.config.eventCbs, fn)
	}
}

func WithExternalState(sr StateRepository) Option {
	return func(w *Worker) {
		w.sr = sr
	}
}