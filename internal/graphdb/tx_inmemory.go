package graphdb

import (
	"errors"
	"sync"
)

var ErrTxClosed = errors.New("transaction already closed")

type inMemoryTx struct {
	engine *InMemoryEngine
	data   *InMemoryEngine

	mu     sync.Mutex
	closed bool
}

func (tx *inMemoryTx) CreateNode(labels []string, props map[string]any) (NodeID, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return 0, ErrTxClosed
	}
	return tx.data.CreateNode(labels, props)
}

func (tx *inMemoryTx) CreateEdge(from NodeID, to NodeID, edgeType string, props map[string]any) (EdgeID, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return 0, ErrTxClosed
	}
	return tx.data.CreateEdge(from, to, edgeType, props)
}

func (tx *inMemoryTx) FindNodesByLabel(label string) ([]Node, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, ErrTxClosed
	}
	return tx.data.FindNodesByLabel(label)
}

func (tx *inMemoryTx) FindNodesByProperty(label string, property string, value any) ([]Node, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, ErrTxClosed
	}
	return tx.data.FindNodesByProperty(label, property, value)
}

func (tx *inMemoryTx) FindEdgesByType(edgeType string) ([]Edge, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, ErrTxClosed
	}
	return tx.data.FindEdgesByType(edgeType)
}

func (tx *inMemoryTx) FindOutgoingEdges(from NodeID) ([]Edge, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, ErrTxClosed
	}
	return tx.data.FindOutgoingEdges(from)
}

func (tx *inMemoryTx) FindIncomingEdges(to NodeID) ([]Edge, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return nil, ErrTxClosed
	}
	return tx.data.FindIncomingEdges(to)
}

func (tx *inMemoryTx) Match(query string) (ResultSet, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return ResultSet{}, ErrTxClosed
	}
	return tx.data.Match(query)
}

func (tx *inMemoryTx) Explain(query string) (QueryPlan, error) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return QueryPlan{}, ErrTxClosed
	}
	return tx.data.Explain(query)
}

func (tx *inMemoryTx) CreateIndex(spec IndexSpec) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return ErrTxClosed
	}
	return tx.data.CreateIndex(spec)
}

func (tx *inMemoryTx) DropIndex(spec IndexSpec) error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return ErrTxClosed
	}
	return tx.data.DropIndex(spec)
}

func (tx *inMemoryTx) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return ErrTxClosed
	}

	tx.engine.mu.Lock()
	defer tx.engine.mu.Unlock()

	applyEngineSnapshotNoLock(tx.engine, tx.data)
	tx.closed = true
	tx.data = nil
	return nil
}

func (tx *inMemoryTx) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	if tx.closed {
		return ErrTxClosed
	}
	tx.closed = true
	tx.data = nil
	return nil
}

func cloneEngineNoLock(src *InMemoryEngine) *InMemoryEngine {
	dst := NewInMemoryEngine()
	dst.nextNodeID = src.nextNodeID
	dst.nextEdgeID = src.nextEdgeID

	dst.nodes = make(map[NodeID]Node, len(src.nodes))
	for id, node := range src.nodes {
		dst.nodes[id] = cloneNode(node)
	}

	dst.edges = make(map[EdgeID]Edge, len(src.edges))
	for id, edge := range src.edges {
		dst.edges[id] = cloneEdge(edge)
	}

	dst.labelIndex = cloneNestedSetMap(src.labelIndex)
	dst.propertyIndexes = clonePropertyIndexes(src.propertyIndexes)
	dst.edgeTypeIndex = cloneNestedSetMap(src.edgeTypeIndex)
	dst.outAdj = cloneNestedSetMap(src.outAdj)
	dst.inAdj = cloneNestedSetMap(src.inAdj)
	return dst
}

func applyEngineSnapshotNoLock(dst *InMemoryEngine, src *InMemoryEngine) {
	cloned := cloneEngineNoLock(src)
	dst.nextNodeID = cloned.nextNodeID
	dst.nextEdgeID = cloned.nextEdgeID
	dst.nodes = cloned.nodes
	dst.edges = cloned.edges
	dst.labelIndex = cloned.labelIndex
	dst.propertyIndexes = cloned.propertyIndexes
	dst.edgeTypeIndex = cloned.edgeTypeIndex
	dst.outAdj = cloned.outAdj
	dst.inAdj = cloned.inAdj
}

func cloneNestedSetMap[K comparable, V comparable](in map[K]map[V]struct{}) map[K]map[V]struct{} {
	out := make(map[K]map[V]struct{}, len(in))
	for k, set := range in {
		cloned := make(map[V]struct{}, len(set))
		for v := range set {
			cloned[v] = struct{}{}
		}
		out[k] = cloned
	}
	return out
}

func clonePropertyIndexes(in map[propertyIndexKey]map[string]map[NodeID]struct{}) map[propertyIndexKey]map[string]map[NodeID]struct{} {
	out := make(map[propertyIndexKey]map[string]map[NodeID]struct{}, len(in))
	for key, valueMap := range in {
		clonedValueMap := make(map[string]map[NodeID]struct{}, len(valueMap))
		for valueKey, ids := range valueMap {
			clonedIDs := make(map[NodeID]struct{}, len(ids))
			for id := range ids {
				clonedIDs[id] = struct{}{}
			}
			clonedValueMap[valueKey] = clonedIDs
		}
		out[key] = clonedValueMap
	}
	return out
}
