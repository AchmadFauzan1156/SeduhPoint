# SeduhPoint — Sistem Poin Member & Promo Kedai Kopi Seduh

Project Akhir Pemrograman Web · Kelompok 5 · Topik 5
Stack: **Go 1.22+ (`net/http`, `html/template`) + MySQL 8** + Bootstrap 5 (CDN).
Struktur mengikuti materi #10 *Struktur Aplikasi & Kickoff Project*.

## Cara menjalankan (Windows + Laragon)
1. Nyalakan Laragon → **Start All**.
2. Pertama kali (atau setelah skema berubah):
   `powershell -ExecutionPolicy Bypass -File scripts\setup.ps1 -Reset`
   (membuat database dari `db/schema.sql` + `db/seed.sql`, `.env`, lalu vet/test/build)
3. Sehari-hari: double-click **`jalankan.bat`**, atau `go run ./cmd/web`
   (auto reload: `wgo run ./cmd/web`).
4. Buka http://localhost:8080 — login `admin` / `kasir1` / `kasir2` / `kasir3`, password `password`.

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
db/schema.sql, seed.sql  skema MySQL + data contoh
templates/               layout, header, sidebar, footer + satu folder per modul
static/  docs/  scripts/
```
Alur request: **Handler → Service → Repository → MySQL**.

**M4 (27 Okt):** ringkasan admin (omzet, transaksi, biaya promo, member Gold; hari ini / 7 / 30 hari),
laporan biaya promo per promo & per cabang dengan filter tanggal & cabang (F-28), laporan poin aktif & yang akan
kedaluwarsa ≤ 30 hari (F-29), laporan perkembangan member per bulan (F-30). Semua hanya menghitung transaksi selesai.

**M4.0 (27 Okt):** deskripsi produk (klik nama/foto menu → kartu detail di kasir & admin), panel
"Promo berlaku sekarang" berisi ketentuan tiap promo (kode voucher disembunyikan dari kasir), catatan promo untuk kasir.
Aturan baru: diskon 1–90%, jam promo tidak boleh melewati tengah malam, beli 2 gratis 1 tidak boleh tumpang tindih
di kategori & cabang yang sama, promo yang sudah dipakai transaksi terkunci jenis/nilai/kode/targetnya, dan jika
potongan minimal belanja & voucher sama-sama memenuhi syarat sistem memakai yang lebih besar (voucher tidak terpakai).
**Perlu reset DB** (`scripts\cek.bat`) karena ada kolom baru `produk.deskripsi` dan `promo.catatan`.

**Fitur yang sudah jalan (M3):** kasir (menu bergambar, promo otomatis, voucher, tukar poin FIFO), checkout dalam
satu DB transaction (anti bayar ganda), struk, riwayat, pembatalan (poin ditarik/dikembalikan, kuota voucher kembali),
job harian kedaluwarsa poin & evaluasi Silver/Gold.

**M2:** login & peran, Cabang, Kategori, Produk (dengan unggah foto), Member
(daftar, cari, detail poin & lot, menuju Gold), Promo & voucher 4 jenis (cabang, jam, kuota, aktif/nonaktif).
Akun contoh di `db/seed.sql`, termasuk 4 promo contoh (P1–P4 di `docs/data-uji.md`).

**Contoh lengkap = modul Cabang** (`repository/cabang.go`, `service/cabang.go`, `handler/cabang.go`,
`templates/cabang/`): daftar, tambah, ubah, hapus (soft delete), validasi, Post/Redirect/Get.
Bagian yang harus dikerjakan: cari teks `TODO M2`, `TODO M3`, `TODO M4`.

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

Dokumen pendukung: `docs/use-case.md`, `docs/data-uji.md`.
