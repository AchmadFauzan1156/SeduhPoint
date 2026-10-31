package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"seduhpoint/internal/model"
)

type PenggunaRepo struct {
	DB *sql.DB
}

func (r *PenggunaRepo) CariByUsername(ctx context.Context, username string) (model.Pengguna, error) {
	var p model.Pengguna
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, username, password_hash, peran, cabang_id FROM pengguna
		 WHERE username = ? AND aktif = TRUE AND deleted_at IS NULL`, username,
	).Scan(&p.ID, &p.Username, &p.PasswordHash, &p.Peran, &p.CabangID)
	if errors.Is(err, sql.ErrNoRows) {
		return p, model.ErrNotFound
	}
	if err != nil {
		return p, fmt.Errorf("cari pengguna: %w", err)
	}
	return p, nil
}

func (r *PenggunaRepo) CariByID(ctx context.Context, id int) (model.Pengguna, error) {
	var p model.Pengguna
	err := r.DB.QueryRowContext(ctx,
		`SELECT id, username, password_hash, peran, cabang_id FROM pengguna
		 WHERE id = ? AND aktif = TRUE AND deleted_at IS NULL`, id,
	).Scan(&p.ID, &p.Username, &p.PasswordHash, &p.Peran, &p.CabangID)
	if errors.Is(err, sql.ErrNoRows) {
		return p, model.ErrNotFound
	}
	if err != nil {
		return p, fmt.Errorf("cari pengguna %d: %w", id, err)
	}
	return p, nil
}

func (r *PenggunaRepo) UbahPassword(ctx context.Context, id int, hash string) error {
	if _, err := r.DB.ExecContext(ctx, "UPDATE pengguna SET password_hash = ? WHERE id = ?", hash, id); err != nil {
		return fmt.Errorf("ubah password %d: %w", id, err)
	}
	return nil
}

// DaftarKasir: akun kasir beserta nama cabangnya (halaman kelola password admin).
func (r *PenggunaRepo) DaftarKasir(ctx context.Context) ([]model.Pengguna, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT p.id, p.username, p.peran, p.cabang_id, COALESCE(c.nama, '')
		FROM pengguna p LEFT JOIN cabang c ON c.id = p.cabang_id
		WHERE p.peran = 'KASIR' AND p.aktif = TRUE AND p.deleted_at IS NULL ORDER BY c.nama, p.username`)
	if err != nil {
		return nil, fmt.Errorf("daftar kasir: %w", err)
	}
	defer rows.Close()
	var hasil []model.Pengguna
	for rows.Next() {
		var p model.Pengguna
		if err := rows.Scan(&p.ID, &p.Username, &p.Peran, &p.CabangID, &p.CabangNama); err != nil {
			return nil, fmt.Errorf("scan kasir: %w", err)
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}
