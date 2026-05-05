package graphdb

type NodeID uint64
type EdgeID uint64

type Node struct {
	ID         NodeID
	Labels     []string
	Properties map[string]any
}

type Edge struct {
	ID         EdgeID
	From       NodeID
	To         NodeID
	Type       string
	Properties map[string]any
}
