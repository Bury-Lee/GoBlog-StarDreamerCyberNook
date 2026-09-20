@echo off
rem GoBlog (StarDreamerCyberNook) multi-platform build script
rem Usage:
rem   build.bat                     Build windows/linux/macos (amd64)
rem   set BUILD_ARM64=1 && build.bat   Also build linux/macos (arm64)
rem   set OUT_DIR=D:\path && build.bat  Custom output dir (default dist\)
setlocal

set "SCRIPT_DIR=%~dp0"
set "ROOT_DIR=%SCRIPT_DIR%.."
if "%OUT_DIR%"=="" set "OUT_DIR=%SCRIPT_DIR%..\dist"
set "LDFLAGS=-s -w"

where go >nul 2>nul
if errorlevel 1 (
    echo Error: go command not found, please install Go and add it to PATH
    exit /b 1
)

if not exist "%OUT_DIR%" mkdir "%OUT_DIR%"

pushd "%ROOT_DIR%"
set CGO_ENABLED=0

echo ==^> build windows/amd64
set GOOS=windows
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -trimpath -o "%OUT_DIR%\main_windows_amd64.exe" .
if errorlevel 1 goto :error

echo ==^> build linux/amd64
set GOOS=linux
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -trimpath -o "%OUT_DIR%\main_linux_amd64" .
if errorlevel 1 goto :error

echo ==^> build darwin/amd64
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="%LDFLAGS%" -trimpath -o "%OUT_DIR%\main_macos_amd64" .
if errorlevel 1 goto :error

if "%BUILD_ARM64%"=="1" (
    echo ==^> build linux/arm64
    set GOOS=linux
    set GOARCH=arm64
    go build -ldflags="%LDFLAGS%" -trimpath -o "%OUT_DIR%\main_linux_arm64" .
    if errorlevel 1 goto :error

    echo ==^> build darwin/arm64
    set GOOS=darwin
    set GOARCH=arm64
    go build -ldflags="%LDFLAGS%" -trimpath -o "%OUT_DIR%\main_macos_arm64" .
    if errorlevel 1 goto :error
)

popd
echo.
echo Build finished, output dir: %OUT_DIR%
dir /b "%OUT_DIR%"
exit /b 0

:error
popd
echo.
echo Build FAILED!
exit /b 1
