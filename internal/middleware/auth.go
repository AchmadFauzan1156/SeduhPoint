package middleware

// Session sederhana yang disimpan di memori server.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
)

type Sesi struct {
	PenggunaID int
	Username   string
	Peran      string // ADMIN | KASIR
	CabangID   int
}

var (
	mu   sync.Mutex // lindungi map dari data race
	sesi = map[string]Sesi{}
)

const namaCookie = "seduh_sesi"

type ctxKey struct{}

func BuatSesi(w http.ResponseWriter, s Sesi) {
	b := make([]byte, 32)
	rand.Read(b)
	id := hex.EncodeToString(b)
	mu.Lock()
	sesi[id] = s
	mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: namaCookie, Value: id, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func HapusSesi(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(namaCookie); err == nil {
		mu.Lock()
		delete(sesi, c.Value)
		mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: namaCookie, Value: "", Path: "/", MaxAge: -1})
}

// Ambil mengembalikan sesi pengguna yang sedang login.
func Ambil(r *http.Request) (Sesi, bool) {
	s, ok := r.Context().Value(ctxKey{}).(Sesi)
	return s, ok
}

// Login: hanya pengguna yang sudah login.
func Login(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(namaCookie)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		mu.Lock()
		s, ok := sesi[c.Value]
		mu.Unlock()
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, s)))
	}
}

// Peran: hanya peran tertentu (ADMIN / KASIR).
func Peran(peran string, next http.HandlerFunc) http.HandlerFunc {
	return Login(func(w http.ResponseWriter, r *http.Request) {
		if s, _ := Ambil(r); s.Peran != peran {
			http.Error(w, "Akses ditolak", http.StatusForbidden)
			return
		}
		next(w, r)
	})
}

func Admin(next http.HandlerFunc) http.HandlerFunc { return Peran("ADMIN", next) }
func Kasir(next http.HandlerFunc) http.HandlerFunc { return Peran("KASIR", next) }
