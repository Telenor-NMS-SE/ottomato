package manager

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"sync"
	"time"
)

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
	ctx, cancel := context.WithCancel(m.ctx)
	defer cancel()

	workers, err := m.state.GetAllWorkers(ctx)
	if err != nil {
		m.signal.Error(fmt.Errorf("failed to distribute: failed to get workers: %w", err))
		return
	}

	var wm = make(map[string]Worker, len(workers))
	var current = make(map[string][]string, len(workers))
	for _, w := range workers {
		wm[w.GetID()] = w

		workloads, err := m.state.GetAssociations(ctx, w)
		if err != nil {
			m.signal.Error(fmt.Errorf("failed to distribute: failed to get worker associations: %w", err))
			return
		}

		current[w.GetID()] = make([]string, 0, len(workloads))
		for _, wl := range workloads {
			current[w.GetID()] = append(current[w.GetID()], wl.GetID())
		}
	}

	workloads, err := m.state.GetAllWorkloads(ctx)
	if err != nil {
		m.signal.Error(fmt.Errorf("failed to distribute: failed to get workloads: %w", err))
		return
	}

	var wlm = make(map[string]Workload, len(workloads))
	wanted := make([]string, 0, len(workloads))
	for _, wl := range workloads {
		wlm[wl.GetID()] = wl
		wanted = append(wanted, wl.GetID())
	}

	deletes := map[string][]string{}
	for w, wls := range current {
		deletes[w] = []string{}

		c := []string{}
		for _, wl := range wls {
			if slices.Contains(wanted, wl) {
				c = append(c, wl)
				continue
			}

			deletes[w] = append(deletes[w], wl)
		}

		current[w] = c
	}

	var wg sync.WaitGroup
	for w, dels := range deletes {
		for _, del := range dels {
			wg.Go(func() {
				if err := ctx.Err(); err != nil {
					m.signal.Error(fmt.Errorf("failed to delete workload '%s' from '%s': %w", del, w, err))
					return
				}

				if err := wm[w].Unload(&workload{id: del}); err != nil {
					m.signal.Error(fmt.Errorf("failed to unload unwanted workload '%s' from '%s': %w", del, w, err))
				}
			})
		}
	}
	wg.Wait()

	distribute := []string{}
outer:
	for _, wl := range wanted {
		for _, wls := range current {
			if slices.Contains(wls, wl) {
				continue outer
			}
		}

		distribute = append(distribute, wl)
	}

	load := make(map[string]int, len(current))
	for w, wls := range current {
		load[w] = len(wls)
	}

	distribution := map[string]string{}
	for _, wl := range distribute {
		var wid = ""
		var min = 999_999_999

		for w, l := range load {
			if l < min || wid == "" {
				wid = w
				min = l
			}
		}

		distribution[wl] = wid
		load[wid] += 1
	}

	for wl, w := range distribution {
		wg.Go(func() {
			if err := ctx.Err(); err != nil {
				m.signal.Error(fmt.Errorf("failed to distribute workload '%s' to '%s': %w", wl, w, err))
				return
			}

			if err := wm[w].Load(wlm[wl]); err != nil {
				m.signal.Error(fmt.Errorf("failed to load workload '%s' on to worker '%s': %w", wl, w, err))
				return
			}

			wlm[wl].SetStatus(StatusRunning)
			if err := m.state.UpdateWorkload(ctx, wlm[wl]); err != nil {
				m.signal.Error(fmt.Errorf("failed to update workload state on '%s' after distribution: %w", wl, err))
			}

			if err := m.state.Associate(ctx, wlm[wl], wm[w]); err != nil {
				m.signal.Error(fmt.Errorf("failed to associate workload '%s' with worker '%s': %w", wl, w, err))
			}
		})
	}
	wg.Wait()
}

func (m *Manager) rebalance() {
	ctx, cancel := context.WithCancel(m.ctx)
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
		if delta <= m.maxDelta {
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

		for i := 0; i <= (delta - m.maxDelta); i++ {
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
