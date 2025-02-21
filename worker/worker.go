package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Worker struct {
	ctx context.Context
	sc  gocron.Scheduler
	sr  StateRepository

	config struct {
		id          string
		splayHi     time.Duration
		splayLo     time.Duration
		maxPingDown int
		pingTimeout time.Duration
		eventCbs    []func(context.Context, Event)
		errCb       func(error)
	}

	workloadsMu sync.RWMutex
	workloads   map[string]workload

	failMu      sync.Mutex
	failCounter map[string]int

	EventCh chan Event
}

type (
	Workload interface {
		Init(context.Context)
		Name() string
		Stop() error
		Ping(context.Context) error
		Info() map[string]any
		RunTask(context.Context, string, *Task) (Result, error)
	}
	workload struct {
		Object Workload
		jid    uuid.UUID
	}
)

type (
	Task struct {
		Command string         `json:"command"`
		Args    []any          `json:"arguments"`
		Kwargs  map[string]any `json:"kwargs"`
	}
	Result struct {
		JobID         string         `json:"jobId"`
		WorkerID      string         `json:"workerId"`
		Tags          []string       `json:"tags"`
		Hostname      string         `json:"hostname"`
		Command       string         `json:"command"`
		Args          []any          `json:"arguments"`
		Kwargs        map[string]any `json:"kwargs"`
		Success       bool           `json:"success"`
		Return        any            `json:"return"`
		Timestamp     time.Time      `json:"timestamp"`
		ExecutionTime int64          `json:"executionTime"`
	}
)

type (
	StateRepository interface {
		RegisterWorker(string)
		RegisterWorkload(string, string)
		DeleteWorkload(string, string)
		UpdateWorkload(string, string)
	}
)

const (
	DEFAULT_SPLAY_LO     = 8 * time.Second
	DEFAULT_SPLAY_HI     = 10 * time.Second
	DEFAULT_PING_TIMEOUT = 10 * time.Second
	DEFAULT_MAX_PINGDOWN = 2
)

var (
	ErrWorkloadNotFound = errors.New("managed object does not exist")
	ErrWorkloadExists   = errors.New("managed object already exist")
	ErrScheduleCleanup  = errors.New("failed to clean up scheduler")
)

// Create a new worker instance with default options, override with []Option
func New(ctx context.Context, sr StateRepository, opts ...Option) (*Worker, error) {
	var err error

	worker := &Worker{
		ctx:         ctx,
		sr:          sr,
		workloads:   make(map[string]workload),
		EventCh:     make(chan Event),
		failCounter: map[string]int{},
	}

	worker.config.eventCbs = append(worker.config.eventCbs, worker.stateUpdateCb)

	for _, opt := range opts {
		opt(worker)
	}

	// set sane defaults if no options has been provided
	if worker.config.id == "" {
		worker.config.id = uuid.NewString()
	}

	if worker.config.splayHi == 0 {
		worker.config.splayHi = DEFAULT_SPLAY_HI
	}

	if worker.config.splayLo == 0 {
		worker.config.splayLo = DEFAULT_SPLAY_LO
	}

	if worker.config.pingTimeout == 0 {
		worker.config.pingTimeout = DEFAULT_PING_TIMEOUT
	}

	if worker.config.maxPingDown == 0 {
		worker.config.maxPingDown = DEFAULT_MAX_PINGDOWN
	}

	if worker.sc, err = gocron.NewScheduler(); err != nil {
		return worker, err
	}

	if _, err = worker.sc.NewJob(
		gocron.DurationJob(10*time.Second),
		gocron.NewTask(worker.garbageCollector),
		gocron.WithContext(ctx),
	); err != nil {
		return worker, err
	}

	worker.sc.Start()

	go worker.eventLoop()
	return worker, nil
}

// Stop the worker
func (w *Worker) Stop() error {
	return w.sc.Shutdown()
}

// Run a task on the given target
func (w *Worker) RunTask(ctx context.Context, target string, task *Task) (Result, error) {
	start := time.Now()

	wl, exists := w.workloads[target]
	if !exists {
		return Result{}, ErrWorkloadNotFound
	}

	job, err := wl.Object.RunTask(ctx, target, task)

	// add a bit of metadata
	job.Timestamp = start
	job.ExecutionTime = time.Since(start).Milliseconds()
	job.WorkerID = w.config.id

	return job, err
}

// Adds a new workload to the worker
func (w *Worker) AddWorkload(ctx context.Context, wl Workload) error {
	w.workloadsMu.Lock()

	if _, exists := w.workloads[wl.Name()]; exists {
		return ErrWorkloadExists
	}

	mo := workload{
		Object: wl,
	}

	job, err := w.sc.NewJob(
		gocron.DurationRandomJob(w.config.splayLo, w.config.splayHi),
		gocron.NewTask(w.stateCheck(wl.Name())),
		gocron.WithContext(ctx),
	)
	if err != nil {
		return err
	}

	mo.jid = job.ID()
	w.workloads[wl.Name()] = mo

	go wl.Init(ctx)
	w.workloadsMu.Unlock()

	w.EventCh <- *NewWorkloadInitiatedEvent(w.config.id, wl.Name())

	return nil
}

// Stops and deletes a workload from the worker
func (w *Worker) DeleteWorkload(name string) (err error) {
	w.workloadsMu.Lock()
	defer w.workloadsMu.Unlock()

	mo, exists := w.workloads[name]
	if !exists {
		return ErrWorkloadNotFound
	}

	if err := mo.Object.Stop(); err != nil {
		return err
	}

	// clean up schedule, but do not return on error
	if err := w.sc.RemoveJob(mo.jid); err != nil {
		err = fmt.Errorf("%w: %v", ErrScheduleCleanup, err) //nolint:all
	}

	delete(w.workloads, name)
	w.EventCh <- *NewWorkloadDeadEvent(w.config.id, name)

	return err
}

// This should return []map[string]string with a bunch of metadata
func (w *Worker) Workloads() []string {
	keys := make([]string, 0, len(w.workloads))

	for k := range w.workloads {
		keys = append(keys, k)
	}

	return keys
}

// Lists all of the current jobs in the task scheduler
func (w *Worker) Tasks() []map[string]string {
	jobs := w.sc.Jobs()

	tasks := make([]map[string]string, 0, len(jobs))

	for _, job := range jobs {
		task := map[string]string{
			"id":   job.ID().String(),
			"name": job.Name(),
		}

		if lr, err := job.LastRun(); err == nil {
			task["lastRun"] = lr.UTC().String()
		}

		if nr, err := job.NextRun(); err == nil {
			task["nextRun"] = nr.UTC().String()
		}

		tasks = append(tasks, task)
	}

	return tasks
}

// Returns the WorkerID
func (w *Worker) GetManagerID() string {
	return w.config.id
}

// Checks current state on a workload
func (w *Worker) stateCheck(host string) func(context.Context) {
	return func(ctx context.Context) {
		mo, exists := w.workloads[host]
		if !exists {
			return
		}

		w.failMu.Lock()
		defer w.failMu.Unlock()

		if err := mo.Object.Ping(ctx); err != nil {
			w.EventCh <- *NewWorkloadUnreachableEvent(w.config.id, host)
			w.failCounter[host] += 1
		} else {
			w.EventCh <- *NewWorkloadReachableEvent(w.config.id, host)
			delete(w.failCounter, host)
		}
	}
}

// Collects all workloads which has exceeded their failure thresholds
// and removes them from the worker
func (w *Worker) garbageCollector() {
	w.failMu.Lock()
	defer w.failMu.Unlock()

	for host, counter := range w.failCounter {
		if counter >= w.config.maxPingDown {
			delete(w.failCounter, host)

			if err := w.DeleteWorkload(host); err != nil && w.config.errCb != nil {
				w.config.errCb(err)
			}
		}
	}
}

// Executes all of the event callbacks on any event received from the worker
func (w *Worker) eventLoop() {
	for {
		select {
		case e := <-w.EventCh:
			for _, fn := range w.config.eventCbs {
				go fn(w.ctx, e)
			}
		case <-w.ctx.Done():
			return
		}
	}
}

// An event callback that makes sure state is up to date
func (w *Worker) stateUpdateCb(ctx context.Context, e Event) {
	switch e.EventType {
	case EventInitialized:
		w.sr.RegisterWorkload(e.WorkloadName, w.config.id)
	case EventDead:
		w.sr.DeleteWorkload(e.WorkloadName, w.config.id)
	case EventReachable:
		w.sr.UpdateWorkload(e.WorkloadName, w.config.id)
	}
}
