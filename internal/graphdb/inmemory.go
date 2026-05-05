package graphdb

import "sync"

// InMemoryEngine is a minimal scaffold for iterative implementation.
type InMemoryEngine struct {
	mu sync.RWMutex
}

func NewInMemoryEngine() *InMemoryEngine {
	return &InMemoryEngine{}
}

func (e *InMemoryEngine) CreateNode(_ []string, _ map[string]any) (NodeID, error) {
	return 0, ErrNotImplemented
}

func (e *InMemoryEngine) CreateEdge(_ NodeID, _ NodeID, _ string, _ map[string]any) (EdgeID, error) {
	return 0, ErrNotImplemented
}

func (e *InMemoryEngine) FindNodesByLabel(_ string) ([]Node, error) {
	return nil, ErrNotImplemented
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
