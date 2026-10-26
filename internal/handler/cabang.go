package handler

// CONTOH LENGKAP (F-01) — handler lain mengikuti pola ini.

import (
	"fmt"
	"net/http"
	"strings"

	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type CabangForm struct {
	Action string
	Cabang model.Cabang
}

type CabangHandler struct {
	Svc  *service.CabangService
	View *Renderer
}

func (h *CabangHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	daftar, err := h.Svc.Daftar(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	h.View.Tampil(w, "cabang/list.html", halaman(r, "Cabang", daftar))
}

func (h *CabangHandler) FormTambah(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, "cabang/form.html", halaman(r, "Tambah Cabang",
		CabangForm{Action: "/cabang", Cabang: model.Cabang{Aktif: true}}))
}

func (h *CabangHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	c := cabangDariForm(r)
	if _, err := h.Svc.Buat(r.Context(), c); err != nil {
		h.formGagal(w, r, err, "Tambah Cabang", CabangForm{Action: "/cabang", Cabang: c})
		return
	}
	http.Redirect(w, r, "/cabang?pesan=Cabang+ditambahkan", http.StatusSeeOther)
}

func (h *CabangHandler) FormEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	c, err := h.Svc.Detail(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.View.Tampil(w, "cabang/form.html", halaman(r, "Ubah Cabang",
		CabangForm{Action: fmt.Sprintf("/cabang/%d/edit", id), Cabang: c}))
}

func (h *CabangHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	c := cabangDariForm(r)
	c.ID = id
	if err := h.Svc.Ubah(r.Context(), c); err != nil {
		h.formGagal(w, r, err, "Ubah Cabang", CabangForm{Action: fmt.Sprintf("/cabang/%d/edit", id), Cabang: c})
		return
	}
	http.Redirect(w, r, "/cabang?pesan=Cabang+diperbarui", http.StatusSeeOther)
}

func (h *CabangHandler) Hapus(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	if err := h.Svc.Hapus(r.Context(), id); err != nil {
		gagal(w, err)
		return
	}
	http.Redirect(w, r, "/cabang?pesan=Cabang+dihapus", http.StatusSeeOther)
}

// formGagal: salah input -> tampilkan ulang form beserta pesan (422), selain itu gagal().
func (h *CabangHandler) formGagal(w http.ResponseWriter, r *http.Request, err error, judul string, f CabangForm) {
	pesan, ok := pesanValidasi(err)
	if !ok {
		gagal(w, err)
		return
	}
	hal := halaman(r, judul, f)
	hal.Error = pesan
	w.WriteHeader(http.StatusUnprocessableEntity)
	h.View.Tampil(w, "cabang/form.html", hal)
}

func cabangDariForm(r *http.Request) model.Cabang {
	return model.Cabang{
		Nama:   strings.TrimSpace(r.PostFormValue("nama")),
		Alamat: strings.TrimSpace(r.PostFormValue("alamat")),
		Aktif:  r.PostFormValue("aktif") == "on",
	}
}
