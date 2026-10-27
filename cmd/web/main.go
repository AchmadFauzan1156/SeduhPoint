// SeduhPoint — Sistem Poin Member & Promo Kedai Kopi Seduh (Kelompok 5)
// Jalankan dari root project:  go run ./cmd/web   (atau: wgo run ./cmd/web)
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"seduhpoint/internal/config"
	"seduhpoint/internal/handler"
	mw "seduhpoint/internal/middleware"
	"seduhpoint/internal/repository"
	"seduhpoint/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := bukaDB(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Repository
	cabangRepo := &repository.CabangRepo{DB: db}
	penggunaRepo := &repository.PenggunaRepo{DB: db}
	kategoriRepo := &repository.KategoriRepo{DB: db}
	produkRepo := &repository.ProdukRepo{DB: db}
	memberRepo := &repository.MemberRepo{DB: db}
	promoRepo := &repository.PromoRepo{DB: db}
	transaksiRepo := &repository.TransaksiRepo{DB: db}
	poinRepo := &repository.PoinRepo{DB: db}
	laporanRepo := &repository.LaporanRepo{DB: db}

	// Service
	authSvc := &service.AuthService{Repo: penggunaRepo}
	cabangSvc := &service.CabangService{Repo: cabangRepo}
	produkSvc := &service.ProdukService{Repo: produkRepo, KategoriRepo: kategoriRepo}
	memberSvc := &service.MemberService{DB: db, Repo: memberRepo, PoinRepo: poinRepo}
	promoSvc := &service.PromoService{Repo: promoRepo}
	poinSvc := &service.PoinService{DB: db, PoinRepo: poinRepo, MemberRepo: memberRepo}
	transaksiSvc := &service.TransaksiService{DB: db, Repo: transaksiRepo, PromoRepo: promoRepo,
		PoinRepo: poinRepo, ProdukRepo: produkRepo, MemberRepo: memberRepo, Tingkat: poinSvc}
	laporanSvc := &service.LaporanService{Repo: laporanRepo}

	// Handler
	view := handler.NewRenderer()
	halaman := &handler.HalamanHandler{LaporanSvc: laporanSvc, MemberSvc: memberSvc, TransaksiSvc: transaksiSvc, View: view}
	auth := &handler.AuthHandler{Svc: authSvc, View: view}
	cabang := &handler.CabangHandler{Svc: cabangSvc, View: view}
	produk := &handler.ProdukHandler{Svc: produkSvc, PromoSvc: promoSvc, View: view}
	member := &handler.MemberHandler{Svc: memberSvc, View: view}
	promo := &handler.PromoHandler{Svc: promoSvc, CabangSvc: cabangSvc, ProdukSvc: produkSvc, View: view}
	transaksi := &handler.TransaksiHandler{Svc: transaksiSvc, ProdukSvc: produkSvc, MemberSvc: memberSvc, CabangSvc: cabangSvc, View: view}
	laporan := &handler.LaporanHandler{Svc: laporanSvc, CabangSvc: cabangSvc, View: view}

	// Job harian: kedaluwarsa poin & evaluasi tingkat (F-25, F-26)
	go poinSvc.JalankanJobHarian()

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /login", auth.FormLogin)
	mux.HandleFunc("POST /login", auth.Login)
	mux.HandleFunc("POST /logout", auth.Logout)
	mux.HandleFunc("GET /{$}", mw.Login(halaman.Dashboard))

	// Data master (F-01 – F-04)
	mux.HandleFunc("GET /cabang", mw.Admin(cabang.Daftar))
	mux.HandleFunc("GET /cabang/form", mw.Admin(cabang.FormTambah))
	mux.HandleFunc("POST /cabang", mw.Admin(cabang.Simpan))
	mux.HandleFunc("GET /cabang/{id}/edit", mw.Admin(cabang.FormEdit))
	mux.HandleFunc("POST /cabang/{id}/edit", mw.Admin(cabang.Update))
	mux.HandleFunc("POST /cabang/{id}/hapus", mw.Admin(cabang.Hapus))

	mux.HandleFunc("GET /kategori", mw.Admin(produk.DaftarKategori))
	mux.HandleFunc("POST /kategori", mw.Admin(produk.SimpanKategori))
	mux.HandleFunc("GET /kategori/{id}/edit", mw.Admin(produk.FormEditKategori))
	mux.HandleFunc("POST /kategori/{id}/edit", mw.Admin(produk.UpdateKategori))
	mux.HandleFunc("POST /kategori/{id}/hapus", mw.Admin(produk.HapusKategori))
	mux.HandleFunc("GET /produk", mw.Admin(produk.Daftar))
	mux.HandleFunc("GET /produk/form", mw.Admin(produk.FormTambah))
	mux.HandleFunc("POST /produk", mw.Admin(produk.Simpan))
	mux.HandleFunc("GET /produk/{id}/edit", mw.Admin(produk.FormEdit))
	mux.HandleFunc("POST /produk/{id}/edit", mw.Admin(produk.Update))
	mux.HandleFunc("POST /produk/{id}/hapus", mw.Admin(produk.Hapus))

	mux.HandleFunc("GET /member", mw.Login(member.Daftar))
	mux.HandleFunc("GET /member/form", mw.Login(member.FormTambah))
	mux.HandleFunc("POST /member", mw.Login(member.Simpan))
	mux.HandleFunc("GET /member/{id}", mw.Login(member.Detail))
	mux.HandleFunc("GET /member/{id}/edit", mw.Login(member.FormEdit))
	mux.HandleFunc("POST /member/{id}/edit", mw.Login(member.Update))

	// Promo & voucher (F-06 – F-11)
	mux.HandleFunc("GET /promo", mw.Admin(promo.Daftar))
	mux.HandleFunc("GET /promo/form", mw.Admin(promo.FormTambah))
	mux.HandleFunc("POST /promo", mw.Admin(promo.Simpan))
	mux.HandleFunc("GET /promo/{id}/edit", mw.Admin(promo.FormEdit))
	mux.HandleFunc("POST /promo/{id}/edit", mw.Admin(promo.Update))
	mux.HandleFunc("POST /promo/{id}/toggle", mw.Admin(promo.Toggle))

	// Transaksi & pembatalan (F-12 – F-24)
	mux.HandleFunc("GET /kasir", mw.Kasir(transaksi.Kasir))
	mux.HandleFunc("POST /kasir/hitung", mw.Kasir(transaksi.Hitung))
	mux.HandleFunc("POST /kasir/bayar", mw.Kasir(transaksi.Bayar))
	mux.HandleFunc("GET /transaksi", mw.Login(transaksi.Daftar))
	mux.HandleFunc("GET /transaksi/{id}", mw.Login(transaksi.Struk))
	mux.HandleFunc("GET /transaksi/{id}/batal", mw.Login(transaksi.FormBatal))
	mux.HandleFunc("POST /transaksi/{id}/batal", mw.Login(transaksi.Batal))

	// Laporan pemilik (F-28 – F-30)
	mux.HandleFunc("GET /laporan/promo", mw.Admin(laporan.Promo))
	mux.HandleFunc("GET /laporan/poin", mw.Admin(laporan.Poin))
	mux.HandleFunc("GET /laporan/member", mw.Admin(laporan.Member))

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mw.Logging(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("server jalan di http://%s", cfg.Addr)
	log.Fatal(srv.ListenAndServe())
}

func bukaDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	log.Println("database terhubung")
	return db, nil
}
