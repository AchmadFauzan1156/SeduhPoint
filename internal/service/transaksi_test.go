package service

import (
	"testing"
	"time"

	"seduhpoint/internal/config"
	"seduhpoint/internal/model"
)

// Pembatalan transaksi tukar poin: poin kembali ke lot asal, kecuali lot itu sudah DITARIK.
func TestStatusLotKembali(t *testing.T) {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, config.WIB)
	besok, kemarin := now.AddDate(0, 0, 1), now.AddDate(0, 0, -1)
	cases := []struct {
		nama    string
		lot     model.PoinLot
		status  string
		kembali bool
	}{
		{"lot aktif", model.PoinLot{Status: "AKTIF", TglKedaluwarsa: besok}, "AKTIF", true},
		{"lot habis terpakai", model.PoinLot{Status: "HABIS", TglKedaluwarsa: besok}, "AKTIF", true},
		{"lot lewat masa berlaku", model.PoinLot{Status: "HABIS", TglKedaluwarsa: kemarin}, "KEDALUWARSA", true},
		{"lot sudah dikedaluwarsakan job", model.PoinLot{Status: "KEDALUWARSA", TglKedaluwarsa: besok}, "KEDALUWARSA", true},
		// Transaksi asal poin sudah dibatalkan: poin tidak boleh hidup lagi.
		{"lot ditarik", model.PoinLot{Status: "DITARIK", TglKedaluwarsa: besok}, "", false},
	}
	for _, c := range cases {
		status, kembali := StatusLotKembali(c.lot, now)
		if status != c.status || kembali != c.kembali {
			t.Errorf("%s: dapat (%q, %v), harusnya (%q, %v)", c.nama, status, kembali, c.status, c.kembali)
		}
	}
}

func TestBolehBatal(t *testing.T) {
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, config.WIB)
	cases := []struct {
		nama  string
		peran string
		waktu time.Time
		boleh bool
	}{
		{"kasir, transaksi hari ini", "KASIR", now.Add(-2 * time.Hour), true},
		{"kasir, transaksi kemarin", "KASIR", now.AddDate(0, 0, -1), false},
		// 23.30 WIB kemarin = 16.30 UTC: tetap dihitung kemarin menurut WIB.
		{"kasir, kemarin malam (UTC)", "KASIR", time.Date(2026, 11, 9, 16, 30, 0, 0, time.UTC), false},
		{"admin, transaksi kemarin", "ADMIN", now.AddDate(0, 0, -1), true},
	}
	for _, c := range cases {
		if got := BolehBatal(c.peran, c.waktu, now); got != c.boleh {
			t.Errorf("%s: dapat %v, harusnya %v", c.nama, got, c.boleh)
		}
	}
}

func TestValidasiPasswordBaru(t *testing.T) {
	cases := []struct {
		nama, lama, baru, ulang, field string // field "" = lolos
	}{
		{"ok", "password", "kopiSeduh8", "kopiSeduh8", ""},
		{"terlalu pendek", "password", "kopi1", "kopi1", "baru"},
		{"sama dengan lama", "kopiSeduh8", "kopiSeduh8", "kopiSeduh8", "baru"},
		{"password bawaan", "lama12345", "password", "password", "baru"},
		{"konfirmasi beda", "password", "kopiSeduh8", "kopiSeduh9", "ulang"},
	}
	for _, c := range cases {
		err := ValidasiPasswordBaru(c.lama, c.baru, c.ulang)
		var field string
		if ve, ok := err.(*ValidasiError); ok {
			field = ve.Field
		} else if err != nil {
			t.Fatalf("%s: error tak terduga %v", c.nama, err)
		}
		if field != c.field {
			t.Errorf("%s: field %q, harusnya %q", c.nama, field, c.field)
		}
	}
}
