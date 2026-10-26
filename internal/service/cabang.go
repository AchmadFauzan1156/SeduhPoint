package service

// CONTOH LENGKAP (F-01) — service lain mengikuti pola ini.

import (
	"context"
	"strings"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type CabangService struct {
	Repo *repository.CabangRepo
}

func (s *CabangService) Daftar(ctx context.Context) ([]model.Cabang, error) {
	return s.Repo.Semua(ctx)
}

func (s *CabangService) Detail(ctx context.Context, id int) (model.Cabang, error) {
	return s.Repo.CariByID(ctx, id)
}

func (s *CabangService) Buat(ctx context.Context, c model.Cabang) (int64, error) {
	if err := validasiCabang(c); err != nil {
		return 0, err
	}
	return s.Repo.Tambah(ctx, c)
}

func (s *CabangService) Ubah(ctx context.Context, c model.Cabang) error {
	if err := validasiCabang(c); err != nil {
		return err
	}
	return s.Repo.Ubah(ctx, c)
}

func (s *CabangService) Hapus(ctx context.Context, id int) error {
	return s.Repo.Hapus(ctx, id)
}

func validasiCabang(c model.Cabang) error {
	if strings.TrimSpace(c.Nama) == "" {
		return &ValidasiError{"Nama cabang wajib diisi"}
	}
	if len([]rune(c.Nama)) > 100 {
		return &ValidasiError{"Nama cabang maksimal 100 karakter"}
	}
	return nil
}
