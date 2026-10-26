# Use Case SeduhPoint (M1.4)

Aktor: **Admin/Pemilik**, **Kasir**, **Sistem (otomatis)** (Tabel 1.1 laporan).

## Daftar use case
| Kode | Use case | Aktor | Kebutuhan |
|---|---|---|---|
| UC-01 | Login & logout | Admin, Kasir | F-05 |
| UC-02 | Kelola cabang | Admin | F-01 |
| UC-03 | Kelola kategori & produk | Admin | F-02 |
| UC-04 | Daftar / ubah member | Admin, Kasir | F-03 |
| UC-05 | Cari member & lihat saldo poin | Admin, Kasir | F-04, F-27 |
| UC-06 | Kelola promo & voucher | Admin | F-06 – F-11 |
| UC-07 | Buat transaksi penjualan | Kasir | F-12 – F-19 |
| UC-08 | Batalkan transaksi | Kasir, Admin | F-20 – F-24 |
| UC-09 | Kedaluwarsakan poin | Sistem | F-25 |
| UC-10 | Evaluasi tingkat member | Sistem | F-26 |
| UC-11 | Lihat laporan biaya promo | Admin | F-28 |
| UC-12 | Lihat laporan poin aktif | Admin | F-29 |
| UC-13 | Lihat laporan perkembangan member | Admin | F-30 |

## Skenario use case utama

### UC-07 Buat transaksi penjualan
**Prasyarat:** kasir sudah login dan terikat ke satu cabang.

**Alur utama**
1. Kasir membuka halaman Transaksi Baru.
2. Kasir (opsional) mencari member berdasarkan nama atau no. HP.
3. Kasir memilih produk dan jumlahnya.
4. Sistem menerapkan otomatis promo produk yang berlaku (periode, cabang, jam WIB): diskon persen dan beli 2 gratis 1.
5. Kasir (opsional) memilih **satu** dari: potongan minimal belanja, atau kode voucher.
6. Jika pelanggan member, kasir (opsional) mengisi jumlah poin yang ditukar.
7. Kasir menekan **Hitung**; sistem menampilkan subtotal, rincian potongan per promo, potongan poin, total bayar, dan poin yang akan didapat.
8. Kasir menekan **Bayar**; sistem menyimpan transaksi dalam satu database transaction, mengurangi poin FIFO, menambah lot poin baru, menambah pemakaian voucher, lalu mengevaluasi tingkat member.
9. Sistem menampilkan struk.

**Alur alternatif / pengecualian**
- 5a. Potongan minimal belanja dan voucher dipilih bersamaan → ditolak (AB09).
- 5b. Voucher tidak ada / di luar periode / bukan untuk cabang ini / di luar jam / kuota habis → ditolak dengan pesan sesuai penyebab.
- 5c. Syarat minimal belanja tidak terpenuhi (dihitung setelah promo produk) → potongan tidak diterapkan.
- 6a. Poin diminta melebihi batas (saldo aktif atau 50% total setelah promo) → sistem membatasi ke nilai maksimum dan menampilkannya.
- 6b. Pelanggan non-member → kolom poin tidak tersedia.
- 8a. Tombol Bayar ditekan dua kali → token form mencegah transaksi ganda.

### UC-08 Batalkan transaksi
**Prasyarat:** transaksi berstatus SELESAI.
1. Kasir/admin membuka struk lalu memilih **Batalkan**.
2. Sistem menampilkan dampak: poin yang akan ditarik, poin yang dikembalikan, kuota voucher yang dikembalikan.
3. Kasir mengisi alasan (wajib) dan mengonfirmasi.
4. Dalam satu database transaction: status → DIBATALKAN; lot poin dari transaksi ini ditarik; poin yang ditukar dikembalikan ke lot asal; kuota voucher dikurangi 1; tingkat member dievaluasi ulang.

**Pengecualian:** transaksi sudah DIBATALKAN → ditolak. Poin perolehan sudah terpakai di transaksi lain → lihat asumsi 8 (perlu diskusi).

### UC-09 / UC-10 Proses otomatis harian
Saat aplikasi menyala dan setiap 24 jam: lot AKTIF yang lewat `tgl_kedaluwarsa` → KEDALUWARSA + mutasi; semua member dievaluasi ulang Silver/Gold dari total belanja SELESAI 3 bulan terakhir.

## Asumsi tambahan (lanjutan Tabel 1.11, perlu dikonfirmasi ke dosen)
| No | Hal yang ambigu | Asumsi sementara |
|---|---|---|
| 12 | Tingkat yang dipakai untuk menghitung poin jika transaksi itu membuat member naik ke Gold | Poin dihitung dengan tingkat **sebelum** transaksi; tingkat baru berlaku untuk transaksi berikutnya |
| 13 | Satu item kena diskon persen **dan** termasuk kategori beli 2 gratis 1 | Item yang sudah kena diskon persen tidak ikut dihitung pada beli 2 gratis 1 |
| 14 | Pembulatan diskon persen | Dibulatkan ke bawah ke rupiah per baris item: floor(harga × qty × persen / 100) |
| 15 | Kasir meminta tukar poin melebihi batas | Dibatasi otomatis ke maksimum, bukan ditolak |
