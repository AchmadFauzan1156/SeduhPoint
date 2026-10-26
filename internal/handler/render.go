package handler

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// rupiah memformat angka jadi "Rp39.000.000" (sama dengan materi #10, versi int64)
func rupiah(n int64) string {
	s := strconv.FormatInt(n, 10)
	hasil := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 && s[i-1] != '-' {
			hasil += "."
		}
		hasil += string(c)
	}
	return "Rp" + hasil
}

func addOne(n int) int { return n + 1 }

// tgl: "26 Okt 2026, 22.35"
func tgl(t time.Time) string { return tglPendek(t) + ", " + t.Format("15.04") }

// has: apakah id ada di daftar (untuk checkbox terpilih)
func has(daftar []int, id int) bool {
	for _, x := range daftar {
		if x == id {
			return true
		}
	}
	return false
}

// jam: "14:00:00" -> "14:00"
func jam(s string) string {
	if len(s) >= 5 {
		return s[:5]
	}
	return s
}

func tglInput(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func tglPendek(t time.Time) string {
	bulan := [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	return strconv.Itoa(t.Day()) + " " + bulan[t.Month()-1] + " " + strconv.Itoa(t.Year())
}

// rentang: "1 Okt – 31 Des 2026" (tahun hanya sekali jika sama)
func rentang(a, b time.Time) string {
	bulan := [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	kiri := strconv.Itoa(a.Day()) + " " + bulan[a.Month()-1]
	if a.Year() != b.Year() {
		kiri += " " + strconv.Itoa(a.Year())
	}
	return kiri + " – " + tglPendek(b)
}

// kali: harga × qty untuk subtotal baris
func kali(h int64, q int) int64 { return h * int64(q) }

var funcs = template.FuncMap{"rentang": rentang, "kali": kali, "rupiah": rupiah, "add_one": addOne, "tgl": tgl, "upper": strings.ToUpper,
	"has": has, "jam": jam, "tglInput": tglInput, "tglPendek": tglPendek}

type Renderer struct {
	tpl map[string]*template.Template
}

// NewRenderer mem-parse semua halaman di templates/ (kecuali layout & potongannya)
// saat program start. Kunci map = path relatif, mis. "cabang/list.html".
func NewRenderer() *Renderer {
	dasar := []string{"templates/layout.html", "templates/header.html", "templates/sidebar.html", "templates/footer.html"}
	halaman, _ := filepath.Glob("templates/*.html")
	sub, _ := filepath.Glob("templates/*/*.html")
	halaman = append(halaman, sub...)

	rd := &Renderer{tpl: make(map[string]*template.Template)}
	for _, f := range halaman {
		kunci := strings.TrimPrefix(filepath.ToSlash(f), "templates/")
		switch kunci {
		case "layout.html", "header.html", "sidebar.html", "footer.html":
			continue
		}
		rd.tpl[kunci] = template.Must(template.New("layout").Funcs(funcs).ParseFiles(append(dasar, f)...))
	}
	return rd
}

// Tampil me-render ke buffer dulu supaya error template tidak menghasilkan halaman setengah jadi.
func (rd *Renderer) Tampil(w http.ResponseWriter, file string, data any) {
	t, ok := rd.tpl[file]
	if !ok {
		log.Printf("template %s tidak ada", file)
		http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		log.Printf("render %s: %v", file, err)
		http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
