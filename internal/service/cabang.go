package service

// Cabang. Setiap cabang punya satu akun kasir yang dipakai bersama di mesin kasir cabang itu;
// akun dibuat admin bersamaan dengan cabangnya. Jika kasirnya berganti orang, admin cukup
// me-reset password akun tersebut di halaman Pengguna.

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type CabangService struct {
	DB           *sql.DB
	Repo         *repository.CabangRepo
	PenggunaRepo *repository.PenggunaRepo
}

// AkunKasir: isian akun kasir pada form cabang.
type AkunKasir struct {
	Username string
	Password string
	Ulang    string
}

func (a AkunKasir) Kosong() bool { return a.Username == "" && a.Password == "" && a.Ulang == "" }

func (s *CabangService) Daftar(ctx context.Context) ([]model.Cabang, error) {
	return s.Repo.Semua(ctx)
}

func (s *CabangService) Detail(ctx context.Context, id int) (model.Cabang, error) {
	return s.Repo.CariByID(ctx, id)
}

// KasirCabang: username akun kasir cabang ini ("" jika belum punya).
func (s *CabangService) KasirCabang(ctx context.Context, id int) (string, error) {
	return s.PenggunaRepo.KasirCabang(ctx, id)
}

// Buat menyimpan cabang baru sekaligus akun kasirnya dalam satu database transaction:
// kalau salah satu gagal, keduanya batal, jadi tidak ada cabang tanpa kasir.
func (s *CabangService) Buat(ctx context.Context, c model.Cabang, akun AkunKasir) (int64, error) {
	errs := map[string]string{}
	validasiCabang(&c, errs)
	if err := s.validasiAkun(ctx, &akun, errs); err != nil {
		return 0, err
	}
	if err := hasilValidasi(errs); err != nil {
		return 0, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(akun.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	var id int64
	err = repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		var err error
		if id, err = s.Repo.Tambah(ctx, tx, c); err != nil {
			return err
		}
		return s.PenggunaRepo.TambahKasir(ctx, tx, akun.Username, string(hash), id)
	})
	return id, err
}

// Ubah menyimpan perubahan cabang. Untuk cabang lama yang belum punya kasir,
// akun kasir boleh dibuat di sini (opsional: isian kosong berarti tidak membuat akun).
func (s *CabangService) Ubah(ctx context.Context, c model.Cabang, akun AkunKasir) error {
	errs := map[string]string{}
	validasiCabang(&c, errs)
	buatAkun := !akun.Kosong()
	if buatAkun {
		if ada, err := s.PenggunaRepo.KasirCabang(ctx, c.ID); err != nil {
			return err
		} else if ada != "" {
			errs["username"] = "Cabang ini sudah punya akun kasir (" + ada + ")"
		} else if err := s.validasiAkun(ctx, &akun, errs); err != nil {
			return err
		}
	}
	if err := hasilValidasi(errs); err != nil {
		return err
	}
	if !buatAkun {
		return s.Repo.Ubah(ctx, c)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(akun.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.Repo.Ubah(ctx, c); err != nil {
		return err
	}
	return s.PenggunaRepo.TambahKasir(ctx, s.DB, akun.Username, string(hash), int64(c.ID))
}

// Hapus cabang (soft delete) sekaligus menonaktifkan akun kasirnya, supaya akun itu tidak bisa login lagi.
func (s *CabangService) Hapus(ctx context.Context, id int) error {
	return repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		if err := s.Repo.Hapus(ctx, tx, id); err != nil {
			return err
		}
		return s.PenggunaRepo.NonaktifkanKasirCabang(ctx, tx, id)
	})
}

func validasiCabang(c *model.Cabang, errs map[string]string) {
	c.Nama, c.Alamat = strings.TrimSpace(c.Nama), strings.TrimSpace(c.Alamat)
	if c.Nama == "" {
		errs["nama"] = "Nama cabang wajib diisi"
	} else if utf8.RuneCountInString(c.Nama) > 100 {
		errs["nama"] = "Nama cabang maksimal 100 karakter"
	}
	if utf8.RuneCountInString(c.Alamat) > 255 {
		errs["alamat"] = "Alamat maksimal 255 karakter"
	}
}

var reUsername = regexp.MustCompile(`^[a-z0-9_]{4,30}$`)

// validasiAkun: username 4–30 huruf kecil/angka/garis bawah dan belum dipakai; password sesuai aturan password baru.
func (s *CabangService) validasiAkun(ctx context.Context, a *AkunKasir, errs map[string]string) error {
	a.Username = strings.ToLower(strings.TrimSpace(a.Username))
	if !reUsername.MatchString(a.Username) {
		errs["username"] = "Username 4–30 karakter: huruf kecil, angka, atau garis bawah (mis. kasir_pleburan)"
	} else if dipakai, err := s.PenggunaRepo.UsernameDipakai(ctx, a.Username); err != nil {
		return err
	} else if dipakai {
		errs["username"] = "Username " + a.Username + " sudah dipakai"
	}
	if ve, ok := ValidasiPasswordBaru(a.Password, a.Ulang).(*ValidasiError); ok {
		for f, p := range ve.Fields {
			errs["kasir_"+f] = p // kasir_baru, kasir_ulang
		}
	}
	return nil
}
