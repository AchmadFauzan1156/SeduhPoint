package service

// F-02 — TODO M2: ikuti pola cabang.go

import (
	"context"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type ProdukService struct {
	Repo         *repository.ProdukRepo
	KategoriRepo *repository.KategoriRepo
}

func (s *ProdukService) DaftarKategori(ctx context.Context) ([]model.Kategori, error) {
	return s.KategoriRepo.Semua(ctx)
}
func (s *ProdukService) BuatKategori(ctx context.Context, k model.Kategori) (int64, error) {
	// TODO: nama wajib & unik
	return s.KategoriRepo.Tambah(ctx, k)
}
func (s *ProdukService) Daftar(ctx context.Context) ([]model.Produk, error) {
	return s.Repo.Semua(ctx, false)
}
func (s *ProdukService) Detail(ctx context.Context, id int) (model.Produk, error) {
	return s.Repo.CariByID(ctx, id)
}

// Buat/Ubah — TODO: validasi nama wajib, harga >= 0, kategori ada.
func (s *ProdukService) Buat(ctx context.Context, p model.Produk) (int64, error) {
	return s.Repo.Tambah(ctx, p)
}
func (s *ProdukService) Ubah(ctx context.Context, p model.Produk) error { return s.Repo.Ubah(ctx, p) }
func (s *ProdukService) Hapus(ctx context.Context, id int) error        { return s.Repo.Hapus(ctx, id) }
