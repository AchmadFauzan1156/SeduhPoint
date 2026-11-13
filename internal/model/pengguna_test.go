package model

import (
	"testing"
	"time"

	"seduhpoint/internal/config"
)

// Materi #12: uji kasus negatif, bukan hanya "admin bisa" tetapi juga "kasir tidak bisa".
func TestAturanAksesTransaksi(t *testing.T) {
	sekarang := time.Date(2026, 11, 10, 9, 0, 0, 0, config.WIB)
	admin := Pengguna{ID: 1, Peran: PeranAdmin}
	kasir1 := Pengguna{ID: 2, Peran: PeranKasir, CabangID: 1}
	hariIni := Transaksi{CabangID: 1, Status: "SELESAI", Waktu: sekarang.Add(-2 * time.Hour)}
	kemarin := Transaksi{CabangID: 1, Status: "SELESAI", Waktu: sekarang.AddDate(0, 0, -1)}
	// 23.30 WIB kemarin = 16.30 UTC: tetap dihitung kemarin menurut WIB.
	kemarinMalam := Transaksi{CabangID: 1, Status: "SELESAI", Waktu: time.Date(2026, 11, 9, 16, 30, 0, 0, time.UTC)}
	cabangLain := Transaksi{CabangID: 2, Status: "SELESAI", Waktu: sekarang}
	sudahBatal := Transaksi{CabangID: 1, Status: "DIBATALKAN", Waktu: sekarang}

	cases := []struct {
		nama         string
		u            Pengguna
		t            Transaksi
		lihat, batal bool
	}{
		{"admin, transaksi kemarin", admin, kemarin, true, true},
		{"admin, cabang mana pun", admin, cabangLain, true, true},
		{"kasir, transaksi hari ini cabangnya", kasir1, hariIni, true, true},
		{"kasir, transaksi kemarin", kasir1, kemarin, true, false},
		{"kasir, kemarin malam (UTC)", kasir1, kemarinMalam, true, false},
		{"kasir, cabang lain", kasir1, cabangLain, false, false},
		{"transaksi sudah dibatalkan", admin, sudahBatal, true, false},
		{"belum login (peran kosong)", Pengguna{CabangID: 1}, hariIni, false, false},
	}
	for _, c := range cases {
		if got := c.u.BolehLihat(c.t); got != c.lihat {
			t.Errorf("%s: BolehLihat = %v, harusnya %v", c.nama, got, c.lihat)
		}
		if got := c.u.BolehBatal(c.t, sekarang); got != c.batal {
			t.Errorf("%s: BolehBatal = %v, harusnya %v", c.nama, got, c.batal)
		}
	}
}
