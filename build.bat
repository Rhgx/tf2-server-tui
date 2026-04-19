@echo off
setlocal
cd /d "%~dp0"

set GOEXE=C:\Program Files\Go\bin\go.exe
if not exist "%GOEXE%" (
  echo Go was not found at "%GOEXE%".
  echo Install Go from https://go.dev/dl/ and then run this script again.
  exit /b 1
)

echo Building tf2-server-tui.exe...
"%GOEXE%" build -trimpath -ldflags="-s -w" -o tf2-server-tui.exe .
if errorlevel 1 exit /b 1

echo.
echo Build complete: tf2-server-tui.exe
