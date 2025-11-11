package manager

import "fmt"

type State uint8

const (
	StateInit State = iota
	StateDistributing
	StateRunning
	StateDown
	StateErr
)

func (s State) String() string {
	switch s {
	case StateInit:
		return "initializing"
	case StateDistributing:
		return "distributing"
	case StateRunning:
		return "running"
	case StateDown:
		return "down"
	case StateErr:
		return "error"
	default:
		return "invalid"
	}
}

func (s State) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == "invalid" {
		return []byte{}, fmt.Errorf("invalid state")
	}

	return []byte(`"` + str + `"`), nil
}

func (s *State) UnmarshalJSON(val []byte) error {
	switch string(val) {
	case `"initializing"`:
		*s = StateInit
	case `"distributing"`:
		*s = StateDistributing
	case `"running"`:
		*s = StateRunning
	case `"down"`:
		*s = StateDown
	case `"error"`:
		*s = StateErr
	default:
		return fmt.Errorf("invalid state")
	}

	return nil
}
