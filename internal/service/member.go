package service

// F-03, F-04, F-27 Member

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type MemberService struct {
	DB       *sql.DB
	Repo     *repository.MemberRepo
	PoinRepo *repository.PoinRepo
}

// DetailMember: data untuk panel detail (F-27).
type DetailMember struct {
	Member     model.Member
	Lot        []model.PoinLot
	MenujuGold int64 // sisa belanja agar Gold; 0 jika sudah memenuhi
	PersenGold int   // 0–100, untuk bar progres
}

func (s *MemberService) Cari(ctx context.Context, q string) ([]model.Member, error) {
	return s.Repo.Cari(ctx, strings.TrimSpace(q))
}

func (s *MemberService) Hitung(ctx context.Context) (total, gold int, err error) {
	return s.Repo.Hitung(ctx)
}

func (s *MemberService) Detail(ctx context.Context, id int) (model.Member, error) {
	return s.Repo.CariByID(ctx, id)
}

func (s *MemberService) DetailLengkap(ctx context.Context, id int) (DetailMember, error) {
	var d DetailMember
	var err error
	if d.Member, err = s.Repo.CariByID(ctx, id); err != nil {
		return d, err
	}
	if d.Lot, err = s.PoinRepo.LotMember(ctx, id); err != nil {
		return d, err
	}
	if d.Member.Belanja3Bulan < BatasGold {
		d.MenujuGold = BatasGold - d.Member.Belanja3Bulan
	}
	d.PersenGold = int(min(100, d.Member.Belanja3Bulan*100/BatasGold))
	return d, nil
}

var reHP = regexp.MustCompile(`^(08|628)[0-9]{8,12}$`)

// NormalisasiHP: buang spasi/tanda hubung, "+62" -> "62".
func NormalisasiHP(s string) string {
	s = strings.NewReplacer(" ", "", "-", "", ".", "").Replace(strings.TrimSpace(s))
	return strings.TrimPrefix(s, "+")
}

func (s *MemberService) validasi(ctx context.Context, m *model.Member) error {
	m.Nama = strings.TrimSpace(m.Nama)
	m.NoHP = NormalisasiHP(m.NoHP)
	switch {
	case m.Nama == "":
		return salah("nama", "Nama wajib diisi")
	case len([]rune(m.Nama)) > 100:
		return salah("nama", "Nama maksimal 100 karakter")
	case m.NoHP == "":
		return salah("no_hp", "Nomor HP wajib diisi")
	case !reHP.MatchString(m.NoHP):
		return salah("no_hp", "Nomor HP tidak valid (contoh: 081234567890)")
	}
	if dipakai, err := s.Repo.NoHPDipakai(ctx, m.NoHP, m.ID); err != nil {
		return err
	} else if dipakai {
		return salah("no_hp", "Nomor HP sudah terdaftar sebagai member")
	}
	return nil
}

// Daftar: member baru otomatis SILVER (F-03).
func (s *MemberService) Daftar(ctx context.Context, m model.Member) (int64, error) {
	if err := s.validasi(ctx, &m); err != nil {
		return 0, err
	}
	var id int64
	err := repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		var err error
		id, err = s.Repo.Tambah(ctx, tx, m)
		return err
	})
	return id, err
}

func (s *MemberService) Ubah(ctx context.Context, m model.Member) error {
	if err := s.validasi(ctx, &m); err != nil {
		return err
	}
	return s.Repo.Ubah(ctx, m)
}
