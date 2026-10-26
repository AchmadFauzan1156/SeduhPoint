package service

// F-28 – F-30 — TODO M4

import (
	"context"
	"time"

	"seduhpoint/internal/repository"
)

type LaporanService struct {
	Repo *repository.LaporanRepo
}

func (s *LaporanService) BiayaPromo(ctx context.Context, dari, sampai time.Time, cabangID int) ([]repository.BarisBiayaPromo, error) {
	if sampai.Before(dari) {
		return nil, salah("sampai", "Tanggal 'sampai' tidak boleh sebelum 'dari'")
	}
	return s.Repo.BiayaPromo(ctx, dari, sampai, cabangID)
}
func (s *LaporanService) Poin(ctx context.Context) (repository.RingkasanPoin, error) {
	return s.Repo.Poin(ctx)
}
func (s *LaporanService) Member(ctx context.Context, tahun int) ([]repository.BarisMember, error) {
	return s.Repo.Member(ctx, tahun)
}
