package graphdb

type IndexKind string

const (
	IndexKindProperty IndexKind = "property" // v2
	IndexKindEdgeType IndexKind = "edgeType" // v3
)

type IndexSpec struct {
	Kind     IndexKind
	Label    string
	Property string
	EdgeType string
}
