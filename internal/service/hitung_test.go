package service

import (
	"testing"
	"time"
)

// Skenario uji dari Laporan 1.10 & 1.11. Hapus t.Skip setelah fungsi diisi.
func TestPoinDiperoleh(t *testing.T) {
	cases := []struct {
		total int64
		gold  bool
		want  int
	}{
		{9999, false, 0},
		{10000, false, 1},
		{95000, false, 9},
		{95000, true, 13}, // floor(9 × 1,5)
	}
	for _, c := range cases {
		if got := PoinDiperoleh(c.total, c.gold); got != c.want {
			t.Errorf("PoinDiperoleh(%d,%v)=%d, want %d", c.total, c.gold, got, c.want)
		}
	}
}

func TestMaksPoinTukar(t *testing.T) {
	if got := MaksPoinTukar(100, 50000); got != 25 {
		t.Errorf("got %d, want 25", got)
	}
	if got := MaksPoinTukar(10, 50000); got != 10 {
		t.Errorf("got %d, want 10", got)
	}
}

func TestBeliDuaGratisSatu(t *testing.T) {
	p := PromoAktif{Jenis: "B2G1", KategoriTarget: map[int]bool{1: true}}
	items := []Item{{ProdukID: 1, KategoriID: 1, Harga: 18000, Qty: 2}, {ProdukID: 2, KategoriID: 1, Harga: 22000, Qty: 1}}
	if got := BeliDuaGratisSatu(items, p); got != 18000 {
		t.Errorf("got %d, want 18000", got)
	}
}

func TestMinBelanjaDanVoucherDitolak(t *testing.T) {
	_, err := Hitung(Input{PromoMin: &PromoAktif{}, Voucher: &PromoAktif{}})
	if err == nil {
		t.Error("harus error: AB09")
	}
}

// Skenario perhitungan promo & poin (data uji bagian A di laporan).
func TestHitungSkenario(t *testing.T) {

	americano := Item{ProdukID: 1, KategoriID: 1, Harga: 18000}
	kopiSusu := Item{ProdukID: 2, KategoriID: 1, Harga: 22000}
	cappuccino := Item{ProdukID: 3, KategoriID: 1, Harga: 25000}
	matcha := Item{ProdukID: 4, KategoriID: 2, Harga: 27000}
	coklat := Item{ProdukID: 5, KategoriID: 2, Harga: 23000}
	croissant := Item{ProdukID: 6, KategoriID: 3, Harga: 20000}
	roti := Item{ProdukID: 7, KategoriID: 3, Harga: 15000}
	qty := func(it Item, n int) Item { it.Qty = n; return it }

	p1 := PromoAktif{ID: 1, Nama: "P1", Jenis: "DISKON_PERSEN", Persen: 20, ProdukTarget: map[int]bool{3: true}}
	p2 := &PromoAktif{ID: 2, Nama: "P2", Jenis: "MIN_BELANJA", MinBelanja: 100000, Nominal: 15000}
	p3 := PromoAktif{ID: 3, Nama: "P3", Jenis: "B2G1", KategoriTarget: map[int]bool{1: true}}
	p4 := &PromoAktif{ID: 4, Nama: "P4", Jenis: "VOUCHER", Nominal: 10000}

	cases := []struct {
		kode      string
		in        Input
		subtotal  int64
		promo     int64
		poinTukar int
		bayar     int64
		poinDapat int
	}{
		{"T02", Input{Items: []Item{qty(kopiSusu, 2), qty(roti, 1)}, IsMember: true}, 59000, 0, 0, 59000, 5},
		{"T03", Input{Items: []Item{qty(kopiSusu, 2), qty(roti, 1)}, IsMember: true, IsGold: true}, 59000, 0, 0, 59000, 7},
		{"T04", Input{Items: []Item{qty(cappuccino, 2), qty(matcha, 1)}, PromoProduk: []PromoAktif{p1}, IsMember: true}, 77000, 10000, 0, 67000, 6},
		{"T05", Input{Items: []Item{qty(americano, 2), qty(kopiSusu, 1)}, PromoProduk: []PromoAktif{p3}, IsMember: true}, 58000, 18000, 0, 40000, 4},
		{"T06", Input{Items: []Item{qty(americano, 1), qty(kopiSusu, 2), qty(cappuccino, 2)}, PromoProduk: []PromoAktif{p3}, IsMember: true}, 112000, 18000, 0, 94000, 9},
		{"T07", Input{Items: []Item{qty(matcha, 3), qty(croissant, 1)}, PromoMin: p2, IsMember: true}, 101000, 15000, 0, 86000, 8},
		{"T08", Input{Items: []Item{qty(cappuccino, 4)}, PromoProduk: []PromoAktif{p1}, PromoMin: p2, IsMember: true}, 100000, 20000, 0, 80000, 8},
		{"T09", Input{Items: []Item{qty(coklat, 2), qty(roti, 1)}, Voucher: p4, IsMember: true}, 61000, 10000, 0, 51000, 5},
		{"T10", Input{Items: []Item{qty(cappuccino, 2), qty(croissant, 1)}, PromoProduk: []PromoAktif{p1}, Voucher: p4, IsMember: true}, 70000, 20000, 0, 50000, 5},
		{"T11", Input{Items: []Item{qty(matcha, 3), qty(croissant, 1)}, PromoMin: p2, IsMember: true, IsGold: true}, 101000, 15000, 0, 86000, 12},
		{"T12", Input{Items: []Item{qty(coklat, 2), qty(roti, 1)}, IsMember: true, SaldoPoin: 40, PoinDiminta: 40}, 61000, 0, 30, 31000, 6},
		{"T13", Input{Items: []Item{qty(matcha, 3), qty(croissant, 1)}, IsMember: true, SaldoPoin: 12, PoinDiminta: 50}, 101000, 0, 12, 89000, 10},
		{"T15", Input{Items: []Item{qty(cappuccino, 2), qty(americano, 2)}, PromoProduk: []PromoAktif{p1, p3}, IsMember: true}, 86000, 10000, 0, 76000, 7},
	}
	for _, c := range cases {
		h, err := Hitung(c.in)
		if err != nil {
			t.Errorf("%s: error %v", c.kode, err)
			continue
		}
		if h.Subtotal != c.subtotal || h.TotalPromo != c.promo || h.PoinDitukar != c.poinTukar ||
			h.TotalBayar != c.bayar || h.PoinDiperoleh != c.poinDapat {
			t.Errorf("%s: dapat subtotal=%d promo=%d tukar=%d bayar=%d poin=%d; harap %d %d %d %d %d",
				c.kode, h.Subtotal, h.TotalPromo, h.PoinDitukar, h.TotalBayar, h.PoinDiperoleh,
				c.subtotal, c.promo, c.poinTukar, c.bayar, c.poinDapat)
		}
	}
}

// V04–V06: promo dibatasi jam dan cabang.
func TestPromoBerlaku(t *testing.T) {
	wib := time.FixedZone("WIB", 7*3600)
	mulai, selesai := time.Date(2026, 11, 1, 0, 0, 0, 0, wib), time.Date(2026, 11, 30, 0, 0, 0, 0, wib)
	jam := func(h, m int) time.Time { return time.Date(2026, 11, 2, h, m, 0, 0, wib) }
	cases := []struct {
		nama   string
		t      time.Time
		cabang int
		want   bool
	}{
		{"V04 13.59 sebelum jam", jam(13, 59), 1, false},
		{"V05 15.00 dalam jam", jam(15, 0), 1, true},
		{"17.00 sudah lewat", jam(17, 0), 1, false},
		{"V06 cabang lain", jam(15, 0), 2, false},
		{"sebelum periode", time.Date(2026, 10, 31, 15, 0, 0, 0, wib), 1, false},
		{"hari terakhir periode", time.Date(2026, 11, 30, 15, 0, 0, 0, wib), 1, true},
	}
	for _, c := range cases {
		if got := PromoBerlaku(c.t, mulai, selesai, "14:00:00", "17:00:00", []int{1, 3}, c.cabang); got != c.want {
			t.Errorf("%s: dapat %v, harap %v", c.nama, got, c.want)
		}
	}
	if !PromoBerlaku(jam(23, 0), mulai, selesai, "", "", []int{1}, 1) {
		t.Error("tanpa jam harus berlaku sepanjang hari")
	}
}

// T14: diskon persen dibulatkan ke bawah per baris (asumsi 14).
func TestDiskonDibulatkan(t *testing.T) {
	p := PromoAktif{Persen: 15, ProdukTarget: map[int]bool{3: true}}
	if got := DiskonPersen([]Item{{ProdukID: 3, Harga: 25000, Qty: 1}}, p, nil); got != 3750 {
		t.Errorf("got %d, want 3750", got)
	}
	if got := DiskonPersen([]Item{{ProdukID: 3, Harga: 23333, Qty: 1}}, p, nil); got != 3499 {
		t.Errorf("got %d, want 3499 (floor dari 3499,95)", got)
	}
}

func TestNonMemberTidakDapatPoin(t *testing.T) {
	h, _ := Hitung(Input{Items: []Item{{ProdukID: 1, Harga: 56000, Qty: 1}}, PoinDiminta: 10})
	if h.PoinDitukar != 0 || h.PoinDiperoleh != 0 || h.TotalBayar != 56000 {
		t.Errorf("non-member: %+v", h)
	}
}

// Temuan review #1: dua diskon persen untuk menu yang sama -> dipakai yang terbesar.
func TestDiskonTerbesarDipakai(t *testing.T) {
	kecil := PromoAktif{ID: 1, Nama: "10%", Jenis: JenisDiskon, Persen: 10, ProdukTarget: map[int]bool{3: true}}
	besar := PromoAktif{ID: 2, Nama: "25%", Jenis: JenisDiskon, Persen: 25, ProdukTarget: map[int]bool{3: true}}
	h, _ := Hitung(Input{Items: []Item{{ProdukID: 3, Harga: 20000, Qty: 1}}, PromoProduk: []PromoAktif{kecil, besar}})
	if h.TotalPromo != 5000 || len(h.RincianPromo) != 1 || h.RincianPromo[0].PromoID != 2 {
		t.Errorf("harus hanya diskon 25%% (Rp5.000), dapat %+v", h.RincianPromo)
	}
}
