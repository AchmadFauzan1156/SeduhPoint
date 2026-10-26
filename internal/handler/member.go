package handler

// F-03, F-04, F-27 — TODO M2

import (
	"net/http"

	"seduhpoint/internal/service"
)

type MemberHandler struct {
	Svc  *service.MemberService
	View *Renderer
}

func (h *MemberHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	// q := r.URL.Query().Get("q") -> h.Svc.Cari
	h.View.Tampil(w, "member/list.html", halaman(r, "Member", nil))
}
func (h *MemberHandler) FormTambah(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "member/form.html", halaman(r, "Daftar Member", nil))
}
func (h *MemberHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/member", http.StatusSeeOther)
}
func (h *MemberHandler) Detail(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "member/detail.html", halaman(r, "Detail Member", nil))
}
func (h *MemberHandler) FormEdit(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "member/form.html", halaman(r, "Ubah Member", nil))
}
func (h *MemberHandler) Update(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/member", http.StatusSeeOther)
}
