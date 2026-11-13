# Hasil Simulasi Pengujian SeduhPoint

**Tanggal:** 13 November 2026, 22.35–23.20 WIB
**Lingkungan:** server lokal; database MySQL 8.4 berisi ±10.000 member dan ±78.000 transaksi data contoh

## Metode

Simulasi dijalankan oleh dua agen penguji otomatis. Masing-masing memakai satu hak akses dan bekerja lewat
aplikasi seperti pengguna sungguhan: request HTTP dengan cookie login, tanpa mengubah database secara langsung.
Database hanya dibaca untuk memverifikasi hasil. Acuan aturannya ada di [aturan-bisnis.md](aturan-bisnis.md).

1. **Admin, fase 1:**
   - menyiapkan data yang valid: cabang, kategori, produk dengan foto asli yang sesuai, promo, dan member
   - menguji semua validasi
   - mencatat angka laporan awal
2. **Kasir:**
   - menjalankan transaksi di cabang 1, 2, dan 3
   - untuk setiap transaksi, hasil yang seharusnya dihitung manual lebih dulu, lalu dicocokkan dengan pratinjau kasir, struk, dan isi tabel database
3. **Admin, fase 2:**
   - mencocokkan laporan dengan hasil kasir
   - menguji promo yang sudah dipakai, pembatalan oleh admin, dan reset password kasir
4. Bug yang ditemukan diperbaiki, lalu diuji ulang.

## Ringkasan hasil

| Bagian | Skenario | Hasil |
|---|---|---|
| Admin fase 1 | 60 | 54 lulus, 3 bug ditemukan (B1, B2, B3) |
| Kasir: transaksi | 22 transaksi | 22 cocok (uang, promo, poin, lot FIFO, tingkat, kuota) |
| Kasir: keranjang, voucher, pembatalan, hak akses, CSRF, member | 29 | 29 lulus |
| Admin fase 2 | 26 | 26 lulus |
| Laporan vs hitungan kasir vs database | 17 angka | 17 cocok |
| Uji ulang setelah perbaikan | 5 bug | semua beres |

## Proses bisnis yang diverifikasi

| Area | Yang dibuktikan |
|---|---|
| Promo produk | <ul><li>Diskon dibulatkan per baris</li><li>Diskon terbesar menang (20% atas 15%)</li><li>Produk yang sudah didiskon tidak ikut B2G1</li><li>B2G1 menggratiskan unit termurah: 4 unit → 1 gratis, 6 unit → 2 gratis</li><li>Promo hanya berlaku di cabang yang ditentukan</li></ul> |
| Promo transaksi | <ul><li>Syarat minimal belanja dihitung setelah diskon produk</li><li>Antara minimal belanja dan voucher, yang potongannya lebih besar dipakai; voucher yang kalah tidak memotong kuota</li><li>Voucher huruf kecil dinormalkan</li><li>Voucher tidak dikenal, voucher di cabang lain, dan voucher dengan kuota habis ditolak</li></ul> |
| Poin | <ul><li>Silver ×1, Gold ×1,5 (dibulatkan ke bawah)</li><li>Tukar poin dibatasi 50% dan saldo; permintaan berlebih dibatasi otomatis</li><li>Pengambilan lot FIFO</li><li>Lot berlaku 6 bulan</li></ul> |
| Tingkat | <ul><li>Naik ke Gold saat belanja 3 bulan ≥ Rp1,5 jt; transaksi itu sendiri masih dihitung sebagai Silver</li><li>Turun ke Silver setelah pembatalan</li><li>Riwayat tingkat tercatat</li></ul> |
| Pembatalan | <ul><li>Alasan wajib, tidak bisa dibatalkan dua kali</li><li>Poin perolehan ditarik, poin tukar dikembalikan ke lot asal</li><li>Kuota voucher kembali dan bisa dipakai lagi</li><li>Poin dari transaksi yang sudah dibatalkan tidak hidup lagi</li><li>Panel dampak sesuai hasil akhir</li></ul> |
| Data master dan promo | <ul><li>Validasi menampilkan semua kesalahan sekaligus</li><li>Foto dicek dari isi file (file teks berekstensi .jpg dan file >2 MB ditolak)</li><li>Foto lama tetap dipakai kalau tidak mengunggah foto baru</li><li>Promo yang sudah dipakai terkunci jenis, nilai, dan targetnya</li><li>Kuota tidak bisa diturunkan di bawah yang sudah terpakai</li></ul> |
| Laporan | <ul><li>Laporan promo, dashboard, poin aktif, dan perkembangan member cocok dengan database</li><li>Transaksi batal tidak dihitung</li><li>Angka langsung berubah setelah pembatalan</li></ul> |
| Hak akses dan keamanan | <ul><li>Kasir 403 di semua halaman dan POST admin</li><li>Admin 403 di halaman kasir</li><li>Transaksi cabang lain 404</li><li>Kasir tidak bisa membatalkan transaksi kemarin</li><li>CSRF ditolak</li><li>Logout dan reset password langsung mencabut sesi</li></ul> |

## Bug yang ditemukan dan diperbaiki

| Kode | Masalah | Perbaikan |
|---|---|---|
| B1 | Promo B2G1 yang disimpan nonaktif bisa diaktifkan lewat tombol toggle tanpa cek bentrok, sehingga pelanggan bisa mendapat gratis dobel | Cek bentrok juga dijalankan saat promo diaktifkan |
| B2 | Nomor `6281…` dianggap berbeda dari `081…`, sehingga satu orang bisa terdaftar dua kali | Awalan `+62`/`62` diubah menjadi `0`; cek unik memeriksa kedua bentuk |
| U2 | Mencari member dengan HP bertanda hubung atau berawalan 62 tidak ketemu | Kata kunci berupa nomor HP dinormalkan dulu |
| K1 | Beberapa request Bayar bersamaan dengan token sama: data benar, tetapi request lain mendapat error 500 | Checkout dicoba ulang untuk error akibat transaksi bersamaan (MySQL 1062/1213/1205). Uji ulang: 5 request paralel → 1 transaksi, 0 error |
| B3 | Dashboard menulis "1760 member perlu diingatkan" padahal member uniknya 1333 | Dihitung per member unik; angkanya kini sama dengan database |

## Catatan yang belum diubah

- Catatan "Belanja RpX lagi untuk …" tidak muncul ketika voucher dipakai.
- Panel pembatalan belum menyebut perubahan tingkat member.
- Poin ×1,5 dari tingkat Gold "sementara" (karena transaksi yang kemudian dibatalkan) tidak dikoreksi. Ini sesuai aturan "tingkat sebelum transaksi", tetapi perlu keputusan tim.
- Nama cabang boleh kembar.
- Kode voucher tidak tampil di daftar promo.
- Pesan untuk harga berupa huruf kurang tepat.
