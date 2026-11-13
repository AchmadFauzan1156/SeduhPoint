package middleware

// RequireLogin (materi #11) dan RequireRole (materi #12).

import (
	"context"
	"errors"
	"log"
	"net/http"

	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

// NamaCookie: nama cookie yang membawa token session.
const NamaCookie = "session"

type ctxKey struct{}

type Auth struct {
	Svc *service.AuthService
}

// RequireLogin: request tanpa session yang valid diarahkan ke /login.
// Pengguna yang login dititipkan di context request.
func (a *Auth) RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(NamaCookie)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		p, err := a.Svc.PenggunaDariToken(r.Context(), c.Value)
		if errors.Is(err, model.ErrNotFound) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if err != nil {
			log.Printf("cek session: %v", err)
			http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PenggunaDari mengembalikan pengguna yang sedang login.
func PenggunaDari(ctx context.Context) (model.Pengguna, bool) {
	p, ok := ctx.Value(ctxKey{}).(model.Pengguna)
	return p, ok
}

// RequireRole dipasang setelah RequireLogin: pengguna di context sudah terisi.
// Allow-list: hanya peran yang sama persis yang lolos.
func RequireRole(peran string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PenggunaDari(r.Context())
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if p.Peran != peran {
				http.Error(w, "Akses ditolak", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
