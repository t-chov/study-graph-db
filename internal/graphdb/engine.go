package graphdb

import "errors"

var ErrNotImplemented = errors.New("not implemented")

// Engine is the top-level API that will be expanded from v1 to v8.
type Engine interface {
	// v1: in-memory nodes, edges, label search
	CreateNode(labels []string, props map[string]any) (NodeID, error)
	CreateEdge(from NodeID, to NodeID, edgeType string, props map[string]any) (EdgeID, error)
	FindNodesByLabel(label string) ([]Node, error)

	// v2: property index
	FindNodesByProperty(label string, property string, value any) ([]Node, error)

	// v3: edge-type index and adjacency list
	FindEdgesByType(edgeType string) ([]Edge, error)
	FindOutgoingEdges(from NodeID) ([]Edge, error)
	FindIncomingEdges(to NodeID) ([]Edge, error)

	// v4: simple MATCH query
	Match(query string) (ResultSet, error)

	// v5: query plan / explain
	Explain(query string) (QueryPlan, error)

	// v6: add/remove index
	CreateIndex(spec IndexSpec) error
	DropIndex(spec IndexSpec) error

	// v8: transaction / WAL entry point
	Begin() (Tx, error)
}
