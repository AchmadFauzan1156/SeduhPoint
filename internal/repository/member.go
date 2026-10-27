package repository

// Member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"seduhpoint/internal/model"
)

type MemberRepo struct {
	DB *sql.DB
}

// kolomMember juga menghitung poin aktif dan belanja 3 bulan terakhir (transaksi SELESAI, setelah promo).
const kolomMember = `m.id, m.nama, m.no_hp, m.tgl_daftar, m.tingkat_saat_ini,
	COALESCE((SELECT SUM(l.poin_sisa) FROM poin_lot l WHERE l.member_id = m.id AND l.status = 'AKTIF' AND l.tgl_kedaluwarsa > NOW()), 0),
	COALESCE((SELECT SUM(t.subtotal - t.total_promo) FROM transaksi t WHERE t.member_id = m.id AND t.status = 'SELESAI'
	          AND t.waktu >= DATE_SUB(NOW(), INTERVAL 3 MONTH)), 0)`

func scanMember(s interface{ Scan(...any) error }, m *model.Member) error {
	return s.Scan(&m.ID, &m.Nama, &m.NoHP, &m.TglDaftar, &m.Tingkat, &m.PoinAktif, &m.Belanja3Bulan)
}

// Cari: nama atau nomor HP. q kosong = semua member.
func (r *MemberRepo) Cari(ctx context.Context, q string) ([]model.Member, error) {
	pola := "%" + q + "%"
	rows, err := r.DB.QueryContext(ctx, "SELECT "+kolomMember+` FROM member m
		WHERE m.deleted_at IS NULL AND (m.nama LIKE ? OR m.no_hp LIKE ?) ORDER BY m.nama LIMIT 200`, pola, pola)
	if err != nil {
		return nil, fmt.Errorf("cari member: %w", err)
	}
	defer rows.Close()
	var hasil []model.Member
	for rows.Next() {
		var m model.Member
		if err := scanMember(rows, &m); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		hasil = append(hasil, m)
	}
	return hasil, rows.Err()
}

func (r *MemberRepo) CariByID(ctx context.Context, id int) (model.Member, error) {
	var m model.Member
	err := scanMember(r.DB.QueryRowContext(ctx, "SELECT "+kolomMember+" FROM member m WHERE m.id = ? AND m.deleted_at IS NULL", id), &m)
	if errors.Is(err, sql.ErrNoRows) {
		return m, model.ErrNotFound
	}
	if err != nil {
		return m, fmt.Errorf("cari member %d: %w", id, err)
	}
	return m, nil
}

func (r *MemberRepo) Hitung(ctx context.Context) (total, gold int, err error) {
	err = r.DB.QueryRowContext(ctx,
		"SELECT COUNT(*), COALESCE(SUM(tingkat_saat_ini = 'GOLD'), 0) FROM member WHERE deleted_at IS NULL").Scan(&total, &gold)
	return
}

func (r *MemberRepo) NoHPDipakai(ctx context.Context, noHP string, kecualiID int) (bool, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM member WHERE no_hp = ? AND id <> ?", noHP, kecualiID).Scan(&n)
	return n > 0, err
}

// Tambah: member baru otomatis SILVER + baris riwayat tingkat (dipanggil di dalam tx).
func (r *MemberRepo) Tambah(ctx context.Context, q Querier, m model.Member) (int64, error) {
	res, err := q.ExecContext(ctx,
		"INSERT INTO member (nama, no_hp, tgl_daftar, tingkat_saat_ini) VALUES (?, ?, CURDATE(), 'SILVER')", m.Nama, m.NoHP)
	if err != nil {
		return 0, fmt.Errorf("tambah member: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	_, err = q.ExecContext(ctx,
		"INSERT INTO member_tingkat_riwayat (member_id, tingkat, berlaku_sejak) VALUES (?, 'SILVER', NOW())", id)
	return id, err
}

func (r *MemberRepo) Ubah(ctx context.Context, m model.Member) error {
	res, err := r.DB.ExecContext(ctx,
		"UPDATE member SET nama = ?, no_hp = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", m.Nama, m.NoHP, m.ID)
	if err != nil {
		return fmt.Errorf("ubah member %d: %w", m.ID, err)
	}
	return pastikanAdaBaris(res)
}

// TotalBelanjaSejak: SUM(subtotal - total_promo) transaksi SELESAI sejak `sejak`. Dipakai evaluasi tingkat.
func (r *MemberRepo) TotalBelanjaSejak(ctx context.Context, q Querier, memberID int, sejak time.Time) (int64, error) {
	var total int64
	err := q.QueryRowContext(ctx, `SELECT COALESCE(SUM(subtotal - total_promo), 0) FROM transaksi
		WHERE member_id = ? AND status = 'SELESAI' AND waktu >= ?`, memberID, sejak).Scan(&total)
	return total, err
}

// UbahTingkat: update tingkat_saat_ini + INSERT member_tingkat_riwayat. Dipakai evaluasi tingkat.
func (r *MemberRepo) UbahTingkat(ctx context.Context, q Querier, memberID int, tingkat string) error {
	if _, err := q.ExecContext(ctx, "UPDATE member SET tingkat_saat_ini = ?, updated_at = NOW() WHERE id = ?", tingkat, memberID); err != nil {
		return err
	}
	_, err := q.ExecContext(ctx,
		"INSERT INTO member_tingkat_riwayat (member_id, tingkat, berlaku_sejak) VALUES (?, ?, NOW())", memberID, tingkat)
	return err
}

// Tingkat membaca tingkat member saat ini (di dalam tx, dikunci).
func (r *MemberRepo) Tingkat(ctx context.Context, q Querier, memberID int) (string, error) {
	var t string
	err := q.QueryRowContext(ctx, "SELECT tingkat_saat_ini FROM member WHERE id = ? FOR UPDATE", memberID).Scan(&t)
	return t, err
}

// SemuaID: id semua member aktif (untuk evaluasi tingkat harian).
func (r *MemberRepo) SemuaID(ctx context.Context) ([]int, error) {
	return ambilID(ctx, r.DB, "SELECT id FROM member WHERE deleted_at IS NULL")
}
