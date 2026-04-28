# File Keepalive Setup Script (Windows PowerShell)
# This script helps you set up the file keepalive service

$ErrorActionPreference = "Stop"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "File Keepalive Service Setup" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "✓ Go is installed: $goVersion" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "Error: Go is not installed" -ForegroundColor Red
    Write-Host "Please install Go 1.21 or later from https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

# Check if .env file exists
if (-not (Test-Path .env)) {
    Write-Host "Creating .env file from template..." -ForegroundColor Yellow
    Copy-Item .env.example .env
    Write-Host "✓ Created .env file" -ForegroundColor Green
    Write-Host ""
    Write-Host "⚠️  IMPORTANT: Edit .env and add your Supabase credentials:" -ForegroundColor Yellow
    Write-Host "   - SUPABASE_URL" -ForegroundColor Yellow
    Write-Host "   - SUPABASE_KEY" -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to open .env in Notepad"
    notepad .env
} else {
    Write-Host "✓ .env file already exists" -ForegroundColor Green
}

Write-Host ""
Write-Host "Building the application..." -ForegroundColor Cyan
go build -o file-keepalive.exe

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Build successful" -ForegroundColor Green
    Write-Host ""
    Write-Host "=========================================" -ForegroundColor Cyan
    Write-Host "Setup Complete!" -ForegroundColor Green
    Write-Host "=========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "1. Test with dry-run mode:" -ForegroundColor White
    Write-Host "   .\file-keepalive.exe -dry-run -once" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "2. Run a real check:" -ForegroundColor White
    Write-Host "   .\file-keepalive.exe -once" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "3. Run as a service (loops every 7 days):" -ForegroundColor White
    Write-Host "   .\file-keepalive.exe" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "4. Set up automated scheduling:" -ForegroundColor White
    Write-Host "   - GitHub Actions: See README.md" -ForegroundColor Yellow
    Write-Host "   - Task Scheduler: See README.md" -ForegroundColor Yellow
    Write-Host ""
} else {
    Write-Host "✗ Build failed" -ForegroundColor Red
    exit 1
}
