# Setup SeduhPoint di Windows + Laragon (M1.6)
# Jalankan dari PowerShell:  powershell -ExecutionPolicy Bypass -File scripts\setup.ps1
# Tambahkan -Reset untuk menghapus dan membuat ulang database seduhpoint.
param([switch]$Reset, [string]$MySql)
$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)
function Langkah($t) { Write-Host "`n== $t" -ForegroundColor Cyan }

Langkah "Cek Go"
go version

Langkah "Cari MySQL Laragon"
$mysql = $MySql
# 1) dari proses mysqld yang sedang berjalan (Laragon sudah Start All)
if (-not $mysql) {
  $proc = Get-Process mysqld -ErrorAction SilentlyContinue | Where-Object { $_.Path } | Select-Object -First 1
  if ($proc) { $cand = Join-Path (Split-Path $proc.Path) "mysql.exe"; if (Test-Path $cand) { $mysql = $cand } }
}
# 2) folder laragon di semua drive / folder user
if (-not $mysql) {
  $roots = @(Get-PSDrive -PSProvider FileSystem | ForEach-Object { $_.Root }) + @("$env:USERPROFILE\", "$env:ProgramFiles\")
  foreach ($r in $roots) {
    $f = Get-ChildItem (Join-Path $r "laragon\bin\mysql\*\bin\mysql.exe") -ErrorAction SilentlyContinue | Select-Object -Last 1
    if ($f) { $mysql = $f.FullName; break }
  }
}
# 3) dari PATH
if (-not $mysql) { $mysql = (Get-Command mysql -ErrorAction SilentlyContinue).Source }
if (-not $mysql) { throw "mysql.exe tidak ditemukan. Nyalakan Laragon (Start All) lalu ulangi, atau jalankan dengan -MySql 'D:\path\ke\mysql.exe'" }
Write-Host $mysql
& $mysql -u root -e "SELECT VERSION();"
if ($LASTEXITCODE -ne 0) { throw "MySQL belum jalan. Buka Laragon lalu klik 'Start All', kemudian jalankan ulang script ini." }

function ImportSql($file) {
  $full = (Resolve-Path $file).Path
  $p = Start-Process -FilePath $mysql -ArgumentList "-u", "root", "--default-character-set=utf8mb4" `
       -RedirectStandardInput $full -NoNewWindow -Wait -PassThru
  if ($p.ExitCode -ne 0) { throw "Gagal import $file (kode $($p.ExitCode))" }
  Write-Host "  import $file OK"
}

Langkah "Database"
if ($Reset) { & $mysql -u root -e "DROP DATABASE IF EXISTS seduhpoint;" }
$ada = & $mysql -u root -N -e "SHOW DATABASES LIKE 'seduhpoint';"
if ($ada) {
  Write-Host "Database seduhpoint sudah ada, import dilewati (pakai -Reset untuk membuat ulang)."
} else {
  ImportSql "database\migrations\001_schema.sql"
  ImportSql "database\migrations\002_seed.sql"
  Write-Host "Database seduhpoint dibuat + data contoh."
}

Langkah "File .env"
if (-not (Test-Path .env)) { Copy-Item .env.example .env; Write-Host ".env dibuat dari .env.example" } else { Write-Host ".env sudah ada" }

Langkah "Unduh library, vet, test, build"
go mod tidy;            if ($LASTEXITCODE -ne 0) { throw "go mod tidy gagal" }
go vet ./...;           if ($LASTEXITCODE -ne 0) { throw "go vet gagal" }
go test ./...;          if ($LASTEXITCODE -ne 0) { throw "go test gagal" }
go build -o seduhpoint.exe .; if ($LASTEXITCODE -ne 0) { throw "go build gagal" }

Langkah "Uji jalan + login admin"
$p = Start-Process .\seduhpoint.exe -WorkingDirectory (Get-Location) -PassThru -WindowStyle Hidden
try {
  Start-Sleep -Seconds 3
  $s = New-Object Microsoft.PowerShell.Commands.WebRequestSession
  $r = Invoke-WebRequest "http://localhost:8080/login" -Method Post -Body @{username="admin";password="password"} -WebSession $s -UseBasicParsing
  if ($r.Content -match "Selamat datang, admin") { Write-Host "OK: aplikasi jalan & login admin berhasil" -ForegroundColor Green }
  else { throw "Login admin gagal (cek hash password di seed)" }
} finally { Stop-Process -Id $p.Id -Force }

Langkah "Git"
if (Get-Command git -ErrorAction SilentlyContinue) {
  if (-not (Test-Path .git)) { git init -b main | Out-Null; Write-Host "Repo git dibuat" }
  git add -A
  if (git config user.name) {
    git commit -q -m "Kerangka awal SeduhPoint (M1)" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`nClaude-Session: https://claude.ai/code/session_01BoDXLG3ARMGdQ28S9Tn3rP"
    git log --oneline -1
  } else { Write-Host "git user.name belum diatur: jalankan git config --global user.name/user.email lalu commit manual" -ForegroundColor Yellow }
} else { Write-Host "git tidak ditemukan, langkah repo dilewati" -ForegroundColor Yellow }

Write-Host "`nSELESAI. Jalankan aplikasi: go run .  lalu buka http://localhost:8080" -ForegroundColor Green
