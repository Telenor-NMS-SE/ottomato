package manager

import (
	"context"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

type Manager struct {
	id        string
	ctx       context.Context
	scheduler gocron.Scheduler

	signal Signals
	state  StateStorage

	// used to guarantee exclusivity between dist and rebalance
	mainJobMu sync.Mutex

	distributionInterval time.Duration
	rebalanceInterval    time.Duration
	cleanupInterval      time.Duration
	cleanupMaxTime       time.Duration // Max time a workload can be in a errornous state

	// Distribution timeout is not needed after gocron update
	// which gives us access to singleton job which sets its
	// next run interval when the current job is finished.
	// This context timeout has been troublesome, as doing
	// distributions over the network means that a context
	// timeout can be exceeded, but we're still distributing
	// workloads to workers, which leads to an uncertain state.
	distributionTimeout time.Duration

	maxDelta int // Max allowed delta for workers' distributed workloads
}

type Signals interface {
	Event(Event)
	Error(error)
}

type StateStorage interface {
	Lock()
	Unlock()

	GetAllWorkers(context.Context) ([]Worker, error)
	GetWorker(context.Context, string) (Worker, error)
	AddWorker(context.Context, Worker) error
	DeleteWorker(context.Context, Worker) error

	GetAllWorkloads(context.Context) ([]Workload, error)
	GetWorkload(context.Context, string) (Workload, error)
	AddWorkload(context.Context, Workload) error
	UpdateWorkload(context.Context, Workload) error
	DeleteWorkload(context.Context, Workload) error

	GetAssociation(context.Context, Workload) (Worker, error)
	GetAssociations(context.Context, Worker) ([]Workload, error)
	Associate(context.Context, Workload, Worker) error
	Disassociate(context.Context, Workload, Worker) error
}

type ctxScope string

const ctxScopeKey ctxScope = "scope"

func New(ctx context.Context, opts ...Option) (*Manager, error) {
	mgr := &Manager{
		id:  uuid.NewString(),
		ctx: context.WithValue(ctx, ctxScopeKey, "local"),

		distributionInterval: time.Minute,
		rebalanceInterval:    time.Minute,
		cleanupInterval:      5 * time.Minute,
		cleanupMaxTime:       5 * time.Minute,

		maxDelta: 5,
	}

	for _, opt := range opts {
		opt(mgr)
	}

	if mgr.signal == nil {
		mgr.signal = NewSlogSignaller(nil)
	}

	if mgr.state == nil {
		mgr.state = NewMemoryStore()
	}

	var err error
	mgr.scheduler, err = gocron.NewScheduler(
		gocron.WithLimitConcurrentJobs(10, gocron.LimitModeReschedule),
	)
	if err != nil {
		return mgr, err
	}

	// Add scheduled job for (re)distribution of workloads
	if _, err := mgr.scheduler.NewJob(
		gocron.DurationJob(mgr.distributionInterval),
		gocron.NewTask(mgr.distributor),
		gocron.WithContext(ctx),
		gocron.WithIntervalFromCompletion(),
	); err != nil {
		return mgr, err
	}

	// Add scheduled job for rebalancing workloads on workers
	if _, err := mgr.scheduler.NewJob(
		gocron.DurationJob(mgr.rebalanceInterval),
		gocron.NewTask(mgr.rebalance),
		gocron.WithContext(ctx),
		gocron.WithIntervalFromCompletion(),
	); err != nil {
		return mgr, err
	}

	// Add scheduled job for workloads stuck in distributing?
	if _, err := mgr.scheduler.NewJob(
		gocron.DurationJob(mgr.cleanupInterval),
		gocron.NewTask(mgr.cleanup),
		gocron.WithContext(ctx),
	); err != nil {
		return mgr, err
	}

	mgr.scheduler.Start()

	return mgr, nil
}

func (m *Manager) Stop() error {
	m.ctx.Done()
	return m.scheduler.Shutdown()
}
