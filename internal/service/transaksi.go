package service

// F-12 – F-24 — TODO M3

import (
	"context"
	"database/sql"

	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type TransaksiService struct {
	DB        *sql.DB
	Repo      *repository.TransaksiRepo
	PromoRepo *repository.PromoRepo
	PoinRepo  *repository.PoinRepo
	Member    *MemberService
	Tingkat   *PoinService
}

type PermintaanCheckout struct {
	CabangID        int
	PenggunaID      int
	MemberID        int // 0 = non-member (asumsi 11)
	Items           []Item
	KodeVoucher     string
	PakaiMinBelanja bool
	PoinDiminta     int
	Token           string // anti submit ganda (V08)
}

// Pratinjau menghitung total tanpa menyimpan (tombol "Hitung" di kasir).
func (s *TransaksiService) Pratinjau(ctx context.Context, p PermintaanCheckout) (Hasil, error) {
	// TODO: ambil promo berlaku + cek jam (PromoBerlaku), saldo poin, lalu Hitung()
	return Hasil{}, nil
}

// Checkout menyimpan transaksi (Laporan 1.8 langkah 8) dalam satu transaction.
func (s *TransaksiService) Checkout(ctx context.Context, p PermintaanCheckout) (int64, error) {
	var trxID int64
	err := repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		// 1. promo berlaku; voucher -> PromoRepo.VoucherForUpdate, cek kuota, TambahPemakaian(+1)
		// 2. lot poin member -> PoinRepo.LotAktifFIFO (FOR UPDATE)
		// 3. Hitung(...)
		// 4. Repo.Tambah, TambahItem (snapshot harga), TambahPromo
		// 5. tukar poin: kurangi lot FIFO + TambahMutasi(TUKAR) per lot
		// 6. perolehan: TambahLot (kedaluwarsa +6 bln) + TambahMutasi(PEROLEHAN)
		// 7. Tingkat.Evaluasi(ctx, tx, memberID) — setelah poin dihitung (asumsi 12)
		return nil
	})
	return trxID, err
}

func (s *TransaksiService) Daftar(ctx context.Context, cabangID int) ([]model.Transaksi, error) {
	return s.Repo.Semua(ctx, cabangID)
}
func (s *TransaksiService) Detail(ctx context.Context, id int) (model.Transaksi, error) {
	return s.Repo.CariByID(ctx, id)
}

// Batalkan (F-20 – F-24, AB10).
func (s *TransaksiService) Batalkan(ctx context.Context, id int, alasan string) error {
	if alasan == "" {
		return salah("alasan", "Alasan pembatalan wajib diisi")
	}
	return repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		// 1. pastikan status SELESAI -> Repo.TandaiBatal
		// 2. tarik poin perolehan (TARIK) — asumsi 8
		// 3. kembalikan poin tukar ke lot asal (KEMBALI); lot kedaluwarsa tetap kedaluwarsa (asumsi 9)
		// 4. kembalikan kuota voucher (TambahPemakaian -1)
		// 5. Tingkat.Evaluasi ulang
		return nil
	})
}
