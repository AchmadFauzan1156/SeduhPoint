// Package service: aturan bisnis. Tidak tahu http.Request dan tidak ada SQL.
package service

// ValidasiError: salah input dari pengguna (handler -> 422 + pesan di dekat field),
// dibedakan dari error server (500). Field = nama input yang salah (boleh kosong).
type ValidasiError struct {
	Field string
	Pesan string
}

func (e *ValidasiError) Error() string { return e.Pesan }

func salah(field, pesan string) error { return &ValidasiError{Field: field, Pesan: pesan} }
