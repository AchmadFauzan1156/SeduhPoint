# Akun Uji

Akun untuk mencoba dan menguji aplikasi di lingkungan lokal. **Jangan dipakai di server sungguhan.**
Ganti password lewat menu **Pengguna** (admin) sebelum aplikasi dipakai untuk kedai.

## Akun data contoh

Dibuat oleh `db/seduhpoint.sql`, jadi tersedia di setiap database baru.

| Username | Password | Peran | Cabang |
|---|---|---|---|
| `admin` | `password` | Admin / pemilik | semua cabang |
| `kasir1` | `password` | Kasir | Seduh Sekaran |
| `kasir2` | `password` | Kasir | Seduh Banaran |
| `kasir3` | `password` | Kasir | Seduh Tembalang |

## Akun tambahan dari simulasi pengujian

Dibuat lewat aplikasi saat simulasi (lihat [hasil-simulasi.md](hasil-simulasi.md)): setiap cabang baru langsung
dibuatkan satu akun kasir. Akun-akun ini **hanya ada di database yang dipakai untuk simulasi**, tidak ikut
`db/seduhpoint.sql`.

| Username | Password | Peran | Cabang | Status |
|---|---|---|---|---|
| `kasir4` | `Kasir4#Seduh` | Kasir | Seduh Pleburan | aktif |
| `kasir5` | `Kasir5#Seduh` | Kasir | Seduh Ngaliyan | aktif |
| `kasir6` | `Kasir6#Seduh` | Kasir | Seduh Uji Tutup (cabang sudah dihapus) | nonaktif, tidak bisa login |

`kasir6` sengaja dibiarkan nonaktif untuk menunjukkan bahwa menghapus cabang otomatis menonaktifkan
akun kasirnya.

## Kegunaan untuk pengujian

| Yang diuji | Akun |
|---|---|
| Semua fitur admin: data master, promo, laporan, pengguna, gabung/hapus member | `admin` |
| Transaksi dengan semua promo contoh (termasuk beli 2 gratis 1) | `kasir1`, `kasir2` |
| Promo yang tidak berlaku di cabang tertentu (B2G1 dan voucher cabang 1/2 saja) | `kasir3` |
| Kasir hanya melihat dan membatalkan transaksi cabangnya sendiri (404 untuk cabang lain) | `kasir1` lawan `kasir2` |
| Cabang baru yang dibuat bersama akun kasirnya | `kasir4`, `kasir5` |
| Akun dari cabang yang dihapus tidak bisa login | `kasir6` |
