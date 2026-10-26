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
