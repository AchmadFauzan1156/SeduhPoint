package service

// F-05, NF-03 (SEMENTARA — disesuaikan dengan materi Authentication dari dosen)

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type AuthService struct {
	Repo *repository.PenggunaRepo
}

var ErrLoginGagal = &ValidasiError{"Username atau password salah"}

func (s *AuthService) Login(ctx context.Context, username, password string) (model.Pengguna, error) {
	p, err := s.Repo.CariByUsername(ctx, username)
	if errors.Is(err, model.ErrNotFound) {
		return p, ErrLoginGagal
	}
	if err != nil {
		return p, err
	}
	if bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(password)) != nil {
		return p, ErrLoginGagal
	}
	return p, nil
}
