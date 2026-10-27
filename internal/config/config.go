package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	_ "time/tzdata" // agar zona waktu Asia/Jakarta tersedia di Windows

	"github.com/joho/godotenv"
)

// WIB dipakai untuk pengecekan jam promo (NF-06).
var WIB, _ = time.LoadLocation("Asia/Jakarta")

type Config struct {
	Addr  string
	DBDSN string
}

func Load() (Config, error) {
	keAkarProyek()
	_ = godotenv.Load() // baca .env jika ada
	cfg := Config{
		Addr:  getEnv("ADDR", "127.0.0.1:8080"),
		DBDSN: os.Getenv("DB_DSN"),
	}
	if cfg.DBDSN == "" {
		wd, _ := os.Getwd()
		return cfg, errors.New("DB_DSN belum di-set: file .env tidak ditemukan di " + wd +
			" (salin .env.example menjadi .env di folder SeduhPoint)")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// keAkarProyek: .env, templates/ dan static/ dibaca relatif terhadap folder kerja.
// Jika aplikasi dijalankan dari subfolder (mis. cmd/web), pindah ke folder yang berisi go.mod.
func keAkarProyek() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for i := 0; i < 4; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			_ = os.Chdir(dir)
			return
		}
		induk := filepath.Dir(dir)
		if induk == dir {
			return
		}
		dir = induk
	}
}
