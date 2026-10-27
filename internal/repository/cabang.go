package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"seduhpoint/internal/model"
)

type CabangRepo struct {
	DB *sql.DB
}

const kolomCabang = "id, nama, alamat, aktif"

func (r *CabangRepo) Semua(ctx context.Context) ([]model.Cabang, error) {
	rows, err := r.DB.QueryContext(ctx,
		"SELECT "+kolomCabang+" FROM cabang WHERE deleted_at IS NULL ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("daftar cabang: %w", err)
	}
	defer rows.Close()
	var hasil []model.Cabang
	for rows.Next() {
		var c model.Cabang
		if err := rows.Scan(&c.ID, &c.Nama, &c.Alamat, &c.Aktif); err != nil {
			return nil, fmt.Errorf("scan cabang: %w", err)
		}
		hasil = append(hasil, c)
	}
	return hasil, rows.Err()
}

func (r *CabangRepo) CariByID(ctx context.Context, id int) (model.Cabang, error) {
	var c model.Cabang
	err := r.DB.QueryRowContext(ctx,
		"SELECT "+kolomCabang+" FROM cabang WHERE id = ? AND deleted_at IS NULL", id,
	).Scan(&c.ID, &c.Nama, &c.Alamat, &c.Aktif)
	if errors.Is(err, sql.ErrNoRows) {
		return c, model.ErrNotFound
	}
	if err != nil {
		return c, fmt.Errorf("cari cabang %d: %w", id, err)
	}
	return c, nil
}

func (r *CabangRepo) Tambah(ctx context.Context, c model.Cabang) (int64, error) {
	res, err := r.DB.ExecContext(ctx,
		"INSERT INTO cabang (nama, alamat, aktif) VALUES (?, ?, ?)", c.Nama, c.Alamat, c.Aktif)
	if err != nil {
		return 0, fmt.Errorf("tambah cabang: %w", err)
	}
	return res.LastInsertId()
}

func (r *CabangRepo) Ubah(ctx context.Context, c model.Cabang) error {
	res, err := r.DB.ExecContext(ctx,
		"UPDATE cabang SET nama=?, alamat=?, aktif=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL",
		c.Nama, c.Alamat, c.Aktif, c.ID)
	if err != nil {
		return fmt.Errorf("ubah cabang %d: %w", c.ID, err)
	}
	return pastikanAdaBaris(res)
}

// Hapus = soft delete (data lama tetap ada untuk riwayat transaksi).
func (r *CabangRepo) Hapus(ctx context.Context, id int) error {
	res, err := r.DB.ExecContext(ctx,
		"UPDATE cabang SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return fmt.Errorf("hapus cabang %d: %w", id, err)
	}
	return pastikanAdaBaris(res)
}
