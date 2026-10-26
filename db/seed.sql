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
-- TODO M2: tambahkan contoh promo (4 jenis) + promo_cabang/produk/kategori
