package handler

import "net/http"

// F-12 – F-24 — TODO M3
func (h *Handler) KasirForm(w http.ResponseWriter, r *http.Request) {
	// kirim produk aktif; kasir pilih member (opsional), item, voucher, poin
	h.render(w, r, "transaksi/kasir", "Kasir", nil)
}
func (h *Handler) KasirHitung(w http.ResponseWriter, r *http.Request) {
	// parse form -> Layanan.Pratinjau -> tampilkan rincian potongan (F-13, F-19)
	h.render(w, r, "transaksi/kasir", "Kasir", nil)
}
func (h *Handler) KasirBayar(w http.ResponseWriter, r *http.Request) {
	// Layanan.Checkout -> redirect ke struk (PRG)
	http.Redirect(w, r, "/transaksi", http.StatusSeeOther)
}
func (h *Handler) TransaksiList(w http.ResponseWriter, r *http.Request) {
	// kasir hanya melihat cabangnya sendiri
	h.render(w, r, "transaksi/list", "Riwayat Transaksi", nil)
}
func (h *Handler) Struk(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "transaksi/struk", "Struk", nil)
}
func (h *Handler) BatalForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "transaksi/batal", "Batalkan Transaksi", nil)
}
func (h *Handler) BatalProses(w http.ResponseWriter, r *http.Request) {
	// alasan wajib -> Layanan.BatalkanTransaksi
	http.Redirect(w, r, "/transaksi", http.StatusSeeOther)
}
