package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"seduhpoint/internal/model"
)

// Materi #4: semua field diperiksa sekaligus, bukan berhenti di error pertama.
func TestValidasiMengumpulkanSemuaError(t *testing.T) {
	c := model.Cabang{Nama: "  ", Alamat: strings.Repeat("a", 256)}
	var ve *ValidasiError
	if err := validasiCabang(&c); !errors.As(err, &ve) {
		t.Fatalf("harusnya ValidasiError, dapat %v", err)
	}
	for _, field := range []string{"nama", "alamat"} {
		if ve.Fields[field] == "" {
			t.Errorf("pesan untuk %q tidak ada: %v", field, ve.Fields)
		}
	}

	// Promo kosong: nama, jenis, tanggal, dan cabang dilaporkan bersamaan.
	p := model.Promo{}
	if err := (&PromoService{}).Validasi(context.Background(), &p); !errors.As(err, &ve) {
		t.Fatalf("harusnya ValidasiError, dapat %v", err)
	}
	for _, field := range []string{"nama", "jenis", "tgl_mulai", "tgl_selesai", "cabang"} {
		if ve.Fields[field] == "" {
			t.Errorf("promo: pesan untuk %q tidak ada: %v", field, ve.Fields)
		}
	}
}

func TestValidasiCabangLolos(t *testing.T) {
	c := model.Cabang{Nama: " Seduh Sekaran ", Alamat: "Jl. Raya Sekaran"}
	if err := validasiCabang(&c); err != nil {
		t.Fatal(err)
	}
	if c.Nama != "Seduh Sekaran" {
		t.Errorf("nama harus dirapikan (TrimSpace), dapat %q", c.Nama)
	}
}
