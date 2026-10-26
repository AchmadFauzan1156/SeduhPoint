package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"seduhpoint/internal/model"
)

func n(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

// Validasi tanpa voucher tidak menyentuh database, jadi bisa dites tanpa MySQL.
func TestValidasiPromo(t *testing.T) {
	s := &PromoService{}
	mulai := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	dasar := func() model.Promo {
		return model.Promo{Nama: "Promo", TglMulai: mulai, TglSelesai: mulai.AddDate(0, 0, 29), CabangIDs: []int{1}}
	}
	cases := []struct {
		nama  string
		ubah  func(p *model.Promo)
		field string // "" = lolos
	}{
		{"diskon ok", func(p *model.Promo) { p.Jenis, p.NilaiPersen, p.ProdukIDs = JenisDiskon, n(20), []int{3} }, ""},
		{"diskon 0%", func(p *model.Promo) { p.Jenis, p.NilaiPersen, p.ProdukIDs = JenisDiskon, n(0), []int{3} }, "nilai_persen"},
		{"diskon tanpa produk", func(p *model.Promo) { p.Jenis, p.NilaiPersen = JenisDiskon, n(20) }, "produk"},
		{"min belanja ok", func(p *model.Promo) { p.Jenis, p.MinBelanja, p.NilaiNominal = JenisMinBelanja, n(100000), n(15000) }, ""},
		{"potongan >= minimal", func(p *model.Promo) { p.Jenis, p.MinBelanja, p.NilaiNominal = JenisMinBelanja, n(10000), n(15000) }, "nilai_nominal"},
		{"b2g1 tanpa kategori", func(p *model.Promo) { p.Jenis = JenisB2G1 }, "kategori"},
		{"tanggal terbalik", func(p *model.Promo) {
			p.Jenis, p.KategoriIDs, p.TglSelesai = JenisB2G1, []int{1}, mulai.AddDate(0, 0, -1)
		}, "tgl_selesai"},
		{"jam hanya satu", func(p *model.Promo) {
			p.Jenis, p.KategoriIDs, p.JamMulai = JenisB2G1, []int{1}, sql.NullString{String: "14:00", Valid: true}
		}, "jam_selesai"},
		{"jam terbalik", func(p *model.Promo) {
			p.Jenis, p.KategoriIDs = JenisB2G1, []int{1}
			p.JamMulai, p.JamSelesai = sql.NullString{String: "17:00", Valid: true}, sql.NullString{String: "14:00", Valid: true}
		}, "jam_selesai"},
		{"tanpa cabang", func(p *model.Promo) { p.Jenis, p.KategoriIDs, p.CabangIDs = JenisB2G1, []int{1}, nil }, "cabang"},
		{"jenis kosong", func(p *model.Promo) {}, "jenis"},
	}
	for _, c := range cases {
		p := dasar()
		c.ubah(&p)
		err := s.Validasi(context.Background(), &p)
		var ve *ValidasiError
		switch {
		case c.field == "" && err != nil:
			t.Errorf("%s: harusnya lolos, dapat %v", c.nama, err)
		case c.field != "" && (!errors.As(err, &ve) || ve.Field != c.field):
			t.Errorf("%s: harusnya salah di %q, dapat %v", c.nama, c.field, err)
		}
	}
}

func TestDiskonMengosongkanFieldLain(t *testing.T) {
	p := model.Promo{Nama: "X", Jenis: JenisDiskon, NilaiPersen: n(10), ProdukIDs: []int{1}, MinBelanja: n(5),
		KategoriIDs: []int{2}, TglMulai: time.Now(), TglSelesai: time.Now(), CabangIDs: []int{1}}
	if err := (&PromoService{}).Validasi(context.Background(), &p); err != nil {
		t.Fatal(err)
	}
	if p.MinBelanja.Valid || p.KategoriIDs != nil {
		t.Error("field jenis lain harus dikosongkan")
	}
}

func TestNormalisasiHP(t *testing.T) {
	for in, want := range map[string]string{"0812-3456 7890": "081234567890", "+6281234567890": "6281234567890"} {
		if got := NormalisasiHP(in); got != want {
			t.Errorf("%q -> %q, harap %q", in, got, want)
		}
	}
}
