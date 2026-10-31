package service

// Login: cek username & password (bcrypt). Password dikelola admin:
// GantiPassword untuk akun admin sendiri, ResetPasswordKasir untuk akun kasir.

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

var ErrLoginGagal = &ValidasiError{Pesan: "Username atau password salah"}

// hashPalsu dipakai saat username tidak ada, supaya waktu respons sama dengan password salah
// (username yang terdaftar tidak bisa ditebak dari lama respons).
var hashPalsu, _ = bcrypt.GenerateFromPassword([]byte("bukan-password"), bcrypt.DefaultCost)

func (s *AuthService) Login(ctx context.Context, username, password string) (model.Pengguna, error) {
	p, err := s.Repo.CariByUsername(ctx, username)
	if errors.Is(err, model.ErrNotFound) {
		bcrypt.CompareHashAndPassword(hashPalsu, []byte(password))
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

// ValidasiPasswordBaru: aturan password baru (tanpa database, bisa dites).
func ValidasiPasswordBaru(lama, baru, ulang string) error {
	switch {
	case len(baru) < 8:
		return salah("baru", "Password baru minimal 8 karakter")
	case len(baru) > 72: // batas bcrypt
		return salah("baru", "Password baru maksimal 72 karakter")
	case baru == lama:
		return salah("baru", "Password baru harus berbeda dari password lama")
	case baru == "password":
		return salah("baru", "Password ini terlalu mudah ditebak")
	case baru != ulang:
		return salah("ulang", "Konfirmasi password tidak sama")
	}
	return nil
}

func (s *AuthService) GantiPassword(ctx context.Context, penggunaID int, lama, baru, ulang string) error {
	p, err := s.Repo.CariByID(ctx, penggunaID)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(lama)) != nil {
		return salah("lama", "Password lama salah")
	}
	if err := ValidasiPasswordBaru(lama, baru, ulang); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(baru), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Repo.UbahPassword(ctx, penggunaID, string(hash))
}

func (s *AuthService) DaftarKasir(ctx context.Context) ([]model.Pengguna, error) {
	return s.Repo.DaftarKasir(ctx)
}

func (s *AuthService) Kasir(ctx context.Context, id int) (model.Pengguna, error) {
	p, err := s.Repo.CariByID(ctx, id)
	if err == nil && p.Peran != "KASIR" {
		return p, model.ErrNotFound
	}
	return p, err
}

// ResetPasswordKasir: admin mengatur password baru kasir tanpa perlu password lamanya.
func (s *AuthService) ResetPasswordKasir(ctx context.Context, kasirID int, baru, ulang string) error {
	if _, err := s.Kasir(ctx, kasirID); err != nil {
		return err
	}
	if err := ValidasiPasswordBaru("", baru, ulang); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(baru), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Repo.UbahPassword(ctx, kasirID, string(hash))
}
