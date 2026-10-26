package service

// F-06 – F-11 — TODO M2

import (
	"context"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type PromoService struct {
	Repo *repository.PromoRepo
}

func (s *PromoService) Daftar(ctx context.Context) ([]model.Promo, error) { return s.Repo.Semua(ctx) }
func (s *PromoService) Detail(ctx context.Context, id int) (model.Promo, error) {
	return s.Repo.CariByID(ctx, id)
}

// Buat/Ubah — TODO validasi per jenis:
// DISKON_PERSEN: persen 1–100 + minimal 1 produk; MIN_BELANJA: min belanja & nominal > 0;
// B2G1: minimal 1 kategori; VOUCHER: kode unik, nominal > 0, kuota >= 1;
// semua: tgl_mulai <= tgl_selesai, minimal 1 cabang, jam_mulai < jam_selesai jika diisi.
func (s *PromoService) Buat(ctx context.Context, p model.Promo) (int64, error) {
	return s.Repo.Tambah(ctx, p)
}
func (s *PromoService) Ubah(ctx context.Context, p model.Promo) error { return s.Repo.Ubah(ctx, p) }
func (s *PromoService) SetAktif(ctx context.Context, id int, aktif bool) error {
	return s.Repo.SetAktif(ctx, id, aktif)
}
