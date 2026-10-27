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
