// Package model berisi struct data yang dipakai bersama oleh repository,
// service, dan handler. Tidak berisi business logic.
package model

import (
	"database/sql"
	"errors"
	"time"
)

var (
	// ErrNotFound: repository mengubah sql.ErrNoRows menjadi error ini.
	ErrNotFound = errors.New("data tidak ditemukan")
	// ErrLoginGagal: username tidak ada atau password salah (pesannya sengaja sama, materi #11).
	ErrLoginGagal = errors.New("username atau password salah")
)

type Cabang struct {
	ID     int
	Nama   string
	Alamat string
	Aktif  bool
}

// Pengguna tidak memuat password hash, jadi aman dikirim ke handler dan template (materi #11).
type Pengguna struct {
	ID         int
	Username   string
	Peran      string // PeranAdmin | PeranKasir
	CabangID   int    // 0 untuk admin (semua cabang)
	CabangNama string // diisi di daftar pengguna
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
	Deskripsi    string
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
	Catatan       string
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
	WaktuBatal    sql.NullTime
	Token         sql.NullString

	// diisi query daftar/struk
	CabangNama string
	KasirNama  string
	MemberNama string
}

// ItemStruk: baris item pada struk (harga = snapshot saat transaksi).
type ItemStruk struct {
	Nama  string
	Qty   int
	Harga int64
}

// PromoStruk: potongan per promo pada sebuah transaksi.
type PromoStruk struct {
	PromoID int
	Nama    string
	Jenis   string
	Nilai   int64
}

// MutasiPoin: satu baris buku besar poin.
type MutasiPoin struct {
	ID     int
	LotID  int
	Jenis  string
	Jumlah int
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
