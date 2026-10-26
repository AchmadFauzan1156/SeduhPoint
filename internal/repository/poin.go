package repository

// Poin per lot (FIFO) + buku besar mutasi — TODO M3

import (
	"context"
	"database/sql"
	"fmt"

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

// LotMember: semua lot poin member, yang aktif & paling dekat kedaluwarsa di atas (F-27).
func (r *PoinRepo) LotMember(ctx context.Context, memberID int) ([]model.PoinLot, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, member_id, transaksi_id, poin_awal, poin_sisa, tgl_perolehan, tgl_kedaluwarsa,
		       CASE WHEN status = 'AKTIF' AND tgl_kedaluwarsa <= NOW() THEN 'KEDALUWARSA' ELSE status END
		FROM poin_lot WHERE member_id = ?
		ORDER BY (status = 'AKTIF' AND tgl_kedaluwarsa > NOW()) DESC, tgl_kedaluwarsa`, memberID)
	if err != nil {
		return nil, fmt.Errorf("lot poin member %d: %w", memberID, err)
	}
	defer rows.Close()
	var hasil []model.PoinLot
	for rows.Next() {
		var l model.PoinLot
		if err := rows.Scan(&l.ID, &l.MemberID, &l.TransaksiID, &l.PoinAwal, &l.PoinSisa, &l.TglPerolehan, &l.TglKedaluwarsa, &l.Status); err != nil {
			return nil, err
		}
		hasil = append(hasil, l)
	}
	return hasil, rows.Err()
}
