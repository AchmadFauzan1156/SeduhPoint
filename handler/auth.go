package handler

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"

	mw "seduhpoint/middleware"
)

// F-05, NF-03
func (h *Handler) LoginForm(w http.ResponseWriter, r *http.Request) {
	h.View.Render(w, "login", Page{Judul: "Login"})
}

func (h *Handler) LoginProses(w http.ResponseWriter, r *http.Request) {
	username, password := r.PostFormValue("username"), r.PostFormValue("password")
	p, err := h.Store.GetPenggunaByUsername(username)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(password)) != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		h.View.Render(w, "login", Page{Judul: "Login", Flash: "Username atau password salah",
			Form: map[string]string{"username": username}})
		return
	}
	mw.BuatSesi(w, mw.Sesi{PenggunaID: p.ID, Username: p.Username, Peran: p.Peran, CabangID: int(p.CabangID.Int64)})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	mw.HapusSesi(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	// TODO M4: ringkasan hari ini (jumlah transaksi, omzet, poin aktif)
	h.render(w, r, "dashboard", "Dashboard", nil)
}
