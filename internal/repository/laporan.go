package repository

// F-28 – F-30 — TODO M4 (hanya transaksi SELESAI, F-24)

import (
	"context"
	"database/sql"
	"time"
)

type LaporanRepo struct {
	DB *sql.DB
}

type BarisBiayaPromo struct {
	PromoID       int
	Nama          string
	Jenis         string
	JumlahPakai   int
	TotalPotongan int64
}

type RingkasanPoin struct {
	TotalPoinAktif    int
	NilaiRupiah       int64
	AkanKedaluwarsa30 int
}

type BarisMember struct {
	Bulan       string
	MemberBaru  int
	Silver      int
	Gold        int
	NaikGold    int
	TurunSilver int
}

func (r *LaporanRepo) BiayaPromo(ctx context.Context, dari, sampai time.Time, cabangID int) ([]BarisBiayaPromo, error) {
	return nil, nil
}
func (r *LaporanRepo) Poin(ctx context.Context) (RingkasanPoin, error)              { return RingkasanPoin{}, nil }
func (r *LaporanRepo) Member(ctx context.Context, tahun int) ([]BarisMember, error) { return nil, nil }
