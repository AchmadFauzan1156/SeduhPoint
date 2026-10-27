# SeduhPoint — Sistem Poin Member & Promo Kedai Kopi Seduh

Project Akhir Pemrograman Web · Kelompok 5 · Topik 5
Stack: **Go 1.22+ (`net/http`, `html/template`) + MySQL 8**, CSS & JS sendiri tanpa framework.
Struktur mengikuti materi #10 *Struktur Aplikasi & Kickoff Project*.

## Cara menjalankan (Windows + Laragon, lewat terminal)
Semua perintah dijalankan dari folder `SeduhPoint`.

1. Nyalakan Laragon → **Start All**.
2. Buat database (pertama kali, atau untuk mengulang dari awal):
   ```
   mysql -u root -e "DROP DATABASE IF EXISTS seduhpoint"
   mysql -u root --default-character-set=utf8mb4 -e "source db/seduhpoint.sql"
   ```
   Jika `mysql` tidak dikenali, pakai Terminal Laragon (Menu → Terminal) atau path lengkapnya,
   mis. `E:\laragon\bin\mysql\mysql-8.4.3-winx64\bin\mysql.exe`.
3. Salin konfigurasi: `copy .env.example .env` (isi `DB_DSN` jika password MySQL tidak kosong).
4. Jalankan aplikasi:
   ```
   go run ./cmd/web
   ```
   atau dengan auto reload (materi #6): `wgo run -file .html -file .css ./cmd/web`
   (pasang dulu: `go install github.com/bokwoon95/wgo@latest`).
5. Buka http://localhost:8080 — login `admin` / `kasir1` / `kasir2` / `kasir3`, password `password`.

Tes aturan bisnis: `go test ./...`

## Struktur
```
cmd/web/main.go          hanya menyambung: config → db → repository → service → handler → route
internal/
  config/                baca .env (ADDR, DB_DSN), zona waktu WIB
  model/                 struct data + ErrNotFound (tanpa business logic)
  repository/            satu-satunya tempat SQL; satu struct per entitas (CabangRepo, ProdukRepo, ...)
  service/               aturan bisnis + validasi (ValidasiError); hitung.go = mesin promo & poin + test
  handler/               urusan HTTP; render.go (Renderer.Tampil), helper.go (gagal, idDariPath)
  middleware/            logging, login & peran
db/seduhpoint.sql        seluruh query database: membuat database + 15 tabel + data contoh
templates/               layout, header, sidebar, footer + satu folder per modul
static/                  css, js, logo, foto menu
```
Alur request: **Handler → Service → Repository → MySQL**.

## Fitur
- **Login & peran:** admin (pemilik) dan kasir per cabang.
- **Data master:** cabang, kategori, produk (foto & deskripsi), member (cari, detail poin & lot, menuju Gold).
- **Promo & voucher 4 jenis:** diskon persen, potongan minimal belanja, beli 2 gratis 1, voucher berkode
  (periode, cabang, jam, kuota). Diskon 1–90%, jam promo tidak boleh lewat tengah malam, beli 2 gratis 1 tidak boleh
  tumpang tindih, promo yang sudah dipakai terkunci jenis/nilai/kode/targetnya.
- **Kasir:** menu bergambar, kartu detail menu, ketentuan promo berlaku, promo dihitung otomatis; minimal belanja
  dan voucher tidak digabung, sistem memakai potongan yang lebih besar; tukar poin FIFO (maks 50%).
- **Checkout** dalam satu DB transaction (anti bayar ganda), struk, riwayat, pembatalan (poin ditarik/dikembalikan,
  kuota voucher kembali), job harian kedaluwarsa poin & evaluasi Silver/Gold.
- **Ringkasan admin** (hari ini / 7 / 30 hari, grafik omzet, omzet per cabang) dan **laporan pemilik**:
  biaya promo (F-28), poin aktif (F-29), perkembangan member (F-30). Hanya transaksi selesai yang dihitung.

## Pola route (sama untuk tiap modul)
| Method | Path | Fungsi |
|---|---|---|
| GET | `/cabang` | Daftar |
| GET | `/cabang/form` | Form tambah |
| POST | `/cabang` | Simpan |
| GET | `/cabang/{id}/edit` | Form ubah (pre-fill) |
| POST | `/cabang/{id}/edit` | Update |
| POST | `/cabang/{id}/hapus` | Hapus (soft delete) |

## Usulan pembagian modul (silakan diubah)
| Modul | File utama | Milestone |
|---|---|---|
| A. Login + Cabang + Kategori/Produk | `handler|service|repository/{auth,cabang,produk}.go` | M1–M2 |
| B. Member + saldo poin | `*/member.go`, `templates/member/` | M2 |
| C. Promo & voucher | `*/promo.go`, `templates/promo/` | M2 |
| D. Mesin hitung + test | `service/hitung.go`, `hitung_test.go` | M3 |
| E. Kasir, checkout, pembatalan | `service/transaksi.go`, `handler/transaksi.go` | M3 |
| F. Poin kedaluwarsa, tingkat, laporan | `service/poin.go`, `*/laporan.go` | M3–M4 |

## Aturan main kode
- Handler tidak berisi SQL; service tidak tahu `http.Request`; SQL hanya di repository.
- Query selalu `...Context(ctx, ...)` dengan placeholder `?`.
- Uang selalu `int64` rupiah (NF-04). Gold 1,5× → `poin*3/2`.
- Perubahan poin/transaksi/kuota **wajib** dalam `repository.WithTx` + `SELECT ... FOR UPDATE`.
- `poin_mutasi` hanya INSERT (NF-05). Transaksi batal tidak dihapus, status `DIBATALKAN` (F-20).
- Data master dihapus dengan soft delete (`deleted_at`).
- Setelah POST berhasil → `http.Redirect(..., 303)`.

