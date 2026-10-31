package handler

import (
	"net/http"

	"seduhpoint/internal/middleware"
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

func (h *AuthHandler) FormPassword(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "password.html", halaman(r, "Ganti password", "password", nil))
}

func (h *AuthHandler) GantiPassword(w http.ResponseWriter, r *http.Request) {
	s, _ := middleware.Ambil(r)
	err := h.Svc.GantiPassword(r.Context(), s.PenggunaID,
		r.PostFormValue("lama"), r.PostFormValue("baru"), r.PostFormValue("ulang"))
	if err != nil {
		formGagal(w, h.View, "password.html", halaman(r, "Ganti password", "password", nil), err)
		return
	}
	middleware.HapusSesiLain(r, s.PenggunaID)
	redirect(w, r, "/password", "Password berhasil diganti")
}
