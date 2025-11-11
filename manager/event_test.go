package manager

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestEventTypeSerialize(t *testing.T) {
	cases := map[EventType][]byte{
		EventWorkerAdded:     []byte(`"worker.added"`),
		EventWorkerDeleted:   []byte(`"worker.deleted"`),
		EventWorkloadAdded:   []byte(`"workload.added"`),
		EventWorkloadDeleted: []byte(`"workload.deleted"`),
	}

	for input, exp := range cases {
		recv, err := json.Marshal(input)
		if err != nil {
			t.Fatalf("unexpected error when serializing event type: %v", err)
		}

		if !bytes.Equal(exp, recv) {
			t.Errorf("expected to get '%s', but got: '%s'", exp, recv)
		}
	}
}

func TestEventTypeSerializeError(t *testing.T) {
	_, err := json.Marshal(EventType(-1))
	if err == nil {
		t.Fatalf("expected an error, but got nothing")
	}

	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected an ErrInvalidEvent, but got: %v", err)
	}
}

func TestEventTypeDeserialize(t *testing.T) {
	cases := map[string]EventType{
		`"worker.added"`:     EventWorkerAdded,
		`"worker.deleted"`:   EventWorkerDeleted,
		`"workload.added"`:   EventWorkloadAdded,
		`"workload.deleted"`: EventWorkloadDeleted,
	}

	for input, exp := range cases {
		var recv EventType
		if err := json.Unmarshal([]byte(input), &recv); err != nil {
			t.Fatalf("received an unexpected error: %v", err)
		}

		if exp != recv {
			t.Errorf("expected to get '%s', but got: '%s'", exp, recv)
		}
	}
}

func TestEventTypeDeserializeError(t *testing.T) {
	var recv EventType
	err := json.Unmarshal([]byte(`"unknown"`), &recv)
	if err == nil {
		t.Fatalf("expected an error, but got nothing")
	}

	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected an ErrInvalidEvent, but got: %v", err)
	}
}
