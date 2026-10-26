package store

import "time"

// F-28 – F-30 — TODO M4 (hanya transaksi status SELESAI, F-24)

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

func (s *Store) LaporanBiayaPromo(dari, sampai time.Time, cabangID int) ([]BarisBiayaPromo, error) {
	return nil, nil
}
func (s *Store) LaporanPoin() (RingkasanPoin, error)            { return RingkasanPoin{}, nil }
func (s *Store) LaporanMember(tahun int) ([]BarisMember, error) { return nil, nil }
