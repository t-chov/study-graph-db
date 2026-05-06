package graphdb

import (
	"regexp"
	"strings"
)

// Supported query grammar (v4 minimal MATCH):
//
//  1. Node pattern
//     MATCH (<nodeVar>:<Label>)
//
//  2. Single-hop edge pattern
//     MATCH (<leftVar>:<LeftLabel>)-[:<EdgeType>]->(<rightVar>:<RightLabel>)
//
// Notes:
//   - Leading/trailing spaces are ignored.
//   - A trailing semicolon is allowed.
//   - Variable, label, and edge type must match [A-Za-z_][A-Za-z0-9_]*.
//   - Property predicates, WHERE, RETURN, multi-hop patterns, and undirected edges
//     are not supported in v4.
var (
	nodeMatchPattern = regexp.MustCompile(`^MATCH\s+\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*:\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)$`)
	edgeMatchPattern = regexp.MustCompile(`^MATCH\s+\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*:\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)\s*-\s*\[\s*:\s*([A-Za-z_][A-Za-z0-9_]*)\s*\]\s*->\s*\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*:\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)$`)
)

func executeSimpleMatch(e *InMemoryEngine, query string) (ResultSet, error) {
	parsed, err := parseSimpleMatchQuery(query)
	if err != nil {
		return ResultSet{}, err
	}

	switch parsed.kind {
	case simpleMatchKindNode:
		return matchSingleNode(e, parsed.nodeVar, parsed.nodeLabel), nil
	case simpleMatchKindSingleHop:
		return matchSingleHop(
			e,
			parsed.leftVar,
			parsed.leftLabel,
			parsed.edgeType,
			parsed.rightVar,
			parsed.rightLabel,
		), nil
	default:
		return ResultSet{}, ErrInvalidMatchQuery
	}
}

func matchSingleNode(e *InMemoryEngine, nodeVar string, label string) ResultSet {
	nodes := e.findNodesByLabelNoLock(label)
	rows := make([]ResultRow, 0, len(nodes))
	for _, node := range nodes {
		rows = append(rows, ResultRow{
			nodeVar: node,
		})
	}
	return ResultSet{Rows: rows}
}

func matchSingleHop(e *InMemoryEngine, leftVar string, leftLabel string, edgeType string, rightVar string, rightLabel string) ResultSet {
	edges := e.findEdgesByTypeNoLock(edgeType)
	rows := make([]ResultRow, 0, len(edges))
	for _, edge := range edges {
		fromNode := e.nodes[edge.From]
		toNode := e.nodes[edge.To]
		if !hasLabel(fromNode, leftLabel) || !hasLabel(toNode, rightLabel) {
			continue
		}

		rows = append(rows, ResultRow{
			leftVar:  cloneNode(fromNode),
			rightVar: cloneNode(toNode),
		})
	}
	return ResultSet{Rows: rows}
}

func hasLabel(node Node, label string) bool {
	for _, l := range node.Labels {
		if l == label {
			return true
		}
	}
	return false
}

type simpleMatchKind string

const (
	simpleMatchKindNode      simpleMatchKind = "node"
	simpleMatchKindSingleHop simpleMatchKind = "singleHop"
)

type simpleMatchQuery struct {
	kind simpleMatchKind

	nodeVar   string
	nodeLabel string

	leftVar    string
	leftLabel  string
	edgeType   string
	rightVar   string
	rightLabel string
}

func parseSimpleMatchQuery(query string) (simpleMatchQuery, error) {
	normalized := strings.TrimSpace(query)
	normalized = strings.TrimSuffix(normalized, ";")
	normalized = strings.TrimSpace(normalized)

	if m := nodeMatchPattern.FindStringSubmatch(normalized); m != nil {
		return simpleMatchQuery{
			kind:      simpleMatchKindNode,
			nodeVar:   m[1],
			nodeLabel: m[2],
		}, nil
	}
	if m := edgeMatchPattern.FindStringSubmatch(normalized); m != nil {
		return simpleMatchQuery{
			kind:       simpleMatchKindSingleHop,
			leftVar:    m[1],
			leftLabel:  m[2],
			edgeType:   m[3],
			rightVar:   m[4],
			rightLabel: m[5],
		}, nil
	}
	return simpleMatchQuery{}, ErrInvalidMatchQuery
}
