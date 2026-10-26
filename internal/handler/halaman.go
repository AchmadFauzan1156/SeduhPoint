package handler

import (
	"net/http"

	"seduhpoint/internal/middleware"
)

type HalamanHandler struct {
	View *Renderer
}

// Dashboard: kasir langsung ke halaman transaksi baru; admin ke ringkasan.
func (h *HalamanHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if s, _ := middleware.Ambil(r); s.Peran == "KASIR" {
		http.Redirect(w, r, "/kasir", http.StatusSeeOther)
		return
	}
	// TODO M4: ringkasan hari ini (omzet, transaksi, biaya promo, member Gold)
	h.View.Tampil(w, "dashboard.html", halaman(r, "Ringkasan", "ringkasan", nil))
}
