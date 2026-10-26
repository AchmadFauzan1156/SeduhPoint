package store

import "seduhpoint/models"

// F-05
func (s *Store) GetPenggunaByUsername(username string) (models.Pengguna, error) {
	var p models.Pengguna
	err := s.DB.QueryRow(`SELECT id, username, password_hash, peran, cabang_id
		FROM pengguna WHERE username = ? AND aktif = TRUE`, username).
		Scan(&p.ID, &p.Username, &p.PasswordHash, &p.Peran, &p.CabangID)
	return p, err
}
