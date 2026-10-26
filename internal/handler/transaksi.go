package handler

// F-12 – F-24 — TODO M3

import (
	"net/http"

	"seduhpoint/internal/service"
)

type TransaksiHandler struct {
	Svc  *service.TransaksiService
	View *Renderer
}

func (h *TransaksiHandler) Kasir(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "transaksi/kasir.html", halaman(r, "Transaksi Baru", "kasir", nil))
}
func (h *TransaksiHandler) Hitung(w http.ResponseWriter, r *http.Request) {
	// parse form -> h.Svc.Pratinjau -> tampilkan rincian (F-13, F-19)
	h.View.Tampil(w, "transaksi/kasir.html", halaman(r, "Transaksi Baru", "kasir", nil))
}
func (h *TransaksiHandler) Bayar(w http.ResponseWriter, r *http.Request) {
	// h.Svc.Checkout -> redirect ke struk (PRG)
	http.Redirect(w, r, "/transaksi", http.StatusSeeOther)
}
func (h *TransaksiHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	// kasir hanya melihat cabangnya sendiri
	h.View.Tampil(w, "transaksi/list.html", halaman(r, "Riwayat Transaksi", "transaksi", nil))
}
func (h *TransaksiHandler) Struk(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "transaksi/struk.html", halaman(r, "Struk", "transaksi", nil))
}
func (h *TransaksiHandler) FormBatal(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "transaksi/batal.html", halaman(r, "Batalkan Transaksi", "transaksi", nil))
}
func (h *TransaksiHandler) Batal(w http.ResponseWriter, r *http.Request) {
	// h.Svc.Batalkan(ctx, id, alasan)
	http.Redirect(w, r, "/transaksi", http.StatusSeeOther)
}
