@echo off
cd /d "%~dp0"
title SeduhPoint - jangan tutup jendela ini selama aplikasi dipakai
echo Membangun SeduhPoint...
go build -o seduhpoint.exe ./cmd/web || (pause & exit /b 1)
start "" http://localhost:8080
seduhpoint.exe
pause
