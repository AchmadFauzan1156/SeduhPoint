package handler

// F-12 – F-24 Kasir, struk, riwayat, pembatalan.
// Halaman kasir sepenuhnya server-side: setiap tombol mengirim form, keranjang disimpan di input tersembunyi.

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"seduhpoint/internal/config"
	"seduhpoint/internal/middleware"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

type TransaksiHandler struct {
	Svc       *service.TransaksiService
	ProdukSvc *service.ProdukService
	MemberSvc *service.MemberService
	CabangSvc *service.CabangService
	View      *Renderer
}

type KasirPage struct {
	Cabang     string
	Produk     []model.Produk
	Kategori   []model.Kategori
	Kat        int
	Q          string
	Label      map[int]string
	Pr         service.Pratinjau
	MemberID   int
	CariMember string
	HasilCari  []model.Member
	Kode       string
	Poin       int
	Token      string
	Detail     map[int]DetailProduk
	PromoHari  []PromoInfo
}

// keranjang dibaca dari form kasir.
type keranjang struct {
	Qty        map[int]int
	MemberID   int
	Kode, Q    string
	CariMember string
	Poin, Kat  int
	Token      string
}

func tokenBaru() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func bacaKeranjang(r *http.Request) keranjang {
	k := keranjang{Qty: map[int]int{}, MemberID: atoi(r.FormValue("member_id")), Kode: strings.TrimSpace(r.FormValue("kode_voucher")),
		Q: strings.TrimSpace(r.FormValue("q")), CariMember: strings.TrimSpace(r.FormValue("cari_member")),
		Poin: atoi(r.FormValue("poin")), Kat: atoi(r.FormValue("kat_aktif")), Token: r.FormValue("token")}
	if v, ok := r.PostForm["kat"]; ok && len(v) > 0 { // tombol filter kategori ditekan
		k.Kat = atoi(v[0])
	}
	for _, v := range r.PostForm["item"] {
		if id, qty, ok := strings.Cut(v, ":"); ok {
			if i, q := atoi(id), atoi(qty); i > 0 && q > 0 {
				k.Qty[i] = min(q, 99)
			}
		}
	}
	if id := atoi(r.FormValue("tambah")); id > 0 {
		k.Qty[id] = min(k.Qty[id]+1, 99)
	}
	if id := atoi(r.FormValue("kurang")); id > 0 {
		if k.Qty[id]--; k.Qty[id] <= 0 {
			delete(k.Qty, id)
		}
	}
	if id := atoi(r.FormValue("pilih_member")); id > 0 {
		k.MemberID, k.CariMember = id, ""
	}
	switch r.FormValue("aksi") {
	case "kosong":
		k = keranjang{Qty: map[int]int{}, Kat: k.Kat, Token: tokenBaru()}
	case "lepas_member":
		k.MemberID, k.Poin = 0, 0
	case "hapus_voucher":
		k.Kode = ""
	}
	if k.Token == "" {
		k.Token = tokenBaru()
	}
	return k
}

func (h *TransaksiHandler) permintaan(r *http.Request, k keranjang) service.PermintaanCheckout {
	s, _ := middleware.Ambil(r)
	return service.PermintaanCheckout{CabangID: s.CabangID, PenggunaID: s.PenggunaID, MemberID: k.MemberID, Qty: k.Qty,
		KodeVoucher: k.Kode, PoinDiminta: k.Poin, Token: k.Token}
}

// tampilKasir menghitung pratinjau dan merender halaman kasir. errAwal = error dari checkout (jika ada).
func (h *TransaksiHandler) tampilKasir(w http.ResponseWriter, r *http.Request, k keranjang, errAwal error) {
	ctx := r.Context()
	s, _ := middleware.Ambil(r)
	pg := KasirPage{Kat: k.Kat, Q: k.Q, MemberID: k.MemberID, CariMember: k.CariMember, Kode: k.Kode, Poin: k.Poin,
		Token: k.Token, Label: map[int]string{}}
	hal := halaman(r, "Transaksi baru", "kasir", nil)
	catat := func(err error) bool {
		var ve *service.ValidasiError
		if !errors.As(err, &ve) {
			return false
		}
		if ve.Field != "" {
			hal.Errors[ve.Field] = ve.Pesan
		} else {
			hal.Error = ve.Pesan
		}
		return true
	}
	if errAwal != nil && !catat(errAwal) {
		gagal(w, errAwal)
		return
	}
	if c, err := h.CabangSvc.Detail(ctx, s.CabangID); err == nil {
		pg.Cabang = c.Nama
	}
	semua, err := h.ProdukSvc.DaftarDijual(ctx)
	if err != nil {
		gagal(w, err)
		return
	}
	if pg.Kategori, err = h.ProdukSvc.DaftarKategori(ctx); err != nil {
		gagal(w, err)
		return
	}
	for _, p := range semua {
		if (k.Kat == 0 || p.KategoriID == k.Kat) && (k.Q == "" || strings.Contains(strings.ToLower(p.Nama), strings.ToLower(k.Q))) {
			pg.Produk = append(pg.Produk, p)
		}
	}
	pg.Pr, err = h.Svc.Pratinjau(ctx, h.permintaan(r, k))
	if err != nil {
		if !catat(err) {
			gagal(w, err)
			return
		}
		// Voucher tidak valid: tampilkan pesan, hitung ulang tanpa voucher.
		k.Kode, pg.Kode = "", ""
		if pg.Pr, err = h.Svc.Pratinjau(ctx, h.permintaan(r, k)); err != nil {
			gagal(w, err)
			return
		}
	}
	// Label promo di kartu menu
	for _, pm := range pg.Pr.Promo {
		switch pm.Jenis {
		case service.JenisDiskon:
			for _, id := range pm.ProdukIDs {
				pg.Label[id] = "Diskon " + strconv.FormatInt(pm.NilaiPersen.Int64, 10) + "%"
			}
		case service.JenisB2G1:
			for _, p := range semua {
				if has(pm.KategoriIDs, p.KategoriID) && pg.Label[p.ID] == "" {
					pg.Label[p.ID] = "Beli 2 gratis 1"
				}
			}
		}
	}
	// Kartu detail produk & panel "Promo hari ini" (kode voucher tidak ditampilkan ke kasir).
	pg.Detail = detailProduk(pg.Produk, pg.Pr.Promo, false, true)
	pg.PromoHari = infoPromo(pg.Pr.Promo, false)
	if k.CariMember != "" {
		if pg.HasilCari, err = h.MemberSvc.Cari(ctx, k.CariMember); err != nil {
			gagal(w, err)
			return
		}
		if len(pg.HasilCari) > 5 {
			pg.HasilCari = pg.HasilCari[:5]
		}
	}
	hal.Data = pg
	if len(hal.Errors) > 0 || hal.Error != "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	h.View.Tampil(w, "transaksi/kasir.html", hal)
}

func (h *TransaksiHandler) Kasir(w http.ResponseWriter, r *http.Request) {
	h.tampilKasir(w, r, keranjang{Qty: map[int]int{}, Kat: atoi(r.URL.Query().Get("kat")), Token: tokenBaru()}, nil)
}

// Hitung: semua tombol di halaman kasir (tambah, kurang, pilih member, voucher, poin) kembali ke sini.
func (h *TransaksiHandler) Hitung(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	h.tampilKasir(w, r, bacaKeranjang(r), nil)
}

func (h *TransaksiHandler) Bayar(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	k := bacaKeranjang(r)
	id, err := h.Svc.Checkout(r.Context(), h.permintaan(r, k))
	if err != nil {
		h.tampilKasir(w, r, k, err)
		return
	}
	redirect(w, r, fmt.Sprintf("/transaksi/%d", id), "Transaksi berhasil disimpan")
}

// bolehAkses: kasir hanya boleh melihat/membatalkan transaksi cabangnya sendiri.
func bolehAkses(r *http.Request, t model.Transaksi) bool {
	s, _ := middleware.Ambil(r)
	return s.Peran == "ADMIN" || s.CabangID == t.CabangID
}

func (h *TransaksiHandler) Struk(w http.ResponseWriter, r *http.Request) {
	id, ok := idDariPath(w, r)
	if !ok {
		return
	}
	st, err := h.Svc.Struk(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return
	}
	if !bolehAkses(r, st.Transaksi) {
		http.Error(w, "Akses ditolak", http.StatusForbidden)
		return
	}
	h.View.Tampil(w, "transaksi/struk.html", halaman(r, st.Transaksi.NoTransaksi, "transaksi", st))
}

type RiwayatPage struct {
	Daftar []model.Transaksi
	Tgl    time.Time
	Batal  *BatalPanel
	Alasan string
}

type BatalPanel struct {
	Transaksi model.Transaksi
	Dampak    service.Dampak
}

func (h *TransaksiHandler) riwayat(w http.ResponseWriter, r *http.Request, batal *BatalPanel, alasan string, err error) {
	s, _ := middleware.Ambil(r)
	tgl, e := time.ParseInLocation("2006-01-02", r.URL.Query().Get("tgl"), config.WIB)
	if e != nil {
		tgl = time.Now().In(config.WIB)
		if batal != nil {
			tgl = batal.Transaksi.Waktu
		}
	}
	cabang := 0
	if s.Peran == "KASIR" {
		cabang = s.CabangID
	}
	daftar, e := h.Svc.Daftar(r.Context(), cabang, tgl)
	if e != nil {
		gagal(w, e)
		return
	}
	judul := "Riwayat transaksi"
	if batal != nil {
		judul = "Batalkan " + batal.Transaksi.NoTransaksi
	}
	hal := halaman(r, judul, "transaksi", RiwayatPage{Daftar: daftar, Tgl: tgl, Batal: batal, Alasan: alasan})
	if err != nil {
		formGagal(w, h.View, "transaksi/list.html", hal, err)
		return
	}
	h.View.Tampil(w, "transaksi/list.html", hal)
}

func (h *TransaksiHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	h.riwayat(w, r, nil, "", nil)
}

func (h *TransaksiHandler) panelBatal(w http.ResponseWriter, r *http.Request) (*BatalPanel, bool) {
	id, ok := idDariPath(w, r)
	if !ok {
		return nil, false
	}
	t, err := h.Svc.Detail(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return nil, false
	}
	if !bolehAkses(r, t) {
		http.Error(w, "Akses ditolak", http.StatusForbidden)
		return nil, false
	}
	d, err := h.Svc.DampakBatal(r.Context(), id)
	if err != nil {
		gagal(w, err)
		return nil, false
	}
	return &BatalPanel{Transaksi: t, Dampak: d}, true
}

func (h *TransaksiHandler) FormBatal(w http.ResponseWriter, r *http.Request) {
	if p, ok := h.panelBatal(w, r); ok {
		h.riwayat(w, r, p, "", nil)
	}
}

func (h *TransaksiHandler) Batal(w http.ResponseWriter, r *http.Request) {
	p, ok := h.panelBatal(w, r)
	if !ok {
		return
	}
	alasan := r.PostFormValue("alasan")
	if err := h.Svc.Batalkan(r.Context(), p.Transaksi.ID, alasan); err != nil {
		h.riwayat(w, r, p, alasan, err)
		return
	}
	redirect(w, r, fmt.Sprintf("/transaksi/%d", p.Transaksi.ID), "Transaksi "+p.Transaksi.NoTransaksi+" dibatalkan")
}
