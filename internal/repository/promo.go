package repository

// F-06 – F-11 Promo & voucher (F-13, F-14 dipakai kasir di M3)

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"seduhpoint/internal/model"
)

type PromoRepo struct {
	DB *sql.DB
}

const kolomPromo = `p.id, p.nama, p.jenis, p.nilai_persen, p.nilai_nominal, p.min_belanja, p.kode_voucher,
	p.kuota, p.kuota_terpakai, p.tgl_mulai, p.tgl_selesai, p.jam_mulai, p.jam_selesai, p.aktif,
	COALESCE((SELECT GROUP_CONCAT(c.nama ORDER BY c.id SEPARATOR ', ') FROM promo_cabang pc JOIN cabang c ON c.id = pc.cabang_id WHERE pc.promo_id = p.id), ''),
	COALESCE((SELECT GROUP_CONCAT(x.nama ORDER BY x.nama SEPARATOR ', ') FROM promo_produk pp JOIN produk x ON x.id = pp.produk_id WHERE pp.promo_id = p.id),
	         (SELECT GROUP_CONCAT(k.nama ORDER BY k.nama SEPARATOR ', ') FROM promo_kategori pk JOIN kategori k ON k.id = pk.kategori_id WHERE pk.promo_id = p.id), '')`

func scanPromo(s interface{ Scan(...any) error }, p *model.Promo) error {
	return s.Scan(&p.ID, &p.Nama, &p.Jenis, &p.NilaiPersen, &p.NilaiNominal, &p.MinBelanja, &p.KodeVoucher,
		&p.Kuota, &p.KuotaTerpakai, &p.TglMulai, &p.TglSelesai, &p.JamMulai, &p.JamSelesai, &p.Aktif,
		&p.CabangNama, &p.TargetNama)
}

func (r *PromoRepo) Semua(ctx context.Context) ([]model.Promo, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT "+kolomPromo+" FROM promo p WHERE p.deleted_at IS NULL ORDER BY p.aktif DESC, p.tgl_mulai DESC, p.id DESC")
	if err != nil {
		return nil, fmt.Errorf("daftar promo: %w", err)
	}
	defer rows.Close()
	var hasil []model.Promo
	for rows.Next() {
		var p model.Promo
		if err := scanPromo(rows, &p); err != nil {
			return nil, fmt.Errorf("scan promo: %w", err)
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

func (r *PromoRepo) CariByID(ctx context.Context, id int) (model.Promo, error) {
	var p model.Promo
	err := scanPromo(r.DB.QueryRowContext(ctx, "SELECT "+kolomPromo+" FROM promo p WHERE p.id = ? AND p.deleted_at IS NULL", id), &p)
	if errors.Is(err, sql.ErrNoRows) {
		return p, model.ErrNotFound
	}
	if err != nil {
		return p, fmt.Errorf("cari promo %d: %w", id, err)
	}
	for _, rel := range []struct {
		q   string
		dst *[]int
	}{
		{"SELECT cabang_id FROM promo_cabang WHERE promo_id = ?", &p.CabangIDs},
		{"SELECT produk_id FROM promo_produk WHERE promo_id = ?", &p.ProdukIDs},
		{"SELECT kategori_id FROM promo_kategori WHERE promo_id = ?", &p.KategoriIDs},
	} {
		if *rel.dst, err = ambilID(ctx, r.DB, rel.q, id); err != nil {
			return p, err
		}
	}
	return p, nil
}

func ambilID(ctx context.Context, q Querier, query string, args ...any) ([]int, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *PromoRepo) KodeDipakai(ctx context.Context, kode string, kecualiID int) (bool, error) {
	var n int
	err := r.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM promo WHERE kode_voucher = ? AND id <> ?", kode, kecualiID).Scan(&n)
	return n > 0, err
}

// Simpan: tambah (p.ID == 0) atau ubah, termasuk tabel relasi, dalam satu transaction.
func (r *PromoRepo) Simpan(ctx context.Context, p model.Promo) (int64, error) {
	id := int64(p.ID)
	err := WithTx(ctx, r.DB, func(tx *sql.Tx) error {
		args := []any{p.Nama, p.Jenis, p.NilaiPersen, p.NilaiNominal, p.MinBelanja, p.KodeVoucher, p.Kuota,
			p.TglMulai, p.TglSelesai, p.JamMulai, p.JamSelesai, p.Aktif}
		if p.ID == 0 {
			res, err := tx.ExecContext(ctx, `INSERT INTO promo (nama, jenis, nilai_persen, nilai_nominal, min_belanja,
				kode_voucher, kuota, tgl_mulai, tgl_selesai, jam_mulai, jam_selesai, aktif)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, args...)
			if err != nil {
				return fmt.Errorf("tambah promo: %w", err)
			}
			if id, err = res.LastInsertId(); err != nil {
				return err
			}
		} else {
			res, err := tx.ExecContext(ctx, `UPDATE promo SET nama = ?, jenis = ?, nilai_persen = ?, nilai_nominal = ?,
				min_belanja = ?, kode_voucher = ?, kuota = ?, tgl_mulai = ?, tgl_selesai = ?, jam_mulai = ?,
				jam_selesai = ?, aktif = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL`, append(args, p.ID)...)
			if err != nil {
				return fmt.Errorf("ubah promo %d: %w", p.ID, err)
			}
			if err := pastikanAdaBaris(res); err != nil {
				return err
			}
		}
		for _, rel := range []struct {
			tabel, kolom string
			ids          []int
		}{{"promo_cabang", "cabang_id", p.CabangIDs}, {"promo_produk", "produk_id", p.ProdukIDs}, {"promo_kategori", "kategori_id", p.KategoriIDs}} {
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+rel.tabel+" WHERE promo_id = ?", id); err != nil {
				return err
			}
			if len(rel.ids) == 0 {
				continue
			}
			nilai := strings.TrimSuffix(strings.Repeat("(?, ?),", len(rel.ids)), ",")
			var a []any
			for _, x := range rel.ids {
				a = append(a, id, x)
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO "+rel.tabel+" (promo_id, "+rel.kolom+") VALUES "+nilai, a...); err != nil {
				return fmt.Errorf("simpan %s: %w", rel.tabel, err)
			}
		}
		return nil
	})
	return id, err
}

func (r *PromoRepo) SetAktif(ctx context.Context, id int, aktif bool) error {
	res, err := r.DB.ExecContext(ctx, "UPDATE promo SET aktif = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", aktif, id)
	if err != nil {
		return fmt.Errorf("aktif promo %d: %w", id, err)
	}
	return pastikanAdaBaris(res)
}

// BerlakuUntuk: promo aktif yang tanggalnya mencakup `t` dan berlaku di cabang ini, beserta targetnya.
// Pengecekan jam dilakukan di service (PromoBerlaku) agar aturan WIB ada di satu tempat.
func (r *PromoRepo) BerlakuUntuk(ctx context.Context, cabangID int, t time.Time) ([]model.Promo, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT "+kolomPromo+` FROM promo p
		JOIN promo_cabang pc ON pc.promo_id = p.id AND pc.cabang_id = ?
		WHERE p.deleted_at IS NULL AND p.aktif = TRUE AND ? BETWEEN p.tgl_mulai AND p.tgl_selesai
		ORDER BY p.id`, cabangID, t.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("promo berlaku: %w", err)
	}
	var hasil []model.Promo
	for rows.Next() {
		var p model.Promo
		if err := scanPromo(rows, &p); err != nil {
			rows.Close()
			return nil, err
		}
		hasil = append(hasil, p)
	}
	rows.Close()
	for i := range hasil {
		if err := r.isiRelasi(ctx, r.DB, &hasil[i]); err != nil {
			return nil, err
		}
	}
	return hasil, nil
}

func (r *PromoRepo) isiRelasi(ctx context.Context, q Querier, p *model.Promo) error {
	var err error
	if p.CabangIDs, err = ambilID(ctx, q, "SELECT cabang_id FROM promo_cabang WHERE promo_id = ?", p.ID); err != nil {
		return err
	}
	if p.ProdukIDs, err = ambilID(ctx, q, "SELECT produk_id FROM promo_produk WHERE promo_id = ?", p.ID); err != nil {
		return err
	}
	p.KategoriIDs, err = ambilID(ctx, q, "SELECT kategori_id FROM promo_kategori WHERE promo_id = ?", p.ID)
	return err
}

// VoucherByKode: kunci = true memakai SELECT ... FOR UPDATE agar kuota aman saat checkout (F-14).
func (r *PromoRepo) VoucherByKode(ctx context.Context, q Querier, kode string, kunci bool) (model.Promo, error) {
	query := "SELECT " + kolomPromo + " FROM promo p WHERE p.kode_voucher = ? AND p.jenis = 'VOUCHER' AND p.deleted_at IS NULL"
	if kunci {
		query += " FOR UPDATE"
	}
	var p model.Promo
	err := scanPromo(q.QueryRowContext(ctx, query, kode), &p)
	if errors.Is(err, sql.ErrNoRows) {
		return p, model.ErrNotFound
	}
	if err != nil {
		return p, err
	}
	return p, r.isiRelasi(ctx, q, &p)
}

// TambahPemakaian: +1 saat voucher dipakai, -1 saat transaksinya dibatalkan.
func (r *PromoRepo) TambahPemakaian(ctx context.Context, q Querier, id, delta int) error {
	_, err := q.ExecContext(ctx, "UPDATE promo SET kuota_terpakai = GREATEST(kuota_terpakai + ?, 0) WHERE id = ?", delta, id)
	return err
}
