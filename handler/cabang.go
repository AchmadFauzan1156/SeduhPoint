package handler

import (
	"net/http"
	"strings"

	"seduhpoint/models"
)

// CONTOH LENGKAP: list, form, validasi server, Post/Redirect/Get (materi #4). F-01

func (h *Handler) CabangList(w http.ResponseWriter, r *http.Request) {
	list, err := h.Store.ListCabang()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.render(w, r, "cabang/list", "Cabang", list)
}

func (h *Handler) CabangForm(w http.ResponseWriter, r *http.Request) {
	c := models.Cabang{Aktif: true}
	if id := atoi(r.PathValue("id")); id > 0 {
		var err error
		if c, err = h.Store.GetCabang(id); err != nil {
			http.NotFound(w, r)
			return
		}
	}
	h.render(w, r, "cabang/form", "Form Cabang", c)
}

func (h *Handler) CabangSimpan(w http.ResponseWriter, r *http.Request) {
	c := models.Cabang{
		ID:     atoi(r.PostFormValue("id")),
		Nama:   strings.TrimSpace(r.PostFormValue("nama")),
		Alamat: strings.TrimSpace(r.PostFormValue("alamat")),
		Aktif:  r.PostFormValue("aktif") == "on",
	}
	errs := map[string]string{}
	if c.Nama == "" {
		errs["nama"] = "Nama cabang wajib diisi"
	}
	if len(errs) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		h.View.Render(w, "cabang/form", Page{Judul: "Form Cabang", Errors: errs, Data: c})
		return
	}
	if err := h.Store.SimpanCabang(c); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/cabang?pesan=Cabang+tersimpan", http.StatusSeeOther)
}
