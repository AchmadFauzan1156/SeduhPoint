package handler

// Pola handler:
// daftar + panel form di halaman yang sama, validasi di service, Post/Redirect/Get.

import (
	"fmt"
	"net/http"

	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type CabangForm struct {
	Action   string
	Cabang   model.Cabang
	Kasir    string // username kasir cabang ini; kosong = belum punya (form menampilkan isian akun kasir)
	Username string // isian username kasir yang ditampilkan ulang saat validasi gagal (password tidak)
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
		formGagal(w, r, h.View, "cabang/list.html", hal, err)
		return
	}
	h.View.Tampil(w, r, "cabang/list.html", hal)
}

func (h *CabangHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.tampil(w, r, "Cabang", CabangForm{Action: "/cabang", Cabang: model.Cabang{Aktif: true}}, nil)
}

func (h *CabangHandler) FormTambah(w http.ResponseWriter, r *http.Request) { h.Daftar(w, r) }

func (h *CabangHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	c, akun := cabangDariForm(r), akunDariForm(r)
	if _, err := h.Svc.Buat(r.Context(), c, akun); err != nil {
		h.tampil(w, r, "Cabang", CabangForm{Action: "/cabang", Cabang: c, Username: akun.Username}, err)
		return
	}
	catatAksi(r, "menambah cabang %s dengan akun kasir %s", c.Nama, akun.Username)
	redirect(w, r, "/cabang", "cabang-dibuat")
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
	kasir, err := h.Svc.KasirCabang(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.tampil(w, r, "Ubah cabang", CabangForm{Action: fmt.Sprintf("/cabang/%d/edit", id), Cabang: c, Kasir: kasir}, nil)
}

func (h *CabangHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	c, akun := cabangDariForm(r), akunDariForm(r)
	c.ID = id
	if err := h.Svc.Ubah(r.Context(), c, akun); err != nil {
		kasir, _ := h.Svc.KasirCabang(r.Context(), id)
		h.tampil(w, r, "Ubah cabang", CabangForm{Action: fmt.Sprintf("/cabang/%d/edit", id), Cabang: c,
			Kasir: kasir, Username: akun.Username}, err)
		return
	}
	if !akun.Kosong() {
		catatAksi(r, "membuat akun kasir %s untuk cabang %d", akun.Username, id)
	}
	redirect(w, r, "/cabang", "cabang-diubah")
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
	catatAksi(r, "menghapus cabang %d", id)
	redirect(w, r, "/cabang", "cabang-dihapus")
}

func cabangDariForm(r *http.Request) model.Cabang {
	return model.Cabang{
		Nama:   r.PostFormValue("nama"),
		Alamat: r.PostFormValue("alamat"),
		Aktif:  r.PostFormValue("aktif") == "on",
	}
}

// akunDariForm: isian akun kasir pada form cabang.
func akunDariForm(r *http.Request) service.AkunKasir {
	return service.AkunKasir{
		Username: r.PostFormValue("username"),
		Password: r.PostFormValue("kasir_baru"),
		Ulang:    r.PostFormValue("kasir_ulang"),
	}
}
