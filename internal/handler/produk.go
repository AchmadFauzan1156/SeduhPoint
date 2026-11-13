package handler

// Kategori & Produk (pola sama dengan cabang.go) + unggah foto produk

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"seduhpoint/internal/config"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type ProdukHandler struct {
	Svc      *service.ProdukService
	PromoSvc *service.PromoService
	View     *Renderer
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
		formGagal(w, r, h.View, "kategori/list.html", hal, err)
		return
	}
	h.View.Tampil(w, r, "kategori/list.html", hal)
}

func (h *ProdukHandler) DaftarKategori(w http.ResponseWriter, r *http.Request) {
	h.tampilKategori(w, r, "/kategori", model.Kategori{}, nil)
}

func (h *ProdukHandler) SimpanKategori(w http.ResponseWriter, r *http.Request) {
	k := model.Kategori{Nama: r.PostFormValue("nama")}
	if err := h.Svc.BuatKategori(r.Context(), k); err != nil {
		h.tampilKategori(w, r, "/kategori", k, err)
		return
	}
	redirect(w, r, "/kategori", "kategori-dibuat")
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
	if err := h.Svc.UbahKategori(r.Context(), k); err != nil {
		h.tampilKategori(w, r, fmt.Sprintf("/kategori/%d/edit", id), k, err)
		return
	}
	redirect(w, r, "/kategori", "kategori-diubah")
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
	redirect(w, r, "/kategori", "kategori-dihapus")
}

// ---------- Produk ----------

type ProdukPage struct {
	Daftar   []model.Produk
	Kategori []model.Kategori
	Action   string
	Form     model.Produk
	Detail   map[int]DetailProduk
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
	// Promo yang masih aktif (belum berakhir) untuk kartu detail produk
	var berjalan []model.Promo
	if semua, e := h.PromoSvc.Daftar(r.Context()); e == nil {
		hariIni := time.Now().In(config.WIB).Format("2006-01-02")
		for _, pm := range semua {
			if pm.Aktif && pm.TglSelesai.Format("2006-01-02") >= hariIni {
				berjalan = append(berjalan, pm)
			}
		}
	}
	hal := halaman(r, judul, "produk", ProdukPage{Daftar: daftar, Kategori: kat, Action: action, Form: p,
		Detail: detailProduk(daftar, berjalan, true, false)})
	if err != nil {
		formGagal(w, r, h.View, "produk/list.html", hal, err)
		return
	}
	h.View.Tampil(w, r, "produk/list.html", hal)
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
		h.tampil(w, r, "Produk", action, model.Produk{ID: id}, service.SalahInput("gambar", "Foto terlalu besar (maks 2 MB)"))
		return
	}
	p := produkDariForm(r)
	p.ID = id
	// Foto lama diambil dari database, bukan dari hidden input, supaya path gambar
	// tidak bisa diubah lewat form (mass assignment, materi #12).
	if id > 0 {
		lama, err := h.Svc.Detail(r.Context(), id)
		if err != nil {
			gagal(w, err)
			return
		}
		p.Gambar = lama.Gambar
	}
	path, err := simpanFoto(r, "gambar")
	if err != nil {
		h.tampil(w, r, "Produk", action, p, err)
		return
	}
	if path != "" {
		p.Gambar = path
	}
	kode := "produk-dibuat"
	if id == 0 {
		_, err = h.Svc.Buat(r.Context(), p)
	} else {
		err, kode = h.Svc.Ubah(r.Context(), p), "produk-diubah"
	}
	if err != nil {
		h.tampil(w, r, "Produk", action, p, err)
		return
	}
	redirect(w, r, "/produk", kode)
}

// produkDariForm hanya membaca field yang memang boleh diisi dari form (materi #10 & #12).
func produkDariForm(r *http.Request) model.Produk {
	return model.Produk{
		Nama:       r.PostFormValue("nama"),
		KategoriID: atoi(r.PostFormValue("kategori_id")),
		Harga:      angka(r.PostFormValue("harga")),
		Deskripsi:  r.PostFormValue("deskripsi"),
		Aktif:      r.PostFormValue("aktif") == "on",
	}
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
	redirect(w, r, "/produk", "produk-dihapus")
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
		return "", service.SalahInput(field, "Foto maksimal 2 MB")
	}
	kepala := make([]byte, 512)
	n, _ := io.ReadFull(f, kepala)
	ext, ok := ekstensiFoto[http.DetectContentType(kepala[:n])]
	if !ok {
		return "", service.SalahInput(field, "Foto harus JPG, PNG, atau WebP")
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
