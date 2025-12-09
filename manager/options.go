package manager

import (
	"time"
)

type Option func(*Manager)

// Set an explicit Manager ID, default: UUID
func WithManagerID(id string) Option {
	return func(m *Manager) {
		m.id = id
	}
}

func WithSignaller(signaller Signals) Option {
	return func(m *Manager) {
		m.signal = signaller
	}
}

func WithStateStorage(state StateStorage) Option {
	return func(m *Manager) {
		m.state = state
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

func WithCleanupInterval(t time.Duration) Option {
	return func(m *Manager) {
		m.cleanupInterval = t
	}
}

func WithCleanupMaxTime(t time.Duration) Option {
	return func(m *Manager) {
		m.cleanupMaxTime = t
	}
}
