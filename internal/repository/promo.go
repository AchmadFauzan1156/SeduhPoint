package repository

// F-06 – F-14 — TODO M2/M3

import (
	"context"
	"database/sql"
	"time"

	"seduhpoint/internal/model"
)

type PromoRepo struct {
	DB *sql.DB
}

func (r *PromoRepo) Semua(ctx context.Context) ([]model.Promo, error) { return nil, nil }
func (r *PromoRepo) CariByID(ctx context.Context, id int) (model.Promo, error) {
	return model.Promo{}, model.ErrNotFound
}

// Tambah/Ubah juga menulis promo_cabang, promo_produk, promo_kategori (pakai tx).
func (r *PromoRepo) Tambah(ctx context.Context, p model.Promo) (int64, error) { return 0, nil }
func (r *PromoRepo) Ubah(ctx context.Context, p model.Promo) error            { return nil }
func (r *PromoRepo) SetAktif(ctx context.Context, id int, aktif bool) error   { return nil }

// BerlakuUntuk: promo aktif pada cabang & tanggal tertentu (cek jam di service).
func (r *PromoRepo) BerlakuUntuk(ctx context.Context, cabangID int, t time.Time) ([]model.Promo, error) {
	return nil, nil
}

// VoucherForUpdate: SELECT ... WHERE kode_voucher = ? FOR UPDATE (kunci kuota).
func (r *PromoRepo) VoucherForUpdate(ctx context.Context, q Querier, kode string) (model.Promo, error) {
	return model.Promo{}, model.ErrNotFound
}
func (r *PromoRepo) TambahPemakaian(ctx context.Context, q Querier, id, delta int) error { return nil }
