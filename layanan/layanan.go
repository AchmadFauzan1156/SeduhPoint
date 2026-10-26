// Package layanan berisi alur bisnis yang melibatkan beberapa tabel sekaligus
// dan harus berjalan dalam satu database transaction (NF-02).
package layanan

import (
	"database/sql"

	"seduhpoint/store"
)

type Layanan struct {
	DB    *sql.DB
	Store *store.Store
}

func New(db *sql.DB, st *store.Store) *Layanan { return &Layanan{DB: db, Store: st} }
