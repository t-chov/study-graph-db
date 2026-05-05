package graphdb

import "testing"

func TestNewInMemoryEngine(t *testing.T) {
	e := NewInMemoryEngine()
	if e == nil {
		t.Fatal("engine must not be nil")
	}
}

func TestCreateNodeAndFindByLabel(t *testing.T) {
	e := NewInMemoryEngine()

	labels := []string{"User", "Person"}
	props := map[string]any{"name": "alice"}
	nodeID, err := e.CreateNode(labels, props)
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	if nodeID != 1 {
		t.Fatalf("unexpected node id: %d", nodeID)
	}

	// Ensure external input mutation does not affect stored data.
	labels[0] = "Hacked"
	props["name"] = "mallory"

	nodes, err := e.FindNodesByLabel("User")
	if err != nil {
		t.Fatalf("find nodes failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got: %d", len(nodes))
	}
	if nodes[0].ID != nodeID {
		t.Fatalf("unexpected node in result: %d", nodes[0].ID)
	}
	if got := nodes[0].Properties["name"]; got != "alice" {
		t.Fatalf("unexpected stored property value: %v", got)
	}
}

func TestFindNodesByMissingLabelReturnsEmptySlice(t *testing.T) {
	e := NewInMemoryEngine()

	nodes, err := e.FindNodesByLabel("missing")
	if err != nil {
		t.Fatalf("find nodes failed: %v", err)
	}
	if len(nodes) != 0 {
		t.Fatalf("expected 0 nodes, got: %d", len(nodes))
	}
}

func TestCreateEdge(t *testing.T) {
	e := NewInMemoryEngine()

	from, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"})
	if err != nil {
		t.Fatalf("create from node failed: %v", err)
	}
	to, err := e.CreateNode([]string{"User"}, map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("create to node failed: %v", err)
	}

	edgeProps := map[string]any{"since": 2024}
	edgeID, err := e.CreateEdge(from, to, "FOLLOWS", edgeProps)
	if err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if edgeID != 1 {
		t.Fatalf("unexpected edge id: %d", edgeID)
	}

	edgeProps["since"] = 2000
	edge, ok := e.edges[edgeID]
	if !ok {
		t.Fatalf("edge %d not found in store", edgeID)
	}
	if edge.From != from || edge.To != to || edge.Type != "FOLLOWS" {
		t.Fatalf("unexpected edge content: %+v", edge)
	}
	if got := edge.Properties["since"]; got != 2024 {
		t.Fatalf("unexpected stored edge property value: %v", got)
	}
}

func TestCreateEdgeValidatesInputs(t *testing.T) {
	e := NewInMemoryEngine()
	nodeID, err := e.CreateNode([]string{"User"}, nil)
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	if _, err := e.CreateEdge(nodeID, nodeID, "", nil); err != ErrEmptyEdgeType {
		t.Fatalf("expected ErrEmptyEdgeType, got: %v", err)
	}
	if _, err := e.CreateEdge(nodeID, 999, "FOLLOWS", nil); err != ErrNodeNotFound {
		t.Fatalf("expected ErrNodeNotFound, got: %v", err)
	}
}

func TestFindNodesByPropertyWithoutIndexFallsBackToScan(t *testing.T) {
	e := NewInMemoryEngine()
	if _, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	if _, err := e.CreateNode([]string{"User"}, map[string]any{"name": "bob"}); err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	nodes, err := e.FindNodesByProperty("User", "name", "alice")
	if err != nil {
		t.Fatalf("find nodes by property failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got: %d", len(nodes))
	}
	if got := nodes[0].Properties["name"]; got != "alice" {
		t.Fatalf("unexpected node property: %v", got)
	}
}

func TestPropertyIndexBackfillAndIncrementalUpdate(t *testing.T) {
	e := NewInMemoryEngine()
	if _, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	spec := IndexSpec{
		Kind:     IndexKindProperty,
		Label:    "User",
		Property: "name",
	}
	if err := e.CreateIndex(spec); err != nil {
		t.Fatalf("create index failed: %v", err)
	}

	nodes, err := e.FindNodesByProperty("User", "name", "alice")
	if err != nil {
		t.Fatalf("find nodes by property failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 backfilled node, got: %d", len(nodes))
	}

	if _, err := e.CreateNode([]string{"User"}, map[string]any{"name": "bob"}); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	nodes, err = e.FindNodesByProperty("User", "name", "bob")
	if err != nil {
		t.Fatalf("find nodes by property failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 incrementally indexed node, got: %d", len(nodes))
	}
}

func TestDropPropertyIndex(t *testing.T) {
	e := NewInMemoryEngine()
	spec := IndexSpec{
		Kind:     IndexKindProperty,
		Label:    "User",
		Property: "name",
	}
	if err := e.CreateIndex(spec); err != nil {
		t.Fatalf("create index failed: %v", err)
	}
	if err := e.DropIndex(spec); err != nil {
		t.Fatalf("drop index failed: %v", err)
	}
	if _, ok := e.propertyIndexes[propertyIndexKey{Label: "User", Property: "name"}]; ok {
		t.Fatal("expected property index to be removed")
	}
}

func TestCreateIndexValidation(t *testing.T) {
	e := NewInMemoryEngine()
	if err := e.CreateIndex(IndexSpec{Kind: IndexKindEdgeType, Label: "User", Property: "name"}); err != ErrUnsupportedIndexKind {
		t.Fatalf("expected ErrUnsupportedIndexKind, got: %v", err)
	}
	if err := e.CreateIndex(IndexSpec{Kind: IndexKindProperty, Label: "", Property: "name"}); err != ErrInvalidIndexSpec {
		t.Fatalf("expected ErrInvalidIndexSpec, got: %v", err)
	}
	if err := e.DropIndex(IndexSpec{Kind: IndexKindEdgeType, Label: "User", Property: "name"}); err != ErrUnsupportedIndexKind {
		t.Fatalf("expected ErrUnsupportedIndexKind, got: %v", err)
	}
}
