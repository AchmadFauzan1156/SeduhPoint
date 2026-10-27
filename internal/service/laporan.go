package service

// laporan.go: laporan pemilik (F-28 biaya promo, F-29 poin aktif, F-30 perkembangan member)
// dan ringkasan dashboard admin. Semua hanya menghitung transaksi SELESAI (F-24).

import (
	"context"
	"time"

	"seduhpoint/internal/config"
	"seduhpoint/internal/repository"
)

type LaporanService struct {
	Repo     *repository.LaporanRepo
	Sekarang func() time.Time // untuk pengujian; nil = waktu sekarang (WIB)
}

func (s *LaporanService) sekarang() time.Time {
	if s.Sekarang != nil {
		return s.Sekarang()
	}
	return time.Now().In(config.WIB)
}

// HariKedaluwarsaDekat: batas "akan kedaluwarsa dalam waktu dekat" (F-29).
const HariKedaluwarsaDekat = 30

// persen menghitung lebar batang grafik (0–100) terhadap nilai terbesar.
func persen(nilai, maks int64) int {
	if maks <= 0 || nilai <= 0 {
		return 0
	}
	return max(2, int(nilai*100/maks))
}

// ---------- F-28 Biaya promo ----------

type PromoBiaya struct {
	repository.BarisBiayaPromo
	NamaJenis string
	Persen    int
}

type CabangBiaya struct {
	repository.BarisCabang
	Porsi  float64 // potongan promo / penjualan sebelum potongan, dalam %
	Persen int
}

type LaporanPromo struct {
	Dari, Sampai   time.Time // sampai = inklusif (untuk tampilan)
	CabangID       int
	Promo          []PromoBiaya
	TotalPotongan  int64
	TotalPemakaian int
	Cabang         []CabangBiaya
	TotalOmzet     int64
	TotalSubtotal  int64
	TotalTrx       int
}

func (s *LaporanService) BiayaPromo(ctx context.Context, dari, sampai time.Time, cabangID int) (LaporanPromo, error) {
	lp := LaporanPromo{Dari: dari, Sampai: sampai, CabangID: cabangID}
	if sampai.Before(dari) {
		return lp, salah("sampai", "Tanggal 'sampai' tidak boleh sebelum 'dari'")
	}
	if sampai.Sub(dari) > 366*24*time.Hour {
		return lp, salah("sampai", "Rentang laporan paling lama 1 tahun")
	}
	d, sEks := dari.Format("2006-01-02"), sampai.AddDate(0, 0, 1).Format("2006-01-02")
	baris, err := s.Repo.BiayaPromo(ctx, d, sEks, cabangID)
	if err != nil {
		return lp, err
	}
	var maks int64
	for _, b := range baris {
		lp.TotalPotongan += b.TotalPotongan
		lp.TotalPemakaian += b.JumlahPakai
		maks = max(maks, b.TotalPotongan)
	}
	for _, b := range baris {
		lp.Promo = append(lp.Promo, PromoBiaya{BarisBiayaPromo: b, NamaJenis: NamaJenis[b.Jenis], Persen: persen(b.TotalPotongan, maks)})
	}
	cab, err := s.Repo.PerCabang(ctx, d, sEks, cabangID)
	if err != nil {
		return lp, err
	}
	var maksOmzet int64
	for _, c := range cab {
		maksOmzet = max(maksOmzet, c.Omzet)
	}
	for _, c := range cab {
		cb := CabangBiaya{BarisCabang: c, Persen: persen(c.Omzet, maksOmzet)}
		if c.Subtotal > 0 {
			cb.Porsi = float64(c.TotalPromo) * 100 / float64(c.Subtotal)
		}
		lp.Cabang = append(lp.Cabang, cb)
		lp.TotalOmzet += c.Omzet
		lp.TotalSubtotal += c.Subtotal
		lp.TotalTrx += c.Transaksi
	}
	return lp, nil
}

// ---------- F-29 Poin aktif ----------

type BulanPoin struct {
	Bulan  time.Time
	Poin   int
	Persen int
}

type LaporanPoin struct {
	Per           time.Time
	PoinAktif     int
	NilaiRupiah   int64
	MemberBerpoin int
	PoinDekat     int // akan kedaluwarsa dalam HariKedaluwarsaDekat hari
	NilaiDekat    int64
	HariDekat     int
	Dekat         []repository.LotKedaluwarsa
	PerBulan      []BulanPoin
}

func (s *LaporanService) Poin(ctx context.Context) (LaporanPoin, error) {
	now := s.sekarang()
	lp := LaporanPoin{Per: now, HariDekat: HariKedaluwarsaDekat}
	var err error
	if lp.PoinAktif, lp.MemberBerpoin, err = s.Repo.PoinAktif(ctx, now); err != nil {
		return lp, err
	}
	lp.NilaiRupiah = int64(lp.PoinAktif) * NilaiPoin
	if lp.Dekat, err = s.Repo.AkanKedaluwarsa(ctx, now, now.AddDate(0, 0, HariKedaluwarsaDekat)); err != nil {
		return lp, err
	}
	for _, l := range lp.Dekat {
		lp.PoinDekat += l.Poin
	}
	lp.NilaiDekat = int64(lp.PoinDekat) * NilaiPoin
	bulan, err := s.Repo.PoinPerBulanKedaluwarsa(ctx, now)
	if err != nil {
		return lp, err
	}
	var maks int64
	for _, b := range bulan {
		maks = max(maks, int64(b.Poin))
	}
	for _, b := range bulan {
		lp.PerBulan = append(lp.PerBulan, BulanPoin{Bulan: b.Bulan, Poin: b.Poin, Persen: persen(int64(b.Poin), maks)})
	}
	return lp, nil
}

// ---------- F-30 Perkembangan member ----------

type BarisMember struct {
	Bulan       time.Time
	MemberBaru  int
	NaikGold    int
	TurunSilver int
	Silver      int // jumlah member Silver di akhir bulan
	Gold        int // jumlah member Gold di akhir bulan
}

type LaporanMember struct {
	Tahun      int
	Baris      []BarisMember
	TotalBaru  int
	TotalNaik  int
	TotalTurun int
}

// PerkembanganMember menghitung F-30 dari riwayat tingkat member. Murni tanpa database (lihat laporan_test.go).
// Bulan yang belum terjadi (setelah `sekarang`) tidak ditampilkan.
func PerkembanganMember(data []repository.MemberTingkat, tahun int, sekarang time.Time) []BarisMember {
	var hasil []BarisMember
	for bln := time.January; bln <= time.December; bln++ {
		awal := time.Date(tahun, bln, 1, 0, 0, 0, 0, sekarang.Location())
		if awal.After(sekarang) {
			break
		}
		akhir := awal.AddDate(0, 1, 0)
		b := BarisMember{Bulan: awal}
		for _, m := range data {
			daftar := time.Date(m.TglDaftar.Year(), m.TglDaftar.Month(), m.TglDaftar.Day(), 0, 0, 0, 0, awal.Location())
			if !daftar.Before(akhir) {
				continue // belum terdaftar di akhir bulan ini
			}
			if !daftar.Before(awal) {
				b.MemberBaru++
			}
			tingkat := "SILVER"
			for i, r := range m.Riwayat {
				sejak := time.Date(r.Sejak.Year(), r.Sejak.Month(), r.Sejak.Day(), r.Sejak.Hour(), r.Sejak.Minute(), r.Sejak.Second(), 0, awal.Location())
				if !sejak.Before(akhir) {
					break
				}
				tingkat = r.Tingkat
				if sejak.Before(awal) {
					continue
				}
				switch {
				case r.Tingkat == "GOLD":
					b.NaikGold++
				case i > 0 && m.Riwayat[i-1].Tingkat == "GOLD": // dari Gold kembali ke Silver
					b.TurunSilver++
				}
			}
			if tingkat == "GOLD" {
				b.Gold++
			} else {
				b.Silver++
			}
		}
		// bulan-bulan awal sebelum ada member sama sekali tidak ditampilkan
		if len(hasil) == 0 && b.Silver+b.Gold == 0 {
			continue
		}
		hasil = append(hasil, b)
	}
	return hasil
}

func (s *LaporanService) Member(ctx context.Context, tahun int) (LaporanMember, error) {
	lm := LaporanMember{Tahun: tahun}
	data, err := s.Repo.RiwayatMember(ctx)
	if err != nil {
		return lm, err
	}
	lm.Baris = PerkembanganMember(data, tahun, s.sekarang())
	for _, b := range lm.Baris {
		lm.TotalBaru += b.MemberBaru
		lm.TotalNaik += b.NaikGold
		lm.TotalTurun += b.TurunSilver
	}
	return lm, nil
}

// ---------- Ringkasan dashboard ----------

type Ringkasan struct {
	Hari         int // 1, 7 atau 30
	Dari, Sampai time.Time
	Omzet        int64
	Transaksi    int
	Batal        int
	BiayaPromo   int64
	Cabang       []CabangBiaya
	PromoTeratas []PromoBiaya
}

// Ringkasan: angka untuk dashboard admin pada `hari` hari terakhir (termasuk hari ini).
func (s *LaporanService) Ringkasan(ctx context.Context, hari int) (Ringkasan, error) {
	if hari != 7 && hari != 30 {
		hari = 1
	}
	now := s.sekarang()
	sampai := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dari := sampai.AddDate(0, 0, 1-hari)
	rk := Ringkasan{Hari: hari, Dari: dari, Sampai: sampai}
	lp, err := s.BiayaPromo(ctx, dari, sampai, 0)
	if err != nil {
		return rk, err
	}
	rk.Omzet, rk.Transaksi, rk.BiayaPromo, rk.Cabang = lp.TotalOmzet, lp.TotalTrx, lp.TotalPotongan, lp.Cabang
	for _, c := range lp.Cabang {
		rk.Batal += c.Batal
	}
	if len(lp.Promo) > 3 {
		lp.Promo = lp.Promo[:3]
	}
	rk.PromoTeratas = lp.Promo
	return rk, nil
}
