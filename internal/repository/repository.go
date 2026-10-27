// Package repository: satu-satunya tempat ada SQL.
// Setiap repo berupa struct yang menyimpan *sql.DB.
package repository

import (
	"context"
	"database/sql"

	"seduhpoint/internal/model"
)

// Querier dipenuhi *sql.DB dan *sql.Tx, sehingga method tulis
// bisa dipanggil di dalam database transaction.
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func pastikanAdaBaris(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}

// WithTx menjalankan fn dalam satu transaction; error -> rollback.
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
