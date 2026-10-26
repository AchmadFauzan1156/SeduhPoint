# Unduh foto menu & login dari CDN Unsplash (Unsplash License: gratis, boleh dipakai tanpa izin).
# Alamat CDN diambil dari halaman foto masing-masing (og:image). Jalankan dari folder SeduhPoint:
#   powershell -ExecutionPolicy Bypass -File scripts\unduh-aset.ps1
$ProgressPreference = "SilentlyContinue"
Set-Location (Split-Path $PSScriptRoot -Parent)
New-Item -ItemType Directory -Force static\img\menu | Out-Null
# nama = @(halaman foto, gambar CDN, lebar)
$aset = [ordered]@{
  "menu\americano"  = @("-G07QTORSeE", "photo-1753572736794-1c6d1199aa50", 800)
  "menu\kopi-susu"  = @("nIZl9PuwMXk", "photo-1764298490445-6add12301888", 800)
  "menu\cappuccino" = @("tA90pRfL2gM", "photo-1497636577773-f1231844b336", 800)
  "menu\matcha"     = @("U4kneR_s7bU", "photo-1766426910014-63a302fa137e", 800)
  "menu\coklat"     = @("7GkyzWeM0eg", "photo-1702165639524-252a1c1b1ab5", 800)
  "menu\croissant"  = @("VzUE5RtCuBA", "photo-1691480162735-9b91238080f6", 800)
  "menu\roti-bakar" = @("AlX2WLwPKXk", "photo-1532038331778-7f22c2594ec1", 800)
  "login"           = @("21tOEaJPM_o", "photo-1545418314-7ce0b9b53901", 1600)
}
$kredit = @("# Kredit foto", "", "Semua foto dari Unsplash (https://unsplash.com/license), bebas dipakai.", "", "| File | Foto |", "|---|---|")
foreach ($nama in $aset.Keys) {
  $id, $cdn, $w = $aset[$nama]; $file = "static\img\$nama.jpg"
  $err = & curl.exe -fsSL -o $file "https://images.unsplash.com/$($cdn)?w=$w&q=80&fm=jpg&fit=max" 2>&1
  $ok = (Test-Path $file) -and ((Get-Item $file).Length -gt 10000)
  if ($ok) { Write-Host "OK    $nama" -ForegroundColor Green; $kredit += "| $nama.jpg | https://unsplash.com/photos/$id |" }
  else { Remove-Item $file -ErrorAction SilentlyContinue; Write-Host "GAGAL $nama : $err" -ForegroundColor Yellow }
}
$kredit | Set-Content -Encoding UTF8 static\img\KREDIT.md
Write-Host "`nSelesai. Hasil di static\img\" -ForegroundColor Cyan
