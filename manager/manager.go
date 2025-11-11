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

	eventCbs []func(context.Context, *Event)
	eventCh  chan (*Event)

	workersMu sync.RWMutex
	workers   map[string]Worker // worker1-oiDHawoda:   ..

	workloadsMu sync.RWMutex
	workloads   map[string]Workload // workload1-oiIAODWoa: ..

	distributionsMu sync.RWMutex
	distributions   map[string]string // workload1-oiIAODWoa: worker1-oiDHawoda

	distributionInterval time.Duration
	rebalanceInterval    time.Duration
}

func New(ctx context.Context, opts ...Option) (*Manager, error) {
	mgr := &Manager{
		id:                   uuid.NewString(),
		ctx:                  ctx,
		distributionInterval: time.Minute,
		rebalanceInterval:    time.Minute,
	}

	for _, opt := range opts {
		opt(mgr)
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
	); err != nil {
		return mgr, err
	}

	// Add scheduled job for rebalancing workloads on workers
	if _, err := mgr.scheduler.NewJob(
		gocron.DurationJob(mgr.rebalanceInterval),
		gocron.NewTask(mgr.rebalance),
		gocron.WithContext(ctx),
	); err != nil {
		return mgr, err
	}

	// Add scheduled job for workloads stuck in distributing?

	mgr.scheduler.Start()

	go mgr.eventLoop()

	return mgr, nil
}

func (m *Manager) Stop() error {
	return m.scheduler.Shutdown()
}

func (m *Manager) eventLoop() {
	for {
		select {
		case e := <-m.eventCh:
			for _, cb := range m.eventCbs {
				go cb(m.ctx, e)
			}
		case <-m.ctx.Done():
			return
		}
	}
}
