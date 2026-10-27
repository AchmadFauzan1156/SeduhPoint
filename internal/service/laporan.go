package service

// laporan.go: laporan pemilik
// dan ringkasan dashboard admin. Semua hanya menghitung transaksi SELESAI.

import (
	"context"
	"strconv"
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

// HariKedaluwarsaDekat: batas "akan kedaluwarsa dalam waktu dekat".
const HariKedaluwarsaDekat = 30

// persen menghitung lebar batang grafik (0–100) terhadap nilai terbesar.
func persen(nilai, maks int64) int {
	if maks <= 0 || nilai <= 0 {
		return 0
	}
	return max(2, int(nilai*100/maks))
}

// ---------- Biaya promo ----------

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

// ---------- Poin aktif ----------

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

// ---------- Perkembangan member ----------

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

// PerkembanganMember menghitung perkembangan member dari riwayat tingkatnya. Murni tanpa database (lihat laporan_test.go).
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

// Perubahan: nilai periode ini dibanding periode sebelumnya yang sama panjang.
type Perubahan struct {
	Ada    bool    // false jika periode sebelumnya 0 (tidak ada pembanding)
	Persen float64 // dibulatkan 0,1
	Naik   bool
	Datar  bool
}

func bandingkan(kini, lalu int64) Perubahan {
	if lalu == 0 {
		return Perubahan{}
	}
	p := float64(kini-lalu) * 100 / float64(lalu)
	p = float64(int64(p*10+copysign(0.5, p))) / 10
	return Perubahan{Ada: true, Persen: abs(p), Naik: p > 0, Datar: p == 0}
}

func copysign(v, tanda float64) float64 {
	if tanda < 0 {
		return -v
	}
	return v
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// TitikGrafik: satu titik grafik omzet harian, koordinat sudah dihitung untuk SVG.
type TitikGrafik struct {
	Tanggal time.Time
	Omzet   int64
	X, Y    float64
	HariIni bool
}

type GarisY struct {
	Y     float64
	Label string
}

type Grafik struct {
	Lebar, Tinggi float64
	Titik         []TitikGrafik
	Garis         string // atribut points polyline
	Area          string // atribut points polygon (area di bawah garis)
	GarisY        []GarisY
	Total         int64
	Kosong        bool
}

// IrisanDonat: satu cabang pada donat omzet; Panjang/Offset untuk stroke-dasharray (keliling 100).
type IrisanDonat struct {
	Nama    string
	Omzet   int64
	Persen  float64
	Panjang float64
	Offset  float64
	Warna   string
}

type Ringkasan struct {
	Hari         int // 1, 7 atau 30
	Dari, Sampai time.Time
	Omzet        int64
	Transaksi    int
	Batal        int
	BiayaPromo   int64
	Cabang       []CabangBiaya
	PromoTeratas []PromoBiaya
	// Pembanding periode sebelumnya (panjang sama)
	DariLalu, SampaiLalu time.Time
	POmzet, PTransaksi   Perubahan
	PBiayaPromo          Perubahan
	Grafik               Grafik
	Donat                []IrisanDonat
}

var warnaCabang = []string{"#723E29", "#C08A5B", "#E3CBB4", "#9C6B4E", "#5A2E1C"}

// skalaAtas: batas atas sumbu Y yang "bulat" (1, 2, 2,5 atau 5 × 10^n) di atas nilai maksimum.
func skalaAtas(maks int64) int64 {
	if maks <= 0 {
		return 100000
	}
	langkah := int64(1)
	for langkah*10 <= maks {
		langkah *= 10
	}
	for _, k := range []int64{1, 2, 5, 10} {
		if langkah*k >= maks {
			// beri ruang 4 garis
			return langkah * k
		}
	}
	return langkah * 10
}

func labelRupiahPendek(v int64) string {
	switch {
	case v >= 1000000:
		jt := float64(v) / 1000000
		if jt == float64(int64(jt)) {
			return strconvFormat(int64(jt)) + " jt"
		}
		return trimKoma(jt) + " jt"
	case v >= 1000:
		return strconvFormat(v/1000) + " rb"
	}
	return strconvFormat(v)
}

// BuatGrafik menyusun grafik omzet harian `hari` hari sampai tanggal `akhir` (inklusif).
func BuatGrafik(omzet map[string]int64, akhir time.Time, hari int, lebar, tinggi float64) Grafik {
	g := Grafik{Lebar: lebar, Tinggi: tinggi}
	var maks int64
	for i := hari - 1; i >= 0; i-- {
		t := akhir.AddDate(0, 0, -i)
		v := omzet[t.Format("2006-01-02")]
		g.Titik = append(g.Titik, TitikGrafik{Tanggal: t, Omzet: v, HariIni: i == 0})
		g.Total += v
		maks = max(maks, v)
	}
	g.Kosong = g.Total == 0
	atas := skalaAtas(maks)
	const kiri, bawah, atasPad = 0.0, 0.0, 8.0
	tinggiPlot := tinggi - bawah - atasPad
	for i := range g.Titik {
		x := kiri
		if hari > 1 {
			x = kiri + (lebar-kiri)*float64(i)/float64(hari-1)
		}
		y := atasPad + tinggiPlot*(1-float64(g.Titik[i].Omzet)/float64(atas))
		g.Titik[i].X, g.Titik[i].Y = bulat1(x), bulat1(y)
	}
	for k := 0; k <= 4; k++ {
		v := atas * int64(k) / 4
		g.GarisY = append(g.GarisY, GarisY{Y: bulat1(atasPad + tinggiPlot*(1-float64(k)/4)), Label: labelRupiahPendek(v)})
	}
	var b, a []byte
	for _, t := range g.Titik {
		b = append(b, []byte(fmtTitik(t.X, t.Y))...)
	}
	g.Garis = string(b)
	if len(g.Titik) > 0 {
		a = append(a, []byte(fmtTitik(g.Titik[0].X, tinggi-bawah))...)
		a = append(a, b...)
		a = append(a, []byte(fmtTitik(g.Titik[len(g.Titik)-1].X, tinggi-bawah))...)
	}
	g.Area = string(a)
	return g
}

// BuatDonat membagi keliling 100 menurut omzet tiap cabang.
func BuatDonat(cabang []CabangBiaya) []IrisanDonat {
	var total int64
	for _, c := range cabang {
		total += c.Omzet
	}
	var hasil []IrisanDonat
	offset := 0.0
	for i, c := range cabang {
		d := IrisanDonat{Nama: c.Nama, Omzet: c.Omzet, Warna: warnaCabang[i%len(warnaCabang)]}
		if total > 0 {
			d.Persen = bulat1(float64(c.Omzet) * 100 / float64(total))
			d.Panjang = float64(c.Omzet) * 100 / float64(total)
		}
		d.Offset = -offset
		offset += d.Panjang
		hasil = append(hasil, d)
	}
	return hasil
}

// Ringkasan: angka untuk dashboard admin pada `hari` hari terakhir (termasuk hari ini),
// dibanding periode sebelumnya yang sama panjang.
func (s *LaporanService) Ringkasan(ctx context.Context, hari int) (Ringkasan, error) {
	if hari != 7 && hari != 30 {
		hari = 1
	}
	now := s.sekarang()
	sampai := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dari := sampai.AddDate(0, 0, 1-hari)
	rk := Ringkasan{Hari: hari, Dari: dari, Sampai: sampai,
		DariLalu: dari.AddDate(0, 0, -hari), SampaiLalu: dari.AddDate(0, 0, -1)}
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
	rk.Donat = BuatDonat(lp.Cabang)

	lalu, err := s.BiayaPromo(ctx, rk.DariLalu, rk.SampaiLalu, 0)
	if err != nil {
		return rk, err
	}
	rk.POmzet = bandingkan(rk.Omzet, lalu.TotalOmzet)
	rk.PTransaksi = bandingkan(int64(rk.Transaksi), int64(lalu.TotalTrx))
	rk.PBiayaPromo = bandingkan(rk.BiayaPromo, lalu.TotalPotongan)

	// Grafik: 7 hari terakhir untuk "hari ini" & "7 hari", 30 hari untuk "30 hari".
	nGrafik := max(hari, 7)
	awal := sampai.AddDate(0, 0, 1-nGrafik)
	harian, err := s.Repo.OmzetHarian(ctx, awal.Format("2006-01-02"), sampai.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return rk, err
	}
	rk.Grafik = BuatGrafik(harian, sampai, nGrafik, 640, 200)
	return rk, nil
}

func bulat1(f float64) float64 { return float64(int64(f*10+0.5)) / 10 }

func strconvFormat(v int64) string { return strconv.FormatInt(v, 10) }

func trimKoma(f float64) string {
	s := strconv.FormatFloat(f, 'f', 1, 64)
	return s[:len(s)-2] + "," + s[len(s)-1:]
}

func fmtTitik(x, y float64) string {
	return strconv.FormatFloat(x, 'f', -1, 64) + "," + strconv.FormatFloat(y, 'f', -1, 64) + " "
}
