package handler

// F-02 — TODO M2: ikuti pola cabang.go

import (
	"net/http"

	"seduhpoint/internal/service"
)

type ProdukHandler struct {
	Svc  *service.ProdukService
	View *Renderer
}

func (h *ProdukHandler) DaftarKategori(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "kategori/list.html", halaman(r, "Kategori", nil))
}
func (h *ProdukHandler) SimpanKategori(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/kategori", http.StatusSeeOther)
}
func (h *ProdukHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "produk/list.html", halaman(r, "Produk", nil))
}
func (h *ProdukHandler) FormTambah(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "produk/form.html", halaman(r, "Tambah Produk", nil))
}
func (h *ProdukHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/produk", http.StatusSeeOther)
}
func (h *ProdukHandler) FormEdit(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "produk/form.html", halaman(r, "Ubah Produk", nil))
}
func (h *ProdukHandler) Update(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/produk", http.StatusSeeOther)
}
func (h *ProdukHandler) Hapus(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/produk", http.StatusSeeOther)
}
