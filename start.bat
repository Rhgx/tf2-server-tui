@echo off
title TF2 Server Browser
cd /d "%~dp0"

if exist "tf2-server-tui.exe" (
  tf2-server-tui.exe
  exit /b %errorlevel%
)

echo tf2-server-tui.exe was not found.
echo.
echo Build it first with:
echo   build.bat
echo.
pause
