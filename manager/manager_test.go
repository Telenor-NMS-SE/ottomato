package manager

import (
	"context"
	"testing"
)

func TestManager(t *testing.T) {
	mgr, err := New(context.Background())
	if err != nil {
		t.Fatalf("Creation of manager failed fatally")
	}
	if mgr == nil {
		t.Fatalf("Creation of manager returned nil")
	}
	mgr.Stop()
}
