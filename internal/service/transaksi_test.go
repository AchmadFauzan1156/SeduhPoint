package service

import (
	"testing"
	"time"

	"seduhpoint/internal/config"
	"seduhpoint/internal/model"
)

// Contoh kasus: transaksi A memberi poin (lot L), transaksi B menukar poin dari lot L.
// A dibatalkan lebih dulu (lot L menjadi DITARIK), lalu B dibatalkan:
// poin B tidak boleh kembali ke lot L, karena asal poinnya sudah dibatalkan.
func TestStatusLotKembali(t *testing.T) {
	sekarang := time.Date(2026, 11, 10, 12, 0, 0, 0, config.WIB)
	besok, kemarin := sekarang.AddDate(0, 0, 1), sekarang.AddDate(0, 0, -1)
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
		{"lot ditarik (transaksi asal batal)", model.PoinLot{Status: "DITARIK", TglKedaluwarsa: besok}, "", false},
	}
	for _, c := range cases {
		status, kembali := StatusLotKembali(c.lot, sekarang)
		if status != c.status || kembali != c.kembali {
			t.Errorf("%s: dapat (%q, %v), harusnya (%q, %v)", c.nama, status, kembali, c.status, c.kembali)
		}
	}
}
