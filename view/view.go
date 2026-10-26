// Package view memuat html/template: layout + partials + satu halaman (materi #5).
package view

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type View struct{ pages map[string]*template.Template }

var funcs = template.FuncMap{
	// rupiah: 25000 -> "Rp25.000"
	"rupiah": func(n int64) string {
		s := fmt.Sprint(n)
		neg := strings.HasPrefix(s, "-")
		s = strings.TrimPrefix(s, "-")
		var b strings.Builder
		for i, c := range s {
			if i > 0 && (len(s)-i)%3 == 0 {
				b.WriteByte('.')
			}
			b.WriteRune(c)
		}
		if neg {
			return "-Rp" + b.String()
		}
		return "Rp" + b.String()
	},
	"tgl": func(t time.Time) string { return t.Format("02 Jan 2006 15:04") },
}

// New mem-parse semua halaman saat program start (template.Must -> gagal cepat).
func New(dir string) *View {
	base := []string{filepath.Join(dir, "layout.html")}
	partials, _ := filepath.Glob(filepath.Join(dir, "partials", "*.html"))
	base = append(base, partials...)

	pages, _ := filepath.Glob(filepath.Join(dir, "*", "*.html"))
	root, _ := filepath.Glob(filepath.Join(dir, "*.html"))
	pages = append(pages, root...)

	v := &View{pages: map[string]*template.Template{}}
	for _, p := range pages {
		name := strings.TrimSuffix(filepath.ToSlash(strings.TrimPrefix(p, dir+string(filepath.Separator))), ".html")
		if name == "layout" || strings.HasPrefix(name, "partials/") {
			continue
		}
		files := append(append([]string{}, base...), p)
		v.pages[name] = template.Must(template.New("layout.html").Funcs(funcs).ParseFiles(files...))
	}
	return v
}

// Render menampilkan halaman, mis. Render(w, "cabang/list", data).
func (v *View) Render(w http.ResponseWriter, name string, data any) {
	t, ok := v.pages[name]
	if !ok {
		http.Error(w, "template tidak ditemukan: "+name, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Println("render:", err)
	}
}
