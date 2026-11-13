package service

// Member

import (
	"context"
	"database/sql"
	"regexp"
	"strings"
	"unicode/utf8"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type MemberService struct {
	DB       *sql.DB
	Repo     *repository.MemberRepo
	PoinRepo *repository.PoinRepo
}

// DetailMember: data untuk panel detail.
type DetailMember struct {
	Member     model.Member
	Lot        []model.PoinLot
	MenujuGold int64 // sisa belanja agar Gold; 0 jika sudah memenuhi
	PersenGold int   // 0–100, untuk bar progres
}

// Cari berdasarkan nama atau nomor HP. Jika yang diketik berupa nomor HP (mis. "0812-3456", "+62 812..."),
// nomornya dinormalkan dulu, lalu dicari tanpa angka 0 di depan agar cocok dengan nomor yang tersimpan
// sebagai 08... maupun data lama yang masih 628...
func (s *MemberService) Cari(ctx context.Context, q string) ([]model.Member, error) {
	q = strings.TrimSpace(q)
	if hp := NormalisasiHP(q); reAngka.MatchString(hp) {
		q = strings.TrimPrefix(hp, "0")
	}
	return s.Repo.Cari(ctx, q)
}

var reAngka = regexp.MustCompile(`^[0-9]{4,}$`)

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

var reHP = regexp.MustCompile(`^08[0-9]{8,12}$`)

// NormalisasiHP: buang spasi/tanda hubung/titik, lalu ubah awalan "+62"/"62" menjadi "0",
// sehingga 6281234567890 dan 081234567890 dianggap nomor yang sama (satu orang tidak bisa terdaftar dua kali).
func NormalisasiHP(s string) string {
	s = strings.NewReplacer(" ", "", "-", "", ".", "").Replace(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "+")
	if strings.HasPrefix(s, "62") {
		s = "0" + strings.TrimPrefix(s, "62")
	}
	return s
}

// bentukLamaHP: nomor yang tersimpan sebelum normalisasi di atas bisa masih berawalan 62.
func bentukLamaHP(noHP string) string {
	return "62" + strings.TrimPrefix(noHP, "0")
}

func (s *MemberService) validasi(ctx context.Context, m *model.Member) error {
	m.Nama = strings.TrimSpace(m.Nama)
	m.NoHP = NormalisasiHP(m.NoHP)
	errs := map[string]string{}
	if m.Nama == "" {
		errs["nama"] = "Nama wajib diisi"
	} else if utf8.RuneCountInString(m.Nama) > 100 {
		errs["nama"] = "Nama maksimal 100 karakter"
	}
	if m.NoHP == "" {
		errs["no_hp"] = "Nomor HP wajib diisi"
	} else if !reHP.MatchString(m.NoHP) {
		errs["no_hp"] = "Nomor HP tidak valid (contoh: 081234567890)"
	} else if dipakai, err := s.Repo.NoHPDipakai(ctx, []string{m.NoHP, bentukLamaHP(m.NoHP)}, m.ID); err != nil {
		return err
	} else if dipakai {
		errs["no_hp"] = "Nomor HP sudah terdaftar sebagai member"
	}
	return hasilValidasi(errs)
}

// Buat: mendaftarkan member baru, otomatis SILVER.
func (s *MemberService) Buat(ctx context.Context, m model.Member) (int64, error) {
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

// Hapus member yang salah input atau terdaftar dobel. Hanya admin, dan hanya untuk member yang belum
// pernah bertransaksi: member yang sudah punya transaksi/poin tetap disimpan agar riwayat dan laporan utuh.
// Member dihapus permanen (bukan soft delete) supaya nomor HP-nya bisa dipakai lagi.
func (s *MemberService) Hapus(ctx context.Context, u model.Pengguna, id int) error {
	if !u.IsAdmin() {
		return model.ErrForbidden
	}
	return repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		if _, err := s.Repo.CariByID(ctx, id); err != nil {
			return err
		}
		n, err := s.Repo.JumlahTransaksi(ctx, tx, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return salah("", "Member ini sudah punya riwayat transaksi atau poin, jadi tidak bisa dihapus. Perbaiki datanya lewat Ubah data.")
		}
		return s.Repo.Hapus(ctx, tx, id)
	})
}

func (s *MemberService) Ubah(ctx context.Context, m model.Member) error {
	if err := s.validasi(ctx, &m); err != nil {
		return err
	}
	return s.Repo.Ubah(ctx, m)
}
