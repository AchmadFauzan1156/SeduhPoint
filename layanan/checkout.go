package layanan

import (
	"context"
	"database/sql"

	"seduhpoint/database"
	"seduhpoint/hitung"
)

type PermintaanCheckout struct {
	CabangID        int
	PenggunaID      int
	MemberID        int // 0 = non-member (asumsi 11)
	Items           []hitung.Item
	KodeVoucher     string
	PakaiMinBelanja bool
	PoinDiminta     int
	Token           string // anti submit ganda
}

// Pratinjau menghitung total tanpa menyimpan (untuk halaman kasir). TODO M3
func (l *Layanan) Pratinjau(ctx context.Context, p PermintaanCheckout) (hitung.Hasil, error) {
	return hitung.Hasil{}, nil
}

// Checkout menyimpan transaksi (Laporan 1.8 langkah 8). TODO M3
func (l *Layanan) Checkout(ctx context.Context, p PermintaanCheckout) (int64, error) {
	var trxID int64
	err := database.WithTx(ctx, l.DB, func(tx *sql.Tx) error {
		// 1. ambil promo berlaku; voucher -> GetVoucherForUpdate, cek kuota, kuota_terpakai+1
		// 2. lot poin member -> LotAktifFIFOForUpdate
		// 3. hitung.Hitung(...)
		// 4. InsertTransaksi, InsertItem (snapshot harga), InsertTransaksiPromo
		// 5. tukar poin: kurangi lot FIFO + InsertMutasi(TUKAR) per lot
		// 6. perolehan poin: InsertLot (kedaluwarsa +6 bln) + InsertMutasi(PEROLEHAN)
		// 7. EvaluasiTingkat(tx, memberID)
		return nil
	})
	return trxID, err
}
