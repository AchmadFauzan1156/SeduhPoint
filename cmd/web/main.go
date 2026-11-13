// SeduhPoint: sistem poin member dan promo kedai kopi.
// Jalankan dari root project:  go run ./cmd/web   (atau: wgo run ./cmd/web)
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"

	"seduhpoint/internal/config"
	"seduhpoint/internal/handler"
	mw "seduhpoint/internal/middleware"
	"seduhpoint/internal/model"
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
	authSvc := &service.AuthService{Pengguna: penggunaRepo, Sessions: &repository.SessionRepo{DB: db}}
	cabangSvc := &service.CabangService{DB: db, Repo: cabangRepo, PenggunaRepo: penggunaRepo}
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
	auth := &handler.AuthHandler{Svc: authSvc, View: view, CookieSecure: cfg.CookieSecure}
	cabang := &handler.CabangHandler{Svc: cabangSvc, View: view}
	produk := &handler.ProdukHandler{Svc: produkSvc, PromoSvc: promoSvc, View: view}
	member := &handler.MemberHandler{Svc: memberSvc, View: view}
	promo := &handler.PromoHandler{Svc: promoSvc, CabangSvc: cabangSvc, ProdukSvc: produkSvc, View: view}
	transaksi := &handler.TransaksiHandler{Svc: transaksiSvc, ProdukSvc: produkSvc, MemberSvc: memberSvc, CabangSvc: cabangSvc, View: view}
	pengguna := &handler.PenggunaHandler{Svc: authSvc, View: view}
	laporan := &handler.LaporanHandler{Svc: laporanSvc, CabangSvc: cabangSvc, View: view}

	// Job harian: kedaluwarsa poin & evaluasi tingkat
	go poinSvc.JalankanJobHarian()

	authMw := &mw.Auth{Svc: authSvc}
	khususAdmin := mw.RequireRole(model.PeranAdmin)
	khususKasir := mw.RequireRole(model.PeranKasir)
	// admin(h) dan kasir(h) = khususAdmin(http.HandlerFunc(h)) seperti di materi, disingkat agar daftar route mudah dibaca.
	admin := func(h http.HandlerFunc) http.Handler { return khususAdmin(h) }
	kasir := func(h http.HandlerFunc) http.Handler { return khususKasir(h) }

	// privat: semua route yang wajib login (deny by default, materi #12).
	// Route baru otomatis wajib login; yang khusus peran tertentu dibungkus admin(...) / kasir(...).
	privat := http.NewServeMux()
	privat.HandleFunc("GET /{$}", halaman.Dashboard)
	privat.HandleFunc("POST /logout", auth.Logout)

	// Data master (admin)
	privat.Handle("GET /cabang", admin(cabang.Daftar))
	privat.Handle("GET /cabang/form", admin(cabang.FormTambah))
	privat.Handle("POST /cabang", admin(cabang.Simpan))
	privat.Handle("GET /cabang/{id}/edit", admin(cabang.FormEdit))
	privat.Handle("POST /cabang/{id}/edit", admin(cabang.Update))
	privat.Handle("POST /cabang/{id}/hapus", admin(cabang.Hapus))

	privat.Handle("GET /kategori", admin(produk.DaftarKategori))
	privat.Handle("POST /kategori", admin(produk.SimpanKategori))
	privat.Handle("GET /kategori/{id}/edit", admin(produk.FormEditKategori))
	privat.Handle("POST /kategori/{id}/edit", admin(produk.UpdateKategori))
	privat.Handle("POST /kategori/{id}/hapus", admin(produk.HapusKategori))
	privat.Handle("GET /produk", admin(produk.Daftar))
	privat.Handle("GET /produk/form", admin(produk.FormTambah))
	privat.Handle("POST /produk", admin(produk.Simpan))
	privat.Handle("GET /produk/{id}/edit", admin(produk.FormEdit))
	privat.Handle("POST /produk/{id}/edit", admin(produk.Update))
	privat.Handle("POST /produk/{id}/hapus", admin(produk.Hapus))

	// Member (admin & kasir)
	privat.HandleFunc("GET /member", member.Daftar)
	privat.HandleFunc("GET /member/form", member.FormTambah)
	privat.HandleFunc("POST /member", member.Simpan)
	privat.HandleFunc("GET /member/{id}", member.Detail)
	privat.HandleFunc("GET /member/{id}/edit", member.FormEdit)
	privat.HandleFunc("POST /member/{id}/edit", member.Update)
	privat.Handle("POST /member/{id}/hapus", admin(member.Hapus))

	// Promo & voucher (admin)
	privat.Handle("GET /promo", admin(promo.Daftar))
	privat.Handle("GET /promo/form", admin(promo.FormTambah))
	privat.Handle("POST /promo", admin(promo.Simpan))
	privat.Handle("GET /promo/{id}/edit", admin(promo.FormEdit))
	privat.Handle("POST /promo/{id}/edit", admin(promo.Update))
	privat.Handle("POST /promo/{id}/toggle", admin(promo.Toggle))

	// Transaksi (kasir) & riwayat/pembatalan (admin & kasir; hak per transaksi dicek di service)
	privat.Handle("GET /kasir", kasir(transaksi.Kasir))
	privat.Handle("POST /kasir/hitung", kasir(transaksi.Hitung))
	privat.Handle("POST /kasir/bayar", kasir(transaksi.Bayar))
	privat.HandleFunc("GET /transaksi", transaksi.Daftar)
	privat.HandleFunc("GET /transaksi/{id}", transaksi.Struk)
	privat.HandleFunc("GET /transaksi/{id}/batal", transaksi.FormBatal)
	privat.HandleFunc("POST /transaksi/{id}/batal", transaksi.Batal)

	// Laporan pemilik (admin)
	privat.Handle("GET /laporan/promo", admin(laporan.Promo))
	privat.Handle("GET /laporan/poin", admin(laporan.Poin))
	privat.Handle("GET /laporan/member", admin(laporan.Member))

	// Pengguna & password (admin)
	privat.Handle("GET /pengguna", admin(pengguna.Daftar))
	privat.Handle("POST /pengguna/password", admin(pengguna.GantiPassword))
	privat.Handle("GET /pengguna/{id}/password", admin(pengguna.FormReset))
	privat.Handle("POST /pengguna/{id}/password", admin(pengguna.Reset))

	// publik: boleh diakses tanpa login
	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("GET /login", auth.FormLogin)
	mux.HandleFunc("POST /login", auth.Login)
	mux.Handle("/", authMw.RequireLogin(privat)) // sisanya lewat RequireLogin

	// Tolak POST yang dikirim dari situs lain atas nama pengguna yang sedang login (CSRF, materi #12).
	csrf := http.NewCrossOriginProtection()

	// Alur: header keamanan -> csrf -> Logging -> BatasiBody -> mux -> /login, /static (publik)
	//                                            -> RequireLogin -> privat -> RequireRole (admin/kasir)
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mw.SecureHeaders(csrf.Handler(mw.Logging(mw.BatasiBody(mux)))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("server jalan di http://%s", cfg.Addr)
	log.Fatal(srv.ListenAndServe())
}

func bukaDB(dsn string) (*sql.DB, error) {
	// ClientFoundRows: RowsAffected menghitung baris yang cocok dengan WHERE, bukan hanya yang berubah.
	// Tanpa ini, menyimpan form tanpa perubahan apa pun dianggap "data tidak ditemukan" oleh pastikanAdaBaris.
	c, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	c.ClientFoundRows = true
	db, err := sql.Open("mysql", c.FormatDSN())
	if err != nil {
		return nil, err
	}
	// Batasi ukuran pool (materi #6): default MaxOpenConns tanpa batas bisa menghabiskan
	// jatah koneksi MySQL, dan ConnMaxLifetime mencegah koneksi basi yang sudah diputus MySQL.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	log.Println("database terhubung")
	return db, nil
}
