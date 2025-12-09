package manager

import "log/slog"

type Slog interface {
	Info(string, ...any)
	Error(string, ...any)
}

type SlogSignaller struct {
	instance Slog
}

func NewSlogSignaller(instance Slog) *SlogSignaller {
	s := &SlogSignaller{instance}
	if s.instance == nil {
		s.instance = slog.Default()
	}

	return s
}

func (s *SlogSignaller) Event(e Event) {
	s.instance.Info("received event", "type", e.Type.String(), "managerId", e.ManagerID, "workerId", e.WorkerID, "resourceId", e.ResourceID)
}

func (s *SlogSignaller) Error(err error) {
	s.instance.Error("received error", "error", err)
}
