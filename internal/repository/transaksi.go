package repository

// F-12 – F-24 — TODO M3. Method tulis menerima Querier (dipanggil di dalam tx).

import (
	"context"
	"database/sql"

	"seduhpoint/internal/model"
)

type TransaksiRepo struct {
	DB *sql.DB
}

func (r *TransaksiRepo) Tambah(ctx context.Context, q Querier, t model.Transaksi) (int64, error) {
	return 0, nil
}
func (r *TransaksiRepo) TambahItem(ctx context.Context, q Querier, trxID int64, produkID, qty int, harga int64) error {
	return nil
}
func (r *TransaksiRepo) TambahPromo(ctx context.Context, q Querier, trxID int64, promoID int, nilai int64) error {
	return nil
}
func (r *TransaksiRepo) CariByID(ctx context.Context, id int) (model.Transaksi, error) {
	return model.Transaksi{}, model.ErrNotFound
}

// Semua: cabangID 0 = semua cabang (admin).
func (r *TransaksiRepo) Semua(ctx context.Context, cabangID int) ([]model.Transaksi, error) {
	return nil, nil
}
func (r *TransaksiRepo) TandaiBatal(ctx context.Context, q Querier, id int, alasan string) error {
	return nil
}
