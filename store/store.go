// Package store berisi semua query SQL (akses database).
// Pola: satu file per modul. Selalu pakai placeholder "?" (anti SQL injection).
package store

import "database/sql"

type Store struct{ DB *sql.DB }

func New(db *sql.DB) *Store { return &Store{DB: db} }

// Querier dipenuhi *sql.DB dan *sql.Tx, sehingga fungsi yang sama
// bisa dipanggil di dalam maupun di luar transaction.
type Querier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}
