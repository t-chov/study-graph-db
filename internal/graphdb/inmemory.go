package graphdb

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	ErrNodeNotFound            = errors.New("node not found")
	ErrEmptyEdgeType           = errors.New("edge type must not be empty")
	ErrInvalidIndexSpec        = errors.New("invalid index spec")
	ErrUnsupportedIndexKind    = errors.New("unsupported index kind")
	ErrUnsupportedPropertyType = errors.New("unsupported property value type for index")
	ErrInvalidMatchQuery       = errors.New("invalid MATCH query")
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

	// propertyIndexes[(label, property)] => valueKey => node IDs
	propertyIndexes map[propertyIndexKey]map[string]map[NodeID]struct{}

	// edgeTypeIndex[edgeType] => edge IDs
	edgeTypeIndex map[string]map[EdgeID]struct{}
	// outAdj[fromNode] => edge IDs
	outAdj map[NodeID]map[EdgeID]struct{}
	// inAdj[toNode] => edge IDs
	inAdj map[NodeID]map[EdgeID]struct{}
}

func NewInMemoryEngine() *InMemoryEngine {
	return &InMemoryEngine{
		nodes:           make(map[NodeID]Node),
		edges:           make(map[EdgeID]Edge),
		labelIndex:      make(map[string]map[NodeID]struct{}),
		propertyIndexes: make(map[propertyIndexKey]map[string]map[NodeID]struct{}),
		edgeTypeIndex:   make(map[string]map[EdgeID]struct{}),
		outAdj:          make(map[NodeID]map[EdgeID]struct{}),
		inAdj:           make(map[NodeID]map[EdgeID]struct{}),
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
	e.indexNodeProperties(node)

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
	e.indexEdge(edge)

	return id, nil
}

func (e *InMemoryEngine) FindNodesByLabel(label string) ([]Node, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.findNodesByLabelNoLock(label), nil
}

func (e *InMemoryEngine) FindNodesByProperty(label string, property string, value any) ([]Node, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := propertyIndexKey{
		Label:    label,
		Property: property,
	}
	if valueMap, ok := e.propertyIndexes[key]; ok {
		valueKey, err := propertyValueKey(value)
		if err != nil {
			return nil, err
		}
		ids, ok := valueMap[valueKey]
		if !ok {
			return []Node{}, nil
		}
		return e.collectNodes(ids), nil
	}

	ids, ok := e.labelIndex[label]
	if !ok {
		return []Node{}, nil
	}

	// Fallback scan when index does not exist yet.
	matches := make([]Node, 0)
	for id := range ids {
		node := e.nodes[id]
		if reflect.DeepEqual(node.Properties[property], value) {
			matches = append(matches, cloneNode(node))
		}
	}
	return matches, nil
}

func (e *InMemoryEngine) FindEdgesByType(edgeType string) ([]Edge, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.findEdgesByTypeNoLock(edgeType), nil
}

func (e *InMemoryEngine) FindOutgoingEdges(from NodeID) ([]Edge, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ids, ok := e.outAdj[from]
	if !ok {
		return []Edge{}, nil
	}
	return e.collectEdges(ids), nil
}

func (e *InMemoryEngine) FindIncomingEdges(to NodeID) ([]Edge, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ids, ok := e.inAdj[to]
	if !ok {
		return []Edge{}, nil
	}
	return e.collectEdges(ids), nil
}

func (e *InMemoryEngine) Match(query string) (ResultSet, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return executeSimpleMatch(e, query)
}

func (e *InMemoryEngine) Explain(_ string) (QueryPlan, error) {
	return QueryPlan{}, ErrNotImplemented
}

func (e *InMemoryEngine) CreateIndex(spec IndexSpec) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if spec.Kind != IndexKindProperty {
		return ErrUnsupportedIndexKind
	}
	if spec.Label == "" || spec.Property == "" {
		return ErrInvalidIndexSpec
	}

	key := propertyIndexKey{
		Label:    spec.Label,
		Property: spec.Property,
	}
	if _, ok := e.propertyIndexes[key]; ok {
		return nil
	}

	e.propertyIndexes[key] = make(map[string]map[NodeID]struct{})

	for id := range e.labelIndex[spec.Label] {
		node := e.nodes[id]
		value, ok := node.Properties[spec.Property]
		if !ok {
			continue
		}
		valueKey, err := propertyValueKey(value)
		if err != nil {
			continue
		}
		if _, ok := e.propertyIndexes[key][valueKey]; !ok {
			e.propertyIndexes[key][valueKey] = make(map[NodeID]struct{})
		}
		e.propertyIndexes[key][valueKey][id] = struct{}{}
	}

	return nil
}

func (e *InMemoryEngine) DropIndex(spec IndexSpec) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if spec.Kind != IndexKindProperty {
		return ErrUnsupportedIndexKind
	}
	if spec.Label == "" || spec.Property == "" {
		return ErrInvalidIndexSpec
	}

	delete(e.propertyIndexes, propertyIndexKey{
		Label:    spec.Label,
		Property: spec.Property,
	})
	return nil
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

type propertyIndexKey struct {
	Label    string
	Property string
}

func propertyValueKey(value any) (string, error) {
	switch v := value.(type) {
	case nil:
		return "nil", nil
	case string:
		return "s:" + v, nil
	case bool:
		return fmt.Sprintf("b:%t", v), nil
	case int:
		return fmt.Sprintf("i:%d", v), nil
	case int8:
		return fmt.Sprintf("i:%d", v), nil
	case int16:
		return fmt.Sprintf("i:%d", v), nil
	case int32:
		return fmt.Sprintf("i:%d", v), nil
	case int64:
		return fmt.Sprintf("i:%d", v), nil
	case uint:
		return fmt.Sprintf("u:%d", v), nil
	case uint8:
		return fmt.Sprintf("u:%d", v), nil
	case uint16:
		return fmt.Sprintf("u:%d", v), nil
	case uint32:
		return fmt.Sprintf("u:%d", v), nil
	case uint64:
		return fmt.Sprintf("u:%d", v), nil
	case float32:
		return fmt.Sprintf("f:%g", v), nil
	case float64:
		return fmt.Sprintf("f:%g", v), nil
	default:
		return "", ErrUnsupportedPropertyType
	}
}

func (e *InMemoryEngine) indexNodeProperties(node Node) {
	for _, label := range node.Labels {
		for key, valueMap := range e.propertyIndexes {
			if key.Label != label {
				continue
			}
			value, ok := node.Properties[key.Property]
			if !ok {
				continue
			}
			valueKey, err := propertyValueKey(value)
			if err != nil {
				continue
			}
			if _, ok := valueMap[valueKey]; !ok {
				valueMap[valueKey] = make(map[NodeID]struct{})
			}
			valueMap[valueKey][node.ID] = struct{}{}
		}
	}
}

func (e *InMemoryEngine) collectNodes(ids map[NodeID]struct{}) []Node {
	nodes := make([]Node, 0, len(ids))
	for id := range ids {
		nodes = append(nodes, cloneNode(e.nodes[id]))
	}
	return nodes
}

func (e *InMemoryEngine) findNodesByLabelNoLock(label string) []Node {
	ids, ok := e.labelIndex[label]
	if !ok {
		return []Node{}
	}
	return e.collectNodes(ids)
}

func (e *InMemoryEngine) findEdgesByTypeNoLock(edgeType string) []Edge {
	ids, ok := e.edgeTypeIndex[edgeType]
	if !ok {
		return []Edge{}
	}
	return e.collectEdges(ids)
}

func cloneNode(node Node) Node {
	return Node{
		ID:         node.ID,
		Labels:     cloneLabels(node.Labels),
		Properties: cloneProperties(node.Properties),
	}
}

func cloneEdge(edge Edge) Edge {
	return Edge{
		ID:         edge.ID,
		From:       edge.From,
		To:         edge.To,
		Type:       edge.Type,
		Properties: cloneProperties(edge.Properties),
	}
}

func (e *InMemoryEngine) indexEdge(edge Edge) {
	if _, ok := e.edgeTypeIndex[edge.Type]; !ok {
		e.edgeTypeIndex[edge.Type] = make(map[EdgeID]struct{})
	}
	e.edgeTypeIndex[edge.Type][edge.ID] = struct{}{}

	if _, ok := e.outAdj[edge.From]; !ok {
		e.outAdj[edge.From] = make(map[EdgeID]struct{})
	}
	e.outAdj[edge.From][edge.ID] = struct{}{}

	if _, ok := e.inAdj[edge.To]; !ok {
		e.inAdj[edge.To] = make(map[EdgeID]struct{})
	}
	e.inAdj[edge.To][edge.ID] = struct{}{}
}

func (e *InMemoryEngine) collectEdges(ids map[EdgeID]struct{}) []Edge {
	edges := make([]Edge, 0, len(ids))
	for id := range ids {
		edges = append(edges, cloneEdge(e.edges[id]))
	}
	return edges
}
