# ERD Final SeduhPoint (M1.3)

Gambar: `docs/erd.png` (sumber Mermaid: `docs/erd.mmd`, dibuat otomatis dari `db/schema.sql`).
Untuk laporan: ganti Gambar 1.1 dengan `erd.png`. Keterangan: PK = primary key, FK = foreign key,
UK = unique, "null" = boleh kosong. Semua tabel InnoDB, utf8mb4.

## Ringkasan 15 tabel
| Kelompok | Tabel | Catatan |
|---|---|---|
| Master | cabang, kategori, produk, member, pengguna, promo | soft delete (`created_at`, `updated_at`, `deleted_at`) |
| Relasi promo | promo_cabang, promo_produk, promo_kategori | PK gabungan (promo_id + target) |
| Riwayat tingkat | member_tingkat_riwayat | setiap perubahan Silver/Gold (F-26) |
| Transaksi | transaksi, transaksi_item, transaksi_promo | snapshot harga & potongan; status SELESAI/DIBATALKAN |
| Poin | poin_lot, poin_mutasi | lot untuk FIFO & kedaluwarsa; mutasi = buku besar (hanya INSERT) |

## Indeks (selain PK & FK)
| Tabel | Indeks | Untuk |
|---|---|---|
| pengguna | `username` UNIQUE | login |
| kategori | `nama` UNIQUE | nama kategori tidak dobel |
| member | `no_hp` UNIQUE, `idx_member_nama (nama)` | F-03, pencarian F-04 |
| promo | `kode_voucher` UNIQUE, `idx_promo_periode (aktif, tgl_mulai, tgl_selesai)` | validasi voucher, cari promo berlaku |
| transaksi | `no_transaksi` UNIQUE, `token` UNIQUE, `idx_trx_waktu (waktu, cabang_id)`, `idx_trx_member (member_id, status, waktu)` | struk, anti submit ganda, laporan, total belanja 3 bulan |
| transaksi_promo | `idx_tp_promo (promo_id)` | laporan biaya promo |
| poin_lot | `idx_lot_fifo (member_id, status, tgl_kedaluwarsa)` | FIFO (AB04), job kedaluwarsa |
| poin_mutasi | `idx_mutasi_member (member_id, waktu)`, `idx_mutasi_trx (transaksi_id)` | riwayat poin, pembatalan |
| member_tingkat_riwayat | `idx_mtr_member (member_id, berlaku_sejak)` | laporan perkembangan member |

## Hasil review (perubahan dari ERD kasar di laporan)
| No | Temuan | Keputusan |
|---|---|---|
| 1 | Data master di laporan belum punya jejak waktu | Ditambah `created_at`, `updated_at`, `deleted_at` (soft delete, materi #10) |
| 2 | Belum ada pencegah transaksi ganda saat tombol Bayar ditekan dua kali | Ditambah `transaksi.token` UNIQUE |
| 3 | Uang di laporan belum ditentukan tipenya | Semua nilai uang `BIGINT` rupiah (NF-04) |
| 4 | Status lot poin belum dirinci | `poin_lot.status`: AKTIF, HABIS, KEDALUWARSA, DITARIK |
| 5 | Pengembalian poin ke "lot asal" saat pembatalan | Cukup dari `poin_mutasi` jenis TUKAR (berisi `lot_id` + `transaksi_id`), tidak perlu tabel baru |
| 6 | "Total setelah promo" untuk tingkat Gold | Tidak disimpan terpisah: dihitung `subtotal - total_promo` (asumsi 1 & 4) |

## Catatan untuk didiskusikan tim
- `member.no_hp` UNIQUE juga berlaku untuk member yang sudah dihapus (soft delete). Jika nomor yang sama mendaftar ulang, pulihkan data lama alih-alih membuat baru.
- Format `no_transaksi` usulan: `SP-<id cabang>-<yyyymmdd>-<nomor urut>`, contoh `SP-1-20261102-0007`.
