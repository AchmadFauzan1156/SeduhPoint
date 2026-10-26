package handler

import "net/http"

// F-28 – F-30 — TODO M4
func (h *Handler) LaporanPromo(w http.ResponseWriter, r *http.Request) {
	// filter: dari, sampai, cabang
	h.render(w, r, "laporan/promo", "Laporan Biaya Promo", nil)
}
func (h *Handler) LaporanPoin(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "laporan/poin", "Laporan Poin Aktif", nil)
}
func (h *Handler) LaporanMember(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "laporan/member", "Laporan Perkembangan Member", nil)
}
