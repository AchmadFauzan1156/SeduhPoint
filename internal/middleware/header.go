package middleware

import "net/http"

// csp: sumber yang boleh dimuat halaman. Script hanya dari server sendiri (/static/js/app.js),
// font dari Google Fonts, gambar dari server sendiri ditambah blob: untuk pratinjau foto produk.
// style 'unsafe-inline' dibutuhkan karena template memakai atribut style="...".
const csp = "default-src 'self'; script-src 'self'; " +
	"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; " +
	"img-src 'self' data: blob:; form-action 'self'; frame-ancestors 'none'; base-uri 'self'; object-src 'none'"

// SecureHeaders menambahkan header keamanan di setiap respons. Ini tambahan dari materi:
// html/template (materi #5) sudah mencegah XSS dari data, header ini lapisan kedua di browser.
//   - Content-Security-Policy: script dari situs lain atau script sisipan tidak dijalankan
//   - X-Content-Type-Options: browser tidak menebak jenis file (mis. foto unggahan dibaca sebagai HTML)
//   - X-Frame-Options: halaman tidak bisa dimuat di dalam iframe situs lain (clickjacking)
//   - Referrer-Policy: URL lengkap tidak dikirim ke situs lain
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", csp)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}
