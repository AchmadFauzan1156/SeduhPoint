package middleware

import (
	"net/http"
	"strings"
)

const (
	maksForm = 64 << 10 // 64 KB, cukup untuk form teks
	maksFoto = 3 << 20  // form produk membawa foto (maks 2 MB) + isian lain
)

// BatasiBody membatasi ukuran body request dengan http.MaxBytesReader (materi #4).
// Di materi ditulis di dalam handler; di sini dijadikan middleware (materi #10) karena
// pengecekannya sama untuk semua form, jadi tidak perlu ditulis ulang di setiap handler.
func BatasiBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batas := int64(maksForm)
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			batas = maksFoto
		}
		r.Body = http.MaxBytesReader(w, r.Body, batas)
		next.ServeHTTP(w, r)
	})
}
