package graphdb

import "testing"

func TestExecuteSimpleMatch_NodePatternWithSemicolon(t *testing.T) {
	e := NewInMemoryEngine()
	if _, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	result, err := executeSimpleMatch(e, "  MATCH (n:User);  ")
	if err != nil {
		t.Fatalf("execute simple match failed: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got: %d", len(result.Rows))
	}
	node, ok := result.Rows[0]["n"].(Node)
	if !ok {
		t.Fatalf("expected Node type, got: %T", result.Rows[0]["n"])
	}
	if got := node.Properties["name"]; got != "alice" {
		t.Fatalf("unexpected node property: %v", got)
	}
}

func TestExecuteSimpleMatch_SingleHopPattern(t *testing.T) {
	e := NewInMemoryEngine()
	alice, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	bob, err := e.CreateNode([]string{"User"}, map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	project, err := e.CreateNode([]string{"Project"}, map[string]any{"name": "study"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	if _, err := e.CreateEdge(alice, bob, "FOLLOWS", nil); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if _, err := e.CreateEdge(alice, project, "FOLLOWS", nil); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	result, err := executeSimpleMatch(e, "MATCH (a:User)-[:FOLLOWS]->(b:User)")
	if err != nil {
		t.Fatalf("execute simple match failed: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got: %d", len(result.Rows))
	}
	left := result.Rows[0]["a"].(Node)
	right := result.Rows[0]["b"].(Node)
	if left.ID != alice || right.ID != bob {
		t.Fatalf("unexpected matched pair: a=%d b=%d", left.ID, right.ID)
	}
}

func TestExecuteSimpleMatch_InvalidQuery(t *testing.T) {
	e := NewInMemoryEngine()
	if _, err := executeSimpleMatch(e, "MATCH (n)"); err != ErrInvalidMatchQuery {
		t.Fatalf("expected ErrInvalidMatchQuery, got: %v", err)
	}
}

func TestHasLabel(t *testing.T) {
	node := Node{
		ID:     1,
		Labels: []string{"User", "Person"},
	}
	if !hasLabel(node, "User") {
		t.Fatal("expected label User to be found")
	}
	if hasLabel(node, "Project") {
		t.Fatal("did not expect label Project to be found")
	}
}
