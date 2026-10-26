package handler

import (
	"errors"
	"log"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"

	"seduhpoint/internal/middleware"
	"seduhpoint/internal/model"
	"seduhpoint/internal/service"
)

// Halaman adalah data standar yang dikirim ke setiap template.
type Halaman struct {
	Judul  string
	Aktif  string // menu sidebar yang sedang aktif
	Sesi   middleware.Sesi
	Flash  string            // pesan sukses (?pesan=...)
	Error  string            // pesan ValidasiError umum
	Errors map[string]string // pesan per field (ditampilkan di dekat input)
	Form   map[string]string // isi form untuk ditampilkan ulang
	Data   any
}

func halaman(r *http.Request, judul, aktif string, data any) Halaman {
	s, _ := middleware.Ambil(r)
	return Halaman{Judul: judul, Aktif: aktif, Sesi: s, Data: data,
		Flash: r.URL.Query().Get("pesan"), Errors: map[string]string{}}
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

// formGagal: salah input -> tampilkan ulang halaman form beserta pesan di dekat field (422);
// error lain -> gagal().
func formGagal(w http.ResponseWriter, view *Renderer, file string, hal Halaman, err error) {
	var ve *service.ValidasiError
	if !errors.As(err, &ve) {
		gagal(w, err)
		return
	}
	if ve.Field != "" {
		hal.Errors[ve.Field] = ve.Pesan
	} else {
		hal.Error = ve.Pesan
	}
	w.WriteHeader(http.StatusUnprocessableEntity)
	view.Tampil(w, file, hal)
}

// redirect dengan pesan sukses (Post/Redirect/Get, materi #4).
func redirect(w http.ResponseWriter, r *http.Request, url, pesan string) {
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	http.Redirect(w, r, url+sep+"pesan="+neturl.QueryEscape(pesan), http.StatusSeeOther)
}

func angka(s string) int64 {
	s = strings.NewReplacer(".", "", ",", "", " ", "", "Rp", "").Replace(strings.TrimSpace(s))
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func atoi(s string) int { n, _ := strconv.Atoi(strings.TrimSpace(s)); return n }

// idsDariForm: semua nilai checkbox bernama `nama` sebagai []int.
func idsDariForm(r *http.Request, nama string) []int {
	var out []int
	for _, v := range r.PostForm[nama] {
		if n, err := strconv.Atoi(v); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// pesanValidasi mengembalikan pesan jika err adalah ValidasiError.
func pesanValidasi(err error) (string, bool) {
	var ve *service.ValidasiError
	if errors.As(err, &ve) {
		return ve.Pesan, true
	}
	return "", false
}
