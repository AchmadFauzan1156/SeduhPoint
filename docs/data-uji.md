# Data Uji Perhitungan (M1.4)

Dipakai untuk pengujian manual dan unit test (`internal/service/hitung_test.go`).
Semua angka dihitung manual mengikuti alur 1.8 laporan dan asumsi 1–15.

## Data acuan (sama dengan `db/seed.sql`)
| ID | Produk | Kategori | Harga |
|---|---|---|---|
| 1 | Americano | Kopi | 18.000 |
| 2 | Kopi Susu Seduh | Kopi | 22.000 |
| 3 | Cappuccino | Kopi | 25.000 |
| 4 | Matcha Latte | Non-Kopi | 27.000 |
| 5 | Coklat | Non-Kopi | 23.000 |
| 6 | Croissant | Makanan | 20.000 |
| 7 | Roti Bakar | Makanan | 15.000 |

| Kode | Promo | Ketentuan |
|---|---|---|
| P1 | Diskon persen | 20% untuk Cappuccino |
| P2 | Potongan minimal belanja | Min. Rp100.000 → potong Rp15.000 |
| P3 | Beli 2 gratis 1 | Kategori Kopi |
| P4 | Voucher `SEDUH10` | Potong Rp10.000, kuota 2 |

## A. Perhitungan promo & poin
| No | Kasus | Member | Promo aktif | Item | Subtotal | Potongan promo | Total setelah promo | Tukar poin | Total bayar | Poin didapat |
|---|---|---|---|---|---|---|---|---|---|---|
| T01 | Non-member tanpa promo | – | – | 2 Americano, 1 Croissant | 56.000 | 0 | 56.000 | – | 56.000 | 0 |
| T02 | Silver tanpa promo | Silver | – | 2 Kopi Susu, 1 Roti Bakar | 59.000 | 0 | 59.000 | 0 | 59.000 | 5 |
| T03 | Gold 1,5× dibulatkan ke bawah | Gold | – | 2 Kopi Susu, 1 Roti Bakar | 59.000 | 0 | 59.000 | 0 | 59.000 | 7 (5×1,5=7,5) |
| T04 | Diskon persen | Silver | P1 | 2 Cappuccino, 1 Matcha | 77.000 | P1 10.000 | 67.000 | 0 | 67.000 | 6 |
| T05 | Beli 2 gratis 1 | Silver | P3 | 2 Americano, 1 Kopi Susu | 58.000 | P3 18.000 | 40.000 | 0 | 40.000 | 4 |
| T06 | B2G1, 5 item → 1 gratis | Silver | P3 | 1 Americano, 2 Kopi Susu, 2 Cappuccino | 112.000 | P3 18.000 | 94.000 | 0 | 94.000 | 9 |
| T07 | Min. belanja terpenuhi | Silver | P2 | 3 Matcha, 1 Croissant | 101.000 | P2 15.000 | 86.000 | 0 | 86.000 | 8 |
| T08 | Min. belanja gugur setelah promo produk | Silver | P1, P2 | 4 Cappuccino | 100.000 | P1 20.000 (P2 tidak berlaku: 80.000 < 100.000) | 80.000 | 0 | 80.000 | 8 |
| T09 | Voucher | Silver | P4 | 2 Coklat, 1 Roti Bakar | 61.000 | P4 10.000 | 51.000 | 0 | 51.000 | 5 |
| T10 | Promo bertumpuk: produk + voucher | Silver | P1, P4 | 2 Cappuccino, 1 Croissant | 70.000 | P1 10.000 + P4 10.000 | 50.000 | 0 | 50.000 | 5 |
| T11 | Gold + min. belanja | Gold | P2 | 3 Matcha, 1 Croissant | 101.000 | P2 15.000 | 86.000 | 0 | 86.000 | 12 (8×1,5) |
| T12 | Tukar poin dibatasi 50% | Silver, saldo 40 | – | 2 Coklat, 1 Roti Bakar | 61.000 | 0 | 61.000 | diminta 40 → maks 30 (Rp30.000) | 31.000 | 6 (dari 61.000, asumsi 1) |
| T13 | Tukar poin dibatasi saldo | Silver, saldo 12 | – | 3 Matcha, 1 Croissant | 101.000 | 0 | 101.000 | diminta 50 → maks 12 (Rp12.000) | 89.000 | 10 |
| T14 | Diskon persen dibulatkan | Silver | P1 diubah 15% | 1 Cappuccino | 25.000 | P1 3.750 | 21.250 | 0 | 21.250 | 2 |
| T15 | Item diskon tidak ikut B2G1 (asumsi 13) | Silver | P1, P3 | 2 Cappuccino, 2 Americano | 86.000 | P1 10.000; P3 0 (hanya 2 item Kopi non-diskon) | 76.000 | 0 | 76.000 | 7 |

## B. Validasi yang harus ditolak
| No | Kasus | Kondisi | Hasil yang diharapkan |
|---|---|---|---|
| V01 | Min. belanja + voucher bersamaan | 4 Matcha (108.000), pilih P2 dan `SEDUH10` | Ditolak: "Potongan minimal belanja tidak dapat digabung dengan voucher" |
| V02 | Kuota voucher habis | `SEDUH10` kuota 2, terpakai 2 | Ditolak: "Kuota voucher habis" |
| V03 | Voucher di luar periode | Voucher berakhir 25 Okt 2026, transaksi 26 Okt 2026 | Ditolak: "Voucher tidak berlaku" |
| V04 | Promo di luar jam | P2 berlaku 14:00–17:00 WIB, transaksi 13:59 | P2 tidak diterapkan |
| V05 | Promo di jam berlaku | P2 berlaku 14:00–17:00 WIB, transaksi 15:00 | P2 diterapkan |
| V06 | Promo beda cabang | P2 hanya cabang Sekaran, transaksi di Banaran | P2 tidak diterapkan |
| V07 | Non-member tukar poin | Tanpa member, isi poin 10 | Kolom poin diabaikan / ditolak |
| V08 | Submit ganda | Token form yang sama dikirim 2× | Hanya 1 transaksi tersimpan |

## C. Poin FIFO, kedaluwarsa, dan pembatalan
Kondisi awal member Budi (tanggal uji 26 Okt 2026):

| Lot | Poin sisa | Kedaluwarsa | Status |
|---|---|---|---|
| L0 | 5 | 20 Okt 2026 | KEDALUWARSA |
| L1 | 10 | 15 Nov 2026 | AKTIF |
| L2 | 20 | 1 Jan 2027 | AKTIF |
| L3 | 15 | 1 Mar 2027 | AKTIF |

Saldo aktif = **45 poin** (L0 tidak dihitung) = Rp45.000.

| No | Kasus | Aksi | Hasil yang diharapkan |
|---|---|---|---|
| F01 | Tukar FIFO | Transaksi TX-A, total setelah promo 101.000, tukar 25 poin | L1 10→0 (HABIS), L2 20→5, L3 tetap 15. Mutasi: TUKAR −10 (L1), TUKAR −15 (L2). Total bayar 76.000. Poin didapat 10 → lot baru L4 (kedaluwarsa 26 Apr 2027). Saldo = 0+5+15+10 = **30** |
| F02 | Batalkan TX-A | Alasan "salah input" | L4 10→0 (DITARIK), mutasi TARIK −10; L1 0→10 dan L2 5→20, mutasi KEMBALI +10 dan +15. Saldo kembali **45** |
| F03 | Batal saat lot asal sudah kedaluwarsa | Seperti F02 tetapi dibatalkan 20 Nov 2026 (L1 lewat 15 Nov) | L1 dikembalikan +10 tetapi tetap KEDALUWARSA (asumsi 9); saldo aktif = 20+15 = **35** |
| F04 | Proses harian | Job jalan 16 Nov 2026 | L1 (masih AKTIF) → KEDALUWARSA, mutasi KEDALUWARSA −sisa |
| F05 | Kuota voucher kembali | Batalkan transaksi yang memakai `SEDUH10` (terpakai 2) | kuota_terpakai 2 → 1 |
| F06 | Transaksi batal tidak dihitung | Laporan & tingkat setelah F02 | TX-A tidak muncul di laporan biaya promo maupun total belanja 3 bulan |

## D. Tingkat member
| No | Kasus | Kondisi | Hasil yang diharapkan |
|---|---|---|---|
| G01 | Naik ke Gold | Total belanja SELESAI 3 bulan terakhir 1.480.000; transaksi baru total setelah promo 25.000 | Total 1.505.000 ≥ 1.500.000 → GOLD, riwayat dicatat. Poin transaksi ini masih dihitung Silver = 2 (asumsi 12) |
| G02 | Transaksi berikutnya sebagai Gold | Total setelah promo 40.000 | Poin = 4 × 1,5 = **6** |
| G03 | Pembatalan tetapi masih memenuhi | Batalkan transaksi G01 | Total 1.480.000 + 40.000 = 1.520.000 → tetap GOLD |
| G04 | Turun karena jendela 3 bulan | Belanja lama keluar dari jendela 3 bulan sehingga total < 1.500.000 | Job harian mengubah ke SILVER, riwayat dicatat |
| G05 | Tepat di batas | Total tepat 1.500.000 | GOLD (syarat "minimal") |

## E. Laporan pemilik (M4)

Data: 5 transaksi tanggal 27 Okt 2026 (2 di Sekaran, 2 di Banaran, 1 di Tembalang), 1 transaksi Banaran dibatalkan.

| ID | Skenario | Langkah | Hasil yang diharapkan |
|---|---|---|---|
| L01 | Biaya promo semua cabang (F-28) | Laporan biaya promo 1–27 Okt, semua cabang | Total potongan Rp65.000: Beli 2 gratis 1 Rp40.000 (2×), Diskon 20% Rp15.000 (2×), Voucher Rp10.000 (1×) |
| L02 | Transaksi dibatalkan tidak dihitung (F-24) | Bandingkan sebelum & sesudah pembatalan | Transaksi batal hanya muncul di kolom "Batal"; omzet & potongan tidak berubah karena batal |
| L03 | Filter cabang | Pilih Seduh Banaran | Hanya potongan transaksi Banaran (Rp22.000) |
| L04 | Rentang terbalik | Dari 27 Okt, sampai 1 Okt | Ditolak: "Tanggal 'sampai' tidak boleh sebelum 'dari'" |
| L05 | Poin aktif (F-29) | Buka laporan poin aktif | Total poin AKTIF yang belum kedaluwarsa, nilai = poin × Rp1.000 |
| L06 | Poin akan kedaluwarsa | Satu lot kedaluwarsa 12 hari lagi | Muncul di "Akan kedaluwarsa dalam 30 hari" beserta nama member & tanggalnya |
| L07 | Perkembangan member (F-30) | Member daftar Agustus, naik Gold September, turun Silver Oktober | Agustus baru 1; September naik +1, Gold 1; Oktober turun −1 (diuji otomatis di `laporan_test.go`) |
