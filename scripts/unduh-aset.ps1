# Unduh foto menu & login dari Unsplash (Unsplash License: gratis, boleh dipakai tanpa izin).
# Jalankan dari folder SeduhPoint:  powershell -ExecutionPolicy Bypass -File scripts\unduh-aset.ps1
$ErrorActionPreference = "Stop"; $ProgressPreference = "SilentlyContinue"
Set-Location (Split-Path $PSScriptRoot -Parent)
New-Item -ItemType Directory -Force static\img\menu | Out-Null
# nama file = @(id utama, id cadangan, lebar)
$aset = [ordered]@{
  "menu\americano"  = @("-G07QTORSeE", "G0DKq4yz5AA", 800)
  "menu\kopi-susu"  = @("nIZl9PuwMXk", "L-sm1B4L1Ns", 800)
  "menu\cappuccino" = @("tA90pRfL2gM", "1q-Q_Os-WLQ", 800)
  "menu\matcha"     = @("U4kneR_s7bU", "N1rmKN_EOaA", 800)
  "menu\coklat"     = @("Pe_2h7aMb3g", "wgka_XBROhc", 800)
  "menu\croissant"  = @("VzUE5RtCuBA", "a4xoMVKzbak", 800)
  "menu\roti-bakar" = @("AlX2WLwPKXk", "HZlQhqh0zIk", 800)
  "login"           = @("21tOEaJPM_o", "qE1jxYXiwOA", 1600)
}
$kredit = @("# Kredit foto", "", "Semua foto dari Unsplash (https://unsplash.com/license), bebas dipakai.", "", "| File | Foto |", "|---|---|")
foreach ($nama in $aset.Keys) {
  $ok = $false
  foreach ($id in $aset[$nama][0..1]) {
    try {
      $r = Invoke-WebRequest -Uri "https://unsplash.com/photos/$id/download?force=true&w=$($aset[$nama][2])" -OutFile "static\img\$nama.jpg" -PassThru -UseBasicParsing
      if ($r.Headers["Content-Type"] -like "image/*" -and (Get-Item "static\img\$nama.jpg").Length -gt 10000) {
        Write-Host "OK   $nama  ($id)" -ForegroundColor Green; $kredit += "| $nama.jpg | https://unsplash.com/photos/$id |"; $ok = $true; break
      }
    } catch { }
  }
  if (-not $ok) { Write-Host "GAGAL $nama (dua-duanya bukan foto gratis/terhubung)" -ForegroundColor Yellow }
}
$kredit | Set-Content -Encoding UTF8 static\img\KREDIT.md
Write-Host "`nSelesai. Hasil di static\img\" -ForegroundColor Cyan
