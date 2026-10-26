package layanan

import (
	"database/sql"
	"log"
	"time"
)

// KedaluwarsakanPoin: lot AKTIF yang lewat tgl_kedaluwarsa -> KEDALUWARSA + mutasi (F-25). TODO M3
func (l *Layanan) KedaluwarsakanPoin() error { return nil }

// EvaluasiTingkat: total belanja SELESAI 3 bulan terakhir >= 1.500.000 -> GOLD, selain itu SILVER.
// Jika berubah, catat ke member_tingkat_riwayat (AB06, F-26, asumsi 4–5). TODO M3
func (l *Layanan) EvaluasiTingkat(tx *sql.Tx, memberID int) error { return nil }

// EvaluasiSemuaMember dipanggil job harian. TODO M3
func (l *Layanan) EvaluasiSemuaMember() error { return nil }

// JalankanJobHarian berjalan di goroutine terpisah (lihat main.go).
func (l *Layanan) JalankanJobHarian() {
	jalankan := func() {
		if err := l.KedaluwarsakanPoin(); err != nil {
			log.Println("job kedaluwarsa:", err)
		}
		if err := l.EvaluasiSemuaMember(); err != nil {
			log.Println("job tingkat:", err)
		}
	}
	jalankan() // sekali saat start
	for range time.Tick(24 * time.Hour) {
		jalankan()
	}
}
