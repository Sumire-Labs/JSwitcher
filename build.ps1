# JavaSwitcher Build Script
# Version: 0.3.1

$version = "0.3.1"
$outputDir = "dist"

# Create output directory
if (Test-Path $outputDir) {
    Remove-Item $outputDir -Recurse -Force
}
New-Item -ItemType Directory -Path $outputDir | Out-Null

Write-Host "Building JavaSwitcher v$version for multiple platforms..." -ForegroundColor Green

# Windows (64-bit)
Write-Host "`nBuilding for Windows (amd64)..." -ForegroundColor Cyan
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o "$outputDir/JavaSwitcher-$version-windows-amd64.exe" ./cmd
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Windows build successful" -ForegroundColor Green
} else {
    Write-Host "✗ Windows build failed" -ForegroundColor Red
}

# macOS (Intel)
Write-Host "`nBuilding for macOS (amd64)..." -ForegroundColor Cyan
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -o "$outputDir/JavaSwitcher-$version-darwin-amd64" ./cmd
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ macOS (Intel) build successful" -ForegroundColor Green
} else {
    Write-Host "✗ macOS (Intel) build failed" -ForegroundColor Red
}

# macOS (Apple Silicon)
Write-Host "`nBuilding for macOS (arm64)..." -ForegroundColor Cyan
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -o "$outputDir/JavaSwitcher-$version-darwin-arm64" ./cmd
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ macOS (Apple Silicon) build successful" -ForegroundColor Green
} else {
    Write-Host "✗ macOS (Apple Silicon) build failed" -ForegroundColor Red
}

# Linux (64-bit)
Write-Host "`nBuilding for Linux (amd64)..." -ForegroundColor Cyan
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o "$outputDir/JavaSwitcher-$version-linux-amd64" ./cmd
if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Linux build successful" -ForegroundColor Green
} else {
    Write-Host "✗ Linux build failed" -ForegroundColor Red
}

# Reset environment variables
Remove-Item Env:\GOOS
Remove-Item Env:\GOARCH

Write-Host "`nBuild complete! Files are in the '$outputDir' directory." -ForegroundColor Green
Write-Host "`nNote: For Windows code signing, you need a code signing certificate." -ForegroundColor Yellow
Write-Host "See SIGNING.md for instructions on how to sign the Windows executable." -ForegroundColor Yellow

# List built files
Write-Host "`nBuilt files:" -ForegroundColor Cyan
Get-ChildItem $outputDir | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    Write-Host "  $($_.Name) - $size MB"
}
