# SeduhPoint

Aplikasi web untuk mengelola **program member dan promo kedai kopi dengan banyak cabang**. Kasir mencatat
transaksi, sistem menghitung promo dan poin member secara otomatis, dan pemilik memantau biaya promo,
poin yang beredar, serta perkembangan member lewat laporan.

Dibangun dengan **Go** (`net/http`, `html/template`) dan **MySQL/MariaDB**, dirender di server tanpa framework
front-end.

## Fitur

**Kasir**
- Menu bergambar per kategori, pencarian menu, dan kartu detail menu (deskripsi, harga, promo yang berlaku).
- Promo dihitung otomatis; daftar promo yang sedang berlaku beserta ketentuannya bisa dilihat kapan saja.
- Pilih member, pakai kode voucher, dan tukar poin sebagai potongan pembayaran.
- Pembayaran disimpan dalam satu transaksi database (aman dari klik ganda), lalu struk ditampilkan.
- Riwayat transaksi dan pembatalan transaksi beserta alasannya.

**Member & poin**
- Member mendapat 1 poin untuk setiap Rp10.000 belanja setelah promo; 1 poin bernilai Rp1.000.
- Poin berlaku 6 bulan. Saat ditukar, poin yang paling dekat kedaluwarsa dipakai lebih dulu (maks. 50% total belanja).
- Dua tingkat member: **Silver** dan **Gold**. Member dengan belanja minimal Rp1.500.000 dalam 3 bulan terakhir
  menjadi Gold dan mendapat poin 1,5 kali lipat. Tingkat dievaluasi otomatis setiap hari.
- Pembatalan transaksi menarik poin yang didapat dan mengembalikan poin yang ditukar.

**Promo & voucher**
- Empat jenis promo: diskon persen untuk produk tertentu, potongan untuk minimal belanja, beli 2 gratis 1 untuk
  kategori tertentu, dan voucher berkode dengan kuota pemakaian.
- Setiap promo punya masa berlaku dan bisa dibatasi per cabang dan per jam.
- Promo produk bisa digabung dengan promo lain. Potongan minimal belanja tidak digabung dengan voucher; jika
  keduanya memenuhi syarat, sistem memakai potongan yang lebih besar.
- Kode voucher tidak ditampilkan di layar kasir.

**Admin / pemilik**
- Kelola cabang, kategori, produk (dengan foto), member, promo, dan voucher.
- Ringkasan: omzet, jumlah transaksi, biaya promo, member Gold, grafik omzet harian, dan omzet per cabang
  (hari ini, 7 hari, atau 30 hari) dibanding periode sebelumnya.
- Laporan biaya promo (filter tanggal & cabang), laporan poin aktif dan poin yang akan kedaluwarsa,
  serta laporan perkembangan member per bulan.

## Kebutuhan

- [Go](https://go.dev/dl/) 1.22 atau lebih baru
- MySQL 8 atau MariaDB 10.6+ — pilih salah satu:
  - **Laragon** (Windows): klik *Start All*
  - **XAMPP** (Windows/Linux/macOS): nyalakan *MySQL* di XAMPP Control Panel
  - MySQL/MariaDB yang dipasang sendiri, atau lewat Docker:
    `docker run -d --name seduhpoint-db -p 3306:3306 -e MYSQL_ALLOW_EMPTY_PASSWORD=yes mysql:8`

## Menjalankan

Semua perintah dijalankan dari folder proyek.

**1. Buat database** (sekali di awal; ulangi untuk mengembalikan ke data contoh)
```bash
mysql -u root -e "DROP DATABASE IF EXISTS seduhpoint"
mysql -u root --default-character-set=utf8mb4 -e "source db/seduhpoint.sql"
```
Tambahkan `-p` jika user MySQL memakai password. Jika perintah `mysql` tidak dikenali, pakai path lengkapnya:

| Server | Lokasi `mysql` |
|---|---|
| Laragon | `C:\laragon\bin\mysql\mysql-<versi>\bin\mysql.exe` (atau buka *Menu → Terminal* di Laragon) |
| XAMPP Windows | `C:\xampp\mysql\bin\mysql.exe` |
| XAMPP Linux / macOS | `/opt/lampp/bin/mysql` · `/Applications/XAMPP/xamppfiles/bin/mysql` |

Bisa juga lewat phpMyAdmin/HeidiSQL: buka tab SQL atau *Import*, lalu jalankan isi `db/seduhpoint.sql`.

**2. Atur koneksi**
```bash
cp .env.example .env        # Windows CMD: copy .env.example .env
```
Ubah `DB_DSN` di `.env` jika user, password, atau port MySQL berbeda.

**3. Jalankan**
```bash
go run ./cmd/web
```
Buka http://localhost:8080.

Untuk muat ulang otomatis saat kode berubah, pakai [wgo](https://github.com/bokwoon95/wgo):
```bash
go install github.com/bokwoon95/wgo@latest
wgo run -file .html -file .css ./cmd/web
```

**Akun contoh** (password semua: `password`)

| Username | Peran |
|---|---|
| `admin` | Admin / pemilik — semua cabang |
| `kasir1`, `kasir2`, `kasir3` | Kasir cabang Sekaran, Banaran, Tembalang |

Data contoh juga berisi 7 menu, 2 member, dan 4 promo (satu dari tiap jenis; kode voucher `SEDUH10`) yang berlaku
1 Okt – 31 Des 2026; ubah tanggalnya di menu *Promo & voucher* bila perlu.
Ganti password akun sebelum dipakai sungguhan.

## Pengujian

```bash
go test ./...
```
Menguji perhitungan promo, poin, tingkat member, validasi promo, dan data laporan.

## Struktur proyek

```
cmd/web/main.go     titik masuk: menyambung konfigurasi, database, repository, service, handler, dan route
internal/
  config/           membaca .env (ADDR, DB_DSN) dan zona waktu WIB
  model/            struct data
  repository/       semua query SQL, satu file per entitas
  service/          aturan bisnis dan validasi; hitung.go = perhitungan promo & poin
  handler/          urusan HTTP: membaca form, memanggil service, merender template
  middleware/       logging, login, dan pembatasan peran
db/seduhpoint.sql   membuat database, seluruh tabel, dan data contoh
templates/          template HTML per halaman
static/             CSS, JavaScript, logo, dan foto menu
```
Alur request: **Handler → Service → Repository → database**. SQL hanya ada di repository, aturan bisnis hanya
di service, dan setiap perubahan saldo poin, transaksi, atau kuota voucher berjalan di dalam satu transaksi
database dengan penguncian baris (`SELECT ... FOR UPDATE`).

## Catatan teknis

- Nilai uang disimpan sebagai bilangan bulat rupiah (`BIGINT`), tanpa angka desimal.
- Data master dihapus secara *soft delete* (`deleted_at`); transaksi yang dibatalkan tetap tersimpan dengan status
  `DIBATALKAN`, dan riwayat mutasi poin tidak pernah diubah.
- Jam promo dan tanggal mengikuti WIB (Asia/Jakarta).
- Sesi login disimpan di memori server, jadi pengguna perlu login ulang setelah server dimulai ulang.

## Kredit

Foto menu dari [Unsplash](https://unsplash.com/license); font Inter, Plus Jakarta Sans, Lexend, dan Montserrat
berlisensi SIL Open Font License. Rincian ada di `static/img/KREDIT.md`.
