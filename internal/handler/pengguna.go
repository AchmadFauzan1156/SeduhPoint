package handler

// Halaman Pengguna khusus admin (materi #12): daftar akun, ganti password admin, reset password kasir.

import (
	"fmt"
	"net/http"

	"seduhpoint/internal/middleware"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type PenggunaHandler struct {
	Svc  *service.AuthService
	View *Renderer
}

// PenggunaPage: Target = kasir yang password-nya sedang di-reset (kosong = form password admin).
type PenggunaPage struct {
	Daftar []model.Pengguna
	Target model.Pengguna
	Action string
}

func (h *PenggunaHandler) tampil(w http.ResponseWriter, r *http.Request, target model.Pengguna, err error) {
	daftar, e := h.Svc.DaftarPengguna(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	pg := PenggunaPage{Daftar: daftar, Target: target, Action: "/pengguna/password"}
	judul := "Pengguna"
	if target.ID > 0 {
		pg.Action = fmt.Sprintf("/pengguna/%d/password", target.ID)
		judul = "Reset password " + target.Username
	}
	hal := halaman(r, judul, "pengguna", pg)
	if err != nil {
		formGagal(w, r, h.View, "pengguna/list.html", hal, err)
		return
	}
	h.View.Tampil(w, r, "pengguna/list.html", hal)
}

func (h *PenggunaHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.tampil(w, r, model.Pengguna{}, nil)
}

// GantiPassword: admin mengganti password akunnya sendiri (wajib password lama).
func (h *PenggunaHandler) GantiPassword(w http.ResponseWriter, r *http.Request) {
	u, _ := middleware.PenggunaDari(r.Context())
	token := ""
	if c, err := r.Cookie(middleware.NamaCookie); err == nil {
		token = c.Value
	}
	err := h.Svc.GantiPassword(r.Context(), u, token,
		r.PostFormValue("lama"), r.PostFormValue("baru"), r.PostFormValue("ulang"))
	if err != nil {
		h.tampil(w, r, model.Pengguna{}, err)
		return
	}
	catatAksi(r, "mengganti password akunnya sendiri")
	redirect(w, r, "/pengguna", "password-diubah")
}

func (h *PenggunaHandler) FormReset(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	u, _ := middleware.PenggunaDari(r.Context())
	k, err := h.Svc.KasirUntukReset(r.Context(), u, id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.tampil(w, r, k, nil)
}

func (h *PenggunaHandler) Reset(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	u, _ := middleware.PenggunaDari(r.Context())
	k, err := h.Svc.KasirUntukReset(r.Context(), u, id)
	if err != nil {
		gagal(w, err)
		return
	}
	if err := h.Svc.ResetPasswordKasir(r.Context(), u, id, r.PostFormValue("baru"), r.PostFormValue("ulang")); err != nil {
		h.tampil(w, r, k, err)
		return
	}
	catatAksi(r, "me-reset password kasir %s", k.Username)
	redirect(w, r, "/pengguna", "password-direset")
}
