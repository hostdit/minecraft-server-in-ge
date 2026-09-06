@echo off
setlocal
cd /d "%~dp0"
where go >nul 2>&1
if errorlevel 1 (
  echo Go is not installed. Get it from https://go.dev/dl/ then rerun.
  exit /b 1
)
set CGO_ENABLED=0
go build -trimpath -o gemc.exe .
if errorlevel 1 exit /b 1
echo Built gemc.exe
echo Run it with: gemc.exe
