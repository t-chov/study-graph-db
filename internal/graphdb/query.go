package graphdb

type ResultRow map[string]any

type ResultSet struct {
	Rows []ResultRow
}

type QueryPlan struct {
	Steps []string
}
