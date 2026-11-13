# Akun Uji

Akun untuk mencoba dan menguji aplikasi di lingkungan lokal. **Jangan dipakai di server sungguhan.**
Ganti password lewat menu **Pengguna** (admin) sebelum aplikasi dipakai untuk kedai.

Pola penamaan: akun kasir bernama `kasir_<cabang>` dengan password `<Cabang>#2026`; admin memakai `admin123`.

## Akun data contoh

Dibuat oleh `db/seduhpoint.sql`, jadi tersedia di setiap database baru.

| Username | Password | Peran | Cabang |
|---|---|---|---|
| `admin` | `admin123` | Admin / pemilik | semua cabang |
| `kasir_sekaran` | `Sekaran#2026` | Kasir | Seduh Sekaran |
| `kasir_banaran` | `Banaran#2026` | Kasir | Seduh Banaran |
| `kasir_tembalang` | `Tembalang#2026` | Kasir | Seduh Tembalang |

## Akun tambahan dari simulasi pengujian

Dibuat lewat aplikasi saat simulasi (lihat [hasil-simulasi.md](hasil-simulasi.md)): setiap cabang baru langsung
dibuatkan satu akun kasir. Akun-akun ini **hanya ada di database yang dipakai untuk simulasi**, tidak ikut
`db/seduhpoint.sql`.

| Username | Password | Peran | Cabang | Status |
|---|---|---|---|---|
| `kasir_pleburan` | `Pleburan#2026` | Kasir | Seduh Pleburan | aktif |
| `kasir_ngaliyan` | `Ngaliyan#2026` | Kasir | Seduh Ngaliyan | aktif |
| `kasir_uji_tutup` | `UjiTutup#2026` | Kasir | Seduh Uji Tutup (cabang sudah dihapus) | nonaktif, tidak bisa login |

`kasir_uji_tutup` sengaja dibiarkan nonaktif untuk menunjukkan bahwa menghapus cabang otomatis menonaktifkan
akun kasirnya.

## Kegunaan untuk pengujian

| Yang diuji | Akun |
|---|---|
| Semua fitur admin: data master, promo, laporan, pengguna, gabung/hapus member | `admin` |
| Transaksi dengan semua promo contoh (termasuk beli 2 gratis 1) | `kasir_sekaran`, `kasir_banaran` |
| Promo yang tidak berlaku di cabang tertentu (B2G1 hanya untuk cabang Sekaran dan Banaran) | `kasir_tembalang` |
| Kasir hanya melihat dan membatalkan transaksi cabangnya sendiri (404 untuk cabang lain) | `kasir_sekaran` lawan `kasir_banaran` |
| Cabang baru yang dibuat bersama akun kasirnya | `kasir_pleburan`, `kasir_ngaliyan` |
| Akun dari cabang yang dihapus tidak bisa login | `kasir_uji_tutup` |
