// Package middleware: logging, session login, dan pembatasan peran (NF-03).
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"
	"time"
)

type Sesi struct {
	PenggunaID int
	Username   string
	Peran      string // ADMIN | KASIR
	CabangID   int
}

// Session disimpan di memori dan dilindungi Mutex (materi #4: data race).
var (
	mu   sync.Mutex
	sesi = map[string]Sesi{}
)

const cookieName = "seduh_sesi"

type ctxKey struct{}

func BuatSesi(w http.ResponseWriter, s Sesi) {
	b := make([]byte, 32)
	rand.Read(b)
	id := hex.EncodeToString(b)
	mu.Lock()
	sesi[id] = s
	mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: id, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
}

func HapusSesi(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil {
		mu.Lock()
		delete(sesi, c.Value)
		mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1})
}

// Ambil mengembalikan sesi pengguna yang sedang login.
func Ambil(r *http.Request) (Sesi, bool) {
	s, ok := r.Context().Value(ctxKey{}).(Sesi)
	return s, ok
}

// Login mewajibkan pengguna sudah login.
func Login(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieName)
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

// Peran membatasi handler hanya untuk peran tertentu (dipakai setelah Login).
func Peran(peran string, next http.HandlerFunc) http.HandlerFunc {
	return Login(func(w http.ResponseWriter, r *http.Request) {
		if s, _ := Ambil(r); s.Peran != peran {
			http.Error(w, "Akses ditolak", http.StatusForbidden)
			return
		}
		next(w, r)
	})
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(t))
	})
}
