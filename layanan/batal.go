package layanan

import (
	"context"
	"database/sql"

	"seduhpoint/database"
)

// BatalkanTransaksi (F-20 – F-24, AB10). TODO M3
func (l *Layanan) BatalkanTransaksi(ctx context.Context, trxID int, alasan string) error {
	return database.WithTx(ctx, l.DB, func(tx *sql.Tx) error {
		// 1. pastikan status SELESAI, lalu TandaiBatal
		// 2. tarik poin perolehan dari lot transaksi ini -> InsertMutasi(TARIK) (asumsi 8)
		// 3. kembalikan poin yang ditukar ke lot asal -> InsertMutasi(KEMBALI)
		//    lot yang sudah lewat masa berlaku tetap KEDALUWARSA (asumsi 9)
		// 4. kembalikan kuota voucher (kuota_terpakai - 1)
		// 5. EvaluasiTingkat ulang
		return nil
	})
}
