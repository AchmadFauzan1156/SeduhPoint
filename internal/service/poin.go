package service

// F-25 – F-27 — TODO M3

import (
	"context"
	"database/sql"
	"log"
	"time"

	"seduhpoint/internal/repository"
)

const BatasGold int64 = 1_500_000 // AB06

type PoinService struct {
	DB         *sql.DB
	PoinRepo   *repository.PoinRepo
	MemberRepo *repository.MemberRepo
}

// Evaluasi: total belanja SELESAI 3 bulan terakhir >= 1.500.000 -> GOLD, selain itu SILVER.
// Catat ke riwayat hanya jika berubah (AB06, asumsi 4–5). TODO
func (s *PoinService) Evaluasi(ctx context.Context, q repository.Querier, memberID int) error {
	return nil
}

// EvaluasiSemua dipanggil job harian. TODO
func (s *PoinService) EvaluasiSemua(ctx context.Context) error { return nil }

// JalankanJobHarian: kedaluwarsa poin + evaluasi tingkat (dipanggil sebagai goroutine dari main).
func (s *PoinService) JalankanJobHarian() {
	jalan := func() {
		ctx := context.Background()
		if _, err := s.PoinRepo.KedaluwarsakanSemua(ctx); err != nil {
			log.Println("job kedaluwarsa:", err)
		}
		if err := s.EvaluasiSemua(ctx); err != nil {
			log.Println("job tingkat:", err)
		}
	}
	jalan()
	for range time.Tick(24 * time.Hour) {
		jalan()
	}
}
