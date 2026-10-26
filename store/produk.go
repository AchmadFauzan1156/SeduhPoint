package store

import "seduhpoint/models"

// F-02 — TODO M2: ikuti pola store/cabang.go
func (s *Store) ListKategori() ([]models.Kategori, error)            { return nil, nil }
func (s *Store) SimpanKategori(k models.Kategori) error              { return nil }
func (s *Store) ListProduk(hanyaAktif bool) ([]models.Produk, error) { return nil, nil }
func (s *Store) GetProduk(id int) (models.Produk, error)             { return models.Produk{}, nil }
func (s *Store) SimpanProduk(p models.Produk) error                  { return nil }
