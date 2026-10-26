package repository

// Poin per lot (FIFO & kedaluwarsa) + buku besar mutasi (NF-05: hanya INSERT).

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"seduhpoint/internal/model"
)

type PoinRepo struct {
	DB *sql.DB
}

const kolomLot = "id, member_id, transaksi_id, poin_awal, poin_sisa, tgl_perolehan, tgl_kedaluwarsa, status"

func scanLot(s interface{ Scan(...any) error }, l *model.PoinLot) error {
	return s.Scan(&l.ID, &l.MemberID, &l.TransaksiID, &l.PoinAwal, &l.PoinSisa, &l.TglPerolehan, &l.TglKedaluwarsa, &l.Status)
}

func ambilLot(ctx context.Context, q Querier, query string, args ...any) ([]model.PoinLot, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hasil []model.PoinLot
	for rows.Next() {
		var l model.PoinLot
		if err := scanLot(rows, &l); err != nil {
			return nil, err
		}
		hasil = append(hasil, l)
	}
	return hasil, rows.Err()
}

// LotAktifFIFO: lot aktif yang belum kedaluwarsa, paling dekat kedaluwarsa dulu, dikunci (AB04).
func (r *PoinRepo) LotAktifFIFO(ctx context.Context, q Querier, memberID int, sekarang time.Time) ([]model.PoinLot, error) {
	return ambilLot(ctx, q, "SELECT "+kolomLot+` FROM poin_lot
		WHERE member_id = ? AND status = 'AKTIF' AND poin_sisa > 0 AND tgl_kedaluwarsa > ?
		ORDER BY tgl_kedaluwarsa, id FOR UPDATE`, memberID, sekarang)
}

func (r *PoinRepo) LotForUpdate(ctx context.Context, q Querier, lotID int) (model.PoinLot, error) {
	var l model.PoinLot
	err := scanLot(q.QueryRowContext(ctx, "SELECT "+kolomLot+" FROM poin_lot WHERE id = ? FOR UPDATE", lotID), &l)
	return l, err
}

func (r *PoinRepo) TambahLot(ctx context.Context, q Querier, l model.PoinLot) (int64, error) {
	res, err := q.ExecContext(ctx, `INSERT INTO poin_lot (member_id, transaksi_id, poin_awal, poin_sisa, tgl_perolehan, tgl_kedaluwarsa, status)
		VALUES (?, ?, ?, ?, ?, ?, 'AKTIF')`, l.MemberID, l.TransaksiID, l.PoinAwal, l.PoinSisa, l.TglPerolehan, l.TglKedaluwarsa)
	if err != nil {
		return 0, fmt.Errorf("tambah lot poin: %w", err)
	}
	return res.LastInsertId()
}

func (r *PoinRepo) UbahSisaLot(ctx context.Context, q Querier, lotID, sisa int, status string) error {
	_, err := q.ExecContext(ctx, "UPDATE poin_lot SET poin_sisa = ?, status = ? WHERE id = ?", sisa, status, lotID)
	return err
}

// TambahMutasi: PEROLEHAN | TUKAR | KEMBALI | TARIK | KEDALUWARSA. jumlah + menambah, - mengurangi.
func (r *PoinRepo) TambahMutasi(ctx context.Context, q Querier, memberID, lotID int, trxID *int64, jenis string, jumlah int) error {
	_, err := q.ExecContext(ctx, `INSERT INTO poin_mutasi (member_id, lot_id, transaksi_id, jenis, jumlah, waktu)
		VALUES (?, ?, ?, ?, ?, NOW())`, memberID, lotID, trxID, jenis, jumlah)
	if err != nil {
		return fmt.Errorf("mutasi poin %s: %w", jenis, err)
	}
	return nil
}

// MutasiTransaksi: mutasi PEROLEHAN dan TUKAR milik sebuah transaksi (untuk pembatalan).
func (r *PoinRepo) MutasiTransaksi(ctx context.Context, q Querier, trxID int) ([]model.MutasiPoin, error) {
	rows, err := q.QueryContext(ctx, `SELECT id, lot_id, jenis, jumlah FROM poin_mutasi
		WHERE transaksi_id = ? AND jenis IN ('PEROLEHAN', 'TUKAR') ORDER BY id`, trxID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hasil []model.MutasiPoin
	for rows.Next() {
		var m model.MutasiPoin
		if err := rows.Scan(&m.ID, &m.LotID, &m.Jenis, &m.Jumlah); err != nil {
			return nil, err
		}
		hasil = append(hasil, m)
	}
	return hasil, rows.Err()
}

// KedaluwarsakanSemua: lot AKTIF yang lewat masa berlaku -> KEDALUWARSA, sisa dicatat sebagai mutasi (F-25).
func (r *PoinRepo) KedaluwarsakanSemua(ctx context.Context, sekarang time.Time) (int, error) {
	n := 0
	err := WithTx(ctx, r.DB, func(tx *sql.Tx) error {
		lot, err := ambilLot(ctx, tx, "SELECT "+kolomLot+" FROM poin_lot WHERE status = 'AKTIF' AND tgl_kedaluwarsa <= ? FOR UPDATE", sekarang)
		if err != nil {
			return err
		}
		for _, l := range lot {
			if l.PoinSisa > 0 {
				if err := r.TambahMutasi(ctx, tx, l.MemberID, l.ID, nil, "KEDALUWARSA", -l.PoinSisa); err != nil {
					return err
				}
			}
			if err := r.UbahSisaLot(ctx, tx, l.ID, 0, "KEDALUWARSA"); err != nil {
				return err
			}
			n++
		}
		return nil
	})
	return n, err
}

// LotMember: semua lot poin member, yang aktif & paling dekat kedaluwarsa di atas (F-27).
func (r *PoinRepo) LotMember(ctx context.Context, memberID int) ([]model.PoinLot, error) {
	return ambilLot(ctx, r.DB, `SELECT id, member_id, transaksi_id, poin_awal, poin_sisa, tgl_perolehan, tgl_kedaluwarsa,
		CASE WHEN status = 'AKTIF' AND tgl_kedaluwarsa <= NOW() THEN 'KEDALUWARSA' ELSE status END
		FROM poin_lot WHERE member_id = ?
		ORDER BY (status = 'AKTIF' AND tgl_kedaluwarsa > NOW()) DESC, tgl_kedaluwarsa`, memberID)
}
