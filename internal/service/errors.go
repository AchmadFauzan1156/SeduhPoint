// Package service: aturan bisnis. Tidak tahu http.Request dan tidak ada SQL (materi #10).
package service

// ValidasiError: salah input dari pengguna (handler -> 400/422),
// dibedakan dari error server (500).
type ValidasiError struct {
	Pesan string
}

func (e *ValidasiError) Error() string { return e.Pesan }
