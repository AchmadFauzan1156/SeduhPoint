package service

// hitung.go: mesin perhitungan transaksi.
// Murni tanpa database, sehingga bisa diuji dengan go test (lihat hitung_test.go).

import (
	"errors"
	"sort"
	"time"
)

const (
	NilaiPoin     int64 = 1000  // 1 poin = Rp1.000
	KelipatanPoin int64 = 10000 // 1 poin tiap Rp10.000
)

type Item struct {
	ProdukID   int
	KategoriID int
	Nama       string
	Harga      int64
	Qty        int
}

type PromoAktif struct {
	ID             int
	Nama           string
	Jenis          string // DISKON_PERSEN | MIN_BELANJA | B2G1 | VOUCHER
	Persen         int64
	Nominal        int64
	MinBelanja     int64
	ProdukTarget   map[int]bool
	KategoriTarget map[int]bool
}

type Input struct {
	Items       []Item
	PromoProduk []PromoAktif // DISKON_PERSEN & B2G1 yang lolos periode/cabang/jam
	PromoMin    *PromoAktif  // potongan minimal belanja (opsional)
	Voucher     *PromoAktif  // voucher (opsional)
	IsMember    bool
	IsGold      bool
	SaldoPoin   int // poin aktif member
	PoinDiminta int // poin yang ingin ditukar
}

type Potongan struct {
	PromoID int
	Nama    string
	Nilai   int64
}

type Hasil struct {
	Subtotal          int64
	RincianPromo      []Potongan
	TotalPromo        int64
	TotalSetelahPromo int64
	MaksPoin          int // batas tukar poin (ditampilkan di kasir)
	PoinDitukar       int
	PotonganPoin      int64
	TotalBayar        int64
	PoinDiperoleh     int
	MinTidakTercapai  bool // promo minimal belanja dipilih tapi syaratnya belum terpenuhi
}

// ErrMinDanVoucher: potongan minimal belanja tidak bisa digabung dengan voucher.
var ErrMinDanVoucher = &ValidasiError{Field: "kode_voucher", Pesan: "Potongan minimal belanja tidak dapat digabung dengan voucher"}

// Hitung menjalankan langkah 1–7 perhitungan transaksi.
func Hitung(in Input) (Hasil, error) {
	var h Hasil
	if in.PromoMin != nil && in.Voucher != nil {
		return h, ErrMinDanVoucher
	}
	// 1. Subtotal
	for _, it := range in.Items {
		if it.Qty <= 0 || it.Harga < 0 {
			return h, errors.New("item tidak valid")
		}
		h.Subtotal += it.Harga * int64(it.Qty)
	}

	// 2. Promo produk: diskon persen dulu, lalu beli 2 gratis 1 pada item yang tidak didiskon.
	// Diskon terbesar didahulukan, sehingga jika satu menu kena beberapa diskon yang dipakai adalah yang terbesar.
	diskon := make([]PromoAktif, 0, len(in.PromoProduk))
	for _, p := range in.PromoProduk {
		if p.Jenis == JenisDiskon {
			diskon = append(diskon, p)
		}
	}
	sort.SliceStable(diskon, func(i, j int) bool { return diskon[i].Persen > diskon[j].Persen })
	didiskon := map[int]bool{}
	for _, p := range diskon {
		if v := DiskonPersen(in.Items, p, didiskon); v > 0 {
			h.RincianPromo = append(h.RincianPromo, Potongan{p.ID, p.Nama, v})
		}
	}
	for _, p := range in.PromoProduk {
		if p.Jenis == JenisB2G1 {
			var sisa []Item
			for _, it := range in.Items {
				if !didiskon[it.ProdukID] {
					sisa = append(sisa, it)
				}
			}
			if v := BeliDuaGratisSatu(sisa, p); v > 0 {
				h.RincianPromo = append(h.RincianPromo, Potongan{p.ID, p.Nama, v})
			}
		}
	}
	var promoProduk int64
	for _, r := range h.RincianPromo {
		promoProduk += r.Nilai
	}
	total := h.Subtotal - promoProduk

	// 3. Satu promo tingkat transaksi: minimal belanja ATAU voucher.
	//    Syarat minimal belanja dihitung dari subtotal setelah promo produk.
	if p := in.PromoMin; p != nil {
		if total >= p.MinBelanja {
			v := min(p.Nominal, total)
			h.RincianPromo = append(h.RincianPromo, Potongan{p.ID, p.Nama, v})
			total -= v
		} else {
			h.MinTidakTercapai = true
		}
	}
	if p := in.Voucher; p != nil {
		v := min(p.Nominal, total)
		h.RincianPromo = append(h.RincianPromo, Potongan{p.ID, p.Nama, v})
		total -= v
	}

	// 4. Total setelah promo
	for _, r := range h.RincianPromo {
		h.TotalPromo += r.Nilai
	}
	h.TotalSetelahPromo = h.Subtotal - h.TotalPromo

	// 5. Tukar poin (member saja); permintaan di atas batas dibatasi otomatis.
	if in.IsMember {
		h.MaksPoin = MaksPoinTukar(in.SaldoPoin, h.TotalSetelahPromo)
		h.PoinDitukar = max(0, min(in.PoinDiminta, h.MaksPoin))
		h.PotonganPoin = int64(h.PoinDitukar) * NilaiPoin
	}

	// 6. Total bayar
	h.TotalBayar = h.TotalSetelahPromo - h.PotonganPoin

	// 7. Poin diperoleh dari total setelah promo, sebelum potongan poin
	if in.IsMember {
		h.PoinDiperoleh = PoinDiperoleh(h.TotalSetelahPromo, in.IsGold)
	}
	return h, nil
}

// DiskonPersen: diskon % untuk produk target, dibulatkan ke bawah per baris.
// Produk yang sudah kena diskon dicatat di `didiskon` (boleh nil) agar tidak didiskon dua kali.
func DiskonPersen(items []Item, p PromoAktif, didiskon map[int]bool) int64 {
	var total int64
	for _, it := range items {
		if p.ProdukTarget[it.ProdukID] && !didiskon[it.ProdukID] {
			total += it.Harga * int64(it.Qty) * p.Persen / 100
			if didiskon != nil {
				didiskon[it.ProdukID] = true
			}
		}
	}
	return total
}

// BeliDuaGratisSatu: tiap 3 item dari kategori target, 1 item termurah gratis.
func BeliDuaGratisSatu(items []Item, p PromoAktif) int64 {
	var harga []int64
	for _, it := range items {
		if p.KategoriTarget[it.KategoriID] {
			for i := 0; i < it.Qty; i++ {
				harga = append(harga, it.Harga)
			}
		}
	}
	sort.Slice(harga, func(i, j int) bool { return harga[i] < harga[j] })
	var gratis int64
	for i := 0; i < len(harga)/3; i++ {
		gratis += harga[i]
	}
	return gratis
}

// MaksPoinTukar = min(saldo, 50% × totalSetelahPromo / 1.000).
func MaksPoinTukar(saldo int, totalSetelahPromo int64) int {
	return max(0, min(saldo, int(totalSetelahPromo/2/NilaiPoin)))
}

// PoinDiperoleh = floor(total/10.000) × pengali; Gold 1,5× dibulatkan ke bawah.
// Integer saja: poin*3/2.
func PoinDiperoleh(totalSetelahPromo int64, gold bool) int {
	poin := int(totalSetelahPromo / KelipatanPoin)
	if gold {
		poin = poin * 3 / 2
	}
	return poin
}

// TanggalKedaluwarsa = tanggal perolehan + 6 bulan.
func TanggalKedaluwarsa(t time.Time) time.Time {
	return t.AddDate(0, 6, 0)
}

// PromoBerlaku mengecek periode (tanggal), jam (WIB, opsional) dan cabang sebuah promo.
// jamMulai/jamSelesai "" = sepanjang hari. `sekarang` harus sudah dalam WIB.
func PromoBerlaku(sekarang time.Time, mulai, selesai time.Time, jamMulai, jamSelesai string,
	cabangPromo []int, cabangID int) bool {
	hari := sekarang.Format("2006-01-02")
	if hari < mulai.Format("2006-01-02") || hari > selesai.Format("2006-01-02") {
		return false
	}
	if len(jamMulai) >= 5 && len(jamSelesai) >= 5 {
		j := sekarang.Format("15:04")
		if j < jamMulai[:5] || j >= jamSelesai[:5] {
			return false
		}
	}
	for _, c := range cabangPromo {
		if c == cabangID {
			return true
		}
	}
	return false
}
