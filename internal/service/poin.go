package service

// Kedaluwarsa poin & tingkat member.

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"seduhpoint/internal/config"
	"seduhpoint/internal/repository"
)

const BatasGold int64 = 1_500_000 // belanja 3 bulan minimal untuk Gold

type PoinService struct {
	DB         *sql.DB
	PoinRepo   *repository.PoinRepo
	MemberRepo *repository.MemberRepo
}

// Evaluasi: total belanja SELESAI 3 bulan terakhir >= Rp1.500.000 -> GOLD, selain itu SILVER.
// Riwayat hanya dicatat jika tingkat berubah.
func (s *PoinService) Evaluasi(ctx context.Context, q repository.Querier, memberID int) error {
	sekarang := time.Now().In(config.WIB)
	total, err := s.MemberRepo.TotalBelanjaSejak(ctx, q, memberID, sekarang.AddDate(0, -3, 0))
	if err != nil {
		return err
	}
	baru := "SILVER"
	if total >= BatasGold {
		baru = "GOLD"
	}
	lama, err := s.MemberRepo.Tingkat(ctx, q, memberID)
	if err != nil || lama == baru {
		return err
	}
	return s.MemberRepo.UbahTingkat(ctx, q, memberID, baru)
}

// EvaluasiSemua dipanggil job harian (member bisa turun ke Silver karena belanja lama keluar dari jendela 3 bulan).
func (s *PoinService) EvaluasiSemua(ctx context.Context) error {
	ids, err := s.MemberRepo.SemuaID(ctx)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error { return s.Evaluasi(ctx, tx, id) }); err != nil {
			return err
		}
	}
	return nil
}

// JalankanJobHarian: kedaluwarsa poin + evaluasi tingkat, saat aplikasi menyala lalu setiap 24 jam.
func (s *PoinService) JalankanJobHarian() {
	jalan := func() {
		ctx := context.Background()
		if n, err := s.PoinRepo.KedaluwarsakanSemua(ctx, time.Now().In(config.WIB)); err != nil {
			log.Println("job kedaluwarsa:", err)
		} else if n > 0 {
			log.Printf("job kedaluwarsa: %d lot poin kedaluwarsa", n)
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

func itoa(n int) string { return strconv.Itoa(n) }

// rupiah: 25000 -> "Rp25.000" (untuk pesan dari service).
func rupiah(n int64) string {
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return "Rp" + b.String()
}
