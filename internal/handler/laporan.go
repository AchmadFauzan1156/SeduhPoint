package handler

// laporan.go: halaman laporan pemilik (F-28 – F-30), hanya untuk admin.

import (
	"net/http"
	"strconv"
	"time"

	"seduhpoint/internal/config"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type LaporanHandler struct {
	Svc       *service.LaporanService
	CabangSvc *service.CabangService
	View      *Renderer
}

type LaporanPromoPage struct {
	Lap     service.LaporanPromo
	Cabang  []model.Cabang
	NamaCab string
	Dari    string
	Sampai  string
}

// Promo: GET /laporan/promo?dari=YYYY-MM-DD&sampai=YYYY-MM-DD&cabang=ID (F-28).
func (h *LaporanHandler) Promo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now().In(config.WIB)
	hal := halaman(r, "Laporan biaya promo", "lap-promo", nil)
	q := r.URL.Query()
	pg := LaporanPromoPage{Dari: q.Get("dari"), Sampai: q.Get("sampai")}
	if pg.Dari == "" {
		pg.Dari = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, config.WIB).Format("2006-01-02")
	}
	if pg.Sampai == "" {
		pg.Sampai = now.Format("2006-01-02")
	}
	dari, e1 := time.ParseInLocation("2006-01-02", pg.Dari, config.WIB)
	sampai, e2 := time.ParseInLocation("2006-01-02", pg.Sampai, config.WIB)
	cabangID, _ := strconv.Atoi(q.Get("cabang"))
	pg.Cabang, _ = h.CabangSvc.Daftar(ctx)
	pg.NamaCab = "semua cabang"
	for _, c := range pg.Cabang {
		if c.ID == cabangID {
			pg.NamaCab = c.Nama
		}
	}
	switch {
	case e1 != nil:
		hal.Errors["dari"] = "Tanggal tidak valid"
	case e2 != nil:
		hal.Errors["sampai"] = "Tanggal tidak valid"
	default:
		lap, err := h.Svc.BiayaPromo(ctx, dari, sampai, cabangID)
		if pesan, ok := pesanValidasi(err); ok {
			hal.Errors["sampai"] = pesan
		} else if err != nil {
			gagal(w, err)
			return
		}
		pg.Lap = lap
	}
	pg.Lap.CabangID = cabangID
	hal.Data = pg
	h.View.Tampil(w, "laporan/promo.html", hal)
}

// Poin: GET /laporan/poin (F-29).
func (h *LaporanHandler) Poin(w http.ResponseWriter, r *http.Request) {
	lap, err := h.Svc.Poin(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	h.View.Tampil(w, "laporan/poin.html", halaman(r, "Laporan poin aktif", "lap-poin", lap))
}

type LaporanMemberPage struct {
	Lap   service.LaporanMember
	Tahun []int
}

// Member: GET /laporan/member?tahun=YYYY (F-30).
func (h *LaporanHandler) Member(w http.ResponseWriter, r *http.Request) {
	now := time.Now().In(config.WIB)
	tahun, err := strconv.Atoi(r.URL.Query().Get("tahun"))
	if err != nil || tahun < 2020 || tahun > now.Year() {
		tahun = now.Year()
	}
	lap, err := h.Svc.Member(r.Context(), tahun)
	if err != nil {
		gagal(w, err)
		return
	}
	pg := LaporanMemberPage{Lap: lap}
	for t := now.Year(); t >= 2026; t-- {
		pg.Tahun = append(pg.Tahun, t)
	}
	h.View.Tampil(w, "laporan/member.html", halaman(r, "Laporan perkembangan member", "lap-member", pg))
}
