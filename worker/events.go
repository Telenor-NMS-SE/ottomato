package worker

import (
	"encoding/json"
	"errors"
)

type Event struct {
	EventType    EventType `json:"eventType"`
	Worker       string    `json:"manager"`
	WorkloadName string    `json:"managedObject"`
	Message      string    `json:"message"`
}

type EventType int

const (
	EventInitialized EventType = iota
	EventUnreachable
	EventReachable
	EventDead
	EventAdded
	EventDeleted
	EventInitErr
)

func (e EventType) String() string {
	switch e {
	case EventInitialized:
		return "workload.initialized"
	case EventUnreachable:
		return "workload.unreachable"
	case EventReachable:
		return "workload.reachable"
	case EventDead:
		return "workload.dead"
	case EventAdded:
		return "workload.added"
	case EventDeleted:
		return "workload.deleted"
	case EventInitErr:
		return "workload.init.error"
	default:
		return ""
	}
}

func (e EventType) MarshalJSON() ([]byte, error) {
	str := e.String()
	if str == "" {
		return []byte{}, errors.New("invalid event type")
	}

	return json.Marshal(str)
}

func (e *EventType) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case "`workload.added`":
		*e = EventAdded
	case "`workload.initialized`":
		*e = EventInitialized
	case "`workload.unreachable`":
		*e = EventUnreachable
	case "`workload.reachable`":
		*e = EventReachable
	case "`workload.dead`":
		*e = EventDead
	case "`workload.init.error`":
		*e = EventInitErr
	default:
		return errors.New("invalid event type")
	}

	return nil
}

const (
	MsgWorkloadInitiated   = "workload initiated"
	MsgWorkloadReachable   = "workload reachable"
	MsgWorkloadUnreachable = "workload unreachable"
	MsgWorkloadDead        = "workload unresponsive"
)

func NewWorkloadAddedEvent(workerId string, workloadName string) Event {
	return Event{
		EventType:    EventAdded,
		Worker:       workerId,
		WorkloadName: workloadName,
	}
}

func NewWorkloadDeletedEvent(workerId string, workloadName string) Event {
	return Event{
		EventType:    EventDeleted,
		Worker:       workerId,
		WorkloadName: workloadName,
	}
}

func NewWorkloadInitiatedEvent(workerId string, workloadName string) Event {
	return Event{
		EventType:    EventInitialized,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      MsgWorkloadInitiated,
	}
}

func NewWorkloadReachableEvent(workerId string, workloadName string) Event {
	return Event{
		EventType:    EventReachable,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      MsgWorkloadReachable,
	}
}

func NewWorkloadUnreachableEvent(workerId string, workloadName string) Event {
	return Event{
		EventType:    EventUnreachable,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      MsgWorkloadUnreachable,
	}
}

func NewWorkloadDeadEvent(workerId string, workloadName string) Event {
	return Event{
		EventType:    EventDead,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      MsgWorkloadDead,
	}
}

func NewWorkloadInitError(workerId string, workloadName string, msg string) Event {
	return Event{
		EventType:    EventInitErr,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      msg,
	}
}
