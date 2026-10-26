package store

import "seduhpoint/models"

// Poin per lot (FIFO) + buku besar mutasi — TODO M3

// LotAktifFIFOForUpdate: lot AKTIF member urut tgl_kedaluwarsa ASC, FOR UPDATE (AB04).
func (s *Store) LotAktifFIFOForUpdate(q Querier, memberID int) ([]models.PoinLot, error) {
	return nil, nil
}
func (s *Store) SaldoPoin(memberID int) (int, error)                           { return 0, nil }
func (s *Store) InsertLot(q Querier, l models.PoinLot) (int64, error)          { return 0, nil }
func (s *Store) UpdateSisaLot(q Querier, lotID, sisa int, status string) error { return nil }

// InsertMutasi: jenis PEROLEHAN | TUKAR | KEMBALI | TARIK | KEDALUWARSA (NF-05, tidak pernah UPDATE/DELETE).
func (s *Store) InsertMutasi(q Querier, memberID, lotID int, trxID *int64, jenis string, jumlah int) error {
	return nil
}
