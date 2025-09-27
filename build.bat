@echo off
REM JavaSwitcher Build Script for Windows

set APP_NAME=javaswitcher
set VERSION=1.0.0
set BUILD_DIR=build

echo Building JavaSwitcher v%VERSION%...

REM Create build directory
if not exist %BUILD_DIR% mkdir %BUILD_DIR%

REM Build for Windows
echo Building for Windows...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-windows-amd64.exe ./cmd

REM Build for Linux (cross-compile)
echo Building for Linux...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-linux-amd64 ./cmd

REM Build for macOS (cross-compile)
echo Building for macOS...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-X main.version=%VERSION%" -o %BUILD_DIR%\%APP_NAME%-darwin-amd64 ./cmd

echo Build completed! Executables are in the %BUILD_DIR% directory.

REM Reset environment variables
set GOOS=
set GOARCH=

pause