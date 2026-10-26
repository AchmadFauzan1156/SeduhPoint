package repository

// F-02 Kategori & Produk

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"seduhpoint/internal/model"
)

type KategoriRepo struct {
	DB *sql.DB
}

func (r *KategoriRepo) Semua(ctx context.Context) ([]model.Kategori, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT k.id, k.nama, COUNT(p.id)
		FROM kategori k LEFT JOIN produk p ON p.kategori_id = k.id AND p.deleted_at IS NULL
		WHERE k.deleted_at IS NULL GROUP BY k.id, k.nama ORDER BY k.nama`)
	if err != nil {
		return nil, fmt.Errorf("daftar kategori: %w", err)
	}
	defer rows.Close()
	var hasil []model.Kategori
	for rows.Next() {
		var k model.Kategori
		if err := rows.Scan(&k.ID, &k.Nama, &k.JumlahProduk); err != nil {
			return nil, fmt.Errorf("scan kategori: %w", err)
		}
		hasil = append(hasil, k)
	}
	return hasil, rows.Err()
}

func (r *KategoriRepo) CariByID(ctx context.Context, id int) (model.Kategori, error) {
	var k model.Kategori
	err := r.DB.QueryRowContext(ctx, `
		SELECT k.id, k.nama, (SELECT COUNT(*) FROM produk p WHERE p.kategori_id = k.id AND p.deleted_at IS NULL)
		FROM kategori k WHERE k.id = ? AND k.deleted_at IS NULL`, id).Scan(&k.ID, &k.Nama, &k.JumlahProduk)
	if errors.Is(err, sql.ErrNoRows) {
		return k, model.ErrNotFound
	}
	return k, err
}

// NamaDipakai: nama kategori unik di antara kategori yang belum dihapus.
func (r *KategoriRepo) NamaDipakai(ctx context.Context, nama string, kecualiID int) (bool, error) {
	var n int
	err := r.DB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM kategori WHERE nama = ? AND id <> ? AND deleted_at IS NULL", nama, kecualiID).Scan(&n)
	return n > 0, err
}

func (r *KategoriRepo) Tambah(ctx context.Context, k model.Kategori) (int64, error) {
	res, err := r.DB.ExecContext(ctx, "INSERT INTO kategori (nama) VALUES (?)", k.Nama)
	if err != nil {
		return 0, fmt.Errorf("tambah kategori: %w", err)
	}
	return res.LastInsertId()
}

func (r *KategoriRepo) Ubah(ctx context.Context, k model.Kategori) error {
	res, err := r.DB.ExecContext(ctx,
		"UPDATE kategori SET nama = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", k.Nama, k.ID)
	if err != nil {
		return fmt.Errorf("ubah kategori %d: %w", k.ID, err)
	}
	return pastikanAdaBaris(res)
}

// Hapus = soft delete. Nama diberi akhiran agar nama yang sama bisa dipakai lagi (kolom nama UNIQUE).
func (r *KategoriRepo) Hapus(ctx context.Context, id int) error {
	res, err := r.DB.ExecContext(ctx,
		"UPDATE kategori SET deleted_at = NOW(), nama = CONCAT(nama, ' #', id) WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return fmt.Errorf("hapus kategori %d: %w", id, err)
	}
	return pastikanAdaBaris(res)
}

type ProdukRepo struct {
	DB *sql.DB
}

const kolomProduk = "p.id, p.kategori_id, k.nama, p.nama, p.harga, COALESCE(p.gambar, ''), p.aktif"

func scanProduk(s interface{ Scan(...any) error }, p *model.Produk) error {
	return s.Scan(&p.ID, &p.KategoriID, &p.KategoriNama, &p.Nama, &p.Harga, &p.Gambar, &p.Aktif)
}

// Semua: hanyaAktif = true untuk halaman kasir (produk yang dijual saja).
func (r *ProdukRepo) Semua(ctx context.Context, hanyaAktif bool) ([]model.Produk, error) {
	q := "SELECT " + kolomProduk + " FROM produk p JOIN kategori k ON k.id = p.kategori_id WHERE p.deleted_at IS NULL"
	if hanyaAktif {
		q += " AND p.aktif = TRUE"
	}
	rows, err := r.DB.QueryContext(ctx, q+" ORDER BY k.nama, p.nama")
	if err != nil {
		return nil, fmt.Errorf("daftar produk: %w", err)
	}
	defer rows.Close()
	var hasil []model.Produk
	for rows.Next() {
		var p model.Produk
		if err := scanProduk(rows, &p); err != nil {
			return nil, fmt.Errorf("scan produk: %w", err)
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

func (r *ProdukRepo) CariByID(ctx context.Context, id int) (model.Produk, error) {
	var p model.Produk
	err := scanProduk(r.DB.QueryRowContext(ctx,
		"SELECT "+kolomProduk+" FROM produk p JOIN kategori k ON k.id = p.kategori_id WHERE p.id = ? AND p.deleted_at IS NULL", id), &p)
	if errors.Is(err, sql.ErrNoRows) {
		return p, model.ErrNotFound
	}
	if err != nil {
		return p, fmt.Errorf("cari produk %d: %w", id, err)
	}
	return p, nil
}

func nullStr(s string) sql.NullString { return sql.NullString{String: s, Valid: s != ""} }

func (r *ProdukRepo) Tambah(ctx context.Context, p model.Produk) (int64, error) {
	res, err := r.DB.ExecContext(ctx,
		"INSERT INTO produk (kategori_id, nama, harga, gambar, aktif) VALUES (?, ?, ?, ?, ?)",
		p.KategoriID, p.Nama, p.Harga, nullStr(p.Gambar), p.Aktif)
	if err != nil {
		return 0, fmt.Errorf("tambah produk: %w", err)
	}
	return res.LastInsertId()
}

func (r *ProdukRepo) Ubah(ctx context.Context, p model.Produk) error {
	res, err := r.DB.ExecContext(ctx, `
		UPDATE produk SET kategori_id = ?, nama = ?, harga = ?, gambar = ?, aktif = ?, updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL`,
		p.KategoriID, p.Nama, p.Harga, nullStr(p.Gambar), p.Aktif, p.ID)
	if err != nil {
		return fmt.Errorf("ubah produk %d: %w", p.ID, err)
	}
	return pastikanAdaBaris(res)
}

func (r *ProdukRepo) Hapus(ctx context.Context, id int) error {
	res, err := r.DB.ExecContext(ctx, "UPDATE produk SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return fmt.Errorf("hapus produk %d: %w", id, err)
	}
	return pastikanAdaBaris(res)
}

// DijualByID: produk yang masih dijual untuk id tertentu (harga diambil dari database, bukan dari form).
func (r *ProdukRepo) DijualByID(ctx context.Context, ids []int) (map[int]model.Produk, error) {
	hasil := map[int]model.Produk{}
	if len(ids) == 0 {
		return hasil, nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := r.DB.QueryContext(ctx, "SELECT "+kolomProduk+` FROM produk p JOIN kategori k ON k.id = p.kategori_id
		WHERE p.deleted_at IS NULL AND p.aktif = TRUE AND p.id IN (`+ph+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p model.Produk
		if err := scanProduk(rows, &p); err != nil {
			return nil, err
		}
		hasil[p.ID] = p
	}
	return hasil, rows.Err()
}
