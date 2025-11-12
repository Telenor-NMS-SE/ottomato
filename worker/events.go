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
)

func (e EventType) String() string {
	switch e {
	case EventInitialized:
		return "EventInitialized"
	case EventUnreachable:
		return "EventUnreachable"
	case EventReachable:
		return "EventReachable"
	case EventDead:
		return "EventDead"
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
	case "`EventInitialized`":
		*e = EventInitialized
	case "`EventUnreachable`":
		*e = EventUnreachable
	case "`EventReachable`":
		*e = EventReachable
	case "`EventDead`":
		*e = EventDead
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

func NewWorkloadInitiatedEvent(workerId string, workloadName string) *Event {
	return &Event{
		EventType:    EventInitialized,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      MsgWorkloadInitiated,
	}
}

func NewWorkloadReachableEvent(workerId string, workloadName string) *Event {
	return &Event{
		EventType:    EventReachable,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      MsgWorkloadReachable,
	}
}

func NewWorkloadUnreachableEvent(workerId string, workloadName string) *Event {
	return &Event{
		EventType:    EventUnreachable,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      MsgWorkloadUnreachable,
	}
}

func NewWorkloadDeadEvent(workerId string, workloadName string) *Event {
	return &Event{
		EventType:    EventDead,
		Worker:       workerId,
		WorkloadName: workloadName,
		Message:      MsgWorkloadDead,
	}
}
