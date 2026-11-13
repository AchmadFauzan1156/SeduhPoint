# Aturan Bisnis SeduhPoint

Acuan perhitungan transaksi, poin, pembatalan, hak akses, dan validasi. Dokumen ini juga menjadi acuan
saat simulasi pengujian (lihat [hasil-simulasi.md](hasil-simulasi.md)).

## Perhitungan transaksi (urutan)

1. **Subtotal** = Σ harga × qty. Harga selalu diambil dari database, qty maksimal 99 per produk, dan produk nonaktif tidak bisa dijual.
2. **Promo produk**
   - **Diskon persen:** potongan per baris produk target = floor(harga × qty × persen / 100).
     - Jika satu produk kena beberapa diskon, yang dipakai adalah persen terbesar.
     - Produk yang sudah didiskon tidak didiskon lagi.
   - **Beli 2 gratis 1 (B2G1):** berlaku untuk kategori target, hanya pada item yang tidak didiskon.
     - Semua unit diurutkan dari harga termurah.
     - Setiap 3 unit, 1 unit termurah gratis.
3. **Satu promo tingkat transaksi:** minimal belanja **atau** voucher, tidak bisa digabung.
   - Syarat minimal belanja dihitung dari subtotal setelah promo produk.
   - Potongan = min(nominal, total).
   - Jika keduanya memenuhi, yang potongannya lebih besar dipakai.
   - Jika sama besar, minimal belanja yang dipakai dan voucher tidak terpakai (kuotanya tidak berkurang).
   - Voucher harus aktif, dalam tanggal/jam berlaku, berlaku di cabang itu, dan kuotanya belum habis.
4. **Total setelah promo** = subtotal − semua potongan promo.
5. **Tukar poin** (member saja)
   - 1 poin = Rp1.000.
   - Maksimal tukar = min(saldo poin aktif, floor(total setelah promo / 2 / 1.000)). Permintaan di atas batas otomatis dibatasi.
   - Poin diambil dari lot yang paling dekat kedaluwarsa (FIFO).
6. **Total bayar** = total setelah promo − poin ditukar × Rp1.000.
7. **Poin diperoleh** (member saja) = floor(total setelah promo / 10.000).
   - Gold mendapat ×1,5, dibulatkan ke bawah.
   - Tingkat yang dipakai adalah tingkat member sebelum transaksi itu.
   - Poin berlaku 6 bulan.

**Tingkat member:**
- Gold jika total belanja yang SELESAI (subtotal − promo) dalam 3 bulan terakhir ≥ Rp1.500.000; selain itu Silver.
- Tingkat dievaluasi setelah setiap transaksi dan pembatalan, serta oleh job harian.

**Lain-lain:**
- Tombol Bayar yang ditekan berkali-kali (dengan token yang sama) hanya menghasilkan satu transaksi.
- Format nomor transaksi: `SP-<cabang>-<yyyymmdd>-<urut 4 digit>`.

## Pembatalan

- Transaksi tidak dihapus. Statusnya menjadi DIBATALKAN dengan alasan wajib (maksimal 255 karakter).
- Efek pembatalan:
  - Poin yang diperoleh dari transaksi itu ditarik (sisa yang belum terpakai).
  - Poin yang ditukar dikembalikan ke lot asalnya, kecuali lot asal sudah DITARIK karena transaksi asalnya juga dibatalkan.
  - Kuota voucher dikembalikan.
  - Tingkat member dievaluasi ulang.
- Kasir hanya boleh membatalkan transaksi cabangnya sendiri pada hari yang sama (WIB). Admin boleh membatalkan semua transaksi.

## Hak akses

| Aksi | Tamu | Kasir | Admin |
|---|:---:|:---:|:---:|
| Login | ✓ | ✓ | ✓ |
| Halaman kasir dan pembayaran | | ✓ | |
| Riwayat dan struk | | cabang sendiri | semua |
| Batalkan transaksi | | cabang sendiri, hari ini | semua |
| Member: lihat, daftar, ubah | | ✓ | ✓ |
| Hapus member salah input/dobel (belum pernah bertransaksi) | | | ✓ |
| Cabang (beserta akun kasirnya), kategori, produk, promo, laporan, pengguna | | | ✓ |

Setiap cabang punya satu akun kasir yang dibuat admin bersamaan dengan cabangnya dan dipakai bersama di mesin
kasir cabang itu. Kasir berganti orang cukup dengan reset password di halaman Pengguna. Menghapus cabang
menonaktifkan akun kasirnya.

Respons HTTP:
- Belum login: diarahkan ke `/login`.
- Tidak berhak: 403.
- Data cabang lain: 404.
- POST dari situs lain ditolak (CSRF).

## Validasi

- **Member**
  - Nama wajib, maksimal 100 karakter.
  - Nomor HP wajib dan unik. Spasi, tanda hubung, dan titik dibuang; awalan `+62`/`62` diubah menjadi `0`. Format akhirnya `08` diikuti 8–12 angka.
- **Produk**
  - Nama wajib, maksimal 100 karakter.
  - Harga lebih dari 0, maksimal Rp10.000.000, kelipatan Rp100.
  - Kategori wajib; deskripsi maksimal 300 karakter.
  - Foto opsional: JPG/PNG/WebP (dicek dari isi file), maksimal 2 MB.
- **Promo**
  - Nama dan jenis wajib.
  - Diskon: 1–90% dan minimal 1 produk.
  - Minimal belanja: nilai minimal lebih dari 0, potongan lebih dari 0 dan lebih kecil dari nilai minimal.
  - B2G1:
    - Minimal 1 kategori.
    - Tidak boleh bentrok dengan B2G1 aktif lain (kategori dan cabang sama, periode beririsan).
    - Larangan bentrok ini juga dicek saat promo diaktifkan lewat tombol toggle.
  - Voucher: kode 4–20 huruf/angka dan unik, potongan lebih dari 0, kuota minimal 1 dan tidak kurang dari yang sudah terpakai.
  - Tanggal wajib, dan tanggal selesai tidak boleh sebelum tanggal mulai.
  - Jam opsional, tetapi harus diisi berpasangan dan jam mulai < jam selesai.
  - Minimal 1 cabang.
  - Promo yang sudah dipakai transaksi: jenis, nilai, target, dan kodenya terkunci.
- **Password baru:** 8–72 karakter, bukan `password`, dan konfirmasinya harus sama.
- Semua kesalahan isian ditampilkan sekaligus di dekat field-nya (HTTP 422).
