// SeduhPoint — Sistem Poin Member & Promo Kedai Kopi Seduh (Kelompok 5)
// Jalankan: go run .
package main

import (
	"log"
	"net/http"

	"seduhpoint/config"
	"seduhpoint/database"
	"seduhpoint/handler"
	"seduhpoint/layanan"
	"seduhpoint/middleware"
	"seduhpoint/store"
	"seduhpoint/view"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg)
	if err != nil {
		log.Fatal("gagal konek MySQL: ", err)
	}
	defer db.Close()

	st := store.New(db)
	h := &handler.Handler{
		Store:   st,
		View:    view.New("templates"),
		Layanan: layanan.New(db, st),
	}

	// Job harian: kedaluwarsa poin & evaluasi tingkat member (F-25, F-26)
	go h.Layanan.JalankanJobHarian()

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	handler.Routes(mux, h)

	addr := cfg.Host + ":" + cfg.Port
	log.Println("SeduhPoint berjalan di http://localhost:" + cfg.Port)
	log.Fatal(http.ListenAndServe(addr, middleware.Logger(mux)))
}
