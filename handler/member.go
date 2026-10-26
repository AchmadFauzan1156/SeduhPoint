package handler

import "net/http"

// F-03, F-04, F-27 — TODO M2
func (h *Handler) MemberList(w http.ResponseWriter, r *http.Request) {
	// q := r.URL.Query().Get("q") -> Store.CariMember(q)
	h.render(w, r, "member/list", "Member", nil)
}
func (h *Handler) MemberForm(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "member/form", "Form Member", nil)
}
func (h *Handler) MemberSimpan(w http.ResponseWriter, r *http.Request) {
	// validasi: nama wajib, no HP angka & unik
	http.Redirect(w, r, "/member", http.StatusSeeOther)
}
func (h *Handler) MemberDetail(w http.ResponseWriter, r *http.Request) {
	// saldo poin aktif, kedaluwarsa terdekat, riwayat mutasi & tingkat (F-27)
	h.render(w, r, "member/detail", "Detail Member", nil)
}
