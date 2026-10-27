package service

import (
	"testing"
	"time"

	"seduhpoint/internal/repository"
)

func tg(s string) time.Time { t, _ := time.Parse("2006-01-02 15:04", s); return t }

// Perkembangan member: member baru, naik Gold, turun Silver, dan jumlah per tingkat di akhir bulan.
func TestPerkembanganMember(t *testing.T) {
	data := []repository.MemberTingkat{
		{MemberID: 1, TglDaftar: tg("2026-08-05 00:00"), Riwayat: []repository.RiwayatTingkat{
			{Tingkat: "SILVER", Sejak: tg("2026-08-05 10:00")}, {Tingkat: "GOLD", Sejak: tg("2026-09-20 12:00")}, {Tingkat: "SILVER", Sejak: tg("2026-10-03 00:05")}}},
		{MemberID: 2, TglDaftar: tg("2026-09-01 00:00"), Riwayat: []repository.RiwayatTingkat{
			{Tingkat: "SILVER", Sejak: tg("2026-09-01 09:00")}}},
		{MemberID: 3, TglDaftar: tg("2026-10-10 00:00")}, // tanpa riwayat = Silver
	}
	b := PerkembanganMember(data, 2026, tg("2026-10-27 08:00"))
	if len(b) != 3 {
		t.Fatalf("harus Agustus–Oktober saja (sebelum ada member & setelah bulan ini tidak tampil), dapat %d baris", len(b))
	}
	cek := func(i, baru, naik, turun, silver, gold int) {
		t.Helper()
		x := b[i]
		if x.MemberBaru != baru || x.NaikGold != naik || x.TurunSilver != turun || x.Silver != silver || x.Gold != gold {
			t.Errorf("%s: dapat baru=%d naik=%d turun=%d silver=%d gold=%d", x.Bulan.Format("Jan"),
				x.MemberBaru, x.NaikGold, x.TurunSilver, x.Silver, x.Gold)
		}
	}
	cek(0, 1, 0, 0, 1, 0) // Agustus
	cek(1, 1, 1, 0, 1, 1) // September
	cek(2, 1, 0, 1, 3, 0) // Oktober
}

func TestPersenBatang(t *testing.T) {
	if persen(0, 100) != 0 || persen(100, 100) != 100 || persen(1, 1000) != 2 {
		t.Error("persen batang salah")
	}
}

func TestBandingkan(t *testing.T) {
	if p := bandingkan(120, 100); !p.Ada || !p.Naik || p.Persen != 20 {
		t.Errorf("naik 20%%: %+v", p)
	}
	if p := bandingkan(50, 200); p.Naik || p.Persen != 75 {
		t.Errorf("turun 75%%: %+v", p)
	}
	if p := bandingkan(10, 0); p.Ada {
		t.Error("tanpa pembanding harus Ada=false")
	}
}

func TestGrafikDanDonat(t *testing.T) {
	akhir := tg("2026-10-27 00:00")
	g := BuatGrafik(map[string]int64{"2026-10-27": 190000, "2026-10-25": 50000}, akhir, 7, 640, 200)
	if len(g.Titik) != 7 || g.Total != 240000 || !g.Titik[6].HariIni || g.Titik[6].Omzet != 190000 {
		t.Fatalf("grafik salah: %+v", g)
	}
	if g.GarisY[4].Label != "200 rb" || g.Titik[0].Y != 200 {
		t.Errorf("skala Y salah: atas %s, titik nol y=%v", g.GarisY[4].Label, g.Titik[0].Y)
	}
	d := BuatDonat([]CabangBiaya{{BarisCabang: repository.BarisCabang{Nama: "A", Omzet: 75}}, {BarisCabang: repository.BarisCabang{Nama: "B", Omzet: 25}}})
	if d[0].Persen != 75 || d[1].Offset != -75 {
		t.Errorf("donat salah: %+v", d)
	}
}
