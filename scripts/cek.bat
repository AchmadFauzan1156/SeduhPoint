@echo off
cd /d "%~dp0\.."
title Cek SeduhPoint
powershell -ExecutionPolicy Bypass -File scripts\setup.ps1 -Reset -NoGit > scripts\hasil-cek.txt 2>&1
type scripts\hasil-cek.txt
pause
