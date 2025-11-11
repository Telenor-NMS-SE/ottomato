package manager

import "context"

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
