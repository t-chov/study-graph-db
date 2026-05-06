package adjlist

import "slices"

// Graph is a tiny directed graph implementation backed by adjacency lists.
//
// Core idea:
// - out[u] keeps neighbors directly reachable from u (u -> v)
// - in[v] keeps predecessors that can reach v (u -> v)
//
// This mirrors the shape used in many graph databases for fast local traversal.
type Graph struct {
	out map[string]map[string]struct{}
	in  map[string]map[string]struct{}
}

func New() *Graph {
	return &Graph{
		out: make(map[string]map[string]struct{}),
		in:  make(map[string]map[string]struct{}),
	}
}

func (g *Graph) AddNode(id string) {
	if _, ok := g.out[id]; !ok {
		g.out[id] = make(map[string]struct{})
	}
	if _, ok := g.in[id]; !ok {
		g.in[id] = make(map[string]struct{})
	}
}

func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	g.out[from][to] = struct{}{}
	g.in[to][from] = struct{}{}
}

func (g *Graph) OutNeighbors(id string) []string {
	return sortedKeys(g.out[id])
}

func (g *Graph) InNeighbors(id string) []string {
	return sortedKeys(g.in[id])
}

// BFS returns the visit order from start using outgoing edges.
func (g *Graph) BFS(start string) []string {
	if _, ok := g.out[start]; !ok {
		return []string{}
	}

	visited := map[string]struct{}{start: {}}
	queue := []string{start}
	order := make([]string, 0)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, n := range sortedKeys(g.out[node]) {
			if _, ok := visited[n]; ok {
				continue
			}
			visited[n] = struct{}{}
			queue = append(queue, n)
		}
	}
	return order
}

// DFS returns preorder visit order from start using outgoing edges.
func (g *Graph) DFS(start string) []string {
	if _, ok := g.out[start]; !ok {
		return []string{}
	}

	visited := make(map[string]struct{})
	order := make([]string, 0)

	var walk func(node string)
	walk = func(node string) {
		if _, ok := visited[node]; ok {
			return
		}
		visited[node] = struct{}{}
		order = append(order, node)
		for _, n := range sortedKeys(g.out[node]) {
			walk(n)
		}
	}
	walk(start)
	return order
}

func sortedKeys(set map[string]struct{}) []string {
	if len(set) == 0 {
		return []string{}
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
