package service

// F-03, F-04, F-27 — TODO M2

import (
	"context"
	"database/sql"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type MemberService struct {
	DB       *sql.DB
	Repo     *repository.MemberRepo
	PoinRepo *repository.PoinRepo
}

func (s *MemberService) Cari(ctx context.Context, q string) ([]model.Member, error) {
	return s.Repo.Cari(ctx, q)
}
func (s *MemberService) Detail(ctx context.Context, id int) (model.Member, error) {
	return s.Repo.CariByID(ctx, id)
}

// Daftar: validasi nama wajib, no HP angka 10–15 digit & unik; tingkat awal SILVER (F-03).
func (s *MemberService) Daftar(ctx context.Context, m model.Member) (int64, error) {
	var id int64
	err := repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		var err error
		id, err = s.Repo.Tambah(ctx, tx, m)
		return err
	})
	return id, err
}
func (s *MemberService) Ubah(ctx context.Context, m model.Member) error { return s.Repo.Ubah(ctx, m) }
