package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"

	"github.com/Telenor-NMS-SE/ottomato/store"
)

type (
	Worker struct {
		ctx    context.Context
		scOpts []gocron.SchedulerOption
		sc     gocron.Scheduler

		sr StateRepository

		EventCh chan Event

		workloadsMu sync.RWMutex
		workloads   map[string]workload

		failMu      sync.Mutex
		failCounter map[string]int

		initQueueCh chan string

		config config
	}
	config struct {
		id          string
		splayHi     time.Duration
		splayLo     time.Duration
		maxPingDown int
		pingTimeout time.Duration
		eventCbs    []func(context.Context, Event)
		errCb       func(error)
	}
)

type (
	Workload interface {
		Init(context.Context) error
		Ping(context.Context) error
		RunTask(context.Context, *Task) (Result, error)
		Stop() error
		Info() map[string]any
		Name() string
	}
	workload struct {
		object    Workload
		jid       uuid.UUID
		createdAt time.Time
		updatedAt time.Time
	}
)

type Task struct {
	Command string         `json:"command"`
	Args    []string       `json:"args"`
	Kwargs  map[string]any `json:"kwargs"`
}

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
	ErrWorkloadNotFound = errors.New("workload does not exist")
	ErrWorkloadExists   = errors.New("workload already exist")
	ErrScheduleCleanup  = errors.New("failed to clean up scheduler")
)

// Create a new worker instance with default options, override with []Option
func New(ctx context.Context, opts ...Option) (*Worker, error) {
	var err error

	worker := &Worker{
		ctx:         ctx,
		workloads:   make(map[string]workload),
		EventCh:     make(chan Event),
		failCounter: map[string]int{},
		config: config{
			id:          uuid.NewString(),
			splayHi:     DEFAULT_SPLAY_HI,
			splayLo:     DEFAULT_SPLAY_LO,
			pingTimeout: DEFAULT_PING_TIMEOUT,
			maxPingDown: DEFAULT_MAX_PINGDOWN,
			eventCbs:    make([]func(context.Context, Event), 0),
		},
		scOpts: []gocron.SchedulerOption{
			gocron.WithLimitConcurrentJobs(10, gocron.LimitModeReschedule),
		},
		initQueueCh: make(chan string, 192),
	}

	worker.config.eventCbs = append(worker.config.eventCbs, worker.stateUpdateCb)

	for _, opt := range opts {
		opt(worker)
	}

	if worker.sr == nil {
		worker.sr = store.New()
	}

	if worker.sc, err = gocron.NewScheduler(worker.scOpts...); err != nil {
		return worker, err
	}

	if _, err = worker.sc.NewJob(
		gocron.DurationJob(10*time.Second),
		gocron.NewTask(worker.garbageCollector),
		gocron.WithName("garbage collector"),
		gocron.WithContext(ctx),
	); err != nil {
		return worker, err
	}

	worker.sc.Start()

	go worker.eventLoop()
	go worker.runInitQueue()

	return worker, nil
}

// Stop the worker
func (w *Worker) Stop() error {
	w.workloadsMu.Lock()
	defer w.workloadsMu.Unlock()

	for k, wl := range w.workloads {
		if err := wl.object.Stop(); err != nil {
			return fmt.Errorf("failed to stop workload '%s': %w", k, err)
		}
	}

	return w.sc.Shutdown()
}

// Run a task on the given target
func (w *Worker) RunTask(ctx context.Context, target string, task *Task) (Result, error) {
	start := time.Now()

	w.workloadsMu.RLock()
	wl, exists := w.workloads[target]
	if !exists {
		return Result{}, ErrWorkloadNotFound
	}
	w.workloadsMu.RUnlock()

	job, err := wl.object.RunTask(ctx, task)

	// add a bit of metadata
	job.Timestamp = start
	job.ExecutionTime = time.Since(start).Milliseconds()
	job.WorkerID = w.config.id

	return job, err
}

// Adds a new workload to the worker
func (w *Worker) AddWorkload(ctx context.Context, wl Workload) (map[string]any, error) {
	w.workloadsMu.Lock()
	defer w.workloadsMu.Unlock()

	now := time.Now()
	meta := map[string]any{
		"name":      wl.Name(),
		"createdAt": now.UTC().String(),
		"updatedAt": now.UTC().String(),
	}

	if _, exists := w.workloads[wl.Name()]; exists {
		return meta, ErrWorkloadExists
	}

	w.workloads[wl.Name()] = workload{
		object:    wl,
		createdAt: now,
		updatedAt: now,
	}

	w.enqueue(wl.Name())

	w.sr.RegisterWorkload(wl.Name(), w.GetWorkerID())
	w.EventCh <- NewWorkloadAddedEvent(w.config.id, wl.Name())

	return meta, nil

	/*
		job, err := w.sc.NewJob(
			gocron.DurationRandomJob(w.config.splayLo, w.config.splayHi),
			gocron.NewTask(w.stateCheck(wl.Name())),
			gocron.WithName(fmt.Sprintf("worker state check: %s", wl.Name())),
			gocron.WithContext(ctx),
		)
		if err != nil {
			return meta, err
		}
	*/
}

// Stops and deletes a workload from the worker
func (w *Worker) DeleteWorkload(name string) (err error) {
	w.workloadsMu.Lock()
	defer w.workloadsMu.Unlock()

	wl, exists := w.workloads[name]
	if !exists {
		return ErrWorkloadNotFound
	}

	if err := wl.object.Stop(); err != nil {
		return err
	}

	// clean up schedule, but do not return on error
	if err := w.sc.RemoveJob(wl.jid); err != nil {
		err = fmt.Errorf("%w: %v", ErrScheduleCleanup, err) //nolint:all
	}

	delete(w.workloads, name)
	w.sr.DeleteWorkload(name, w.GetWorkerID())
	w.EventCh <- NewWorkloadDeadEvent(w.config.id, name)

	return err
}

// This should return []map[string]string with a bunch of metadata
func (w *Worker) Workloads() []map[string]any {
	w.workloadsMu.RLock()
	defer w.workloadsMu.RUnlock()

	keys := make([]map[string]any, 0, len(w.workloads))

	for k, v := range w.workloads {
		keys = append(keys, map[string]any{
			"name":      k,
			"id":        v.jid,
			"createdAt": v.createdAt,
			"updatedAt": v.updatedAt,
		})
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
func (w *Worker) GetWorkerID() string {
	return w.config.id
}

// Checks current state on a workload
func (w *Worker) stateCheck(host string) func(context.Context) {
	return func(ctx context.Context) {
		ctx, cancel := context.WithTimeout(ctx, w.config.pingTimeout)
		defer cancel()

		w.workloadsMu.RLock()
		wl, exists := w.workloads[host]
		if !exists {
			return
		}
		w.workloadsMu.RUnlock()

		//moving ping before lock, can be longrunning dont want to tie upp the mutex
		err := wl.object.Ping(ctx)

		w.failMu.Lock()
		defer w.failMu.Unlock()

		if err != nil {
			w.EventCh <- NewWorkloadUnreachableEvent(w.config.id, host)
			w.failCounter[host] += 1
		} else {
			w.EventCh <- NewWorkloadReachableEvent(w.config.id, host)
			delete(w.failCounter, host)
			w.sr.UpdateWorkload(host, w.config.id)
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
	case EventInitialized, EventAdded:
		//w.sr.RegisterWorkload(e.WorkloadName, w.config.id)
	case EventDead, EventDeleted:
		//w.sr.DeleteWorkload(e.WorkloadName, w.config.id)
	case EventReachable:
		//w.sr.UpdateWorkload(e.WorkloadName, w.config.id)
	}
}

func (w *Worker) enqueue(k string) {
	w.initQueueCh <- k
}

func (w *Worker) runInitQueue() {
	for {
		select {
		case key := <-w.initQueueCh:
			w.workloadsMu.RLock()
			wl, ok := w.workloads[key]
			w.workloadsMu.RUnlock()
			if !ok {
				continue
			}

			ctx, cancel := context.WithTimeout(w.ctx, 25*time.Second)

			// do stuff here
			if err := wl.object.Init(ctx); err != nil {
				w.workloadsMu.Lock()

				if err := wl.object.Stop(); err != nil {
					w.EventCh <- NewWorkloadStopError(w.config.id, wl.object.Name())
				}

				delete(w.workloads, wl.object.Name())
				w.sr.DeleteWorkload(wl.object.Name(), w.GetWorkerID())
				w.EventCh <- NewWorkloadDeletedEvent(w.config.id, wl.object.Name())

				cancel()
				w.workloadsMu.Unlock()
				continue
			}

			job, err := w.sc.NewJob(
				gocron.DurationRandomJob(w.config.splayLo, w.config.splayHi),
				gocron.NewTask(w.stateCheck(wl.object.Name())),
				gocron.WithName(fmt.Sprintf("worker state check: %s", wl.object.Name())),
				gocron.WithContext(w.ctx),
			)
			if err != nil {
				w.EventCh <- NewWorkloadInitError(w.config.id, wl.object.Name(), "failed to schedule state check")
			} else {
				wl.jid = job.ID()

				w.workloadsMu.Lock()
				w.workloads[key] = wl
				w.workloadsMu.Unlock()
			}

			w.EventCh <- NewWorkloadInitiatedEvent(w.config.id, wl.object.Name())
			cancel()
		case <-w.ctx.Done():
			return
		}
	}
}
