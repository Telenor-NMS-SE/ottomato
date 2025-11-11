package manager

import (
	"context"
	"sync"

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
}

func New(ctx context.Context, opts ...Option) (*Manager, error) {
	mgr := &Manager{
		id:  uuid.NewString(),
		ctx: ctx,
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
