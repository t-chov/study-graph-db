package graphdb

import (
	"errors"
	"sync"
)

var (
	ErrNodeNotFound  = errors.New("node not found")
	ErrEmptyEdgeType = errors.New("edge type must not be empty")
)

// InMemoryEngine is a minimal scaffold for iterative implementation.
type InMemoryEngine struct {
	mu sync.RWMutex

	nextNodeID NodeID
	nextEdgeID EdgeID

	nodes map[NodeID]Node
	edges map[EdgeID]Edge

	// labelIndex[label] => node IDs that have the label.
	labelIndex map[string]map[NodeID]struct{}
}

func NewInMemoryEngine() *InMemoryEngine {
	return &InMemoryEngine{
		nodes:      make(map[NodeID]Node),
		edges:      make(map[EdgeID]Edge),
		labelIndex: make(map[string]map[NodeID]struct{}),
	}
}

func (e *InMemoryEngine) CreateNode(labels []string, props map[string]any) (NodeID, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.nextNodeID++
	id := e.nextNodeID

	node := Node{
		ID:         id,
		Labels:     cloneLabels(labels),
		Properties: cloneProperties(props),
	}
	e.nodes[id] = node

	for _, label := range node.Labels {
		if _, ok := e.labelIndex[label]; !ok {
			e.labelIndex[label] = make(map[NodeID]struct{})
		}
		e.labelIndex[label][id] = struct{}{}
	}

	return id, nil
}

func (e *InMemoryEngine) CreateEdge(from NodeID, to NodeID, edgeType string, props map[string]any) (EdgeID, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if edgeType == "" {
		return 0, ErrEmptyEdgeType
	}
	if _, ok := e.nodes[from]; !ok {
		return 0, ErrNodeNotFound
	}
	if _, ok := e.nodes[to]; !ok {
		return 0, ErrNodeNotFound
	}

	e.nextEdgeID++
	id := e.nextEdgeID

	edge := Edge{
		ID:         id,
		From:       from,
		To:         to,
		Type:       edgeType,
		Properties: cloneProperties(props),
	}
	e.edges[id] = edge

	return id, nil
}

func (e *InMemoryEngine) FindNodesByLabel(label string) ([]Node, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ids, ok := e.labelIndex[label]
	if !ok {
		return []Node{}, nil
	}

	nodes := make([]Node, 0, len(ids))
	for id := range ids {
		node := e.nodes[id]
		nodes = append(nodes, Node{
			ID:         node.ID,
			Labels:     cloneLabels(node.Labels),
			Properties: cloneProperties(node.Properties),
		})
	}
	return nodes, nil
}

func (e *InMemoryEngine) Match(_ string) (ResultSet, error) {
	return ResultSet{}, ErrNotImplemented
}

func (e *InMemoryEngine) Explain(_ string) (QueryPlan, error) {
	return QueryPlan{}, ErrNotImplemented
}

func (e *InMemoryEngine) CreateIndex(_ IndexSpec) error {
	return ErrNotImplemented
}

func (e *InMemoryEngine) DropIndex(_ IndexSpec) error {
	return ErrNotImplemented
}

func (e *InMemoryEngine) Begin() (Tx, error) {
	return nil, ErrNotImplemented
}

func cloneLabels(labels []string) []string {
	if len(labels) == 0 {
		return []string{}
	}
	out := make([]string, len(labels))
	copy(out, labels)
	return out
}

func cloneProperties(props map[string]any) map[string]any {
	if len(props) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(props))
	for k, v := range props {
		out[k] = v
	}
	return out
}
