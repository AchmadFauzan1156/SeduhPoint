package repository

// laporan.go: query laporan pemilik (F-28 – F-30) dan ringkasan dashboard.
// Hanya transaksi berstatus SELESAI yang dihitung (F-24). Rentang tanggal dikirim sebagai
// string "YYYY-MM-DD" dengan batas atas eksklusif, sehingga tidak ada konversi zona waktu.

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type LaporanRepo struct {
	DB *sql.DB
}

// BarisBiayaPromo: satu baris laporan biaya promo (F-28).
type BarisBiayaPromo struct {
	PromoID       int
	Nama          string
	Jenis         string
	JumlahPakai   int
	TotalPotongan int64
}

// BarisCabang: transaksi, omzet dan potongan promo per cabang.
type BarisCabang struct {
	CabangID     int
	Nama         string
	Transaksi    int
	Batal        int
	Subtotal     int64 // penjualan sebelum potongan
	TotalPromo   int64
	PotonganPoin int64
	Omzet        int64 // uang yang diterima (total bayar)
}

// LotKedaluwarsa: poin aktif milik member yang akan kedaluwarsa pada satu tanggal.
type LotKedaluwarsa struct {
	MemberID    int
	Nama        string
	NoHP        string
	Poin        int
	Kedaluwarsa time.Time
}

// PoinPerBulan: sebaran poin aktif menurut bulan kedaluwarsanya.
type PoinPerBulan struct {
	Bulan time.Time
	Poin  int
}

// MemberTingkat: data mentah untuk laporan perkembangan member (F-30).
type MemberTingkat struct {
	MemberID  int
	TglDaftar time.Time
	Riwayat   []RiwayatTingkat // urut berlaku_sejak
}

type RiwayatTingkat struct {
	Tingkat string
	Sejak   time.Time
}

const fmtWaktu = "2006-01-02 15:04:05"

// BiayaPromo: total potongan & jumlah transaksi per promo (F-28).
func (r *LaporanRepo) BiayaPromo(ctx context.Context, dari, sampai string, cabangID int) ([]BarisBiayaPromo, error) {
	q := `
		SELECT p.id, p.nama, p.jenis, COUNT(DISTINCT tp.transaksi_id), COALESCE(SUM(tp.nilai_potongan), 0)
		FROM transaksi_promo tp
		JOIN transaksi t ON t.id = tp.transaksi_id AND t.status = 'SELESAI'
		JOIN promo p ON p.id = tp.promo_id
		WHERE t.waktu >= ? AND t.waktu < ?`
	args := []any{dari, sampai}
	if cabangID > 0 {
		q += " AND t.cabang_id = ?"
		args = append(args, cabangID)
	}
	rows, err := r.DB.QueryContext(ctx, q+" GROUP BY p.id, p.nama, p.jenis ORDER BY 5 DESC, p.nama", args...)
	if err != nil {
		return nil, fmt.Errorf("laporan biaya promo: %w", err)
	}
	defer rows.Close()
	var hasil []BarisBiayaPromo
	for rows.Next() {
		var b BarisBiayaPromo
		if err := rows.Scan(&b.PromoID, &b.Nama, &b.Jenis, &b.JumlahPakai, &b.TotalPotongan); err != nil {
			return nil, err
		}
		hasil = append(hasil, b)
	}
	return hasil, rows.Err()
}

// PerCabang: ringkasan transaksi per cabang pada rentang tanggal (cabang tanpa transaksi tetap tampil).
func (r *LaporanRepo) PerCabang(ctx context.Context, dari, sampai string, cabangID int) ([]BarisCabang, error) {
	q := `
		SELECT c.id, c.nama,
		  COUNT(CASE WHEN t.status = 'SELESAI' THEN 1 END),
		  COUNT(CASE WHEN t.status = 'DIBATALKAN' THEN 1 END),
		  COALESCE(SUM(CASE WHEN t.status = 'SELESAI' THEN t.subtotal END), 0),
		  COALESCE(SUM(CASE WHEN t.status = 'SELESAI' THEN t.total_promo END), 0),
		  COALESCE(SUM(CASE WHEN t.status = 'SELESAI' THEN t.potongan_poin END), 0),
		  COALESCE(SUM(CASE WHEN t.status = 'SELESAI' THEN t.total_bayar END), 0)
		FROM cabang c
		LEFT JOIN transaksi t ON t.cabang_id = c.id AND t.waktu >= ? AND t.waktu < ?
		WHERE c.deleted_at IS NULL`
	args := []any{dari, sampai}
	if cabangID > 0 {
		q += " AND c.id = ?"
		args = append(args, cabangID)
	}
	rows, err := r.DB.QueryContext(ctx, q+" GROUP BY c.id, c.nama ORDER BY 8 DESC, c.nama", args...)
	if err != nil {
		return nil, fmt.Errorf("laporan per cabang: %w", err)
	}
	defer rows.Close()
	var hasil []BarisCabang
	for rows.Next() {
		var b BarisCabang
		if err := rows.Scan(&b.CabangID, &b.Nama, &b.Transaksi, &b.Batal, &b.Subtotal, &b.TotalPromo, &b.PotonganPoin, &b.Omzet); err != nil {
			return nil, err
		}
		hasil = append(hasil, b)
	}
	return hasil, rows.Err()
}

// PoinAktif: total poin yang masih berlaku dan jumlah member pemiliknya (F-29).
func (r *LaporanRepo) PoinAktif(ctx context.Context, sekarang time.Time) (poin, member int, err error) {
	err = r.DB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(l.poin_sisa), 0), COUNT(DISTINCT l.member_id)
		FROM poin_lot l JOIN member m ON m.id = l.member_id AND m.deleted_at IS NULL
		WHERE l.status = 'AKTIF' AND l.poin_sisa > 0 AND l.tgl_kedaluwarsa > ?`,
		sekarang.Format(fmtWaktu)).Scan(&poin, &member)
	if err != nil {
		err = fmt.Errorf("laporan poin aktif: %w", err)
	}
	return
}

// AkanKedaluwarsa: poin aktif yang habis masa berlakunya sampai `batas`, per member & tanggal.
func (r *LaporanRepo) AkanKedaluwarsa(ctx context.Context, sekarang, batas time.Time) ([]LotKedaluwarsa, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT m.id, m.nama, m.no_hp, SUM(l.poin_sisa), DATE(l.tgl_kedaluwarsa) AS tgl
		FROM poin_lot l JOIN member m ON m.id = l.member_id AND m.deleted_at IS NULL
		WHERE l.status = 'AKTIF' AND l.poin_sisa > 0 AND l.tgl_kedaluwarsa > ? AND l.tgl_kedaluwarsa <= ?
		GROUP BY m.id, m.nama, m.no_hp, tgl
		ORDER BY tgl, 4 DESC`,
		sekarang.Format(fmtWaktu), batas.Format(fmtWaktu))
	if err != nil {
		return nil, fmt.Errorf("poin akan kedaluwarsa: %w", err)
	}
	defer rows.Close()
	var hasil []LotKedaluwarsa
	for rows.Next() {
		var l LotKedaluwarsa
		if err := rows.Scan(&l.MemberID, &l.Nama, &l.NoHP, &l.Poin, &l.Kedaluwarsa); err != nil {
			return nil, err
		}
		hasil = append(hasil, l)
	}
	return hasil, rows.Err()
}

// PoinPerBulanKedaluwarsa: sebaran poin aktif per bulan kedaluwarsa (paling jauh 6 bulan ke depan, AB02).
func (r *LaporanRepo) PoinPerBulanKedaluwarsa(ctx context.Context, sekarang time.Time) ([]PoinPerBulan, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT DATE_FORMAT(l.tgl_kedaluwarsa, '%Y-%m-01') AS bln, SUM(l.poin_sisa)
		FROM poin_lot l JOIN member m ON m.id = l.member_id AND m.deleted_at IS NULL
		WHERE l.status = 'AKTIF' AND l.poin_sisa > 0 AND l.tgl_kedaluwarsa > ?
		GROUP BY bln ORDER BY bln`, sekarang.Format(fmtWaktu))
	if err != nil {
		return nil, fmt.Errorf("poin per bulan: %w", err)
	}
	defer rows.Close()
	var hasil []PoinPerBulan
	for rows.Next() {
		var s string
		var p PoinPerBulan
		if err := rows.Scan(&s, &p.Poin); err != nil {
			return nil, err
		}
		p.Bulan, _ = time.Parse("2006-01-02", s)
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

// RiwayatMember: semua member (belum dihapus) beserta riwayat tingkatnya, untuk F-30.
func (r *LaporanRepo) RiwayatMember(ctx context.Context) ([]MemberTingkat, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT m.id, m.tgl_daftar, h.tingkat, h.berlaku_sejak
		FROM member m LEFT JOIN member_tingkat_riwayat h ON h.member_id = m.id
		WHERE m.deleted_at IS NULL
		ORDER BY m.id, h.berlaku_sejak, h.id`)
	if err != nil {
		return nil, fmt.Errorf("riwayat member: %w", err)
	}
	defer rows.Close()
	var hasil []MemberTingkat
	for rows.Next() {
		var id int
		var daftar time.Time
		var tingkat sql.NullString
		var sejak sql.NullTime
		if err := rows.Scan(&id, &daftar, &tingkat, &sejak); err != nil {
			return nil, err
		}
		if len(hasil) == 0 || hasil[len(hasil)-1].MemberID != id {
			hasil = append(hasil, MemberTingkat{MemberID: id, TglDaftar: daftar})
		}
		if tingkat.Valid && sejak.Valid {
			m := &hasil[len(hasil)-1]
			m.Riwayat = append(m.Riwayat, RiwayatTingkat{Tingkat: tingkat.String, Sejak: sejak.Time})
		}
	}
	return hasil, rows.Err()
}
