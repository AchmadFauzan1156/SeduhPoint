-- SeduhPoint — skema MySQL 8 (berdasarkan Tabel 1.10 laporan)
-- Uang = BIGINT rupiah (NF-04). Riwayat tidak dihapus (NF-05).
-- Data master memakai soft delete (deleted_at), sesuai materi #10.
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
  cabang_id     INT NULL,                       -- wajib untuk KASIR (F-05)
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
  jam_mulai      TIME NULL,         -- opsional (F-07)
  jam_selesai    TIME NULL,
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
  tgl_kedaluwarsa DATETIME NOT NULL,   -- tgl_perolehan + 6 bulan (AB02)
  status          ENUM('AKTIF','HABIS','KEDALUWARSA','DITARIK') NOT NULL DEFAULT 'AKTIF',
  FOREIGN KEY (member_id) REFERENCES member(id),
  FOREIGN KEY (transaksi_id) REFERENCES transaksi(id),
  INDEX idx_lot_fifo (member_id, status, tgl_kedaluwarsa)  -- FIFO (AB04)
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
