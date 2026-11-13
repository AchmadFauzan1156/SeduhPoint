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

// HashByUsername hanya dipakai saat login (materi #11): hash password tidak ikut ke struct Pengguna.
func (r *PenggunaRepo) HashByUsername(ctx context.Context, username string) (int, string, error) {
	var id int
	var hash string
	err := r.DB.QueryRowContext(ctx,
		"SELECT id, password_hash FROM pengguna WHERE username = ? AND aktif = TRUE AND deleted_at IS NULL", username,
	).Scan(&id, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", model.ErrNotFound
	}
	if err != nil {
		return 0, "", fmt.Errorf("cari pengguna: %w", err)
	}
	return id, hash, nil
}

// HashByID dipakai saat mengganti password (cek password lama).
func (r *PenggunaRepo) HashByID(ctx context.Context, id int) (string, error) {
	var hash string
	err := r.DB.QueryRowContext(ctx,
		"SELECT password_hash FROM pengguna WHERE id = ? AND aktif = TRUE AND deleted_at IS NULL", id,
	).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", model.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("cari pengguna %d: %w", id, err)
	}
	return hash, nil
}

const kolomPengguna = "p.id, p.username, p.peran, COALESCE(p.cabang_id, 0), COALESCE(c.nama, '')"

func (r *PenggunaRepo) Semua(ctx context.Context) ([]model.Pengguna, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT "+kolomPengguna+` FROM pengguna p
		LEFT JOIN cabang c ON c.id = p.cabang_id
		WHERE p.aktif = TRUE AND p.deleted_at IS NULL ORDER BY p.peran, c.nama, p.username`)
	if err != nil {
		return nil, fmt.Errorf("daftar pengguna: %w", err)
	}
	defer rows.Close()
	var hasil []model.Pengguna
	for rows.Next() {
		var p model.Pengguna
		if err := rows.Scan(&p.ID, &p.Username, &p.Peran, &p.CabangID, &p.CabangNama); err != nil {
			return nil, fmt.Errorf("scan pengguna: %w", err)
		}
		hasil = append(hasil, p)
	}
	return hasil, rows.Err()
}

func (r *PenggunaRepo) CariByID(ctx context.Context, id int) (model.Pengguna, error) {
	var p model.Pengguna
	err := r.DB.QueryRowContext(ctx, "SELECT "+kolomPengguna+` FROM pengguna p
		LEFT JOIN cabang c ON c.id = p.cabang_id
		WHERE p.id = ? AND p.aktif = TRUE AND p.deleted_at IS NULL`, id,
	).Scan(&p.ID, &p.Username, &p.Peran, &p.CabangID, &p.CabangNama)
	if errors.Is(err, sql.ErrNoRows) {
		return p, model.ErrNotFound
	}
	if err != nil {
		return p, fmt.Errorf("cari pengguna %d: %w", id, err)
	}
	return p, nil
}

func (r *PenggunaRepo) UbahPassword(ctx context.Context, id int, hash string) error {
	res, err := r.DB.ExecContext(ctx,
		"UPDATE pengguna SET password_hash = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", hash, id)
	if err != nil {
		return fmt.Errorf("ubah password %d: %w", id, err)
	}
	return pastikanAdaBaris(res)
}
