package manager

import (
	"encoding/json"
	"errors"
)

type Event struct {
	Type       EventType `json:"type"`
	ManagerID  string    `json:"managerId"`
	ResourceID string    `json:"resourceId"`
}

var (
	ErrInvalidEvent = errors.New("invalid event type")
)

type EventType int

const (
	EventWorkerAdded EventType = iota
	EventWorkerDeleted

	EventWorkloadAdded
	EventWorkloadDeleted
)

func (e EventType) String() string {
	switch e {
	case EventWorkerAdded:
		return "worker.added"
	case EventWorkerDeleted:
		return "worker.deleted"
	case EventWorkloadAdded:
		return "workload.added"
	case EventWorkloadDeleted:
		return "workload.deleted"
	default:
		return ""
	}
}

func (e EventType) MarshalJSON() ([]byte, error) {
	str := e.String()
	if str == "" {
		return []byte{}, ErrInvalidEvent
	}

	return json.Marshal(str)
}

func (e *EventType) UnmarshalJSON(val []byte) error {
	switch string(val) {
	case `"worker.added"`:
		*e = EventWorkerAdded
	case `"worker.deleted"`:
		*e = EventWorkerDeleted
	case `"workload.added"`:
		*e = EventWorkloadAdded
	case `"workload.deleted"`:
		*e = EventWorkloadDeleted
	default:
		return ErrInvalidEvent
	}

	return nil
}

func NewWorkerAddedEvent(managerId string, worker Worker) *Event {
	return &Event{
		Type:       EventWorkerAdded,
		ManagerID:  managerId,
		ResourceID: worker.GetID(),
	}
}

func NewWorkerDeletedEvent(managerId string, worker Worker) *Event {
	return &Event{
		Type:       EventWorkerDeleted,
		ManagerID:  managerId,
		ResourceID: worker.GetID(),
	}
}

func NewWorkloadAddedEvent(managerId string, workload Workload) *Event {
	return &Event{
		Type:       EventWorkloadAdded,
		ManagerID:  managerId,
		ResourceID: workload.GetID(),
	}
}

func NewWorkloadDeletedEvent(managerId string, workload Workload) *Event {
	return &Event{
		Type:       EventWorkloadDeleted,
		ManagerID:  managerId,
		ResourceID: workload.GetID(),
	}
}
