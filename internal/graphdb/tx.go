package graphdb

type Tx interface {
	Commit() error
	Rollback() error
}
