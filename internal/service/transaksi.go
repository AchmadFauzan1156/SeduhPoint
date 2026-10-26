package service

// F-12 – F-24 Transaksi: pratinjau (tombol Hitung), checkout, pembatalan.

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"seduhpoint/internal/config"
	"seduhpoint/internal/model"
	"seduhpoint/internal/repository"
)

type TransaksiService struct {
	DB         *sql.DB
	Repo       *repository.TransaksiRepo
	PromoRepo  *repository.PromoRepo
	PoinRepo   *repository.PoinRepo
	ProdukRepo *repository.ProdukRepo
	MemberRepo *repository.MemberRepo
	Tingkat    *PoinService
	Sekarang   func() time.Time // bisa diganti saat test; default waktu WIB sekarang
}

type PermintaanCheckout struct {
	CabangID    int
	PenggunaID  int
	MemberID    int         // 0 = non-member (asumsi 11)
	Qty         map[int]int // produk_id -> jumlah
	KodeVoucher string
	PoinDiminta int
	Token       string // anti submit ganda (V08)
}

// Pratinjau: hasil hitung + data pendukung untuk ditampilkan di halaman kasir.
type Pratinjau struct {
	Hasil   Hasil
	Items   []Item
	Member  *model.Member
	Catatan []string        // info untuk kasir, mis. "Belanja Rp9.000 lagi untuk Hemat sore"
	Label   map[int]string  // produk_id -> label promo di kartu menu
	Promo   []model.Promo   // promo yang berlaku saat ini di cabang
	voucher *model.Promo    // voucher terpakai (untuk checkout)
	lot     []model.PoinLot // lot terkunci (untuk checkout)
}

func (s *TransaksiService) sekarang() time.Time {
	if s.Sekarang != nil {
		return s.Sekarang()
	}
	return time.Now().In(config.WIB)
}

func keAktif(p model.Promo) PromoAktif {
	a := PromoAktif{ID: p.ID, Nama: p.Nama, Jenis: p.Jenis, Persen: p.NilaiPersen.Int64, Nominal: p.NilaiNominal.Int64,
		MinBelanja: p.MinBelanja.Int64, ProdukTarget: map[int]bool{}, KategoriTarget: map[int]bool{}}
	for _, id := range p.ProdukIDs {
		a.ProdukTarget[id] = true
	}
	for _, id := range p.KategoriIDs {
		a.KategoriTarget[id] = true
	}
	return a
}

// PromoSaatIni: promo yang berlaku di cabang pada tanggal & jam (WIB) sekarang.
func (s *TransaksiService) PromoSaatIni(ctx context.Context, cabangID int) ([]model.Promo, error) {
	now := s.sekarang()
	semua, err := s.PromoRepo.BerlakuUntuk(ctx, cabangID, now)
	if err != nil {
		return nil, err
	}
	var hasil []model.Promo
	for _, p := range semua {
		if PromoBerlaku(now, p.TglMulai, p.TglSelesai, p.JamMulai.String, p.JamSelesai.String, p.CabangIDs, cabangID) {
			hasil = append(hasil, p)
		}
	}
	return hasil, nil
}

// siapkan membangun Input mesin hitung dari permintaan kasir. Di dalam checkout (kunci = true)
// voucher dan lot poin dikunci dengan SELECT ... FOR UPDATE.
func (s *TransaksiService) siapkan(ctx context.Context, q repository.Querier, p PermintaanCheckout, kunci bool) (Pratinjau, error) {
	var pr Pratinjau
	var err error
	if pr.Promo, err = s.PromoSaatIni(ctx, p.CabangID); err != nil {
		return pr, err
	}

	// Item: harga & kategori selalu dari database.
	var ids []int
	for id, n := range p.Qty {
		if n > 0 {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	produk, err := s.ProdukRepo.DijualByID(ctx, ids)
	if err != nil {
		return pr, err
	}
	for _, id := range ids {
		if pd, ok := produk[id]; ok {
			pr.Items = append(pr.Items, Item{ProdukID: id, KategoriID: pd.KategoriID, Nama: pd.Nama, Harga: pd.Harga, Qty: p.Qty[id]})
		}
	}

	in := Input{Items: pr.Items}
	var daftarMin []model.Promo
	for _, pm := range pr.Promo {
		switch pm.Jenis {
		case JenisDiskon:
			in.PromoProduk = append(in.PromoProduk, keAktif(pm))
		case JenisB2G1:
			in.PromoProduk = append(in.PromoProduk, keAktif(pm))
		case JenisMinBelanja:
			daftarMin = append(daftarMin, pm)
		}
	}

	// Member & poin
	if p.MemberID > 0 {
		m, err := s.MemberRepo.CariByID(ctx, p.MemberID)
		if err != nil {
			return pr, salah("member", "Member tidak ditemukan")
		}
		pr.Member = &m
		in.IsMember, in.IsGold, in.SaldoPoin = true, m.Tingkat == "GOLD", m.PoinAktif
		if kunci {
			if pr.lot, err = s.PoinRepo.LotAktifFIFO(ctx, q, m.ID, s.sekarang()); err != nil {
				return pr, err
			}
			in.SaldoPoin = 0
			for _, l := range pr.lot {
				in.SaldoPoin += l.PoinSisa
			}
		}
		in.PoinDiminta = p.PoinDiminta
	}

	// Voucher (F-14): ada, aktif, berlaku di cabang & jam ini, kuota masih ada.
	if kode := strings.ToUpper(strings.TrimSpace(p.KodeVoucher)); kode != "" {
		v, err := s.PromoRepo.VoucherByKode(ctx, q, kode, kunci)
		switch {
		case errors.Is(err, model.ErrNotFound):
			return pr, salah("kode_voucher", "Kode voucher "+kode+" tidak ditemukan")
		case err != nil:
			return pr, err
		case !v.Aktif || !PromoBerlaku(s.sekarang(), v.TglMulai, v.TglSelesai, v.JamMulai.String, v.JamSelesai.String, v.CabangIDs, p.CabangID):
			return pr, salah("kode_voucher", "Voucher "+kode+" tidak berlaku di cabang ini saat ini")
		case v.Kuota.Valid && int64(v.KuotaTerpakai) >= v.Kuota.Int64:
			return pr, salah("kode_voucher", "Kuota voucher "+kode+" sudah habis")
		}
		a := keAktif(v)
		in.Voucher, pr.voucher = &a, &v
	}

	// Potongan minimal belanja: pilih yang paling besar yang syaratnya terpenuhi,
	// kecuali ada voucher (AB09: tidak bisa digabung).
	dasar, err := Hitung(Input{Items: in.Items, PromoProduk: in.PromoProduk})
	if err != nil {
		return pr, err
	}
	totalProduk := dasar.TotalSetelahPromo
	var terbaik, berikut *model.Promo
	for i := range daftarMin {
		pm := &daftarMin[i]
		if totalProduk >= pm.MinBelanja.Int64 {
			if terbaik == nil || pm.NilaiNominal.Int64 > terbaik.NilaiNominal.Int64 {
				terbaik = pm
			}
		} else if berikut == nil || pm.MinBelanja.Int64 < berikut.MinBelanja.Int64 {
			berikut = pm
		}
	}
	if terbaik != nil {
		if in.Voucher != nil {
			pr.Catatan = append(pr.Catatan, terbaik.Nama+" tidak digabung dengan voucher (potongan minimal belanja dan voucher pilih salah satu).")
		} else {
			a := keAktif(*terbaik)
			in.PromoMin = &a
		}
	} else if berikut != nil && len(in.Items) > 0 && in.Voucher == nil {
		pr.Catatan = append(pr.Catatan, "Belanja "+rupiah(berikut.MinBelanja.Int64-totalProduk)+" lagi untuk "+berikut.Nama+".")
	}

	if pr.Hasil, err = Hitung(in); err != nil {
		return pr, err
	}
	if in.IsMember && p.PoinDiminta > pr.Hasil.PoinDitukar {
		pr.Catatan = append(pr.Catatan, "Tukar poin dibatasi "+itoa(pr.Hasil.MaksPoin)+" poin (maks 50% belanja dan saldo poin).")
	}
	return pr, nil
}

// Pratinjau menghitung total tanpa menyimpan (tombol "Hitung" / setiap perubahan pesanan).
func (s *TransaksiService) Pratinjau(ctx context.Context, p PermintaanCheckout) (Pratinjau, error) {
	return s.siapkan(ctx, s.DB, p, false)
}

// Checkout menyimpan transaksi (Laporan 1.8 langkah 8) dalam satu database transaction.
func (s *TransaksiService) Checkout(ctx context.Context, p PermintaanCheckout) (int64, error) {
	var trxID int64
	err := repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		// Tombol Bayar ditekan dua kali: kembalikan transaksi yang sudah ada (V08).
		if p.Token != "" {
			if id, ada, err := s.Repo.CariByToken(ctx, tx, p.Token); err != nil || ada {
				trxID = id
				return err
			}
		}
		pr, err := s.siapkan(ctx, tx, p, true)
		if err != nil {
			return err
		}
		if len(pr.Items) == 0 {
			return salah("", "Pesanan masih kosong")
		}
		now := s.sekarang()
		h := pr.Hasil
		t := model.Transaksi{CabangID: p.CabangID, PenggunaID: p.PenggunaID, Waktu: now, Subtotal: h.Subtotal,
			TotalPromo: h.TotalPromo, PoinDitukar: h.PoinDitukar, PotonganPoin: h.PotonganPoin, TotalBayar: h.TotalBayar,
			PoinDiperoleh: h.PoinDiperoleh, Token: sql.NullString{String: p.Token, Valid: p.Token != ""}}
		if pr.Member != nil {
			t.MemberID = sql.NullInt64{Int64: int64(pr.Member.ID), Valid: true}
		}
		if t.NoTransaksi, err = s.Repo.NomorBerikut(ctx, tx, p.CabangID, now); err != nil {
			return err
		}
		if trxID, err = s.Repo.Tambah(ctx, tx, t); err != nil {
			return err
		}
		for _, it := range pr.Items {
			if err := s.Repo.TambahItem(ctx, tx, trxID, it.ProdukID, it.Qty, it.Harga); err != nil {
				return err
			}
		}
		for _, r := range h.RincianPromo {
			if err := s.Repo.TambahPromo(ctx, tx, trxID, r.PromoID, r.Nilai); err != nil {
				return err
			}
		}
		if pr.voucher != nil {
			if err := s.PromoRepo.TambahPemakaian(ctx, tx, pr.voucher.ID, 1); err != nil {
				return err
			}
		}
		if pr.Member == nil {
			return nil
		}
		mid := pr.Member.ID
		// Tukar poin: kurangi lot paling dekat kedaluwarsa dulu (AB04, F-17).
		perlu := h.PoinDitukar
		for _, l := range pr.lot {
			if perlu == 0 {
				break
			}
			ambil := min(l.PoinSisa, perlu)
			status := "AKTIF"
			if l.PoinSisa == ambil {
				status = "HABIS"
			}
			if err := s.PoinRepo.UbahSisaLot(ctx, tx, l.ID, l.PoinSisa-ambil, status); err != nil {
				return err
			}
			if err := s.PoinRepo.TambahMutasi(ctx, tx, mid, l.ID, &trxID, "TUKAR", -ambil); err != nil {
				return err
			}
			perlu -= ambil
		}
		// Perolehan poin: lot baru berlaku 6 bulan (AB01, AB02, F-18).
		if h.PoinDiperoleh > 0 {
			lotID, err := s.PoinRepo.TambahLot(ctx, tx, model.PoinLot{MemberID: mid, TransaksiID: int(trxID),
				PoinAwal: h.PoinDiperoleh, PoinSisa: h.PoinDiperoleh, TglPerolehan: now, TglKedaluwarsa: TanggalKedaluwarsa(now)})
			if err != nil {
				return err
			}
			if err := s.PoinRepo.TambahMutasi(ctx, tx, mid, int(lotID), &trxID, "PEROLEHAN", h.PoinDiperoleh); err != nil {
				return err
			}
		}
		// Tingkat dievaluasi setelah poin dihitung, jadi berlaku untuk transaksi berikutnya (asumsi 12).
		return s.Tingkat.Evaluasi(ctx, tx, mid)
	})
	return trxID, err
}

// Struk: data lengkap untuk halaman struk.
type Struk struct {
	Transaksi model.Transaksi
	Item      []model.ItemStruk
	Promo     []model.PromoStruk
	SaldoPoin int
}

func (s *TransaksiService) Struk(ctx context.Context, id int) (Struk, error) {
	var st Struk
	var err error
	if st.Transaksi, err = s.Repo.CariByID(ctx, id); err != nil {
		return st, err
	}
	if st.Item, err = s.Repo.Item(ctx, id); err != nil {
		return st, err
	}
	if st.Promo, err = s.Repo.Promo(ctx, s.DB, id); err != nil {
		return st, err
	}
	if st.Transaksi.MemberID.Valid {
		m, err := s.MemberRepo.CariByID(ctx, int(st.Transaksi.MemberID.Int64))
		if err == nil {
			st.SaldoPoin = m.PoinAktif
		}
	}
	return st, nil
}

func (s *TransaksiService) Daftar(ctx context.Context, cabangID int, tgl time.Time) ([]model.Transaksi, error) {
	return s.Repo.Semua(ctx, cabangID, tgl)
}

func (s *TransaksiService) Detail(ctx context.Context, id int) (model.Transaksi, error) {
	return s.Repo.CariByID(ctx, id)
}

// Dampak: ringkasan yang ditampilkan sebelum kasir mengonfirmasi pembatalan.
type Dampak struct {
	PoinDitarik      int
	PoinDikembalikan int
	Voucher          string
}

func (s *TransaksiService) DampakBatal(ctx context.Context, id int) (Dampak, error) {
	var d Dampak
	mutasi, err := s.PoinRepo.MutasiTransaksi(ctx, s.DB, id)
	if err != nil {
		return d, err
	}
	for _, m := range mutasi {
		if m.Jenis == "PEROLEHAN" {
			if l, err := s.PoinRepo.LotForUpdate(ctx, s.DB, m.LotID); err == nil {
				d.PoinDitarik += l.PoinSisa
			}
		} else {
			d.PoinDikembalikan += -m.Jumlah
		}
	}
	promo, err := s.Repo.Promo(ctx, s.DB, id)
	if err != nil {
		return d, err
	}
	for _, p := range promo {
		if p.Jenis == JenisVoucher {
			d.Voucher = p.Nama
		}
	}
	return d, nil
}

// Batalkan (F-20 – F-24, AB10): transaksi tidak dihapus, hanya berstatus DIBATALKAN.
func (s *TransaksiService) Batalkan(ctx context.Context, id int, alasan string) error {
	alasan = strings.TrimSpace(alasan)
	if alasan == "" {
		return salah("alasan", "Alasan pembatalan wajib diisi")
	}
	if len([]rune(alasan)) > 255 {
		return salah("alasan", "Alasan maksimal 255 karakter")
	}
	return repository.WithTx(ctx, s.DB, func(tx *sql.Tx) error {
		t, err := s.Repo.CariForUpdate(ctx, tx, id)
		if err != nil {
			return err
		}
		if t.Status != "SELESAI" {
			return salah("", "Transaksi "+t.NoTransaksi+" sudah dibatalkan")
		}
		if err := s.Repo.TandaiBatal(ctx, tx, id, alasan); err != nil {
			return err
		}
		trx := int64(id)
		now := s.sekarang()
		mutasi, err := s.PoinRepo.MutasiTransaksi(ctx, tx, id)
		if err != nil {
			return err
		}
		for _, m := range mutasi {
			l, err := s.PoinRepo.LotForUpdate(ctx, tx, m.LotID)
			if err != nil {
				return err
			}
			switch m.Jenis {
			case "PEROLEHAN": // F-21: tarik poin yang diperoleh; yang sudah terpakai tidak bisa ditarik (asumsi 8)
				if l.PoinSisa > 0 {
					if err := s.PoinRepo.TambahMutasi(ctx, tx, l.MemberID, l.ID, &trx, "TARIK", -l.PoinSisa); err != nil {
						return err
					}
				}
				if err := s.PoinRepo.UbahSisaLot(ctx, tx, l.ID, 0, "DITARIK"); err != nil {
					return err
				}
			case "TUKAR": // F-22: kembalikan ke lot asal; lot yang sudah lewat masa berlaku tetap kedaluwarsa (asumsi 9)
				status := "AKTIF"
				if !l.TglKedaluwarsa.After(now) || l.Status == "KEDALUWARSA" {
					status = "KEDALUWARSA"
				}
				if err := s.PoinRepo.UbahSisaLot(ctx, tx, l.ID, l.PoinSisa-m.Jumlah, status); err != nil {
					return err
				}
				if err := s.PoinRepo.TambahMutasi(ctx, tx, l.MemberID, l.ID, &trx, "KEMBALI", -m.Jumlah); err != nil {
					return err
				}
			}
		}
		// F-23: kembalikan kuota voucher
		promo, err := s.Repo.Promo(ctx, tx, id)
		if err != nil {
			return err
		}
		for _, p := range promo {
			if p.Jenis == JenisVoucher {
				if err := s.PromoRepo.TambahPemakaian(ctx, tx, p.PromoID, -1); err != nil {
					return err
				}
			}
		}
		// F-24: transaksi batal tidak dihitung lagi untuk tingkat member
		if t.MemberID.Valid {
			return s.Tingkat.Evaluasi(ctx, tx, int(t.MemberID.Int64))
		}
		return nil
	})
}
