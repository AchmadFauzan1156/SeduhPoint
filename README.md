# SeduhPoint — Sistem Poin Member & Promo Kedai Kopi Seduh

Project Akhir Pemrograman Web · Kelompok 5 · Topik 5
Stack: **Go 1.22+ (`net/http`, `html/template`) + MySQL 8** + Bootstrap 5 (CDN)

## Cara menjalankan
1. Install Go 1.22+ dan MySQL 8 (Laragon/XAMPP).
2. Import database (urut):
   `database/migrations/001_schema.sql` lalu `database/migrations/002_seed.sql`
3. Salin `.env.example` menjadi `.env` dan sesuaikan user/password MySQL.
4. Jalankan:
   ```
   go mod tidy
   go run .
   ```
5. Buka http://localhost:8080. Login: `admin` / `kasir1` / `kasir2` / `kasir3`, password `password`.

Tes mesin perhitungan: `go test ./...`

## Struktur folder
| Folder | Isi | Rujukan laporan |
|---|---|---|
| `main.go` | Wiring & start server, job harian | — |
| `config/` | Baca `.env`, zona waktu WIB | NF-06 |
| `database/` | Koneksi MySQL, `WithTx`, file SQL | NF-02, Tabel 1.10 |
| `models/` | Struct tabel | Tabel 1.10 |
| `hitung/` | Mesin hitung promo & poin (tanpa DB) + test | 1.8, AB01–AB09 |
| `store/` | Semua query SQL, satu file per modul | — |
| `layanan/` | Checkout, pembatalan, poin & tingkat (pakai transaction) | F-12–F-27 |
| `middleware/` | Session login, peran, logger | F-05, NF-03 |
| `view/` | Loader template + fungsi `rupiah` | — |
| `handler/` | Route & handler HTTP | F-01–F-30 |
| `templates/` | `layout.html`, `partials/`, satu folder per modul | 1.12 |
| `static/` | CSS & JS | — |

**Contoh lengkap yang menjadi acuan:** modul **Cabang** (`store/cabang.go`, `handler/cabang.go`,
`templates/cabang/`): list, form, validasi server, Post/Redirect/Get. Modul lain tinggal mengikuti polanya.
Cari bagian yang harus dikerjakan dengan mencari teks `TODO M2`, `TODO M3`, `TODO M4`.

## Usulan pembagian modul (silakan diubah)
| Modul | File utama | Milestone |
|---|---|---|
| A. Login + Cabang + Kategori/Produk | `handler/auth.go`, `cabang.go`, `produk.go` | M1–M2 |
| B. Member + saldo poin | `handler/member.go`, `store/member.go`, `templates/member/` | M2 |
| C. Promo & voucher | `handler/promo.go`, `store/promo.go`, `templates/promo/` | M2 |
| D. Mesin hitung + test | `hitung/` | M3 |
| E. Kasir, checkout, pembatalan | `layanan/checkout.go`, `layanan/batal.go`, `handler/transaksi.go` | M3 |
| F. Poin kedaluwarsa, tingkat, laporan | `layanan/poin.go`, `store/laporan.go`, `templates/laporan/` | M3–M4 |

## Aturan main kode
- Uang selalu `int64` rupiah, tidak pernah `float` (NF-04). Gold 1,5× → `poin*3/2`.
- Query selalu pakai placeholder `?`.
- Perubahan poin/transaksi/kuota **wajib** di dalam `database.WithTx` + `SELECT ... FOR UPDATE`.
- `poin_mutasi` hanya INSERT (tidak di-UPDATE/DELETE) agar bisa diaudit (NF-05).
- Transaksi dibatalkan tidak dihapus, hanya status `DIBATALKAN` (F-20).
- Setelah form POST berhasil → `http.Redirect(..., 303)` (Post/Redirect/Get).
