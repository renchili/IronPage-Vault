// Package store contains persistence adapters and database-facing code.
//
// Store code owns SQL queries, repository-style functions, and transaction
// helpers. It must not contain HTTP routing, Echo contexts, or response bodies.
package store
