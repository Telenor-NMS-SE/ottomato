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
	m.workloadsMu.Lock()
	defer m.workloadsMu.Unlock()

	m.distributionsMu.Lock()
	defer m.distributionsMu.Unlock()

	m.workloads[wl.GetID()] = wl

	m.distributions[wl.GetID()] = w.GetID()
	wl.SetState(StateRunning)
}

func (m *Manager) cleanup() {
	m.workloadsMu.RLock()
	defer m.workloadsMu.RUnlock()

	for _, wl := range m.workloads {

		if time.Since(wl.LastStateChange()) < m.cleanupMaxTime {
			continue
		}

		switch wl.GetState() {
		case StateDistributing:
			wl.SetState(StateErr)
			m.distributionsMu.Lock()
			delete(m.distributions, wl.GetID())
			m.distributionsMu.Unlock()
		case StateErr:
			wl.SetState(StateInit)
		default:
			continue
		}

	}
}

func (m *Manager) distributor() {
	m.workloadsMu.RLock()
	defer m.workloadsMu.RUnlock()

	m.workersMu.RLock()
	defer m.workersMu.RUnlock()

	for _, workload := range m.workloads {
		if workload.GetState() != StateInit {
			continue
		}

		lo, _, _ := m.sort()
		if worker, ok := m.workers[lo]; ok {
			if err := worker.Load(workload); err != nil {
				if m.eventCh != nil {
					m.eventCh <- NewWorkloadDistributedErrorEvent(m.id, worker.GetID(), workload)
				}
			} else {
				if m.eventCh != nil {
					m.eventCh <- NewWorkloadDistributedEvent(m.id, worker.GetID(), workload)
				}

				m.distributionsMu.Lock()
				m.distributions[workload.GetID()] = worker.GetID()
				m.distributionsMu.Unlock()

				workload.SetState(StateRunning)
			}
		}
	}
}

func (m *Manager) rebalance() {
	m.workersMu.RLock()
	defer m.workersMu.RUnlock()

	for {

		_, hi, delta := m.sort()
		if delta <= DELTA_MAX {
			break
		}

		worker, ok := m.workers[hi]
		if !ok {
			continue
		}

		workloads := m.getRelatedWorkloads(worker)
		sort.Slice(workloads, func(i, j int) bool {
			return workloads[i].LastStateChange().Before(workloads[i].LastStateChange())
		})

		m.workloadsMu.RLock()
		defer m.workloadsMu.RUnlock()

		m.distributionsMu.Lock()
		defer m.distributionsMu.Unlock()

		for i := 0; i <= (delta - DELTA_MAX); i++ {
			wl := workloads[i]

			if err := worker.Unload(workloads[i]); err != nil {
				continue
			}

			delete(m.distributions, wl.GetID())
			wl.SetState(StateInit)
		}
	}
}

// requires, but does not acquire an RLock on m.workersMu
func (m *Manager) sort() (string, string, int) {
	m.distributionsMu.RLock()
	defer m.distributionsMu.RUnlock()

	counters := make(map[string]int, len(m.workers))
	for _, workerId := range m.distributions {
		counters[workerId] += 1
	}

	for workerId := range m.workers {
		if _, ok := counters[workerId]; !ok {
			counters[workerId] = 0
		}
	}

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

func (m *Manager) getRelatedWorkloads(w Worker) []Workload {
	m.distributionsMu.RLock()
	defer m.distributionsMu.RUnlock()

	m.workloadsMu.RLock()
	defer m.workloadsMu.RUnlock()

	res := []Workload{}
	for workloadId, workerId := range m.distributions {
		if w.GetID() != workerId {
			continue
		}

		if wl, ok := m.workloads[workloadId]; ok {
			res = append(res, wl)
		}
	}

	return res
}
