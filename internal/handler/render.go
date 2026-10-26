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

func tgl(t time.Time) string { return t.Format("02 Jan 2006 15:04") }

var funcs = template.FuncMap{"rupiah": rupiah, "add_one": addOne, "tgl": tgl}

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
