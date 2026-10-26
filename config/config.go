package config

import (
	"os"
	"time"

	_ "time/tzdata" // agar zona waktu Asia/Jakarta tersedia di Windows

	"github.com/joho/godotenv"
)

type Config struct {
	Host, Port                             string
	DBUser, DBPass, DBHost, DBPort, DBName string
}

// WIB dipakai untuk pengecekan jam promo (NF-06).
var WIB, _ = time.LoadLocation("Asia/Jakarta")

func Load() Config {
	_ = godotenv.Load() // .env opsional
	return Config{
		Host:   get("APP_HOST", "127.0.0.1"),
		Port:   get("APP_PORT", "8080"),
		DBUser: get("DB_USER", "root"),
		DBPass: get("DB_PASS", ""),
		DBHost: get("DB_HOST", "127.0.0.1"),
		DBPort: get("DB_PORT", "3306"),
		DBName: get("DB_NAME", "seduhpoint"),
	}
}

func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
