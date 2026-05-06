package graphdb

import (
	"encoding/json"
	"os"
	"slices"
)

type persistedGraph struct {
	NextNodeID NodeID `json:"next_node_id"`
	NextEdgeID EdgeID `json:"next_edge_id"`

	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`

	PropertyIndexes []persistedPropertyIndex `json:"property_indexes"`
	EdgeTypeIndexes []string                 `json:"edge_type_indexes"`
}

type persistedPropertyIndex struct {
	Label    string `json:"label"`
	Property string `json:"property"`
}

func (e *InMemoryEngine) Save(path string) error {
	e.mu.RLock()
	snapshot := persistedGraph{
		NextNodeID: e.nextNodeID,
		NextEdgeID: e.nextEdgeID,
		Nodes:      make([]Node, 0, len(e.nodes)),
		Edges:      make([]Edge, 0, len(e.edges)),
	}

	nodeIDs := make([]int, 0, len(e.nodes))
	for id := range e.nodes {
		nodeIDs = append(nodeIDs, int(id))
	}
	slices.Sort(nodeIDs)
	for _, id := range nodeIDs {
		snapshot.Nodes = append(snapshot.Nodes, cloneNode(e.nodes[NodeID(id)]))
	}

	edgeIDs := make([]int, 0, len(e.edges))
	for id := range e.edges {
		edgeIDs = append(edgeIDs, int(id))
	}
	slices.Sort(edgeIDs)
	for _, id := range edgeIDs {
		snapshot.Edges = append(snapshot.Edges, cloneEdge(e.edges[EdgeID(id)]))
	}

	snapshot.PropertyIndexes = make([]persistedPropertyIndex, 0, len(e.propertyIndexes))
	for key := range e.propertyIndexes {
		snapshot.PropertyIndexes = append(snapshot.PropertyIndexes, persistedPropertyIndex{
			Label:    key.Label,
			Property: key.Property,
		})
	}
	slices.SortFunc(snapshot.PropertyIndexes, func(a, b persistedPropertyIndex) int {
		if a.Label != b.Label {
			if a.Label < b.Label {
				return -1
			}
			return 1
		}
		if a.Property < b.Property {
			return -1
		}
		if a.Property > b.Property {
			return 1
		}
		return 0
	})

	snapshot.EdgeTypeIndexes = make([]string, 0, len(e.edgeTypeIndex))
	for edgeType := range e.edgeTypeIndex {
		snapshot.EdgeTypeIndexes = append(snapshot.EdgeTypeIndexes, edgeType)
	}
	slices.Sort(snapshot.EdgeTypeIndexes)
	e.mu.RUnlock()

	payload, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0644)
}

func (e *InMemoryEngine) Load(path string) error {
	payload, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var snapshot persistedGraph
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.nextNodeID = snapshot.NextNodeID
	e.nextEdgeID = snapshot.NextEdgeID

	e.nodes = make(map[NodeID]Node, len(snapshot.Nodes))
	e.edges = make(map[EdgeID]Edge, len(snapshot.Edges))
	e.labelIndex = make(map[string]map[NodeID]struct{})
	e.propertyIndexes = make(map[propertyIndexKey]map[string]map[NodeID]struct{})
	e.edgeTypeIndex = make(map[string]map[EdgeID]struct{})
	e.outAdj = make(map[NodeID]map[EdgeID]struct{})
	e.inAdj = make(map[NodeID]map[EdgeID]struct{})

	for _, node := range snapshot.Nodes {
		cloned := cloneNode(node)
		e.nodes[cloned.ID] = cloned
		if cloned.ID > e.nextNodeID {
			e.nextNodeID = cloned.ID
		}
		for _, label := range cloned.Labels {
			if _, ok := e.labelIndex[label]; !ok {
				e.labelIndex[label] = make(map[NodeID]struct{})
			}
			e.labelIndex[label][cloned.ID] = struct{}{}
		}
	}

	for _, edge := range snapshot.Edges {
		cloned := cloneEdge(edge)
		e.edges[cloned.ID] = cloned
		if cloned.ID > e.nextEdgeID {
			e.nextEdgeID = cloned.ID
		}
		// indexEdge also builds out/in adjacency and edgeType index if enabled.
		e.indexEdge(cloned)
	}

	for _, pi := range snapshot.PropertyIndexes {
		if err := e.createIndexNoLock(IndexSpec{
			Kind:     IndexKindProperty,
			Label:    pi.Label,
			Property: pi.Property,
		}); err != nil {
			return err
		}
	}
	for _, edgeType := range snapshot.EdgeTypeIndexes {
		if err := e.createIndexNoLock(IndexSpec{
			Kind:     IndexKindEdgeType,
			EdgeType: edgeType,
		}); err != nil {
			return err
		}
	}
	return nil
}
