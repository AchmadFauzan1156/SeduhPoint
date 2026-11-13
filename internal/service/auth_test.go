package service

import (
	"context"
	"errors"
	"testing"

	"seduhpoint/internal/model"
)

// Kasir tidak boleh me-reset password akun lain; ditolak sebelum menyentuh database.
func TestResetPasswordHanyaAdmin(t *testing.T) {
	s := &AuthService{}
	kasir := model.Pengguna{ID: 2, Peran: model.PeranKasir, CabangID: 1}
	if _, err := s.KasirUntukReset(context.Background(), kasir, 3); !errors.Is(err, model.ErrForbidden) {
		t.Errorf("kasir harus ditolak (ErrForbidden), dapat %v", err)
	}
	if err := s.ResetPasswordKasir(context.Background(), kasir, 3, "kopiSeduh8", "kopiSeduh8"); !errors.Is(err, model.ErrForbidden) {
		t.Errorf("kasir harus ditolak (ErrForbidden), dapat %v", err)
	}
}

func TestValidasiPasswordBaru(t *testing.T) {
	cases := []struct {
		nama, baru, ulang string
		fields            []string // kosong = lolos
	}{
		{"ok", "kopiSeduh8", "kopiSeduh8", nil},
		{"terlalu pendek", "kopi1", "kopi1", []string{"baru"}},
		{"password bawaan", "password", "password", []string{"baru"}},
		{"konfirmasi beda", "kopiSeduh8", "kopiSeduh9", []string{"ulang"}},
		{"pendek dan beda sekaligus", "kopi", "kopi1", []string{"baru", "ulang"}},
	}
	for _, c := range cases {
		err := ValidasiPasswordBaru(c.baru, c.ulang)
		var ve *ValidasiError
		if len(c.fields) == 0 {
			if err != nil {
				t.Errorf("%s: harusnya lolos, dapat %v", c.nama, err)
			}
			continue
		}
		if !errors.As(err, &ve) {
			t.Fatalf("%s: harusnya ValidasiError, dapat %v", c.nama, err)
		}
		for _, f := range c.fields {
			if ve.Fields[f] == "" {
				t.Errorf("%s: pesan untuk %q tidak ada: %v", c.nama, f, ve.Fields)
			}
		}
	}
}
