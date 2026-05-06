package adjlist

import (
	"reflect"
	"testing"
)

func TestInOutNeighbors(t *testing.T) {
	g := New()
	g.AddEdge("alice", "bob")
	g.AddEdge("alice", "carol")
	g.AddEdge("dave", "bob")

	if got, want := g.OutNeighbors("alice"), []string{"bob", "carol"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("out neighbors mismatch: got=%v want=%v", got, want)
	}
	if got, want := g.InNeighbors("bob"), []string{"alice", "dave"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("in neighbors mismatch: got=%v want=%v", got, want)
	}
}

func TestBFS(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "e")

	if got, want := g.BFS("a"), []string{"a", "b", "c", "d", "e"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bfs order mismatch: got=%v want=%v", got, want)
	}
}

func TestDFS(t *testing.T) {
	g := New()
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "e")

	if got, want := g.DFS("a"), []string{"a", "b", "d", "c", "e"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("dfs order mismatch: got=%v want=%v", got, want)
	}
}

func TestMissingStartNode(t *testing.T) {
	g := New()
	if got := g.BFS("missing"); len(got) != 0 {
		t.Fatalf("expected empty bfs result, got=%v", got)
	}
	if got := g.DFS("missing"); len(got) != 0 {
		t.Fatalf("expected empty dfs result, got=%v", got)
	}
}
