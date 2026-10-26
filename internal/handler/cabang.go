package handler

// CONTOH LENGKAP (F-01) — handler lain mengikuti pola ini:
// daftar + panel form di halaman yang sama, validasi di service, Post/Redirect/Get.

import (
	"fmt"
	"net/http"

	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type CabangForm struct {
	Action string
	Cabang model.Cabang
}

type CabangPage struct {
	Daftar []model.Cabang
	Form   CabangForm
}

type CabangHandler struct {
	Svc  *service.CabangService
	View *Renderer
}

func (h *CabangHandler) tampil(w http.ResponseWriter, r *http.Request, judul string, f CabangForm, err error) {
	daftar, e := h.Svc.Daftar(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	hal := halaman(r, judul, "cabang", CabangPage{Daftar: daftar, Form: f})
	if err != nil {
		formGagal(w, h.View, "cabang/list.html", hal, err)
		return
	}
	h.View.Tampil(w, "cabang/list.html", hal)
}

func (h *CabangHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.tampil(w, r, "Cabang", CabangForm{Action: "/cabang", Cabang: model.Cabang{Aktif: true}}, nil)
}

func (h *CabangHandler) FormTambah(w http.ResponseWriter, r *http.Request) { h.Daftar(w, r) }

func (h *CabangHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	c := cabangDariForm(r)
	if _, err := h.Svc.Buat(r.Context(), c); err != nil {
		h.tampil(w, r, "Cabang", CabangForm{Action: "/cabang", Cabang: c}, err)
		return
	}
	redirect(w, r, "/cabang", "Cabang "+c.Nama+" ditambahkan")
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
	h.tampil(w, r, "Ubah cabang", CabangForm{Action: fmt.Sprintf("/cabang/%d/edit", id), Cabang: c}, nil)
}

func (h *CabangHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	c := cabangDariForm(r)
	c.ID = id
	if err := h.Svc.Ubah(r.Context(), c); err != nil {
		h.tampil(w, r, "Ubah cabang", CabangForm{Action: fmt.Sprintf("/cabang/%d/edit", id), Cabang: c}, err)
		return
	}
	redirect(w, r, "/cabang", "Cabang "+c.Nama+" disimpan")
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
	redirect(w, r, "/cabang", "Cabang dihapus")
}

func cabangDariForm(r *http.Request) model.Cabang {
	return model.Cabang{
		Nama:   r.PostFormValue("nama"),
		Alamat: r.PostFormValue("alamat"),
		Aktif:  r.PostFormValue("aktif") == "on",
	}
}
