package handler

import "net/http"

// F-02 — TODO M2: ikuti pola cabang.go
func (h *Handler) KategoriList(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "kategori/list", "Kategori", nil)
}
func (h *Handler) KategoriSimpan(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/kategori", http.StatusSeeOther)
}
func (h *Handler) ProdukList(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "produk/list", "Produk", nil)
}
func (h *Handler) ProdukForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "produk/form", "Form Produk", nil)
}
func (h *Handler) ProdukSimpan(w http.ResponseWriter, r *http.Request) {
	// validasi: nama wajib, harga >= 0 (integer rupiah), kategori ada
	http.Redirect(w, r, "/produk", http.StatusSeeOther)
}
