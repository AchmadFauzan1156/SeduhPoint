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

// ---- Kelola pengguna & password (khusus admin, materi #12) ----

func (s *AuthService) DaftarPengguna(ctx context.Context) ([]model.Pengguna, error) {
	return s.Pengguna.Semua(ctx)
}

// KasirUntukReset mengambil akun kasir sekaligus memastikan u boleh me-reset password-nya.
func (s *AuthService) KasirUntukReset(ctx context.Context, u model.Pengguna, id int) (model.Pengguna, error) {
	if !u.IsAdmin() {
		return model.Pengguna{}, model.ErrForbidden
	}
	k, err := s.Pengguna.CariByID(ctx, id)
	if err != nil {
		return k, err
	}
	if !k.IsKasir() { // password admin diganti lewat form "password saya" yang meminta password lama
		return k, model.ErrNotFound
	}
	return k, nil
}

// ValidasiPasswordBaru: aturan password baru. Panjang 8–72 sama dengan materi #11 (bcrypt maks 72 byte).
func ValidasiPasswordBaru(baru, ulang string) error {
	errs := map[string]string{}
	switch {
	case len(baru) < 8 || len(baru) > 72:
		errs["baru"] = "Password 8 sampai 72 karakter"
	case baru == "password":
		errs["baru"] = "Password ini terlalu mudah ditebak"
	}
	if baru != ulang {
		errs["ulang"] = "Konfirmasi password tidak sama"
	}
	return hasilValidasi(errs)
}

// GantiPassword: pengguna mengganti password-nya sendiri (wajib password lama).
// Session di perangkat lain dihapus, session yang sedang dipakai (token) tetap.
func (s *AuthService) GantiPassword(ctx context.Context, u model.Pengguna, token, lama, baru, ulang string) error {
	hash, err := s.Pengguna.HashByID(ctx, u.ID)
	if err != nil {
		return err
	}
	errs := map[string]string{}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(lama)) != nil {
		errs["lama"] = "Password lama salah"
	}
	var ve *ValidasiError
	if err := ValidasiPasswordBaru(baru, ulang); errors.As(err, &ve) {
		for f, p := range ve.Fields {
			errs[f] = p
		}
	}
	if _, ada := errs["baru"]; !ada && lama == baru {
		errs["baru"] = "Password baru harus berbeda dari password lama"
	}
	if err := hasilValidasi(errs); err != nil {
		return err
	}
	if err := s.simpanPassword(ctx, u.ID, baru); err != nil {
		return err
	}
	return s.Sessions.HapusMilik(ctx, u.ID, token)
}

// ResetPasswordKasir: admin mengatur password baru kasir tanpa password lama;
// kasir otomatis keluar dari semua perangkat.
func (s *AuthService) ResetPasswordKasir(ctx context.Context, u model.Pengguna, kasirID int, baru, ulang string) error {
	if _, err := s.KasirUntukReset(ctx, u, kasirID); err != nil {
		return err
	}
	if err := ValidasiPasswordBaru(baru, ulang); err != nil {
		return err
	}
	if err := s.simpanPassword(ctx, kasirID, baru); err != nil {
		return err
	}
	return s.Sessions.HapusMilik(ctx, kasirID, "")
}

func (s *AuthService) simpanPassword(ctx context.Context, id int, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Pengguna.UbahPassword(ctx, id, string(hash))
}
