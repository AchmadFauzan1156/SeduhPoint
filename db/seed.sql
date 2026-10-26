-- Data contoh. Password semua akun: "password" (hash bcrypt, ganti setelah login pertama)
USE seduhpoint;
INSERT INTO cabang (nama, alamat) VALUES
 ('Seduh Sekaran','Jl. Raya Sekaran, Gunungpati'),
 ('Seduh Banaran','Jl. Banaran Raya'),
 ('Seduh Tembalang','Jl. Tembalang Baru');
INSERT INTO pengguna (username, password_hash, peran, cabang_id) VALUES
 ('admin',  '$2a$10$afezlsmSerOEjXWG7g0W4uNcIYlFfMZ4FB4D2oTfwKfiEqBcSAUf.', 'ADMIN', NULL),
 ('kasir1', '$2a$10$afezlsmSerOEjXWG7g0W4uNcIYlFfMZ4FB4D2oTfwKfiEqBcSAUf.', 'KASIR', 1),
 ('kasir2', '$2a$10$afezlsmSerOEjXWG7g0W4uNcIYlFfMZ4FB4D2oTfwKfiEqBcSAUf.', 'KASIR', 2),
 ('kasir3', '$2a$10$afezlsmSerOEjXWG7g0W4uNcIYlFfMZ4FB4D2oTfwKfiEqBcSAUf.', 'KASIR', 3);
INSERT INTO kategori (nama) VALUES ('Kopi'),('Non-Kopi'),('Makanan');
INSERT INTO produk (kategori_id, nama, harga) VALUES
 (1,'Americano',18000),(1,'Kopi Susu Seduh',22000),(1,'Cappuccino',25000),
 (2,'Matcha Latte',27000),(2,'Coklat',23000),
 (3,'Croissant',20000),(3,'Roti Bakar',15000);
INSERT INTO member (nama, no_hp, tgl_daftar) VALUES
 ('Budi','081200000001',CURDATE()),('Sari','081200000002',CURDATE());
INSERT INTO member_tingkat_riwayat (member_id, tingkat, berlaku_sejak) VALUES
 (1,'SILVER',NOW()),(2,'SILVER',NOW());
UPDATE produk SET gambar = CONCAT('/static/img/menu/', CASE nama
  WHEN 'Americano' THEN 'americano' WHEN 'Kopi Susu Seduh' THEN 'kopi-susu' WHEN 'Cappuccino' THEN 'cappuccino'
  WHEN 'Matcha Latte' THEN 'matcha' WHEN 'Coklat' THEN 'coklat' WHEN 'Croissant' THEN 'croissant'
  WHEN 'Roti Bakar' THEN 'roti-bakar' END, '.jpg');
-- Contoh promo 4 jenis (M2.5), sama dengan docs/data-uji.md (P1–P4)
INSERT INTO promo (id, nama, jenis, nilai_persen, nilai_nominal, min_belanja, kode_voucher, kuota, tgl_mulai, tgl_selesai, jam_mulai, jam_selesai, aktif) VALUES
 (1, 'Diskon 20% cappuccino', 'DISKON_PERSEN', 20,   NULL,  NULL,   NULL,      NULL, '2026-10-01', '2026-12-31', NULL,       NULL,       TRUE),
 (2, 'Hemat sore',            'MIN_BELANJA',   NULL, 15000, 100000, NULL,      NULL, '2026-10-01', '2026-12-31', '14:00:00', '17:00:00', TRUE),
 (3, 'Beli 2 gratis 1 kopi',  'B2G1',          NULL, NULL,  NULL,   NULL,      NULL, '2026-10-01', '2026-12-31', NULL,       NULL,       TRUE),
 (4, 'Voucher SEDUH10',       'VOUCHER',       NULL, 10000, NULL,   'SEDUH10', 100,  '2026-10-01', '2026-12-31', NULL,       NULL,       TRUE);
INSERT INTO promo_cabang (promo_id, cabang_id) VALUES (1,1),(1,2),(1,3),(2,1),(2,2),(2,3),(3,1),(3,2),(4,1),(4,2),(4,3);
INSERT INTO promo_produk (promo_id, produk_id) VALUES (1,3);
INSERT INTO promo_kategori (promo_id, kategori_id) VALUES (3,1);
