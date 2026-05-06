package graphdb

import "testing"

func TestTxCommitAppliesChanges(t *testing.T) {
	e := NewInMemoryEngine()
	tx, err := e.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}

	alice, err := tx.CreateNode([]string{"User"}, map[string]any{"name": "alice"})
	if err != nil {
		t.Fatalf("tx create node failed: %v", err)
	}
	bob, err := tx.CreateNode([]string{"User"}, map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("tx create node failed: %v", err)
	}
	if _, err := tx.CreateEdge(alice, bob, "FOLLOWS", nil); err != nil {
		t.Fatalf("tx create edge failed: %v", err)
	}

	beforeCommit, err := e.FindNodesByLabel("User")
	if err != nil {
		t.Fatalf("engine find failed: %v", err)
	}
	if len(beforeCommit) != 0 {
		t.Fatalf("expected 0 users before commit, got: %d", len(beforeCommit))
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	afterCommit, err := e.FindNodesByLabel("User")
	if err != nil {
		t.Fatalf("engine find failed: %v", err)
	}
	if len(afterCommit) != 2 {
		t.Fatalf("expected 2 users after commit, got: %d", len(afterCommit))
	}

	edges, err := e.FindEdgesByType("FOLLOWS")
	if err != nil {
		t.Fatalf("find edges failed: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge after commit, got: %d", len(edges))
	}
}

func TestTxRollbackDiscardsChanges(t *testing.T) {
	e := NewInMemoryEngine()
	tx, err := e.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}

	if _, err := tx.CreateNode([]string{"User"}, map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("tx create node failed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	users, err := e.FindNodesByLabel("User")
	if err != nil {
		t.Fatalf("engine find failed: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected 0 users after rollback, got: %d", len(users))
	}
}

func TestTxClosedState(t *testing.T) {
	e := NewInMemoryEngine()
	tx, err := e.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	if err := tx.Commit(); err != ErrTxClosed {
		t.Fatalf("expected ErrTxClosed on second commit, got: %v", err)
	}
	if _, err := tx.CreateNode([]string{"User"}, nil); err != ErrTxClosed {
		t.Fatalf("expected ErrTxClosed for operation after commit, got: %v", err)
	}
}
