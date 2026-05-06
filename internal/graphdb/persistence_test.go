package graphdb

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	e := NewInMemoryEngine()

	alice, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	bob, err := e.CreateNode([]string{"User"}, map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	if _, err := e.CreateEdge(alice, bob, "FOLLOWS", map[string]any{"since": 2026}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if err := e.CreateIndex(IndexSpec{Kind: IndexKindProperty, Label: "User", Property: "name"}); err != nil {
		t.Fatalf("create property index failed: %v", err)
	}
	if err := e.CreateIndex(IndexSpec{Kind: IndexKindEdgeType, EdgeType: "FOLLOWS"}); err != nil {
		t.Fatalf("create edge-type index failed: %v", err)
	}

	path := filepath.Join(t.TempDir(), "graph.json")
	if err := e.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded := NewInMemoryEngine()
	if err := loaded.Load(path); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	users, err := loaded.FindNodesByProperty("User", "name", "alice")
	if err != nil {
		t.Fatalf("find nodes by property failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 loaded user, got: %d", len(users))
	}

	edges, err := loaded.FindEdgesByType("FOLLOWS")
	if err != nil {
		t.Fatalf("find edges by type failed: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 loaded edge, got: %d", len(edges))
	}
	if edges[0].From != alice || edges[0].To != bob {
		t.Fatalf("unexpected loaded edge: %+v", edges[0])
	}

	// next IDs should continue from loaded snapshot.
	nextNodeID, err := loaded.CreateNode([]string{"User"}, map[string]any{"name": "carol"})
	if err != nil {
		t.Fatalf("create node after load failed: %v", err)
	}
	if nextNodeID != 3 {
		t.Fatalf("expected next node id 3, got: %d", nextNodeID)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	e := NewInMemoryEngine()
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0644); err != nil {
		t.Fatalf("write invalid file failed: %v", err)
	}

	if err := e.Load(path); err == nil {
		t.Fatal("expected load to fail for invalid json")
	}
}
