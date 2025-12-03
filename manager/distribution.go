package manager

import (
	"sort"
	"time"
)

const DELTA_MAX = 5

// Assign a workload to a worker, commonly used to backfill
// when a manager is started post worker startup. Bypasses
// distribution steps for the workload.
func (m *Manager) Assign(w Worker, wl Workload) {
	m.state.Lock()
	defer m.state.Unlock()

	wl.SetStatus(StatusRunning)

	m.state.AddWorkload(wl)
	m.state.Associate(wl, w)
}

func (m *Manager) cleanup() {
	m.state.Lock()
	defer m.state.Unlock()

	for _, wl := range m.state.GetAllWorkloads() {
		if time.Since(wl.LastStatusChange()) < m.cleanupMaxTime {
			continue
		}

		switch wl.GetStatus() {
		case StatusDistributing:
			wl.SetStatus(StatusErr)
			m.state.UpdateWorkload(wl)

			if w, ok := m.state.GetAssociation(wl); ok {
				m.state.Disassociate(wl, w)
			}
		case StatusErr:
			wl.SetStatus(StatusInit)
			m.state.UpdateWorkload(wl)
		default:
			continue
		}

	}
}

func (m *Manager) distributor() {
	m.state.Lock()
	defer m.state.Unlock()

	for _, wl := range m.state.GetAllWorkloads() {
		if wl.GetStatus() != StatusInit {
			continue
		}

		counters := map[string]int{}
		for _, w := range m.state.GetAllWorkers() {
			counters[w.GetID()] = len(m.state.GetAssociations(w))
		}

		lo, _, _ := m.sort(counters)
		if w, ok := m.state.GetWorker(lo); ok {
			if err := w.Load(wl); err != nil {
				if m.eventCh != nil {
					m.eventCh <- NewWorkloadDistributedErrorEvent(m.id, w.GetID(), wl)
				}
			} else {
				if m.eventCh != nil {
					m.eventCh <- NewWorkloadDistributedEvent(m.id, w.GetID(), wl)
				}

				m.state.Associate(wl, w)

				wl.SetStatus(StatusRunning)
				m.state.UpdateWorkload(wl)
			}
		}
	}
}

func (m *Manager) rebalance() {
	m.state.Lock()
	defer m.state.Unlock()

	for {
		counters := map[string]int{}
		for _, w := range m.state.GetAllWorkers() {
			counters[w.GetID()] = len(m.state.GetAssociations(w))
		}

		_, hi, delta := m.sort(counters)
		if delta <= DELTA_MAX {
			break
		}

		w, ok := m.state.GetWorker(hi)
		if !ok {
			continue
		}

		wls := m.state.GetAssociations(w)
		sort.Slice(wls, func(i, j int) bool {
			return wls[i].LastStatusChange().Before(wls[i].LastStatusChange())
		})

		for i := 0; i <= (delta - DELTA_MAX); i++ {
			wl := wls[i]

			if err := w.Unload(wl); err != nil {
				continue
			}

			m.state.Disassociate(wl, w)

			wl.SetStatus(StatusInit)
			m.state.UpdateWorkload(wl)
		}
	}
}

// requires, but does not acquire an RLock on m.workersMu
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

