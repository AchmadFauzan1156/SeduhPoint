// Package model berisi struct data yang dipakai bersama oleh repository,
// service, dan handler. Tidak berisi business logic (materi #10).
package model

import (
	"database/sql"
	"errors"
	"time"
)

// ErrNotFound: repository mengubah sql.ErrNoRows menjadi error ini.
var ErrNotFound = errors.New("data tidak ditemukan")

type Cabang struct {
	ID     int
	Nama   string
	Alamat string
	Aktif  bool
}

type Pengguna struct {
	ID           int
	Username     string
	PasswordHash string
	Peran        string // ADMIN | KASIR
	CabangID     sql.NullInt64
}

type Kategori struct {
	ID           int
	Nama         string
	JumlahProduk int
}

type Produk struct {
	ID           int
	KategoriID   int
	KategoriNama string
	Nama         string
	Harga        int64
	Gambar       string // path foto; kosong = pakai gambar bawaan
	Aktif        bool
}

type Member struct {
	ID        int
	Nama      string
	NoHP      string
	TglDaftar time.Time
	Tingkat   string // SILVER | GOLD

	// diisi query daftar/detail
	PoinAktif     int
	Belanja3Bulan int64
}

type Promo struct {
	ID            int
	Nama          string
	Jenis         string // DISKON_PERSEN | MIN_BELANJA | B2G1 | VOUCHER
	NilaiPersen   sql.NullInt64
	NilaiNominal  sql.NullInt64
	MinBelanja    sql.NullInt64
	KodeVoucher   sql.NullString
	Kuota         sql.NullInt64
	KuotaTerpakai int
	TglMulai      time.Time
	TglSelesai    time.Time
	JamMulai      sql.NullString
	JamSelesai    sql.NullString
	Aktif         bool
	CabangIDs     []int
	ProdukIDs     []int
	KategoriIDs   []int
	CabangNama    string // ringkasan untuk daftar, mis. "Sekaran, Banaran"
	TargetNama    string // produk/kategori target, mis. "Cappuccino"
}

type Transaksi struct {
	ID            int
	NoTransaksi   string
	CabangID      int
	PenggunaID    int
	MemberID      sql.NullInt64
	Waktu         time.Time
	Subtotal      int64
	TotalPromo    int64
	PoinDitukar   int
	PotonganPoin  int64
	TotalBayar    int64
	PoinDiperoleh int
	Status        string // SELESAI | DIBATALKAN
	AlasanBatal   sql.NullString
}

type PoinLot struct {
	ID             int
	MemberID       int
	TransaksiID    int
	PoinAwal       int
	PoinSisa       int
	TglPerolehan   time.Time
	TglKedaluwarsa time.Time
	Status         string
}
