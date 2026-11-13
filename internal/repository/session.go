package repository

// Session login di database (materi #11).

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"

	"seduhpoint/internal/model"
)

type SessionRepo struct {
	DB *sql.DB
}

// Buat membuat token baru yang berlaku 8 jam.
// Materi memakai 7 hari; di sini 8 jam (satu shift kerja) karena komputer kasir dipakai bergantian.
func (r *SessionRepo) Buat(ctx context.Context, penggunaID int) (string, error) {
	token := rand.Text() // crypto/rand, 128 bit acak
	// Session yang sudah kedaluwarsa dibersihkan sekalian agar tabel tidak terus membesar.
	if _, err := r.DB.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at <= NOW()"); err != nil {
		return "", fmt.Errorf("bersihkan session: %w", err)
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO sessions (token, pengguna_id, expires_at)
		VALUES (?, ?, NOW() + INTERVAL 8 HOUR)`, token, penggunaID)
	if err != nil {
		return "", fmt.Errorf("buat session: %w", err)
	}
	return token, nil
}

// PenggunaDariToken membaca pengguna (termasuk peran) dari database di setiap request.
// Jadi jika admin menonaktifkan akun, aksesnya langsung berhenti di request berikutnya (materi #12).
func (r *SessionRepo) PenggunaDariToken(ctx context.Context, token string) (model.Pengguna, error) {
	var p model.Pengguna
	err := r.DB.QueryRowContext(ctx, `
		SELECT p.id, p.username, p.peran, COALESCE(p.cabang_id, 0)
		FROM sessions s JOIN pengguna p ON p.id = s.pengguna_id
		WHERE s.token = ? AND s.expires_at > NOW() AND p.aktif = TRUE AND p.deleted_at IS NULL`, token,
	).Scan(&p.ID, &p.Username, &p.Peran, &p.CabangID)
	if errors.Is(err, sql.ErrNoRows) {
		return p, model.ErrNotFound
	}
	return p, err
}

func (r *SessionRepo) Hapus(ctx context.Context, token string) error {
	_, err := r.DB.ExecContext(ctx, "DELETE FROM sessions WHERE token = ?", token)
	return err
}

// HapusMilik mengeluarkan pengguna dari semua perangkat, kecuali token yang sedang dipakai
// (kosongkan kecuali untuk menghapus semuanya). Dipakai setelah password diganti.
func (r *SessionRepo) HapusMilik(ctx context.Context, penggunaID int, kecuali string) error {
	_, err := r.DB.ExecContext(ctx, "DELETE FROM sessions WHERE pengguna_id = ? AND token <> ?", penggunaID, kecuali)
	return err
}
