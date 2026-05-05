# Study Graph DB Roadmap

## v1: In-memory node/edge/label search
- Maintain in-memory node and edge store
- Label based node lookup

## v2: Property index
- Build exact-match property index
- Use index in lookups

## v3: Edge-type index and adjacency list
- Maintain edge-type index
- Add outbound/inbound adjacency lists

## v4: Simple MATCH query
- Parse and execute a very small MATCH syntax

## v5: QueryPlan / Explain
- Build a query plan object
- Return explain output for a query

## v6: Add/remove index
- Support index lifecycle operations

## v7: Persistence
- Save and load graph data from disk

## v8: Transaction or WAL
- Add transaction boundary or write-ahead log
