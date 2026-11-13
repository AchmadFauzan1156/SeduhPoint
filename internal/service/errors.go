// Package service: aturan bisnis. Tidak tahu http.Request dan tidak ada SQL.
package service

// ValidasiError: salah input dari pengguna, dibedakan dari error server (materi #10).
// Handler menampilkan ulang form (422) dengan pesan di dekat field yang salah.
//
// Tambahan dari materi #10: Fields berisi pesan per field. Sesuai materi #4 ("Pengumpulan Error"),
// semua field diperiksa sekaligus sehingga pengguna langsung melihat semua kesalahannya,
// bukan memperbaiki satu per satu lalu submit berulang kali.
type ValidasiError struct {
	Pesan  string            // pesan umum (dipakai jika kesalahannya tidak terkait satu field)
	Fields map[string]string // nama input -> pesan
}

func (e *ValidasiError) Error() string { return e.Pesan }

// salah: satu kesalahan (field boleh kosong untuk pesan umum).
func salah(field, pesan string) error {
	if field == "" {
		return &ValidasiError{Pesan: pesan}
	}
	return &ValidasiError{Pesan: pesan, Fields: map[string]string{field: pesan}}
}

// SalahInput dipakai handler untuk kesalahan input yang dideteksi sebelum masuk service (mis. ukuran foto).
func SalahInput(field, pesan string) error { return salah(field, pesan) }

// hasilValidasi mengubah kumpulan pesan per field menjadi error (nil jika kosong).
func hasilValidasi(errs map[string]string) error {
	if len(errs) == 0 {
		return nil
	}
	return &ValidasiError{Pesan: "Periksa kembali isian yang ditandai", Fields: errs}
}
