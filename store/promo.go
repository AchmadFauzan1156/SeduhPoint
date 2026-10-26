package store

import (
	"time"

	"seduhpoint/models"
)

// F-06 – F-11 — TODO M2
func (s *Store) ListPromo() ([]models.Promo, error)     { return nil, nil }
func (s *Store) GetPromo(id int) (models.Promo, error)  { return models.Promo{}, nil }
func (s *Store) SimpanPromo(p models.Promo) error       { return nil } // + promo_cabang/produk/kategori (pakai tx)
func (s *Store) SetPromoAktif(id int, aktif bool) error { return nil }

// PromoBerlakuSekarang: promo aktif untuk cabang & waktu tertentu (F-13). TODO M3
func (s *Store) PromoBerlakuSekarang(cabangID int, t time.Time) ([]models.Promo, error) {
	return nil, nil
}

// GetVoucherForUpdate: kunci baris voucher di dalam tx agar kuota aman (F-14). TODO M3
// SELECT ... FROM promo WHERE kode_voucher = ? FOR UPDATE
func (s *Store) GetVoucherForUpdate(q Querier, kode string) (models.Promo, error) {
	return models.Promo{}, nil
}
