package repository

// F-02 — TODO M2: isi mengikuti pola cabang.go

import (
	"context"
	"database/sql"

	"seduhpoint/internal/model"
)

type KategoriRepo struct {
	DB *sql.DB
}

func (r *KategoriRepo) Semua(ctx context.Context) ([]model.Kategori, error)         { return nil, nil }
func (r *KategoriRepo) Tambah(ctx context.Context, k model.Kategori) (int64, error) { return 0, nil }
func (r *KategoriRepo) Ubah(ctx context.Context, k model.Kategori) error            { return nil }
func (r *KategoriRepo) Hapus(ctx context.Context, id int) error                     { return nil }

type ProdukRepo struct {
	DB *sql.DB
}

// Semua: JOIN kategori agar KategoriNama terisi.
func (r *ProdukRepo) Semua(ctx context.Context, hanyaAktif bool) ([]model.Produk, error) {
	return nil, nil
}
func (r *ProdukRepo) CariByID(ctx context.Context, id int) (model.Produk, error) {
	return model.Produk{}, model.ErrNotFound
}
func (r *ProdukRepo) Tambah(ctx context.Context, p model.Produk) (int64, error) { return 0, nil }
func (r *ProdukRepo) Ubah(ctx context.Context, p model.Produk) error            { return nil }
func (r *ProdukRepo) Hapus(ctx context.Context, id int) error                   { return nil }
