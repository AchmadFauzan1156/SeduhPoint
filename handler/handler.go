// Package handler menerima request, memanggil store/layanan, lalu render template.
package handler

import (
	"net/http"
	"strconv"

	"seduhpoint/layanan"
	mw "seduhpoint/middleware"
	"seduhpoint/store"
	"seduhpoint/view"
)

type Handler struct {
	Store   *store.Store
	View    *view.View
	Layanan *layanan.Layanan
}

// Page adalah data standar untuk setiap template.
type Page struct {
	Judul  string
	Sesi   mw.Sesi
	Flash  string
	Errors map[string]string // validasi server (materi #4)
	Form   map[string]string // isi form untuk ditampilkan ulang
	Data   any
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, name, judul string, data any) {
	s, _ := mw.Ambil(r)
	p := Page{Judul: judul, Sesi: s, Data: data, Flash: r.URL.Query().Get("pesan")}
	h.View.Render(w, name, p)
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

// Routes: pola Go 1.22 "METHOD /path/{id}".
func Routes(mux *http.ServeMux, h *Handler) {
	const A, K = "ADMIN", "KASIR"
	login, admin := mw.Login, func(f http.HandlerFunc) http.HandlerFunc { return mw.Peran(A, f) }

	mux.HandleFunc("GET /login", h.LoginForm)
	mux.HandleFunc("POST /login", h.LoginProses)
	mux.HandleFunc("POST /logout", h.Logout)
	mux.HandleFunc("GET /{$}", login(h.Dashboard))

	// Data master (F-01 – F-04)
	mux.HandleFunc("GET /cabang", admin(h.CabangList))
	mux.HandleFunc("GET /cabang/baru", admin(h.CabangForm))
	mux.HandleFunc("GET /cabang/{id}/edit", admin(h.CabangForm))
	mux.HandleFunc("POST /cabang", admin(h.CabangSimpan))
	mux.HandleFunc("GET /kategori", admin(h.KategoriList))
	mux.HandleFunc("POST /kategori", admin(h.KategoriSimpan))
	mux.HandleFunc("GET /produk", admin(h.ProdukList))
	mux.HandleFunc("GET /produk/baru", admin(h.ProdukForm))
	mux.HandleFunc("GET /produk/{id}/edit", admin(h.ProdukForm))
	mux.HandleFunc("POST /produk", admin(h.ProdukSimpan))
	mux.HandleFunc("GET /member", login(h.MemberList))
	mux.HandleFunc("GET /member/baru", login(h.MemberForm))
	mux.HandleFunc("GET /member/{id}", login(h.MemberDetail))
	mux.HandleFunc("GET /member/{id}/edit", login(h.MemberForm))
	mux.HandleFunc("POST /member", login(h.MemberSimpan))

	// Promo & voucher (F-06 – F-11)
	mux.HandleFunc("GET /promo", admin(h.PromoList))
	mux.HandleFunc("GET /promo/baru", admin(h.PromoForm))
	mux.HandleFunc("GET /promo/{id}/edit", admin(h.PromoForm))
	mux.HandleFunc("POST /promo", admin(h.PromoSimpan))
	mux.HandleFunc("POST /promo/{id}/toggle", admin(h.PromoToggle))

	// Transaksi & pembatalan (F-12 – F-24)
	mux.HandleFunc("GET /kasir", mw.Peran(K, h.KasirForm))
	mux.HandleFunc("POST /kasir/hitung", mw.Peran(K, h.KasirHitung))
	mux.HandleFunc("POST /kasir/bayar", mw.Peran(K, h.KasirBayar))
	mux.HandleFunc("GET /transaksi", login(h.TransaksiList))
	mux.HandleFunc("GET /transaksi/{id}", login(h.Struk))
	mux.HandleFunc("GET /transaksi/{id}/batal", login(h.BatalForm))
	mux.HandleFunc("POST /transaksi/{id}/batal", login(h.BatalProses))

	// Laporan pemilik (F-28 – F-30)
	mux.HandleFunc("GET /laporan/promo", admin(h.LaporanPromo))
	mux.HandleFunc("GET /laporan/poin", admin(h.LaporanPoin))
	mux.HandleFunc("GET /laporan/member", admin(h.LaporanMember))
}
