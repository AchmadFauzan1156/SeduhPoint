-- =====================================================================
-- SeduhPoint: seluruh query database dalam satu file
--   Bagian 1: membuat database & 15 tabel
--   Bagian 2: data contoh (3 cabang, akun admin/kasir, menu, member, 4 promo)
--
-- Jalankan dari folder SeduhPoint (MySQL Laragon sudah Start All):
--   mysql -u root -e "DROP DATABASE IF EXISTS seduhpoint"        (opsional: mengulang dari awal)
--   mysql -u root --default-character-set=utf8mb4 -e "source db/seduhpoint.sql"
-- Perintah ini sama di PowerShell maupun CMD. Jika "mysql" tidak dikenali, pakai Terminal Laragon
-- (Menu > Terminal) atau path lengkap, mis. E:\laragon\bin\mysql\mysql-8.4.3-winx64\bin\mysql.exe
--
-- Uang = BIGINT rupiah tanpa desimal. Riwayat transaksi & mutasi poin tidak dihapus.
-- Data master memakai soft delete (deleted_at).
-- =====================================================================

-- ---------- Bagian 1: database & tabel ----------
CREATE DATABASE IF NOT EXISTS seduhpoint CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE seduhpoint;

CREATE TABLE cabang (
  id      INT AUTO_INCREMENT PRIMARY KEY,
  nama    VARCHAR(100) NOT NULL,
  alamat  VARCHAR(255) NOT NULL DEFAULT '',
  aktif   BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE pengguna (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  username      VARCHAR(50) NOT NULL UNIQUE,
  password_hash VARCHAR(100) NOT NULL,
  peran         ENUM('ADMIN','KASIR') NOT NULL,
  cabang_id     INT NULL,                       -- wajib untuk KASIR
  aktif         BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (cabang_id) REFERENCES cabang(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE kategori (
  id   INT AUTO_INCREMENT PRIMARY KEY,
  nama VARCHAR(100) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE produk (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  kategori_id INT NOT NULL,
  nama        VARCHAR(100) NOT NULL,
  harga       BIGINT NOT NULL CHECK (harga >= 0),
  gambar      VARCHAR(255) NULL,            -- path foto, mis. /static/img/menu/cappuccino.jpg
  deskripsi   VARCHAR(300) NULL,            -- penjelasan untuk kasir/pelanggan
  aktif       BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (kategori_id) REFERENCES kategori(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE member (
  id               INT AUTO_INCREMENT PRIMARY KEY,
  nama             VARCHAR(100) NOT NULL,
  no_hp            VARCHAR(20) NOT NULL UNIQUE,
  tgl_daftar       DATE NOT NULL,
  tingkat_saat_ini ENUM('SILVER','GOLD') NOT NULL DEFAULT 'SILVER',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  INDEX idx_member_nama (nama)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE member_tingkat_riwayat (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  member_id     INT NOT NULL,
  tingkat       ENUM('SILVER','GOLD') NOT NULL,
  berlaku_sejak DATETIME NOT NULL,
  FOREIGN KEY (member_id) REFERENCES member(id),
  INDEX idx_mtr_member (member_id, berlaku_sejak)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE promo (
  id             INT AUTO_INCREMENT PRIMARY KEY,
  nama           VARCHAR(100) NOT NULL,
  jenis          ENUM('DISKON_PERSEN','MIN_BELANJA','B2G1','VOUCHER') NOT NULL,
  nilai_persen   INT NULL,          -- DISKON_PERSEN (0-100)
  nilai_nominal  BIGINT NULL,       -- MIN_BELANJA / VOUCHER
  min_belanja    BIGINT NULL,       -- MIN_BELANJA
  kode_voucher   VARCHAR(30) NULL UNIQUE,
  kuota          INT NULL,
  kuota_terpakai INT NOT NULL DEFAULT 0,
  tgl_mulai      DATE NOT NULL,
  tgl_selesai    DATE NOT NULL,
  jam_mulai      TIME NULL,         -- opsional
  jam_selesai    TIME NULL,
  catatan        VARCHAR(200) NULL,  -- kalimat tambahan dari admin; ketentuan utama dibuat otomatis
  aktif          BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NULL,
  deleted_at TIMESTAMP NULL,
  INDEX idx_promo_periode (aktif, tgl_mulai, tgl_selesai)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE promo_cabang (
  promo_id  INT NOT NULL,
  cabang_id INT NOT NULL,
  PRIMARY KEY (promo_id, cabang_id),
  FOREIGN KEY (promo_id) REFERENCES promo(id),
  FOREIGN KEY (cabang_id) REFERENCES cabang(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE promo_produk (
  promo_id  INT NOT NULL,
  produk_id INT NOT NULL,
  PRIMARY KEY (promo_id, produk_id),
  FOREIGN KEY (promo_id) REFERENCES promo(id),
  FOREIGN KEY (produk_id) REFERENCES produk(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE promo_kategori (
  promo_id    INT NOT NULL,
  kategori_id INT NOT NULL,
  PRIMARY KEY (promo_id, kategori_id),
  FOREIGN KEY (promo_id) REFERENCES promo(id),
  FOREIGN KEY (kategori_id) REFERENCES kategori(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE transaksi (
  id             INT AUTO_INCREMENT PRIMARY KEY,
  no_transaksi   VARCHAR(30) NOT NULL UNIQUE,   -- format: SP-<cabang>-<yyyymmdd>-<urut>
  token          VARCHAR(64) NULL UNIQUE,       -- anti submit ganda (V08)
  cabang_id      INT NOT NULL,
  pengguna_id    INT NOT NULL,
  member_id      INT NULL,
  waktu          DATETIME NOT NULL,
  subtotal       BIGINT NOT NULL,
  total_promo    BIGINT NOT NULL DEFAULT 0,
  poin_ditukar   INT NOT NULL DEFAULT 0,
  potongan_poin  BIGINT NOT NULL DEFAULT 0,
  total_bayar    BIGINT NOT NULL,
  poin_diperoleh INT NOT NULL DEFAULT 0,
  status         ENUM('SELESAI','DIBATALKAN') NOT NULL DEFAULT 'SELESAI',
  alasan_batal   VARCHAR(255) NULL,
  waktu_batal    DATETIME NULL,
  FOREIGN KEY (cabang_id) REFERENCES cabang(id),
  FOREIGN KEY (pengguna_id) REFERENCES pengguna(id),
  FOREIGN KEY (member_id) REFERENCES member(id),
  INDEX idx_trx_waktu (waktu, cabang_id),
  INDEX idx_trx_member (member_id, status, waktu)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE transaksi_item (
  id            INT AUTO_INCREMENT PRIMARY KEY,
  transaksi_id  INT NOT NULL,
  produk_id     INT NOT NULL,
  qty           INT NOT NULL CHECK (qty > 0),
  harga_satuan  BIGINT NOT NULL,   -- snapshot harga saat transaksi
  FOREIGN KEY (transaksi_id) REFERENCES transaksi(id),
  FOREIGN KEY (produk_id) REFERENCES produk(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE transaksi_promo (
  id             INT AUTO_INCREMENT PRIMARY KEY,
  transaksi_id   INT NOT NULL,
  promo_id       INT NOT NULL,
  nilai_potongan BIGINT NOT NULL,
  FOREIGN KEY (transaksi_id) REFERENCES transaksi(id),
  FOREIGN KEY (promo_id) REFERENCES promo(id),
  INDEX idx_tp_promo (promo_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE poin_lot (
  id              INT AUTO_INCREMENT PRIMARY KEY,
  member_id       INT NOT NULL,
  transaksi_id    INT NOT NULL,
  poin_awal       INT NOT NULL,
  poin_sisa       INT NOT NULL,
  tgl_perolehan   DATETIME NOT NULL,
  tgl_kedaluwarsa DATETIME NOT NULL,   -- tgl_perolehan + 6 bulan
  status          ENUM('AKTIF','HABIS','KEDALUWARSA','DITARIK') NOT NULL DEFAULT 'AKTIF',
  FOREIGN KEY (member_id) REFERENCES member(id),
  FOREIGN KEY (transaksi_id) REFERENCES transaksi(id),
  INDEX idx_lot_fifo (member_id, status, tgl_kedaluwarsa)  -- FIFO
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE poin_mutasi (
  id           INT AUTO_INCREMENT PRIMARY KEY,
  member_id    INT NOT NULL,
  lot_id       INT NOT NULL,
  transaksi_id INT NULL,
  jenis        ENUM('PEROLEHAN','TUKAR','KEMBALI','TARIK','KEDALUWARSA') NOT NULL,
  jumlah       INT NOT NULL,            -- + menambah, - mengurangi
  waktu        DATETIME NOT NULL,
  FOREIGN KEY (member_id) REFERENCES member(id),
  FOREIGN KEY (lot_id) REFERENCES poin_lot(id),
  FOREIGN KEY (transaksi_id) REFERENCES transaksi(id),
  INDEX idx_mutasi_member (member_id, waktu),
  INDEX idx_mutasi_trx (transaksi_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- ---------- Bagian 2: data contoh ----------
-- Data contoh. Password semua akun: "password" (hash bcrypt, ganti setelah login pertama)
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
UPDATE produk SET deskripsi = CASE nama
  WHEN 'Americano' THEN 'Espresso dua shot dan air panas. Pahit ringan, tanpa gula dan susu. Bisa panas atau dingin.'
  WHEN 'Kopi Susu Seduh' THEN 'Menu andalan: espresso, susu segar, dan gula aren. Manis legit, disajikan dingin.'
  WHEN 'Cappuccino' THEN 'Espresso dengan susu dan busa susu tebal, seimbang antara kopi dan susu. Disajikan panas.'
  WHEN 'Matcha Latte' THEN 'Matcha Jepang dan susu segar, manis ringan. Tanpa kopi, cocok untuk yang tidak minum kafein kopi.'
  WHEN 'Coklat' THEN 'Coklat hitam dan susu, kental dan tidak terlalu manis. Tanpa kopi.'
  WHEN 'Croissant' THEN 'Croissant mentega berlapis, renyah di luar dan lembut di dalam. Dipanggang setiap pagi.'
  WHEN 'Roti Bakar' THEN 'Roti tebal dipanggang dengan mentega dan selai blueberry.' END;

-- Contoh promo, satu dari tiap jenis
INSERT INTO promo (id, nama, jenis, nilai_persen, nilai_nominal, min_belanja, kode_voucher, kuota, tgl_mulai, tgl_selesai, jam_mulai, jam_selesai, aktif) VALUES
 (1, 'Diskon 20% cappuccino', 'DISKON_PERSEN', 20,   NULL,  NULL,   NULL,      NULL, '2026-10-01', '2026-12-31', NULL,       NULL,       TRUE),
 (2, 'Hemat sore',            'MIN_BELANJA',   NULL, 15000, 100000, NULL,      NULL, '2026-10-01', '2026-12-31', '14:00:00', '17:00:00', TRUE),
 (3, 'Beli 2 gratis 1 kopi',  'B2G1',          NULL, NULL,  NULL,   NULL,      NULL, '2026-10-01', '2026-12-31', NULL,       NULL,       TRUE),
 (4, 'Voucher Rp10.000',      'VOUCHER',       NULL, 10000, NULL,   'SEDUH10', 100,  '2026-10-01', '2026-12-31', NULL,       NULL,       TRUE);
INSERT INTO promo_cabang (promo_id, cabang_id) VALUES (1,1),(1,2),(1,3),(2,1),(2,2),(2,3),(3,1),(3,2),(4,1),(4,2),(4,3);
INSERT INTO promo_produk (promo_id, produk_id) VALUES (1,3);
INSERT INTO promo_kategori (promo_id, kategori_id) VALUES (3,1);

-- Catatan promo untuk kasir (tampil di panel "Promo berlaku sekarang")
UPDATE promo SET catatan = 'Tawarkan ke pelanggan yang membeli 2 kopi.' WHERE id = 3;
UPDATE promo SET catatan = 'Kode dibagikan ke pelanggan lewat kartu member; kasir tidak menyebutkan kodenya.' WHERE id = 4;
