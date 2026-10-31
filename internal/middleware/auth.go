package middleware

// Session sederhana yang disimpan di memori server.

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Sesi struct {
	PenggunaID int
	Username   string
	Peran      string // ADMIN | KASIR
	CabangID   int
	CSRF       string // token anti-CSRF, wajib dikirim di setiap form POST
	habis      time.Time
}

var (
	mu   sync.Mutex // lindungi map dari data race
	sesi = map[string]Sesi{}
)

const (
	namaCookie = "seduh_sesi"
	LamaSesi   = 8 * time.Hour // satu shift kerja; setelah itu harus login ulang
	maksForm   = 3 << 20       // batas memori form multipart (foto produk maks 2 MB)
)

type ctxKey struct{}

func acak(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func BuatSesi(w http.ResponseWriter, s Sesi) {
	id := acak(32)
	s.CSRF = acak(32)
	s.habis = time.Now().Add(LamaSesi)
	mu.Lock()
	for lama, x := range sesi { // bersihkan sesi kedaluwarsa agar map tidak terus membesar
		if time.Now().After(x.habis) {
			delete(sesi, lama)
		}
	}
	sesi[id] = s
	mu.Unlock()
	http.SetCookie(w, &http.Cookie{Name: namaCookie, Value: id, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: int(LamaSesi.Seconds())})
}

func HapusSesi(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(namaCookie); err == nil {
		mu.Lock()
		delete(sesi, c.Value)
		mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: namaCookie, Value: "", Path: "/", MaxAge: -1})
}

// HapusSesiLain: keluarkan pengguna dari semua perangkat lain (dipakai setelah ganti password).
func HapusSesiLain(r *http.Request, penggunaID int) {
	c, err := r.Cookie(namaCookie)
	if err != nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	for id, s := range sesi {
		if s.PenggunaID == penggunaID && id != c.Value {
			delete(sesi, id)
		}
	}
}

// HapusSesiPengguna: keluarkan pengguna dari semua perangkat (dipakai setelah admin me-reset password kasir).
func HapusSesiPengguna(penggunaID int) {
	mu.Lock()
	defer mu.Unlock()
	for id, s := range sesi {
		if s.PenggunaID == penggunaID {
			delete(sesi, id)
		}
	}
}

// Ambil mengembalikan sesi pengguna yang sedang login.
func Ambil(r *http.Request) (Sesi, bool) {
	s, ok := r.Context().Value(ctxKey{}).(Sesi)
	return s, ok
}

// cari: sesi dari cookie; sesi yang sudah lewat LamaSesi dihapus.
func cari(r *http.Request) (Sesi, bool) {
	c, err := r.Cookie(namaCookie)
	if err != nil {
		return Sesi{}, false
	}
	mu.Lock()
	defer mu.Unlock()
	s, ok := sesi[c.Value]
	if ok && time.Now().After(s.habis) {
		delete(sesi, c.Value)
		return Sesi{}, false
	}
	return s, ok
}

// tokenCocok: form POST harus membawa token CSRF milik sesi.
func tokenCocok(r *http.Request, s Sesi) bool {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(maksForm); err != nil {
			return false
		}
	}
	return subtle.ConstantTimeCompare([]byte(r.PostFormValue("csrf")), []byte(s.CSRF)) == 1
}

// Login: hanya pengguna yang sudah login. Setiap POST juga dicek token CSRF-nya.
func Login(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := cari(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if r.Method == http.MethodPost && !tokenCocok(r, s) {
			http.Error(w, "Sesi formulir tidak valid, muat ulang halaman lalu coba lagi", http.StatusForbidden)
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
