package manager

import "fmt"

type Status uint8

const (
	StatusInit Status = iota
	StatusDistributing
	StatusRunning
	StatusDown
	StatusErr
)

func (s Status) String() string {
	switch s {
	case StatusInit:
		return "initializing"
	case StatusDistributing:
		return "distributing"
	case StatusRunning:
		return "running"
	case StatusDown:
		return "down"
	case StatusErr:
		return "error"
	default:
		return "invalid"
	}
}

func (s Status) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == "invalid" {
		return []byte{}, fmt.Errorf("invalid state")
	}

	return []byte(`"` + str + `"`), nil
}

func (s *Status) UnmarshalJSON(val []byte) error {
	switch string(val) {
	case `"initializing"`:
		*s = StatusInit
	case `"distributing"`:
		*s = StatusDistributing
	case `"running"`:
		*s = StatusRunning
	case `"down"`:
		*s = StatusDown
	case `"error"`:
		*s = StatusErr
	default:
		return fmt.Errorf("invalid state")
	}

	return nil
}
