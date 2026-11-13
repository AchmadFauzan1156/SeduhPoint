package handler

// Member: daftar + cari, panel detail (saldo & lot poin) atau panel form.

import (
	"fmt"
	"net/http"

	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type MemberHandler struct {
	Svc  *service.MemberService
	View *Renderer
}

type MemberPage struct {
	Daftar []model.Member
	Q      string
	Total  int
	Gold   int
	Mode   string // "" | "detail" | "form"
	Detail service.DetailMember
	Action string
	Form   model.Member
}

func (h *MemberHandler) tampil(w http.ResponseWriter, r *http.Request, judul string, pg MemberPage, err error) {
	pg.Q = r.URL.Query().Get("q")
	var e error
	if pg.Daftar, e = h.Svc.Cari(r.Context(), pg.Q); e != nil {
		gagal(w, e)
		return
	}
	if pg.Total, pg.Gold, e = h.Svc.Hitung(r.Context()); e != nil {
		gagal(w, e)
		return
	}
	hal := halaman(r, judul, "member", pg)
	if err != nil {
		formGagal(w, h.View, "member/list.html", hal, err)
		return
	}
	h.View.Tampil(w, "member/list.html", hal)
}

func (h *MemberHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.tampil(w, r, "Member", MemberPage{}, nil)
}

func (h *MemberHandler) FormTambah(w http.ResponseWriter, r *http.Request) {
	h.tampil(w, r, "Daftar member", MemberPage{Mode: "form", Action: "/member"}, nil)
}

func (h *MemberHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	m := memberDariForm(r)
	id, err := h.Svc.Buat(r.Context(), m)
	if err != nil {
		h.tampil(w, r, "Daftar member", MemberPage{Mode: "form", Action: "/member", Form: m}, err)
		return
	}
	redirect(w, r, fmt.Sprintf("/member/%d", id), "member-dibuat")
}

func (h *MemberHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	d, err := h.Svc.DetailLengkap(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.tampil(w, r, d.Member.Nama, MemberPage{Mode: "detail", Detail: d}, nil)
}

func (h *MemberHandler) FormEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	m, err := h.Svc.Detail(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.tampil(w, r, "Ubah "+m.Nama, MemberPage{Mode: "form", Action: fmt.Sprintf("/member/%d/edit", id), Form: m}, nil)
}

func (h *MemberHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	m := memberDariForm(r)
	m.ID = id
	if err := h.Svc.Ubah(r.Context(), m); err != nil {
		h.tampil(w, r, "Ubah member", MemberPage{Mode: "form", Action: fmt.Sprintf("/member/%d/edit", id), Form: m}, err)
		return
	}
	redirect(w, r, fmt.Sprintf("/member/%d", id), "member-diubah")
}

// memberDariForm hanya membaca field yang boleh diisi dari form (materi #10).
func memberDariForm(r *http.Request) model.Member {
	return model.Member{Nama: r.PostFormValue("nama"), NoHP: r.PostFormValue("no_hp")}
}
