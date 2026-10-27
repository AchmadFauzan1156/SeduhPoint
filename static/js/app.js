// Form promo: tampilkan field sesuai jenis promo yang dipilih
const jenis = document.querySelectorAll('input[name="jenis"]');
if (jenis.length) {
  const sync = () => {
    const v = document.querySelector('input[name="jenis"]:checked')?.value;
    document.querySelectorAll("[data-jenis]").forEach(el => {
      const on = el.dataset.jenis.split(" ").includes(v);
      el.hidden = !on;
      el.querySelectorAll("input,select").forEach(i => (i.disabled = !on));
    });
  };
  jenis.forEach(r => r.addEventListener("change", sync));
  sync();
}
// Pratinjau foto produk sebelum disimpan
const foto = document.getElementById("gambar");
if (foto) foto.addEventListener("change", () => {
  const f = foto.files[0], pv = document.getElementById("preview");
  if (f && pv) pv.style.backgroundImage = `url(${URL.createObjectURL(f)})`;
});
// Konfirmasi sebelum menghapus
document.querySelectorAll("form[data-confirm]").forEach(f =>
  f.addEventListener("submit", e => { if (!confirm(f.dataset.confirm)) e.preventDefault(); }));
// Jam & tanggal di bilah atas (WIB), diperbarui tiap 20 detik
const jam = document.querySelector("[data-jam]");
if (jam) {
  const tgl = document.querySelector("[data-tanggal]");
  const fJam = new Intl.DateTimeFormat("id-ID", { timeZone: "Asia/Jakarta", hour: "2-digit", minute: "2-digit", hour12: false });
  const fTgl = new Intl.DateTimeFormat("id-ID", { timeZone: "Asia/Jakarta", weekday: "long", day: "numeric", month: "short", year: "numeric" });
  const perbarui = () => {
    const now = new Date();
    jam.textContent = fJam.format(now).replace(":", ".");
    if (tgl) tgl.textContent = fTgl.format(now).replace(/^(\w+) /, "$1, ").replace(/\./g, "");
  };
  perbarui();
  setInterval(perbarui, 20000);
}
// Tabel lebar yang bisa digeser harus bisa difokus keyboard (WCAG 2.1.1)
const tandaiGeser = () => document.querySelectorAll(".panel").forEach(p => {
  if (p.scrollWidth > p.clientWidth + 1) {
    p.tabIndex = 0; p.setAttribute("role", "region");
    p.setAttribute("aria-label", (p.querySelector("h2")?.textContent || "Tabel") + " (geser ke samping)");
  } else if (p.getAttribute("role") === "region") { p.removeAttribute("tabindex"); p.removeAttribute("role"); p.removeAttribute("aria-label"); }
});
tandaiGeser(); addEventListener("resize", tandaiGeser);
