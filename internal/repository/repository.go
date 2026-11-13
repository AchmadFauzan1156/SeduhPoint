// Package repository: satu-satunya tempat ada SQL.
// Setiap repo berupa struct yang menyimpan *sql.DB.
package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"

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

// BisaDiulang: error karena dua transaksi database berjalan bersamaan dan bisa dicoba lagi.
// Pola cek nomor error MySQL sama dengan materi #11 (1062 duplicate entry).
//   - 1062: duplicate entry, mis. dua request dengan token Bayar yang sama, atau nomor transaksi yang sama
//   - 1213: deadlock, MySQL membatalkan salah satu transaction
//   - 1205: terlalu lama menunggu kunci baris
func BisaDiulang(err error) bool {
	var me *mysql.MySQLError
	if !errors.As(err, &me) {
		return false
	}
	return me.Number == 1062 || me.Number == 1213 || me.Number == 1205
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
