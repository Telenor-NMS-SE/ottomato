package manager

import (
	"context"
	"time"
)

type Option func(*Manager)

// Set an explicit Manager ID, default: UUID
func WithManagerID(id string) Option {
	return func(m *Manager) {
		m.id = id
	}
}

// Add an event callback
func WithEventCallback(cb func(context.Context, *Event)) Option {
	return func(m *Manager) {
		m.eventCbs = append(m.eventCbs, cb)
	}
}

func WithDistributorInterval(t time.Duration) Option {
	return func(m *Manager) {
		m.distributionInterval = t
	}
}

func WithRebalanceInterval(t time.Duration) Option {
	return func(m *Manager) {
		m.rebalanceInterval = t
	}
}

func WithDistributionCleanupInterval(t time.Duration) Option {
	return func(m *Manager) {
		m.distributionCleanupInterval = t
	}
}

func WithMaxDistributionTime(t time.Duration) Option {
	return func(m *Manager) {
		m.distributionMaxTime = t
	}
}
