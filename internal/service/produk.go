package service

// Kategori & Produk

import (
	"context"
	"strings"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type ProdukService struct {
	Repo         *repository.ProdukRepo
	KategoriRepo *repository.KategoriRepo
}

// ---- Kategori ----

func (s *ProdukService) DaftarKategori(ctx context.Context) ([]model.Kategori, error) {
	return s.KategoriRepo.Semua(ctx)
}

func (s *ProdukService) DetailKategori(ctx context.Context, id int) (model.Kategori, error) {
	return s.KategoriRepo.CariByID(ctx, id)
}

func (s *ProdukService) SimpanKategori(ctx context.Context, k model.Kategori) error {
	k.Nama = strings.TrimSpace(k.Nama)
	if k.Nama == "" {
		return salah("nama", "Nama kategori wajib diisi")
	}
	if len([]rune(k.Nama)) > 100 {
		return salah("nama", "Nama kategori maksimal 100 karakter")
	}
	if dipakai, err := s.KategoriRepo.NamaDipakai(ctx, k.Nama, k.ID); err != nil {
		return err
	} else if dipakai {
		return salah("nama", "Kategori \""+k.Nama+"\" sudah ada")
	}
	if k.ID == 0 {
		_, err := s.KategoriRepo.Tambah(ctx, k)
		return err
	}
	return s.KategoriRepo.Ubah(ctx, k)
}

// HapusKategori: hanya boleh jika tidak ada produk di dalamnya.
func (s *ProdukService) HapusKategori(ctx context.Context, id int) error {
	k, err := s.KategoriRepo.CariByID(ctx, id)
	if err != nil {
		return err
	}
	if k.JumlahProduk > 0 {
		return salah("", "Kategori "+k.Nama+" masih berisi produk; pindahkan atau hapus produknya dulu")
	}
	return s.KategoriRepo.Hapus(ctx, id)
}

// ---- Produk ----

func (s *ProdukService) Daftar(ctx context.Context) ([]model.Produk, error) {
	return s.Repo.Semua(ctx, false)
}

func (s *ProdukService) DaftarDijual(ctx context.Context) ([]model.Produk, error) {
	return s.Repo.Semua(ctx, true)
}

func (s *ProdukService) Detail(ctx context.Context, id int) (model.Produk, error) {
	return s.Repo.CariByID(ctx, id)
}

func (s *ProdukService) Simpan(ctx context.Context, p model.Produk) (int64, error) {
	if err := s.validasi(ctx, &p); err != nil {
		return 0, err
	}
	if p.ID == 0 {
		return s.Repo.Tambah(ctx, p)
	}
	return int64(p.ID), s.Repo.Ubah(ctx, p)
}

func (s *ProdukService) Hapus(ctx context.Context, id int) error { return s.Repo.Hapus(ctx, id) }

func (s *ProdukService) validasi(ctx context.Context, p *model.Produk) error {
	p.Nama, p.Deskripsi = strings.TrimSpace(p.Nama), strings.TrimSpace(p.Deskripsi)
	switch {
	case len([]rune(p.Deskripsi)) > 300:
		return salah("deskripsi", "Deskripsi maksimal 300 karakter")
	case p.Nama == "":
		return salah("nama", "Nama produk wajib diisi")
	case len([]rune(p.Nama)) > 100:
		return salah("nama", "Nama produk maksimal 100 karakter")
	case p.Harga <= 0:
		return salah("harga", "Harga harus lebih dari Rp0")
	case p.Harga > 10_000_000:
		return salah("harga", "Harga terlalu besar (maks Rp10.000.000)")
	case p.Harga%100 != 0:
		return salah("harga", "Harga harus kelipatan Rp100")
	}
	if _, err := s.KategoriRepo.CariByID(ctx, p.KategoriID); err != nil {
		return salah("kategori_id", "Pilih kategori")
	}
	return nil
}
