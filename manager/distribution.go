package manager

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
		if worker, ok := m.workers[low]; ok {
			if err := worker.Load(workload); err != nil {
				m.eventCh <- NewWorkloadDistributedErrorEvent(m.id, worker.GetID(), workload)
			} else {
				m.eventCh <- NewWorkloadDistributedEvent(m.id, worker.GetID(), workload)
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
