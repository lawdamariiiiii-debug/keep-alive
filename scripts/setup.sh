#!/bin/bash

# File Keepalive Setup Script
# This script helps you set up the file keepalive service

set -e

echo "========================================="
echo "File Keepalive Service Setup"
echo "========================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Please install Go 1.21 or later from https://go.dev/dl/"
    exit 1
fi

echo "✓ Go is installed: $(go version)"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo "Creating .env file from template..."
    cp .env.example .env
    echo "✓ Created .env file"
    echo ""
    echo "⚠️  IMPORTANT: Edit .env and add your Supabase credentials:"
    echo "   - SUPABASE_URL"
    echo "   - SUPABASE_KEY"
    echo ""
    read -p "Press Enter to open .env in your default editor..."
    ${EDITOR:-nano} .env
else
    echo "✓ .env file already exists"
fi

echo ""
echo "Building the application..."
go build -o file-keepalive

if [ $? -eq 0 ]; then
    echo "✓ Build successful"
    echo ""
    echo "========================================="
    echo "Setup Complete!"
    echo "========================================="
    echo ""
    echo "Next steps:"
    echo ""
    echo "1. Test with dry-run mode:"
    echo "   ./file-keepalive -dry-run -once"
    echo ""
    echo "2. Run a real check:"
    echo "   ./file-keepalive -once"
    echo ""
    echo "3. Run as a service (loops every 7 days):"
    echo "   ./file-keepalive"
    echo ""
    echo "4. Set up automated scheduling:"
    echo "   - GitHub Actions: See README.md"
    echo "   - Cron job: crontab -e"
    echo "   - Systemd: See README.md"
    echo ""
else
    echo "✗ Build failed"
    exit 1
fi
