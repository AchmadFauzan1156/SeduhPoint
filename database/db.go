package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"seduhpoint/config"
)

func Open(c config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FJakarta&multiStatements=true",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

// WithTx menjalankan fn di dalam satu database transaction (NF-02).
// Jika fn mengembalikan error, semua perubahan di-rollback.
func WithTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
