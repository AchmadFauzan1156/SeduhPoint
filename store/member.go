package store

import "seduhpoint/models"

// F-03, F-04 — TODO M2
func (s *Store) CariMember(q string) ([]models.Member, error) { return nil, nil } // LIKE nama / no_hp
func (s *Store) GetMember(id int) (models.Member, error)      { return models.Member{}, nil }

// SimpanMember: member baru otomatis SILVER + catat member_tingkat_riwayat.
func (s *Store) SimpanMember(m models.Member) error { return nil }
