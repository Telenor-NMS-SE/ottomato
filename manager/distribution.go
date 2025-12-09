package manager

import (
	"context"
	"sort"
	"time"
)

const DELTA_MAX = 5

// Assign a workload to a worker, commonly used to backfill
// when a manager is started post worker startup. Bypasses
// distribution steps for the workload.
func (m *Manager) Assign(w Worker, wl Workload) {
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	m.state.Lock()
	defer m.state.Unlock()

	wl.SetStatus(StatusRunning)

	if err := m.state.AddWorkload(ctx, wl); err != nil {
		m.signal.Error(err)
	}

	if err := m.state.Associate(ctx, wl, w); err != nil {
		m.signal.Error(err)
	}
}

func (m *Manager) cleanup() {
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	m.state.Lock()
	defer m.state.Unlock()

	workloads, err := m.state.GetAllWorkloads(ctx)
	if err != nil {
		m.signal.Error(err)
		return
	}

	for _, wl := range workloads {
		if time.Since(wl.LastStatusChange()) < m.cleanupMaxTime {
			continue
		}

		switch wl.GetStatus() {
		case StatusDistributing:
			wl.SetStatus(StatusErr)
			if err := m.state.UpdateWorkload(ctx, wl); err != nil {
				m.signal.Error(err)
				continue
			}

			w, err := m.state.GetAssociation(ctx, wl)
			if err != nil {
				m.signal.Error(err)
				continue
			}

			if err := m.state.Disassociate(ctx, wl, w); err != nil {
				m.signal.Error(err)
				continue
			}
		case StatusErr:
			wl.SetStatus(StatusInit)
			if err := m.state.UpdateWorkload(ctx, wl); err != nil {
				m.signal.Error(err)
			}
		default:
			continue
		}

	}
}

func (m *Manager) distributor() {
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	m.state.Lock()
	defer m.state.Unlock()

	workloads, err := m.state.GetAllWorkloads(ctx)
	if err != nil {
		m.signal.Error(err)
		return
	}

	for _, wl := range workloads {
		if wl.GetStatus() != StatusInit {
			continue
		}

		workers, err := m.state.GetAllWorkers(ctx)
		if err != nil {
			m.signal.Error(err)
			continue
		}

		counters := map[string]int{}
		for _, w := range workers {
			assocs, err := m.state.GetAssociations(ctx, w)
			if err != nil {
				m.signal.Error(err)
				continue
			}

			counters[w.GetID()] = len(assocs)
		}

		lo, _, _ := m.sort(counters)
		w, err := m.state.GetWorker(ctx, lo)
		if err != nil {
			m.signal.Error(err)
			continue
		}

		if err := w.Load(wl); err != nil {
			m.signal.Event(NewWorkloadDistributedErrorEvent(m.id, w.GetID(), wl))
			m.signal.Error(err)
			continue
		}

		m.signal.Event(NewWorkloadDistributedEvent(m.id, w.GetID(), wl))
		if err := m.state.Associate(ctx, wl, w); err != nil {
			m.signal.Error(err)
			continue
		}
		wl.SetStatus(StatusRunning)
		if err := m.state.UpdateWorkload(ctx, wl); err != nil {
			m.signal.Error(err)
			continue
		}
	}
}

func (m *Manager) rebalance() {
	ctx, cancel := context.WithTimeout(m.ctx, 5*time.Second)
	defer cancel()

	m.state.Lock()
	defer m.state.Unlock()

	for {
		workers, err := m.state.GetAllWorkers(ctx)
		if err != nil {
			m.signal.Error(err)
			break // no continue (possible infinite loop)
		}

		counters := map[string]int{}
		for _, w := range workers {
			assocs, err := m.state.GetAssociations(ctx, w)
			if err != nil {
				m.signal.Error(err)
				continue
			}

			counters[w.GetID()] = len(assocs)
		}

		_, hi, delta := m.sort(counters)
		if delta <= DELTA_MAX {
			break
		}

		w, err := m.state.GetWorker(ctx, hi)
		if err != nil {
			m.signal.Error(err)
			continue
		}

		workloads, err := m.state.GetAssociations(ctx, w)
		if err != nil {
			m.signal.Error(err)
			continue
		}

		sort.Slice(workloads, func(i, j int) bool {
			return workloads[i].LastStatusChange().Before(workloads[i].LastStatusChange())
		})

		for i := 0; i <= (delta - DELTA_MAX); i++ {
			wl := workloads[i]

			if err := w.Unload(wl); err != nil {
				continue
			}

			if err := m.state.Disassociate(ctx, wl, w); err != nil {
				m.signal.Error(err)
				continue
			}

			wl.SetStatus(StatusInit)
			if err := m.state.UpdateWorkload(ctx, wl); err != nil {
				m.signal.Error(err)
				continue
			}
		}
	}
}

func (m *Manager) sort(counters map[string]int) (string, string, int) {
	type tmp struct {
		Key   string
		Value int
	}

	s := make([]tmp, 0, len(counters))
	for k, v := range counters {
		s = append(s, tmp{Key: k, Value: v})
	}

	sort.Slice(s, func(i, j int) bool {
		return s[i].Value < s[j].Value
	})

	if len(s) > 0 {
		return s[0].Key, s[len(s)-1].Key, s[len(s)-1].Value - s[0].Value
	}

	return "", "", 0
}
