package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"seduhpoint/internal/model"
)

// RequireRole (materi #12): allow-list, hanya peran yang sama persis yang lolos.
func TestRequireRole(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	khususAdmin := RequireRole(model.PeranAdmin)(ok)

	cases := []struct {
		nama   string
		login  bool
		peran  string
		status int
	}{
		{"admin boleh", true, model.PeranAdmin, http.StatusOK},
		{"kasir ditolak", true, model.PeranKasir, http.StatusForbidden},
		{"peran tidak dikenal ditolak", true, "SUPERVISOR", http.StatusForbidden},
		{"belum login diarahkan ke /login", false, "", http.StatusSeeOther},
	}
	for _, c := range cases {
		r := httptest.NewRequest("GET", "/cabang", nil)
		if c.login {
			r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, model.Pengguna{ID: 1, Peran: c.peran}))
		}
		w := httptest.NewRecorder()
		khususAdmin.ServeHTTP(w, r)
		if w.Code != c.status {
			t.Errorf("%s: status %d, harusnya %d", c.nama, w.Code, c.status)
		}
	}
}

// RequireLogin tanpa cookie session langsung diarahkan ke /login (tanpa menyentuh database).
func TestRequireLoginTanpaCookie(t *testing.T) {
	a := &Auth{}
	h := a.RequireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler tidak boleh dipanggil tanpa login")
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/transaksi", nil))
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login" {
		t.Errorf("harusnya 303 ke /login, dapat %d %q", w.Code, w.Header().Get("Location"))
	}
}

func TestSecureHeaders(t *testing.T) {
	h := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/login", nil))
	for _, nama := range []string{"Content-Security-Policy", "X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy"} {
		if w.Header().Get(nama) == "" {
			t.Errorf("header %s tidak ada", nama)
		}
	}
}
