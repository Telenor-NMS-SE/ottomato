package manager

import (
	"bytes"
	"testing"
)

func TestString(t *testing.T) {
	cases := map[State]string{
		StateInit:         "initializing",
		StateDistributing: "distributing",
		StateRunning:      "running",
		StateDown:         "down",
		StateErr:          "error",
		State(8):          "invalid",
	}

	for input, exp := range cases {
		if recv := input.String(); exp != recv {
			t.Errorf("expected string '%s', but got: %s", exp, recv)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	cases := map[State][]byte{
		StateInit:         []byte(`"initializing"`),
		StateDistributing: []byte(`"distributing"`),
		StateRunning:      []byte(`"running"`),
		StateDown:         []byte(`"down"`),
		StateErr:          []byte(`"error"`),
		State(8):          {},
	}

	for input, exp := range cases {
		state := input.String()

		recv, err := input.MarshalJSON()
		if err != nil && state != "invalid" {
			t.Errorf("unexpected error when serializing value: %v", err)
			continue
		}

		if !bytes.Equal(exp, recv) {
			t.Errorf("expected serialized value to be '%s', but got: %s", exp, recv)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	cases := map[string]State{
		`"initializing"`: StateInit,
		`"distributing"`: StateDistributing,
		`"running"`:      StateRunning,
		`"down"`:         StateDown,
		`"error"`:        StateErr,
	}

	for input, exp := range cases {
		var recv State
		err := recv.UnmarshalJSON([]byte(input))
		if err != nil {
			t.Errorf("unexpected error when deserializing state: %v", err)
			continue
		}

		if exp != recv {
			t.Errorf("expected to get state '%s', but got: %s", exp, recv)
		}
	}

	var recv State
	err := recv.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Errorf("expected to receive an error when deserializing an invalid state")
	}
}
