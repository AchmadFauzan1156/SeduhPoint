package service

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
	if err := validasiCabang(&c); err != nil {
		return 0, err
	}
	return s.Repo.Tambah(ctx, c)
}

func (s *CabangService) Ubah(ctx context.Context, c model.Cabang) error {
	if err := validasiCabang(&c); err != nil {
		return err
	}
	return s.Repo.Ubah(ctx, c)
}

func (s *CabangService) Hapus(ctx context.Context, id int) error {
	return s.Repo.Hapus(ctx, id)
}

func validasiCabang(c *model.Cabang) error {
	c.Nama, c.Alamat = strings.TrimSpace(c.Nama), strings.TrimSpace(c.Alamat)
	if c.Nama == "" {
		return salah("nama", "Nama cabang wajib diisi")
	}
	if len([]rune(c.Nama)) > 100 {
		return salah("nama", "Nama cabang maksimal 100 karakter")
	}
	return nil
}
