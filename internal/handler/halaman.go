package handler

import (
	"net/http"
	"strconv"

	"seduhpoint/internal/middleware"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type HalamanHandler struct {
	LaporanSvc   *service.LaporanService
	MemberSvc    *service.MemberService
	TransaksiSvc *service.TransaksiService
	View         *Renderer
}

type DashboardPage struct {
	R           service.Ringkasan
	MemberTotal int
	MemberGold  int
	Terakhir    []model.Transaksi
}

// Dashboard: kasir langsung ke halaman transaksi baru; admin ke ringkasan (hari ini / 7 / 30 hari).
func (h *HalamanHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if s, _ := middleware.Ambil(r); s.Peran == "KASIR" {
		http.Redirect(w, r, "/kasir", http.StatusSeeOther)
		return
	}
	ctx := r.Context()
	hari, _ := strconv.Atoi(r.URL.Query().Get("hari"))
	rk, err := h.LaporanSvc.Ringkasan(ctx, hari)
	if err != nil {
		gagal(w, err)
		return
	}
	pg := DashboardPage{R: rk}
	if pg.MemberTotal, pg.MemberGold, err = h.MemberSvc.Hitung(ctx); err != nil {
		gagal(w, err)
		return
	}
	if pg.Terakhir, err = h.TransaksiSvc.Daftar(ctx, 0, rk.Sampai); err != nil {
		gagal(w, err)
		return
	}
	if len(pg.Terakhir) > 5 {
		pg.Terakhir = pg.Terakhir[:5]
	}
	h.View.Tampil(w, "dashboard.html", halaman(r, "Ringkasan", "ringkasan", pg))
}
