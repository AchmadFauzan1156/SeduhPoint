package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"seduhpoint/internal/middleware"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

// Halaman adalah data standar yang dikirim ke setiap template.
type Halaman struct {
	Judul string
	Sesi  middleware.Sesi
	Flash string            // pesan sukses (?pesan=...)
	Error string            // pesan ValidasiError
	Form  map[string]string // isi form untuk ditampilkan ulang
	Data  any
}

func halaman(r *http.Request, judul string, data any) Halaman {
	s, _ := middleware.Ambil(r)
	return Halaman{Judul: judul, Sesi: s, Data: data, Flash: r.URL.Query().Get("pesan")}
}

func idDariPath(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

// gagal menerjemahkan error dari service menjadi status HTTP.
func gagal(w http.ResponseWriter, err error) {
	var ve *service.ValidasiError
	switch {
	case errors.As(err, &ve):
		http.Error(w, ve.Pesan, http.StatusBadRequest)
	case errors.Is(err, model.ErrNotFound):
		http.Error(w, "Data tidak ditemukan", http.StatusNotFound)
	default:
		log.Printf("error: %v", err)
		http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
	}
}

// pesanValidasi mengembalikan pesan jika err adalah ValidasiError.
func pesanValidasi(err error) (string, bool) {
	var ve *service.ValidasiError
	if errors.As(err, &ve) {
		return ve.Pesan, true
	}
	return "", false
}

func angka(s string) int64 { n, _ := strconv.ParseInt(s, 10, 64); return n }
