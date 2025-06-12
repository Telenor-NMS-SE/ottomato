package worker

type Event struct {
	EventType    EventType `json:"eventType"`
	Worker       string    `json:"manager"`
	WorkloadName string    `json:"workloadName"`
	Message      string    `json:"msg"`
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
