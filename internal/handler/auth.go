package handler

// Login & logout (materi #11). Akun dibuat oleh admin, jadi tidak ada halaman register:
// kasir tidak boleh mendaftarkan dirinya sendiri (least privilege, materi #12).

import (
	"errors"
	"net/http"

	"seduhpoint/internal/middleware"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type AuthHandler struct {
	Svc          *service.AuthService
	View         *Renderer
	CookieSecure bool
}

func (h *AuthHandler) FormLogin(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, r, "login.html", Halaman{Judul: "Login"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	token, err := h.Svc.Login(r.Context(), username, r.PostFormValue("password"))
	if errors.Is(err, model.ErrLoginGagal) {
		w.WriteHeader(http.StatusUnauthorized)
		h.View.Tampil(w, r, "login.html", Halaman{Judul: "Login", Error: "Username atau password salah",
			Form: map[string]string{"username": username}})
		return
	}
	if err != nil {
		gagal(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     middleware.NamaCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   8 * 60 * 60,    // 8 jam, sama dengan expires_at di database
		HttpOnly: true,           // tidak bisa dibaca JavaScript
		Secure:   h.CookieSecure, // true di production (HTTPS)
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout: hapus session di server dan hapus cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(middleware.NamaCookie); err == nil {
		h.Svc.Logout(r.Context(), c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: middleware.NamaCookie, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
