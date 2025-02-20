package worker

type Event struct {
	Msg          string    `json:"msg"`
	Manager      string    `json:"manager"`
	WorkloadName string    `json:"managed_object"`
	EventType    EventType `json:"event_type"`
}

type EventType int

const (
	EventInitialized EventType = iota
	EventUnreachable
	EventReachable
	EventDead
)

func (e EventType) ToString() string {
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

const (
	WorkloadInitiatedMsg   = "workload initiated"
	WorkloadReachableMsg   = "workload reachable"
	WorkloadUnreachableMsg = "workload unreachable"
	WorkloadDeadMsg        = "workload unresponsive"
)

func NewWorkloadInitiatedEvent(mgrId string, moName string) *Event {
	return &Event{
		Manager:      mgrId,
		Msg:          WorkloadInitiatedMsg,
		WorkloadName: moName,
		EventType:    EventInitialized,
	}
}

func NewWorkloadReachableEvent(mgrId string, moName string) *Event {
	return &Event{
		Manager:      mgrId,
		Msg:          WorkloadReachableMsg,
		WorkloadName: moName,
		EventType:    EventReachable,
	}
}

func NewWorkloadUnreachableEvent(mgrId string, moName string) *Event {
	return &Event{
		Manager:      mgrId,
		Msg:          WorkloadUnreachableMsg,
		WorkloadName: moName,
		EventType:    EventUnreachable,
	}
}
func NewWorkloadDeadEvent(mgrId string, moName string) *Event {
	return &Event{
		Manager:      mgrId,
		Msg:          WorkloadDeadMsg,
		WorkloadName: moName,
		EventType:    EventDead,
	}
}
