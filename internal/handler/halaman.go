package handler

import "net/http"

type HalamanHandler struct {
	View *Renderer
}

func (h *HalamanHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	// TODO M4: ringkasan hari ini (jumlah transaksi, omzet, poin aktif, member Gold)
	h.View.Tampil(w, "dashboard.html", halaman(r, "Dashboard", nil))
}
