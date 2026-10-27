package service

// Ketentuan promo yang dibaca kasir dirakit otomatis dari data promo (F-32),
// sehingga selalu sama dengan yang dihitung sistem dan tidak basi saat promo diubah.

import (
	"strconv"
	"time"

	"seduhpoint/internal/model"
)

var bulanPendek = [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}

func tglIndo(t time.Time) string {
	return strconv.Itoa(t.Day()) + " " + bulanPendek[t.Month()-1] + " " + strconv.Itoa(t.Year())
}

func periode(a, b time.Time) string {
	kiri := strconv.Itoa(a.Day()) + " " + bulanPendek[a.Month()-1]
	if a.Year() != b.Year() {
		kiri += " " + strconv.Itoa(a.Year())
	}
	return kiri + " – " + tglIndo(b)
}

func jamTitik(s string) string { // "14:00:00" -> "14.00"
	if len(s) >= 5 {
		return s[:2] + "." + s[3:5]
	}
	return s
}

// Ketentuan mengembalikan kalimat-kalimat ketentuan sebuah promo.
// tampilkanKode = false menyembunyikan kode voucher (layar kasir: kode harus datang dari pelanggan).
func Ketentuan(p model.Promo, tampilkanKode bool) []string {
	var k []string
	switch p.Jenis {
	case JenisDiskon:
		k = append(k, "Diskon "+strconv.FormatInt(p.NilaiPersen.Int64, 10)+"% untuk "+p.TargetNama+".",
			"Bisa digabung dengan promo lain. Jika satu menu kena beberapa diskon, dipakai yang terbesar.")
	case JenisB2G1:
		k = append(k, "Setiap 3 item "+p.TargetNama+", 1 item termurah gratis.",
			"Item yang sudah kena diskon persen tidak ikut dihitung.")
	case JenisMinBelanja:
		k = append(k, "Potongan "+rupiah(p.NilaiNominal.Int64)+" untuk belanja minimal "+rupiah(p.MinBelanja.Int64)+
			" (dihitung setelah promo produk).",
			"Tidak bisa digabung dengan voucher; sistem otomatis memakai potongan yang lebih besar.")
	case JenisVoucher:
		kode := "kode voucher dari pelanggan"
		if tampilkanKode {
			kode = "kode " + p.KodeVoucher.String
		}
		k = append(k, "Potongan "+rupiah(p.NilaiNominal.Int64)+" dengan "+kode+".")
		if p.Kuota.Valid {
			k = append(k, "Sisa kuota "+strconv.FormatInt(max(0, p.Kuota.Int64-int64(p.KuotaTerpakai)), 10)+
				" dari "+strconv.FormatInt(p.Kuota.Int64, 10)+".")
		}
		k = append(k, "Tidak bisa digabung dengan potongan minimal belanja; sistem otomatis memakai potongan yang lebih besar.")
	}
	waktu := "Berlaku " + periode(p.TglMulai, p.TglSelesai)
	if p.JamMulai.Valid {
		waktu += ", pukul " + jamTitik(p.JamMulai.String) + "–" + jamTitik(p.JamSelesai.String) + " WIB"
	}
	k = append(k, waktu+".")
	if p.CabangNama != "" {
		k = append(k, "Cabang: "+p.CabangNama+".")
	}
	if p.Catatan != "" {
		k = append(k, p.Catatan)
	}
	return k
}

// PromoUntukProduk: promo produk (diskon persen / beli 2 gratis 1) yang mengenai produk ini.
func PromoUntukProduk(promo []model.Promo, pr model.Produk) []model.Promo {
	var hasil []model.Promo
	for _, p := range promo {
		if (p.Jenis == JenisDiskon && ada(p.ProdukIDs, pr.ID)) || (p.Jenis == JenisB2G1 && ada(p.KategoriIDs, pr.KategoriID)) {
			hasil = append(hasil, p)
		}
	}
	return hasil
}

func ada(ids []int, id int) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
