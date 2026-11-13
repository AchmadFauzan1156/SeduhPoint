package handler

import (
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Template dibaca relatif dari root project, jadi test pindah ke sana dulu.
func keRootProject(t *testing.T) {
	t.Helper()
	wd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(wd) })
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
}

// Semua template harus ter-parse (template.Must panic jika ada yang salah tulis)
// dan halaman login bisa dirender tanpa pengguna yang login.
func TestTemplateDanHalamanLogin(t *testing.T) {
	keRootProject(t)
	view := NewRenderer()
	if len(view.tpl) < 10 {
		t.Fatalf("template yang ter-parse terlalu sedikit: %d", len(view.tpl))
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)
	view.Tampil(w, r, "login.html", Halaman{Judul: "Login"})
	if w.Code != 200 || !strings.Contains(w.Body.String(), `name="password"`) {
		t.Fatalf("render login gagal: %d %s", w.Code, w.Body.String())
	}
}
