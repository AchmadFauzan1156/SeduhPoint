package handler

// F-02 Kategori & Produk (pola sama dengan cabang.go) + unggah foto produk

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type ProdukHandler struct {
	Svc  *service.ProdukService
	View *Renderer
}

// ---------- Kategori ----------

type KategoriPage struct {
	Daftar []model.Kategori
	Action string
	Form   model.Kategori
}

func (h *ProdukHandler) tampilKategori(w http.ResponseWriter, r *http.Request, action string, k model.Kategori, err error) {
	daftar, e := h.Svc.DaftarKategori(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	hal := halaman(r, "Kategori", "kategori", KategoriPage{Daftar: daftar, Action: action, Form: k})
	if err != nil {
		formGagal(w, h.View, "kategori/list.html", hal, err)
		return
	}
	h.View.Tampil(w, "kategori/list.html", hal)
}

func (h *ProdukHandler) DaftarKategori(w http.ResponseWriter, r *http.Request) {
	h.tampilKategori(w, r, "/kategori", model.Kategori{}, nil)
}

func (h *ProdukHandler) SimpanKategori(w http.ResponseWriter, r *http.Request) {
	k := model.Kategori{Nama: r.PostFormValue("nama")}
	if err := h.Svc.SimpanKategori(r.Context(), k); err != nil {
		h.tampilKategori(w, r, "/kategori", k, err)
		return
	}
	redirect(w, r, "/kategori", "Kategori ditambahkan")
}

func (h *ProdukHandler) FormEditKategori(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	k, err := h.Svc.DetailKategori(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.tampilKategori(w, r, fmt.Sprintf("/kategori/%d/edit", id), k, nil)
}

func (h *ProdukHandler) UpdateKategori(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	k := model.Kategori{ID: id, Nama: r.PostFormValue("nama")}
	if err := h.Svc.SimpanKategori(r.Context(), k); err != nil {
		h.tampilKategori(w, r, fmt.Sprintf("/kategori/%d/edit", id), k, err)
		return
	}
	redirect(w, r, "/kategori", "Kategori disimpan")
}

func (h *ProdukHandler) HapusKategori(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	if err := h.Svc.HapusKategori(r.Context(), id); err != nil {
		h.tampilKategori(w, r, "/kategori", model.Kategori{}, err)
		return
	}
	redirect(w, r, "/kategori", "Kategori dihapus")
}

// ---------- Produk ----------

type ProdukPage struct {
	Daftar   []model.Produk
	Kategori []model.Kategori
	Action   string
	Form     model.Produk
}

func (h *ProdukHandler) tampil(w http.ResponseWriter, r *http.Request, judul, action string, p model.Produk, err error) {
	daftar, e := h.Svc.Daftar(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	kat, e := h.Svc.DaftarKategori(r.Context())
	if e != nil {
		gagal(w, e)
		return
	}
	hal := halaman(r, judul, "produk", ProdukPage{Daftar: daftar, Kategori: kat, Action: action, Form: p})
	if err != nil {
		formGagal(w, h.View, "produk/list.html", hal, err)
		return
	}
	h.View.Tampil(w, "produk/list.html", hal)
}

func (h *ProdukHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.tampil(w, r, "Produk", "/produk", model.Produk{Aktif: true}, nil)
}

func (h *ProdukHandler) FormTambah(w http.ResponseWriter, r *http.Request) { h.Daftar(w, r) }

func (h *ProdukHandler) FormEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	p, err := h.Svc.Detail(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.tampil(w, r, "Ubah "+p.Nama, fmt.Sprintf("/produk/%d/edit", id), p, nil)
}

func (h *ProdukHandler) Simpan(w http.ResponseWriter, r *http.Request) { h.simpan(w, r, 0) }

func (h *ProdukHandler) Update(w http.ResponseWriter, r *http.Request) {
	if id, ok := idDariPath(w, r); ok {
		h.simpan(w, r, id)
	}
}

func (h *ProdukHandler) simpan(w http.ResponseWriter, r *http.Request, id int) {
	action := "/produk"
	if id > 0 {
		action = fmt.Sprintf("/produk/%d/edit", id)
	}
	if err := r.ParseMultipartForm(maksFoto + 1<<20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		h.tampil(w, r, "Produk", action, model.Produk{ID: id}, &service.ValidasiError{Field: "gambar", Pesan: "Foto terlalu besar (maks 2 MB)"})
		return
	}
	p := model.Produk{
		ID:         id,
		Nama:       r.PostFormValue("nama"),
		KategoriID: atoi(r.PostFormValue("kategori_id")),
		Harga:      angka(r.PostFormValue("harga")),
		Gambar:     r.PostFormValue("gambar_lama"),
		Aktif:      r.PostFormValue("aktif") == "on",
	}
	path, err := simpanFoto(r, "gambar")
	if err != nil {
		h.tampil(w, r, "Produk", action, p, err)
		return
	}
	if path != "" {
		p.Gambar = path
	}
	if _, err := h.Svc.Simpan(r.Context(), p); err != nil {
		h.tampil(w, r, "Produk", action, p, err)
		return
	}
	redirect(w, r, "/produk", p.Nama+" disimpan")
}

func (h *ProdukHandler) Hapus(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	if err := h.Svc.Hapus(r.Context(), id); err != nil {
		gagal(w, err)
		return
	}
	redirect(w, r, "/produk", "Produk dihapus")
}

const maksFoto = 2 << 20 // 2 MB

var ekstensiFoto = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

// simpanFoto menyimpan file unggahan ke static/img/upload dan mengembalikan path URL-nya.
// Tidak ada file = "" tanpa error. Jenis file dicek dari isinya, bukan dari nama file.
func simpanFoto(r *http.Request, field string) (string, error) {
	f, hdr, err := r.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) || errors.Is(err, http.ErrNotMultipart) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	defer f.Close()
	if hdr.Size > maksFoto {
		return "", &service.ValidasiError{Field: field, Pesan: "Foto maksimal 2 MB"}
	}
	kepala := make([]byte, 512)
	n, _ := io.ReadFull(f, kepala)
	ext, ok := ekstensiFoto[http.DetectContentType(kepala[:n])]
	if !ok {
		return "", &service.ValidasiError{Field: field, Pesan: "Foto harus JPG, PNG, atau WebP"}
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	acak := make([]byte, 8)
	rand.Read(acak)
	nama := hex.EncodeToString(acak) + ext
	dst, err := os.Create(filepath.Join("static", "img", "upload", nama))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, f); err != nil {
		return "", err
	}
	return "/static/img/upload/" + nama, nil
}
