// Package adjlist is a small learning package for adjacency lists.
//
// Example:
//
//	g := adjlist.New()
//	g.AddEdge("alice", "bob")
//	g.AddEdge("alice", "carol")
//	g.OutNeighbors("alice") // => ["bob", "carol"]
//
// - OutNeighbors(n): nodes reachable from n
// - InNeighbors(n): nodes that can reach n
// - BFS/DFS: traversal from a start node
package adjlist
