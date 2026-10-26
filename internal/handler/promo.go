package handler

// F-06 – F-11 — TODO M2

import (
	"net/http"

	"seduhpoint/internal/service"
)

type PromoHandler struct {
	Svc  *service.PromoService
	View *Renderer
}

func (h *PromoHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "promo/list.html", halaman(r, "Promo & Voucher", nil))
}
func (h *PromoHandler) FormTambah(w http.ResponseWriter, r *http.Request) {
	// kirim daftar cabang, produk, kategori untuk pilihan target
	h.View.Tampil(w, "promo/form.html", halaman(r, "Tambah Promo", nil))
}
func (h *PromoHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/promo", http.StatusSeeOther)
}
func (h *PromoHandler) FormEdit(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "promo/form.html", halaman(r, "Ubah Promo", nil))
}
func (h *PromoHandler) Update(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/promo", http.StatusSeeOther)
}
func (h *PromoHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/promo", http.StatusSeeOther)
}
