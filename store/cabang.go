package store

import "seduhpoint/models"

// CONTOH LENGKAP — modul lain mengikuti pola ini. (F-01)

func (s *Store) ListCabang() ([]models.Cabang, error) {
	rows, err := s.DB.Query(`SELECT id, nama, alamat, aktif FROM cabang ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Cabang
	for rows.Next() {
		var c models.Cabang
		if err := rows.Scan(&c.ID, &c.Nama, &c.Alamat, &c.Aktif); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetCabang(id int) (models.Cabang, error) {
	var c models.Cabang
	err := s.DB.QueryRow(`SELECT id, nama, alamat, aktif FROM cabang WHERE id = ?`, id).
		Scan(&c.ID, &c.Nama, &c.Alamat, &c.Aktif)
	return c, err
}

func (s *Store) SimpanCabang(c models.Cabang) error {
	if c.ID == 0 {
		_, err := s.DB.Exec(`INSERT INTO cabang (nama, alamat, aktif) VALUES (?, ?, ?)`, c.Nama, c.Alamat, c.Aktif)
		return err
	}
	_, err := s.DB.Exec(`UPDATE cabang SET nama = ?, alamat = ?, aktif = ? WHERE id = ?`, c.Nama, c.Alamat, c.Aktif, c.ID)
	return err
}
