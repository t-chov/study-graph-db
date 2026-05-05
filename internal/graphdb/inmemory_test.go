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

func TestFindEdgesByType(t *testing.T) {
	e := NewInMemoryEngine()
	a, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	b, err := e.CreateNode([]string{"User"}, map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	c, err := e.CreateNode([]string{"User"}, map[string]any{"name": "carol"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	if _, err := e.CreateEdge(a, b, "FOLLOWS", map[string]any{"w": 1}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if _, err := e.CreateEdge(a, c, "LIKES", map[string]any{"w": 2}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if _, err := e.CreateEdge(b, c, "FOLLOWS", map[string]any{"w": 3}); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	edges, err := e.FindEdgesByType("FOLLOWS")
	if err != nil {
		t.Fatalf("find edges by type failed: %v", err)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got: %d", len(edges))
	}

	edges, err = e.FindEdgesByType("UNKNOWN")
	if err != nil {
		t.Fatalf("find edges by type failed: %v", err)
	}
	if len(edges) != 0 {
		t.Fatalf("expected 0 edges, got: %d", len(edges))
	}
}

func TestAdjacencyLists(t *testing.T) {
	e := NewInMemoryEngine()
	a, err := e.CreateNode([]string{"User"}, nil)
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	b, err := e.CreateNode([]string{"User"}, nil)
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	c, err := e.CreateNode([]string{"User"}, nil)
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	if _, err := e.CreateEdge(a, b, "FOLLOWS", nil); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if _, err := e.CreateEdge(a, c, "LIKES", nil); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if _, err := e.CreateEdge(c, a, "FOLLOWS", nil); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	out, err := e.FindOutgoingEdges(a)
	if err != nil {
		t.Fatalf("find outgoing edges failed: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 outgoing edges, got: %d", len(out))
	}
	for _, edge := range out {
		if edge.From != a {
			t.Fatalf("unexpected from node in outgoing edge: %+v", edge)
		}
	}

	in, err := e.FindIncomingEdges(a)
	if err != nil {
		t.Fatalf("find incoming edges failed: %v", err)
	}
	if len(in) != 1 {
		t.Fatalf("expected 1 incoming edge, got: %d", len(in))
	}
	if in[0].To != a {
		t.Fatalf("unexpected to node in incoming edge: %+v", in[0])
	}
}

func TestMatchSingleNodePattern(t *testing.T) {
	e := NewInMemoryEngine()
	if _, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	if _, err := e.CreateNode([]string{"Project"}, map[string]any{"name": "study-graph-db"}); err != nil {
		t.Fatalf("create node failed: %v", err)
	}

	result, err := e.Match("MATCH (n:User)")
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got: %d", len(result.Rows))
	}
	node, ok := result.Rows[0]["n"].(Node)
	if !ok {
		t.Fatalf("expected Node type, got: %T", result.Rows[0]["n"])
	}
	if got := node.Properties["name"]; got != "alice" {
		t.Fatalf("unexpected matched node property: %v", got)
	}
}

func TestMatchSingleHopPattern(t *testing.T) {
	e := NewInMemoryEngine()
	alice, err := e.CreateNode([]string{"User"}, map[string]any{"name": "alice"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	bob, err := e.CreateNode([]string{"User"}, map[string]any{"name": "bob"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	project, err := e.CreateNode([]string{"Project"}, map[string]any{"name": "study-graph-db"})
	if err != nil {
		t.Fatalf("create node failed: %v", err)
	}
	if _, err := e.CreateEdge(alice, bob, "FOLLOWS", nil); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}
	if _, err := e.CreateEdge(alice, project, "WORKS_ON", nil); err != nil {
		t.Fatalf("create edge failed: %v", err)
	}

	result, err := e.Match("MATCH (a:User)-[:FOLLOWS]->(b:User)")
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got: %d", len(result.Rows))
	}
	left, ok := result.Rows[0]["a"].(Node)
	if !ok {
		t.Fatalf("expected Node type for a, got: %T", result.Rows[0]["a"])
	}
	right, ok := result.Rows[0]["b"].(Node)
	if !ok {
		t.Fatalf("expected Node type for b, got: %T", result.Rows[0]["b"])
	}
	if left.ID != alice || right.ID != bob {
		t.Fatalf("unexpected matched pair: a=%d b=%d", left.ID, right.ID)
	}
}

func TestMatchInvalidQuery(t *testing.T) {
	e := NewInMemoryEngine()
	if _, err := e.Match("MATCH (n)"); err != ErrInvalidMatchQuery {
		t.Fatalf("expected ErrInvalidMatchQuery, got: %v", err)
	}
}
