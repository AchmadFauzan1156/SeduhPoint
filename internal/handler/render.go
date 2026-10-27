package handler

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"seduhpoint/internal/config"
	"strconv"
	"strings"
	"time"
)

// rupiah memformat angka jadi "Rp39.000.000"
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

// tglBulan: "Oktober 2026"
func tglBulan(t time.Time) string {
	bulan := [...]string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	return bulan[t.Month()-1] + " " + strconv.Itoa(t.Year())
}

// desimal1: 30,6 (koma desimal Indonesia)
func desimal1(f float64) string {
	return strings.Replace(strconv.FormatFloat(f, 'f', 1, 64), ".", ",", 1)
}

// hariIni: "Selasa, 27 Okt 2026" (WIB); jamSekarang: "12.46", diperbarui tiap menit oleh app.js
func hariIni() string {
	hari := [...]string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	t := time.Now().In(config.WIB)
	return hari[t.Weekday()] + ", " + tglPendek(t)
}

func jamSekarang() string { return time.Now().In(config.WIB).Format("15.04") }

// dict: pasangan kunci-nilai untuk memanggil template bagian dengan beberapa argumen.
func dict(kv ...any) map[string]any {
	m := make(map[string]any, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return m
}

func toInt(f float64) int       { return int(f) }
func mod(a, b int) int          { return a % b }
func sisa100(f float64) float64 { return 100 - f }

// add25: donat dimulai dari arah jam 12 (offset 25 dari keliling 100).
func add25(f float64) float64 { return f + 25 }

func bulanPendek(t time.Time) string {
	return [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}[t.Month()-1]
}

// rupiahPendek: 190 rb, 1,2 jt
func rupiahPendek(n int64) string {
	switch {
	case n >= 1000000:
		return strings.Replace(strconv.FormatFloat(float64(n)/1000000, 'f', 1, 64), ".", ",", 1) + " jt"
	case n >= 1000:
		return strconv.FormatInt(n/1000, 10) + " rb"
	}
	return strconv.FormatInt(n, 10)
}

func add(a, b int) int { return a + b }

// nilaiPoin: poin × Rp1.000
func nilaiPoin(p int) int64 { return int64(p) * 1000 }

// kali: harga × qty untuk subtotal baris
func kali(h int64, q int) int64 { return h * int64(q) }

var funcs = template.FuncMap{"rentang": rentang, "kali": kali, "rupiah": rupiah, "add_one": addOne, "tgl": tgl, "upper": strings.ToUpper,
	"has": has, "tglBulan": tglBulan, "add": add, "dict": dict, "int": toInt, "mod": mod, "sisa100": sisa100, "add25": add25, "bulanPendek": bulanPendek, "rupiahPendek": rupiahPendek, "hariIni": hariIni, "jamSekarang": jamSekarang, "desimal1": desimal1, "nilaiPoin": nilaiPoin, "jam": jam, "tglInput": tglInput, "tglPendek": tglPendek}

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
