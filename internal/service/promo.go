package service

// F-06 – F-11 Promo & voucher

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"unicode/utf8"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type PromoService struct {
	Repo *repository.PromoRepo
}

const (
	JenisDiskon     = "DISKON_PERSEN"
	JenisMinBelanja = "MIN_BELANJA"
	JenisB2G1       = "B2G1"
	JenisVoucher    = "VOUCHER"
)

// UrutJenis: urutan tampil di form.
var UrutJenis = []string{JenisDiskon, JenisMinBelanja, JenisB2G1, JenisVoucher}

var NamaJenis = map[string]string{JenisDiskon: "Diskon persen", JenisMinBelanja: "Minimal belanja", JenisB2G1: "Beli 2 gratis 1", JenisVoucher: "Voucher"}

func (s *PromoService) Daftar(ctx context.Context) ([]model.Promo, error) { return s.Repo.Semua(ctx) }
func (s *PromoService) Detail(ctx context.Context, id int) (model.Promo, error) {
	return s.Repo.CariByID(ctx, id)
}

// Dipakai: jumlah transaksi yang memakai promo ini.
func (s *PromoService) Dipakai(ctx context.Context, id int) (int, error) {
	return s.Repo.Dipakai(ctx, id)
}

func (s *PromoService) Simpan(ctx context.Context, p model.Promo) (int64, error) {
	// Promo yang sudah dipakai transaksi dikunci jenis, nilai, kode, dan targetnya agar laporan tetap jelas;
	// yang boleh diubah: nama, catatan, tanggal, jam, cabang, kuota, status aktif.
	if p.ID > 0 {
		if n, err := s.Repo.Dipakai(ctx, p.ID); err != nil {
			return 0, err
		} else if n > 0 {
			lama, err := s.Repo.CariByID(ctx, p.ID)
			if err != nil {
				return 0, err
			}
			p.Jenis, p.NilaiPersen, p.NilaiNominal, p.MinBelanja = lama.Jenis, lama.NilaiPersen, lama.NilaiNominal, lama.MinBelanja
			p.KodeVoucher, p.ProdukIDs, p.KategoriIDs, p.KuotaTerpakai = lama.KodeVoucher, lama.ProdukIDs, lama.KategoriIDs, lama.KuotaTerpakai
		}
	}
	if err := s.Validasi(ctx, &p); err != nil {
		return 0, err
	}
	return s.Repo.Simpan(ctx, p)
}

func (s *PromoService) SetAktif(ctx context.Context, id int, aktif bool) error {
	return s.Repo.SetAktif(ctx, id, aktif)
}

var reKode = regexp.MustCompile(`^[A-Z0-9]{4,20}$`)
var reJam = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9](:[0-5][0-9])?$`)

func kosong() sql.NullInt64 { return sql.NullInt64{} }

// Validasi menormalkan field sesuai jenis (field jenis lain dikosongkan) lalu memeriksa aturan F-06 – F-10.
func (s *PromoService) Validasi(ctx context.Context, p *model.Promo) error {
	p.Nama = strings.TrimSpace(p.Nama)
	if p.Nama == "" {
		return salah("nama", "Nama promo wajib diisi")
	}
	if len([]rune(p.Nama)) > 100 {
		return salah("nama", "Nama promo maksimal 100 karakter")
	}
	if utf8.RuneCountInString(p.Catatan) > 200 {
		return salah("catatan", "Catatan maksimal 200 karakter")
	}
	if _, ok := NamaJenis[p.Jenis]; !ok {
		return salah("jenis", "Pilih jenis promo")
	}
	// Hanya simpan field milik jenis ini.
	switch p.Jenis {
	case JenisDiskon:
		p.NilaiNominal, p.MinBelanja, p.Kuota = kosong(), kosong(), kosong()
		p.KodeVoucher, p.KategoriIDs = sql.NullString{}, nil
		if !p.NilaiPersen.Valid || p.NilaiPersen.Int64 < 1 || p.NilaiPersen.Int64 > 90 {
			return salah("nilai_persen", "Diskon harus 1–90%")
		}
		if len(p.ProdukIDs) == 0 {
			return salah("produk", "Pilih minimal satu produk yang didiskon")
		}
	case JenisMinBelanja:
		p.NilaiPersen, p.Kuota = kosong(), kosong()
		p.KodeVoucher, p.ProdukIDs, p.KategoriIDs = sql.NullString{}, nil, nil
		if !p.MinBelanja.Valid || p.MinBelanja.Int64 <= 0 {
			return salah("min_belanja", "Minimal belanja harus lebih dari Rp0")
		}
		if !p.NilaiNominal.Valid || p.NilaiNominal.Int64 <= 0 {
			return salah("nilai_nominal", "Potongan harus lebih dari Rp0")
		}
		if p.NilaiNominal.Int64 >= p.MinBelanja.Int64 {
			return salah("nilai_nominal", "Potongan harus lebih kecil dari minimal belanja")
		}
	case JenisB2G1:
		p.NilaiPersen, p.NilaiNominal, p.MinBelanja, p.Kuota = kosong(), kosong(), kosong(), kosong()
		p.KodeVoucher, p.ProdukIDs = sql.NullString{}, nil
		if len(p.KategoriIDs) == 0 {
			return salah("kategori", "Pilih minimal satu kategori")
		}
	case JenisVoucher:
		p.NilaiPersen, p.MinBelanja = kosong(), kosong()
		p.ProdukIDs, p.KategoriIDs = nil, nil
		p.KodeVoucher.String = strings.ToUpper(strings.TrimSpace(p.KodeVoucher.String))
		p.KodeVoucher.Valid = true
		if !reKode.MatchString(p.KodeVoucher.String) {
			return salah("kode_voucher", "Kode 4–20 huruf/angka tanpa spasi, mis. SEDUH10")
		}
		if dipakai, err := s.Repo.KodeDipakai(ctx, p.KodeVoucher.String, p.ID); err != nil {
			return err
		} else if dipakai {
			return salah("kode_voucher", "Kode voucher sudah dipakai promo lain")
		}
		if !p.NilaiNominal.Valid || p.NilaiNominal.Int64 <= 0 {
			return salah("nilai_nominal", "Potongan harus lebih dari Rp0")
		}
		if !p.Kuota.Valid || p.Kuota.Int64 < 1 {
			return salah("kuota", "Kuota minimal 1")
		}
		if p.Kuota.Int64 < int64(p.KuotaTerpakai) {
			return salah("kuota", "Kuota tidak boleh kurang dari yang sudah terpakai")
		}
	}
	if p.TglMulai.IsZero() {
		return salah("tgl_mulai", "Tanggal mulai wajib diisi")
	}
	if p.TglSelesai.IsZero() {
		return salah("tgl_selesai", "Tanggal selesai wajib diisi")
	}
	if p.TglSelesai.Before(p.TglMulai) {
		return salah("tgl_selesai", "Tanggal selesai tidak boleh sebelum tanggal mulai")
	}
	// Jam opsional, tapi harus berpasangan dan mulai < selesai (F-07).
	if p.JamMulai.Valid != p.JamSelesai.Valid {
		return salah("jam_selesai", "Isi jam mulai dan jam selesai, atau kosongkan keduanya")
	}
	if p.JamMulai.Valid {
		if !reJam.MatchString(p.JamMulai.String) || !reJam.MatchString(p.JamSelesai.String) {
			return salah("jam_mulai", "Format jam HH:MM")
		}
		if p.JamMulai.String[:5] >= p.JamSelesai.String[:5] {
			return salah("jam_selesai", "Jam selesai harus setelah jam mulai (promo tidak bisa melewati tengah malam)")
		}
	}
	if len(p.CabangIDs) == 0 {
		return salah("cabang", "Pilih minimal satu cabang")
	}
	// Dua promo beli 2 gratis 1 untuk kategori & cabang yang sama di periode yang beririsan akan memberi gratis dobel.
	if p.Jenis == JenisB2G1 && p.Aktif && s.Repo != nil {
		if nama, err := s.Repo.B2G1Bentrok(ctx, *p); err != nil {
			return err
		} else if nama != "" {
			return salah("kategori", "Bentrok dengan promo \""+nama+"\" (beli 2 gratis 1 untuk kategori dan cabang yang sama di periode yang beririsan)")
		}
	}
	return nil
}
