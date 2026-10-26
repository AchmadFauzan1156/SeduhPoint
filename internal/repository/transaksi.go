package repository

// F-12 – F-24 Transaksi. Method tulis menerima Querier agar dipanggil di dalam satu transaction.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"seduhpoint/internal/model"
)

type TransaksiRepo struct {
	DB *sql.DB
}

const kolomTransaksi = `t.id, t.no_transaksi, t.cabang_id, t.pengguna_id, t.member_id, t.waktu, t.subtotal, t.total_promo,
	t.poin_ditukar, t.potongan_poin, t.total_bayar, t.poin_diperoleh, t.status, t.alasan_batal, t.waktu_batal, t.token,
	c.nama, u.username, COALESCE(m.nama, '')`

const dariTransaksi = ` FROM transaksi t JOIN cabang c ON c.id = t.cabang_id JOIN pengguna u ON u.id = t.pengguna_id
	LEFT JOIN member m ON m.id = t.member_id`

func scanTransaksi(s interface{ Scan(...any) error }, t *model.Transaksi) error {
	return s.Scan(&t.ID, &t.NoTransaksi, &t.CabangID, &t.PenggunaID, &t.MemberID, &t.Waktu, &t.Subtotal, &t.TotalPromo,
		&t.PoinDitukar, &t.PotonganPoin, &t.TotalBayar, &t.PoinDiperoleh, &t.Status, &t.AlasanBatal, &t.WaktuBatal, &t.Token,
		&t.CabangNama, &t.KasirNama, &t.MemberNama)
}

// CariByToken: dipakai mencegah transaksi ganda saat tombol Bayar ditekan dua kali (V08).
func (r *TransaksiRepo) CariByToken(ctx context.Context, q Querier, token string) (int64, bool, error) {
	var id int64
	err := q.QueryRowContext(ctx, "SELECT id FROM transaksi WHERE token = ?", token).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	return id, err == nil, err
}

// NomorBerikut: SP-<cabang>-<yyyymmdd>-<urut 4 digit>.
func (r *TransaksiRepo) NomorBerikut(ctx context.Context, q Querier, cabangID int, waktu time.Time) (string, error) {
	awalan := fmt.Sprintf("SP-%d-%s-", cabangID, waktu.Format("20060102"))
	var n int
	if err := q.QueryRowContext(ctx, "SELECT COUNT(*) FROM transaksi WHERE no_transaksi LIKE ?", awalan+"%").Scan(&n); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%04d", awalan, n+1), nil
}

func (r *TransaksiRepo) Tambah(ctx context.Context, q Querier, t model.Transaksi) (int64, error) {
	res, err := q.ExecContext(ctx, `INSERT INTO transaksi (no_transaksi, token, cabang_id, pengguna_id, member_id, waktu,
		subtotal, total_promo, poin_ditukar, potongan_poin, total_bayar, poin_diperoleh, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'SELESAI')`,
		t.NoTransaksi, t.Token, t.CabangID, t.PenggunaID, t.MemberID, t.Waktu, t.Subtotal, t.TotalPromo,
		t.PoinDitukar, t.PotonganPoin, t.TotalBayar, t.PoinDiperoleh)
	if err != nil {
		return 0, fmt.Errorf("tambah transaksi: %w", err)
	}
	return res.LastInsertId()
}

func (r *TransaksiRepo) TambahItem(ctx context.Context, q Querier, trxID int64, produkID, qty int, harga int64) error {
	_, err := q.ExecContext(ctx, "INSERT INTO transaksi_item (transaksi_id, produk_id, qty, harga_satuan) VALUES (?, ?, ?, ?)",
		trxID, produkID, qty, harga)
	return err
}

func (r *TransaksiRepo) TambahPromo(ctx context.Context, q Querier, trxID int64, promoID int, nilai int64) error {
	_, err := q.ExecContext(ctx, "INSERT INTO transaksi_promo (transaksi_id, promo_id, nilai_potongan) VALUES (?, ?, ?)",
		trxID, promoID, nilai)
	return err
}

func (r *TransaksiRepo) CariByID(ctx context.Context, id int) (model.Transaksi, error) {
	return r.cari(ctx, r.DB, id, "")
}

// CariForUpdate mengunci baris transaksi (dipakai saat pembatalan).
func (r *TransaksiRepo) CariForUpdate(ctx context.Context, q Querier, id int) (model.Transaksi, error) {
	return r.cari(ctx, q, id, " FOR UPDATE")
}

func (r *TransaksiRepo) cari(ctx context.Context, q Querier, id int, kunci string) (model.Transaksi, error) {
	var t model.Transaksi
	err := scanTransaksi(q.QueryRowContext(ctx, "SELECT "+kolomTransaksi+dariTransaksi+" WHERE t.id = ?"+kunci, id), &t)
	if errors.Is(err, sql.ErrNoRows) {
		return t, model.ErrNotFound
	}
	if err != nil {
		return t, fmt.Errorf("cari transaksi %d: %w", id, err)
	}
	return t, nil
}

// Semua: transaksi pada satu tanggal; cabangID 0 = semua cabang (admin).
func (r *TransaksiRepo) Semua(ctx context.Context, cabangID int, tgl time.Time) ([]model.Transaksi, error) {
	q := "SELECT " + kolomTransaksi + dariTransaksi + " WHERE DATE(t.waktu) = ?"
	args := []any{tgl.Format("2006-01-02")}
	if cabangID > 0 {
		q += " AND t.cabang_id = ?"
		args = append(args, cabangID)
	}
	rows, err := r.DB.QueryContext(ctx, q+" ORDER BY t.waktu DESC", args...)
	if err != nil {
		return nil, fmt.Errorf("daftar transaksi: %w", err)
	}
	defer rows.Close()
	var hasil []model.Transaksi
	for rows.Next() {
		var t model.Transaksi
		if err := scanTransaksi(rows, &t); err != nil {
			return nil, err
		}
		hasil = append(hasil, t)
	}
	return hasil, rows.Err()
}

func (r *TransaksiRepo) Item(ctx context.Context, trxID int) ([]model.ItemStruk, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT p.nama, i.qty, i.harga_satuan FROM transaksi_item i
		JOIN produk p ON p.id = i.produk_id WHERE i.transaksi_id = ? ORDER BY i.id`, trxID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hasil []model.ItemStruk
	for rows.Next() {
		var it model.ItemStruk
		if err := rows.Scan(&it.Nama, &it.Qty, &it.Harga); err != nil {
			return nil, err
		}
		hasil = append(hasil, it)
	}
	return hasil, rows.Err()
}

func (r *TransaksiRepo) Promo(ctx context.Context, q Querier, trxID int) ([]model.PromoStruk, error) {
	rows, err := q.QueryContext(ctx, `SELECT p.id, p.nama, p.jenis, tp.nilai_potongan FROM transaksi_promo tp
		JOIN promo p ON p.id = tp.promo_id WHERE tp.transaksi_id = ? ORDER BY tp.id`, trxID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hasil []model.PromoStruk
	for rows.Next() {
		var p model.PromoStruk
		if err := rows.Scan(&p.PromoID, &p.Nama, &p.Jenis, &p.Nilai); err != nil {
			return nil, err
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

func (r *TransaksiRepo) TandaiBatal(ctx context.Context, q Querier, id int, alasan string) error {
	res, err := q.ExecContext(ctx, `UPDATE transaksi SET status = 'DIBATALKAN', alasan_batal = ?, waktu_batal = NOW()
		WHERE id = ? AND status = 'SELESAI'`, alasan, id)
	if err != nil {
		return fmt.Errorf("batalkan transaksi %d: %w", id, err)
	}
	return pastikanAdaBaris(res)
}
