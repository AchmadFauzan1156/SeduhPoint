// Form promo: tampilkan field sesuai jenis promo
const jenis = document.getElementById("jenis");
if (jenis) {
  const sync = () => document.querySelectorAll("[data-jenis]").forEach(el => {
    el.hidden = !el.dataset.jenis.split(" ").includes(jenis.value);
  });
  jenis.addEventListener("change", sync);
  sync();
}
// TODO M3: interaksi keranjang kasir (tambah/hapus item)
