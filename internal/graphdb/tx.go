package graphdb

type Tx interface {
	CreateNode(labels []string, props map[string]any) (NodeID, error)
	CreateEdge(from NodeID, to NodeID, edgeType string, props map[string]any) (EdgeID, error)

	FindNodesByLabel(label string) ([]Node, error)
	FindNodesByProperty(label string, property string, value any) ([]Node, error)
	FindEdgesByType(edgeType string) ([]Edge, error)
	FindOutgoingEdges(from NodeID) ([]Edge, error)
	FindIncomingEdges(to NodeID) ([]Edge, error)

	Match(query string) (ResultSet, error)
	Explain(query string) (QueryPlan, error)

	CreateIndex(spec IndexSpec) error
	DropIndex(spec IndexSpec) error

	Commit() error
	Rollback() error
}
