package manager

import (
	"sort"
)

const DELTA_MAX = 5

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

				workload.SetState(StateDistributing)
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

		workloadIds, err := worker.Unload(delta - DELTA_MAX)
		if err != nil {
			break
		}

		m.workloadsMu.RLock()
		m.distributionsMu.Lock()
		for _, workloadId := range workloadIds {
			wl, ok := m.workloads[workloadId]
			if !ok {
				continue
			}
			delete(m.distributions, workloadId)
			wl.SetState(StateInit)

		}
		m.workloadsMu.RUnlock()
		m.distributionsMu.Unlock()
	}
}

// requires, but does not acquire an RLock on m.workersMu
func (m *Manager) sort() (string, string, uint32) {
	m.distributionsMu.RLock()
	defer m.distributionsMu.RUnlock()

	counters := make(map[string]uint32, len(m.workers))
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
		Value uint32
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
