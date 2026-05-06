package graphdb

import "fmt"

func buildSimpleMatchPlan(e *InMemoryEngine, query string) (QueryPlan, error) {
	parsed, err := parseSimpleMatchQuery(query)
	if err != nil {
		return QueryPlan{}, err
	}

	switch parsed.kind {
	case simpleMatchKindNode:
		candidates := len(e.labelIndex[parsed.nodeLabel])
		return QueryPlan{
			Steps: []string{
				"Parse MATCH query",
				fmt.Sprintf("Lookup nodes by label using labelIndex: label=%s", parsed.nodeLabel),
				fmt.Sprintf("Candidate nodes: %d", candidates),
				fmt.Sprintf("Emit rows with variable: %s", parsed.nodeVar),
			},
		}, nil
	case simpleMatchKindSingleHop:
		edgeCandidates := len(e.edgeTypeIndex[parsed.edgeType])
		return QueryPlan{
			Steps: []string{
				"Parse MATCH query",
				fmt.Sprintf("Lookup edges by type using edgeTypeIndex: type=%s", parsed.edgeType),
				fmt.Sprintf("Candidate edges: %d", edgeCandidates),
				fmt.Sprintf("Filter source node by label: %s", parsed.leftLabel),
				fmt.Sprintf("Filter destination node by label: %s", parsed.rightLabel),
				fmt.Sprintf("Emit rows with variables: %s, %s", parsed.leftVar, parsed.rightVar),
			},
		}, nil
	default:
		return QueryPlan{}, ErrInvalidMatchQuery
	}
}
