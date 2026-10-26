package store

import "seduhpoint/models"

// F-12 – F-24 — TODO M3. Semua fungsi tulis menerima Querier (dipanggil dalam tx).
func (s *Store) InsertTransaksi(q Querier, t models.Transaksi) (int64, error)            { return 0, nil }
func (s *Store) InsertItem(q Querier, trxID int64, produkID, qty int, harga int64) error { return nil }
func (s *Store) InsertTransaksiPromo(q Querier, trxID int64, promoID int, nilai int64) error {
	return nil
}
func (s *Store) GetTransaksi(id int) (models.Transaksi, error)          { return models.Transaksi{}, nil }
func (s *Store) ListTransaksi(cabangID int) ([]models.Transaksi, error) { return nil, nil }
func (s *Store) TandaiBatal(q Querier, id int, alasan string) error     { return nil }
