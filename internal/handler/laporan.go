package handler

// F-28 – F-30 — TODO M4

import (
	"net/http"

	"seduhpoint/internal/service"
)

type LaporanHandler struct {
	Svc  *service.LaporanService
	View *Renderer
}

func (h *LaporanHandler) Promo(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "laporan/promo.html", halaman(r, "Laporan Biaya Promo", "lap-promo", nil))
}
func (h *LaporanHandler) Poin(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "laporan/poin.html", halaman(r, "Laporan Poin Aktif", "lap-poin", nil))
}
func (h *LaporanHandler) Member(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "laporan/member.html", halaman(r, "Laporan Perkembangan Member", "lap-member", nil))
}
