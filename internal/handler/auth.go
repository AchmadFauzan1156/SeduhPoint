package handler

import (
	"net/http"

	"seduhpoint/internal/middleware"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type AuthHandler struct {
	Svc  *service.AuthService
	View *Renderer
}

func (h *AuthHandler) FormLogin(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "login.html", Halaman{Judul: "Login"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	p, err := h.Svc.Login(r.Context(), username, r.PostFormValue("password"))
	if err != nil {
		if pesan, ok := pesanValidasi(err); ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			h.View.Tampil(w, "login.html", Halaman{Judul: "Login", Error: pesan,
				Form: map[string]string{"username": username}})
			return
		}
		gagal(w, err)
		return
	}
	middleware.BuatSesi(w, middleware.Sesi{PenggunaID: p.ID, Username: p.Username,
		Peran: p.Peran, CabangID: int(p.CabangID.Int64)})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	middleware.HapusSesi(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// PasswordPage: halaman kelola password (admin). Target = kasir yang password-nya sedang di-reset.
type PasswordPage struct {
	Kasir  []model.Pengguna
	Target model.Pengguna
}

func (h *AuthHandler) tampilPassword(w http.ResponseWriter, r *http.Request, target model.Pengguna, err error) {
	daftar, e := h.Svc.DaftarKasir(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	judul := "Kelola password"
	if target.ID > 0 {
		judul = "Reset password " + target.Username
	}
	hal := halaman(r, judul, "password", PasswordPage{Kasir: daftar, Target: target})
	if err != nil {
		formGagal(w, h.View, "password.html", hal, err)
		return
	}
	h.View.Tampil(w, "password.html", hal)
}

func (h *AuthHandler) FormPassword(w http.ResponseWriter, r *http.Request) {
	h.tampilPassword(w, r, model.Pengguna{}, nil)
}

// GantiPassword: admin mengganti password akunnya sendiri (wajib password lama).
func (h *AuthHandler) GantiPassword(w http.ResponseWriter, r *http.Request) {
	s, _ := middleware.Ambil(r)
	err := h.Svc.GantiPassword(r.Context(), s.PenggunaID,
		r.PostFormValue("lama"), r.PostFormValue("baru"), r.PostFormValue("ulang"))
	if err != nil {
		h.tampilPassword(w, r, model.Pengguna{}, err)
		return
	}
	middleware.HapusSesiLain(r, s.PenggunaID)
	redirect(w, r, "/password", "Password admin berhasil diganti")
}

func (h *AuthHandler) FormResetKasir(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	k, err := h.Svc.Kasir(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.tampilPassword(w, r, k, nil)
}

// ResetKasir: admin mengatur password baru kasir; kasir otomatis keluar dari semua perangkat.
func (h *AuthHandler) ResetKasir(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	k, err := h.Svc.Kasir(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	if err := h.Svc.ResetPasswordKasir(r.Context(), id, r.PostFormValue("baru"), r.PostFormValue("ulang")); err != nil {
		h.tampilPassword(w, r, k, err)
		return
	}
	middleware.HapusSesiPengguna(id)
	redirect(w, r, "/password", "Password "+k.Username+" berhasil di-reset")
}
