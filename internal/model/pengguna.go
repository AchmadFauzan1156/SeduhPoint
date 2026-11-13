package model

import (
	"errors"
	"time"

	"seduhpoint/internal/config"
)

// ErrForbidden: sudah login, tetapi tidak berhak melakukan aksi ini (materi #12).
var ErrForbidden = errors.New("akses ditolak")

// Konstanta peran (materi #12): ditulis sekali di sini, bukan string "ADMIN"/"KASIR" yang diketik di banyak tempat.
const (
	PeranAdmin = "ADMIN"
	PeranKasir = "KASIR"
)

func (p Pengguna) IsAdmin() bool { return p.Peran == PeranAdmin }
func (p Pengguna) IsKasir() bool { return p.Peran == PeranKasir }

// Aturan akses ditulis sekali di sini lalu dipakai service (pengecekan) dan template (tampilan).
// Materi #12: "Satu aturan, satu tempat".

// BolehLihat: admin melihat semua transaksi; kasir hanya transaksi cabangnya sendiri.
func (p Pengguna) BolehLihat(t Transaksi) bool {
	return p.IsAdmin() || (p.IsKasir() && p.CabangID == t.CabangID)
}

// BolehBatal: transaksi yang masih SELESAI boleh dibatalkan admin kapan saja;
// kasir hanya transaksi cabangnya sendiri pada hari yang sama (WIB).
func (p Pengguna) BolehBatal(t Transaksi, sekarang time.Time) bool {
	if t.Status != "SELESAI" || !p.BolehLihat(t) {
		return false
	}
	if p.IsAdmin() {
		return true
	}
	return t.Waktu.In(config.WIB).Format("2006-01-02") == sekarang.In(config.WIB).Format("2006-01-02")
}
