package handler

import "net/http"

// F-06 – F-11 — TODO M2
func (h *Handler) PromoList(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "promo/list", "Promo & Voucher", nil)
}
func (h *Handler) PromoForm(w http.ResponseWriter, r *http.Request) {
	// kirim daftar cabang, produk, kategori untuk pilihan target
	h.render(w, r, "promo/form", "Form Promo", nil)
}
func (h *Handler) PromoSimpan(w http.ResponseWriter, r *http.Request) {
	// validasi per jenis: DISKON_PERSEN butuh persen+produk, MIN_BELANJA butuh min+nominal,
	// B2G1 butuh kategori, VOUCHER butuh kode unik+nominal+kuota; tgl_mulai <= tgl_selesai
	http.Redirect(w, r, "/promo", http.StatusSeeOther)
}
func (h *Handler) PromoToggle(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/promo", http.StatusSeeOther)
}
