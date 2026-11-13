package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"seduhpoint/internal/config"
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

// cekBolehBatal: transaksi cabang lain -> 404 (tidak membocorkan keberadaan data), hari sebelumnya -> 403.
func TestCekBolehBatal(t *testing.T) {
	sekarang := time.Date(2026, 11, 10, 9, 0, 0, 0, config.WIB)
	kasir := model.Pengguna{ID: 2, Peran: model.PeranKasir, CabangID: 1}
	cases := []struct {
		nama string
		t    model.Transaksi
		want error
	}{
		{"hari ini", model.Transaksi{CabangID: 1, Status: "SELESAI", Waktu: sekarang}, nil},
		{"kemarin", model.Transaksi{CabangID: 1, Status: "SELESAI", Waktu: sekarang.AddDate(0, 0, -1)}, model.ErrForbidden},
		{"cabang lain", model.Transaksi{CabangID: 2, Status: "SELESAI", Waktu: sekarang}, model.ErrNotFound},
		{"sudah dibatalkan (panel tetap bisa dibuka)", model.Transaksi{CabangID: 1, Status: "DIBATALKAN", Waktu: sekarang.AddDate(0, 0, -1)}, nil},
	}
	for _, c := range cases {
		if err := cekBolehBatal(kasir, c.t, sekarang); !errors.Is(err, c.want) {
			t.Errorf("%s: dapat %v, harusnya %v", c.nama, err, c.want)
		}
	}
}
