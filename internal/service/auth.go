package service

// Login, logout, dan pengecekan session (materi #11).

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type AuthService struct {
	Pengguna *repository.PenggunaRepo
	Sessions *repository.SessionRepo
}

// hashPalsu dipakai saat username tidak terdaftar. Tambahan dari materi: tanpa ini, username yang
// tidak ada dijawab jauh lebih cepat (bcrypt tidak dijalankan), sehingga username yang terdaftar
// bisa ditebak dari lamanya respons.
var hashPalsu, _ = bcrypt.GenerateFromPassword([]byte("bukan-password"), bcrypt.DefaultCost)

// Login mengembalikan token session baru jika username dan password cocok.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, error) {
	username = strings.TrimSpace(username)

	id, hash, err := s.Pengguna.HashByUsername(ctx, username)
	if errors.Is(err, model.ErrNotFound) {
		bcrypt.CompareHashAndPassword(hashPalsu, []byte(password))
		return "", model.ErrLoginGagal
	}
	if err != nil {
		return "", err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return "", model.ErrLoginGagal
	}
	return s.Sessions.Buat(ctx, id)
}

func (s *AuthService) PenggunaDariToken(ctx context.Context, token string) (model.Pengguna, error) {
	return s.Sessions.PenggunaDariToken(ctx, token)
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.Sessions.Hapus(ctx, token)
}
