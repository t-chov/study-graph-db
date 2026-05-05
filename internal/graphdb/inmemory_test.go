package graphdb

import "testing"

func TestNewInMemoryEngine(t *testing.T) {
	e := NewInMemoryEngine()
	if e == nil {
		t.Fatal("engine must not be nil")
	}
}

func TestScaffoldReturnsNotImplemented(t *testing.T) {
	e := NewInMemoryEngine()
	if _, err := e.CreateNode(nil, nil); err != ErrNotImplemented {
		t.Fatalf("expected ErrNotImplemented, got: %v", err)
	}
}
