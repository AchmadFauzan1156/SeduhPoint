package handler

// Promo & voucher

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type PromoHandler struct {
	Svc       *service.PromoService
	CabangSvc *service.CabangService
	ProdukSvc *service.ProdukService
	View      *Renderer
}

type PromoPage struct {
	Daftar    []model.Promo
	Action    string
	Form      model.Promo
	Cabang    []model.Cabang
	Produk    []model.Produk
	Kategori  []model.Kategori
	NamaJenis map[string]string
	UrutJenis []string
	Dipakai   int // jumlah transaksi yang memakai promo yang sedang diubah
}

func (h *PromoHandler) tampil(w http.ResponseWriter, r *http.Request, judul, action string, p model.Promo, err error) {
	ctx := r.Context()
	pg := PromoPage{Action: action, Form: p, NamaJenis: service.NamaJenis, UrutJenis: service.UrutJenis}
	var e error
	if pg.Daftar, e = h.Svc.Daftar(ctx); e == nil {
		if pg.Cabang, e = h.CabangSvc.Daftar(ctx); e == nil {
			if pg.Produk, e = h.ProdukSvc.Daftar(ctx); e == nil {
				pg.Kategori, e = h.ProdukSvc.DaftarKategori(ctx)
			}
		}
	}
	if e != nil {
		gagal(w, e)
		return
	}
	if p.ID > 0 {
		pg.Dipakai, _ = h.Svc.Dipakai(ctx, p.ID)
	}
	hal := halaman(r, judul, "promo", pg)
	if err != nil {
		formGagal(w, h.View, "promo/list.html", hal, err)
		return
	}
	h.View.Tampil(w, "promo/list.html", hal)
}

func promoBaru() model.Promo {
	hari := time.Now().Truncate(24 * time.Hour)
	return model.Promo{Jenis: service.JenisDiskon, Aktif: true, TglMulai: hari, TglSelesai: hari.AddDate(0, 1, 0)}
}

func (h *PromoHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.tampil(w, r, "Promo & voucher", "/promo", promoBaru(), nil)
}

func (h *PromoHandler) FormTambah(w http.ResponseWriter, r *http.Request) { h.Daftar(w, r) }

func (h *PromoHandler) FormEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	p, err := h.Svc.Detail(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	h.tampil(w, r, "Ubah "+p.Nama, fmt.Sprintf("/promo/%d/edit", id), p, nil)
}

func (h *PromoHandler) Simpan(w http.ResponseWriter, r *http.Request) { h.simpan(w, r, 0) }

func (h *PromoHandler) Update(w http.ResponseWriter, r *http.Request) {
	if id, ok := idDariPath(w, r); ok {
		h.simpan(w, r, id)
	}
}

func (h *PromoHandler) simpan(w http.ResponseWriter, r *http.Request, id int) {
	action := "/promo"
	if id > 0 {
		action = fmt.Sprintf("/promo/%d/edit", id)
	}
	r.ParseForm()
	p := promoDariForm(r)
	p.ID = id
	if id > 0 { // kuota terpakai dibutuhkan validasi kuota
		if lama, err := h.Svc.Detail(r.Context(), id); err == nil {
			p.KuotaTerpakai = lama.KuotaTerpakai
		}
	}
	if _, err := h.Svc.Simpan(r.Context(), p); err != nil {
		h.tampil(w, r, "Promo & voucher", action, p, err)
		return
	}
	redirect(w, r, "/promo", "Promo "+p.Nama+" disimpan")
}

// Toggle mengaktifkan/menonaktifkan promo tanpa menghapus riwayatnya.
func (h *PromoHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	aktif := r.PostFormValue("aktif") == "1"
	if err := h.Svc.SetAktif(r.Context(), id, aktif); err != nil {
		gagal(w, err)
		return
	}
	pesan := "Promo dinonaktifkan"
	if aktif {
		pesan = "Promo diaktifkan"
	}
	redirect(w, r, "/promo", pesan)
}

func nullAngka(s string) sql.NullInt64 {
	if strings.TrimSpace(s) == "" {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: angka(s), Valid: true}
}

func nullTeks(s string) sql.NullString {
	s = strings.TrimSpace(s)
	return sql.NullString{String: s, Valid: s != ""}
}

func tanggal(s string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02", s, time.Local)
	return t
}

func promoDariForm(r *http.Request) model.Promo {
	return model.Promo{
		Nama:         r.PostFormValue("nama"),
		Jenis:        r.PostFormValue("jenis"),
		NilaiPersen:  nullAngka(r.PostFormValue("nilai_persen")),
		NilaiNominal: nullAngka(r.PostFormValue("nilai_nominal")),
		MinBelanja:   nullAngka(r.PostFormValue("min_belanja")),
		KodeVoucher:  nullTeks(r.PostFormValue("kode_voucher")),
		Kuota:        nullAngka(r.PostFormValue("kuota")),
		TglMulai:     tanggal(r.PostFormValue("tgl_mulai")),
		TglSelesai:   tanggal(r.PostFormValue("tgl_selesai")),
		JamMulai:     nullTeks(r.PostFormValue("jam_mulai")),
		JamSelesai:   nullTeks(r.PostFormValue("jam_selesai")),
		Aktif:        r.PostFormValue("aktif") == "on",
		Catatan:      strings.TrimSpace(r.PostFormValue("catatan")),
		CabangIDs:    idsDariForm(r, "cabang"),
		ProdukIDs:    idsDariForm(r, "produk"),
		KategoriIDs:  idsDariForm(r, "kategori"),
	}
}
