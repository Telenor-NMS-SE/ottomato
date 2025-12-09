package manager

import (
	"errors"
	"testing"
)

type mockSlog struct {
	lastMethod  string
	lastMessage string
}

func (s *mockSlog) Info(msg string, args ...any) {
	s.lastMethod = "Info"
	s.lastMessage = msg
}

func (s *mockSlog) Error(msg string, args ...any) {
	s.lastMethod = "Error"
	s.lastMessage = msg
}

func TestEvent(t *testing.T) {
	ms := &mockSlog{}
	s := NewSlogSignaller(ms)

	s.Event(NewWorkloadAddedEvent("test1", &mockWorkload{id: "test1"}))

	if exp, recv := "Info", ms.lastMethod; exp != recv {
		t.Fatalf("expected method '%s' to be called, but got: %s", exp, recv)
	}

	if exp, recv := "received event", ms.lastMessage; exp != recv {
		t.Fatalf("expected last message to be '%s', but got: %s", exp, recv)
	}
}

func TestError(t *testing.T) {
	ms := &mockSlog{}
	s := NewSlogSignaller(ms)

	s.Error(errors.New("some error"))

	if exp, recv := "Error", ms.lastMethod; exp != recv {
		t.Fatalf("expected method '%s' to be called, but got: %s", exp, recv)
	}

	if exp, recv := "received error", ms.lastMessage; exp != recv {
		t.Fatalf("expected last message to be '%s', but got: %s", exp, recv)
	}
}
