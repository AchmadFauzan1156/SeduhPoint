package repository

// Poin per lot (FIFO) + buku besar mutasi — TODO M3

import (
	"context"
	"database/sql"

	"seduhpoint/internal/model"
)

type PoinRepo struct {
	DB *sql.DB
}

// LotAktifFIFO: lot AKTIF urut tgl_kedaluwarsa ASC ... FOR UPDATE (AB04).
func (r *PoinRepo) LotAktifFIFO(ctx context.Context, q Querier, memberID int) ([]model.PoinLot, error) {
	return nil, nil
}
func (r *PoinRepo) Saldo(ctx context.Context, memberID int) (int, error) { return 0, nil }
func (r *PoinRepo) TambahLot(ctx context.Context, q Querier, l model.PoinLot) (int64, error) {
	return 0, nil
}
func (r *PoinRepo) UbahSisaLot(ctx context.Context, q Querier, lotID, sisa int, status string) error {
	return nil
}

// TambahMutasi: PEROLEHAN | TUKAR | KEMBALI | TARIK | KEDALUWARSA. Hanya INSERT (NF-05).
func (r *PoinRepo) TambahMutasi(ctx context.Context, q Querier, memberID, lotID int, trxID *int64, jenis string, jumlah int) error {
	return nil
}

// KedaluwarsakanSemua: lot AKTIF yang lewat masa berlaku -> KEDALUWARSA + mutasi (F-25).
func (r *PoinRepo) KedaluwarsakanSemua(ctx context.Context) (int, error) { return 0, nil }
