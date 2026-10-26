// hitung.go: "mesin perhitungan" transaksi (Laporan 1.8).
// Murni tanpa database agar mudah diuji dengan go test.
package service

import "time"

const (
	NilaiPoin     int64 = 1000  // 1 poin = Rp1.000 (AB03)
	KelipatanPoin int64 = 10000 // 1 poin tiap Rp10.000 (AB01)
)

type Item struct {
	ProdukID   int
	KategoriID int
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
	PromoProduk []PromoAktif // DISKON_PERSEN & B2G1 yang lolos validasi periode/cabang/jam
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
	PoinDitukar       int
	PotonganPoin      int64
	TotalBayar        int64
	PoinDiperoleh     int
}

// Hitung menjalankan langkah 1–7 pada Laporan 1.8.
func Hitung(in Input) (Hasil, error) {
	var h Hasil
	// TODO M3 langkah 1: Subtotal = Σ harga × qty
	// TODO M3 langkah 2: promo produk -> DiskonPersen(), BeliDuaGratisSatu()
	// TODO M3 langkah 3: PromoMin ATAU Voucher, tolak jika keduanya (AB09 / F-15)
	//                    syarat min. belanja dihitung dari subtotal setelah promo produk (asumsi 7)
	// TODO M3 langkah 4: TotalSetelahPromo = Subtotal - TotalPromo
	// TODO M3 langkah 5: MaksPoinTukar() lalu potongan poin (hanya member)
	// TODO M3 langkah 6: TotalBayar = TotalSetelahPromo - PotonganPoin
	// TODO M3 langkah 7: PoinDiperoleh() (hanya member)
	return h, nil
}

// DiskonPersen: diskon % untuk produk target (AB07a).
func DiskonPersen(items []Item, p PromoAktif) int64 {
	// TODO M3
	return 0
}

// BeliDuaGratisSatu: tiap 3 item kategori target, 1 item termurah gratis (asumsi 6).
func BeliDuaGratisSatu(items []Item, p PromoAktif) int64 {
	// TODO M3
	return 0
}

// MaksPoinTukar = min(saldo, 50% × totalSetelahPromo / 1.000) (AB03).
func MaksPoinTukar(saldo int, totalSetelahPromo int64) int {
	// TODO M3
	return 0
}

// PoinDiperoleh = floor(total/10.000) × pengali; Gold 1,5× dibulatkan ke bawah (AB01, AB06, asumsi 2).
// Gunakan integer: poin*3/2, bukan float (NF-04).
func PoinDiperoleh(totalSetelahPromo int64, gold bool) int {
	// TODO M3
	return 0
}

// TanggalKedaluwarsa = tanggal perolehan + 6 bulan (AB02).
func TanggalKedaluwarsa(t time.Time) time.Time {
	return t.AddDate(0, 6, 0)
}

// PromoBerlaku mengecek periode, cabang, dan jam (WIB) sebuah promo (AB08, F-13, F-14).
func PromoBerlaku(sekarang time.Time, mulai, selesai time.Time, jamMulai, jamSelesai string,
	cabangPromo []int, cabangID int) bool {
	// TODO M2/M3
	return false
}
