package repository

// F-03, F-04, F-26 — TODO M2

import (
	"context"
	"database/sql"
	"time"

	"seduhpoint/internal/model"
)

type MemberRepo struct {
	DB *sql.DB
}

// Cari: nama LIKE ? OR no_hp LIKE ?
func (r *MemberRepo) Cari(ctx context.Context, q string) ([]model.Member, error) { return nil, nil }
func (r *MemberRepo) CariByID(ctx context.Context, id int) (model.Member, error) {
	return model.Member{}, model.ErrNotFound
}

// Tambah: member baru otomatis SILVER + baris member_tingkat_riwayat (pakai tx).
func (r *MemberRepo) Tambah(ctx context.Context, q Querier, m model.Member) (int64, error) {
	return 0, nil
}
func (r *MemberRepo) Ubah(ctx context.Context, m model.Member) error { return nil }

// TotalBelanja3Bulan: SUM(total_setelah_promo) transaksi SELESAI sejak `sejak` (AB06, asumsi 4).
func (r *MemberRepo) TotalBelanjaSejak(ctx context.Context, q Querier, memberID int, sejak time.Time) (int64, error) {
	return 0, nil
}

// UbahTingkat: update tingkat_saat_ini + INSERT member_tingkat_riwayat.
func (r *MemberRepo) UbahTingkat(ctx context.Context, q Querier, memberID int, tingkat string) error {
	return nil
}
