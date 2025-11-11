package manager

import "fmt"
import "sort"

func (m *Manager) distributor() {
	m.workersMu.RLock()
	m.workloadsMu.RLock()
	defer m.workersMu.RUnlock()
	defer m.workloadsMu.RUnlock()

	for _, workload := range m.workloads {
		if workload.GetState() != StateInit {
			continue
		}

		low, _, _ := m.ChatGPTSortDelta()
		fmt.Printf("Low is: %s\n", low)
		if worker, ok := m.workers[low]; ok {
			if err := worker.Load(workload); err != nil {
				m.eventCh <- NewWorkloadDistributedErrorEvent(m.id, worker.GetID(), workload)
			} else {
				m.eventCh <- NewWorkloadDistributedEvent(m.id, worker.GetID(), workload)
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
		_, hi, delta := m.ChatGPTSortDelta()
		if delta <= 5 {
			break
		}

		worker, ok := m.workers[hi]
		if !ok {
			continue
		}

		workloadIds, err := worker.Unload(delta)
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

// Use with caution: This is copy-pasted!
func (m *Manager) ChatGPTSortDelta() (string, string, uint32) {
	m.distributionsMu.RLock()
	defer m.distributionsMu.RUnlock()

	counts := make(map[string]uint32)
	for _, value := range m.distributions {
		counts[value]++
	}

	var minVal, maxVal string
	minCount, maxCount := uint32(^uint32(0)>>1), uint32(0)

	for val, count := range counts {
		if count < minCount {
			minCount = count
			minVal = val
		}
		if count > maxCount {
			maxCount = count
			maxVal = val
		}
	}

	return maxVal, minVal, maxCount - minCount
}

func (m *Manager) sort() (string, string, uint32) {
	counters := map[string]uint32{}
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