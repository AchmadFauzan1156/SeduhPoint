@echo off
cd /d "%~dp0\.."
title Push SeduhPoint ke GitHub
git add -A
git commit -q -m "M2: login, cabang, kategori, produk + foto, member, promo & voucher" -m "Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>" -m "Claude-Session: https://claude.ai/code/session_01BoDXLG3ARMGdQ28S9Tn3rP"
git remote remove origin 2>nul
git remote add origin https://github.com/AchmadFauzan1156/SeduhPoint.git
git branch -M main
git push -u origin main
if errorlevel 1 (
  echo Repo GitHub sudah berisi file, menggabungkan dulu...
  git pull --rebase origin main --allow-unrelated-histories
  git push -u origin main
)
echo.
git log --oneline -3
git status -sb
pause
