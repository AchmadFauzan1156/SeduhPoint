# Sketsa UI SeduhPoint (M1.5)

Warna turunan template Figma Community **"Coffee Shop Website Design – E-commerce Template"**
(primary #723E29, ink #311103, font Inter). Tata letak mengikuti panduan *app-screens* dari
TasteSkill (design-taste-frontend) karena SeduhPoint adalah aplikasi kasir/admin, bukan toko online.
Logo: `static/img/logo.svg` (gelas kopi + "Seduh" / "POINT"). Foto: Unsplash License (`static/img/KREDIT.md`).

| File | Halaman | Kebutuhan |
|---|---|---|
| 01-login.png | Masuk | F-05 |
| 02-ringkasan.png | Ringkasan pemilik | ringkasan |
| 03-kasir.png | Transaksi baru (menu bergambar + panel pesanan) | F-12 – F-18 |
| 04-struk.png | Struk | F-19 |
| 05-riwayat-batal.png | Riwayat + pembatalan | F-20 – F-24 |
| 06-member.png | Member + detail (saldo, lot FIFO, menuju Gold) | F-03, F-04, F-27 |
| 07-produk.png | Produk (pola sama untuk Cabang & Kategori) | F-01, F-02 |
| 08-promo.png | Promo & voucher | F-06 – F-11 |
| 09-laporan.png | Biaya promo, poin, member | F-28 – F-30 |

## Review desain (TasteSkill: inspect.mjs --app + critique app-screens)

Inspector: **0 FAIL** di 9 halaman (1440px & 390px). Sisa WARN dan alasannya:
- `ragged-row` di daftar + panel detail (member, produk, promo, kasir): panel detail sengaja memakai
  tinggi isinya sendiri (`align-items: start`), sesuai saran app-screens §4.2; panel pesanan kasir *sticky*.
- `hero-void` di struk & riwayat: halaman memang pendek; dengan data nyata tabel akan terisi.
- Login `panel-void`: hilang setelah foto login diunduh.

Perbaikan dari putaran review:
1. Tombol & input tanpa bayangan (dulu `box-shadow` sebagai garis tepi) → border 1px datar.
2. Latar krem kekuningan → stone netral #F3F2F0; tint aktif #EAE7E5 (tidak kekuningan).
3. Kontras teks abu (header tabel, placeholder, label grup) dinaikkan ke ≥ 4.5:1.
4. Halaman ringkasan & laporan tidak lagi menyisakan ruang kosong besar (tabel transaksi terakhir & per cabang).
5. Tampilan HP: sidebar dilipat jadi bar + tombol Menu, kolom bertumpuk, menu 2 kolom.
6. Angka memakai digit tabular di seluruh halaman; teks kecil minimal 13px.

Skor kritik (juri: desain 40%, kegunaan 30%, kreativitas 20%, konten 10%), sebelum foto asli:
desain 7 · kegunaan 8 · kreativitas 6.5 · konten 7.5 → **7.2**. Target ≥ 7.5 setelah foto menu & login masuk
(batas TasteSkill: < 4 foto asli membatasi skor desain 6.5).
