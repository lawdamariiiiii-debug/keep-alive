# Quick Start Guide

## 1. Setup Environment

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` and add your credentials:
```
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-api-key-here
```

## 2. Build the Application

```bash
go build -o file-keepalive.exe .
```

## 3. Test Configuration (Dry Run)

```bash
./file-keepalive --dry-run --once
```

This will:
- Test Supabase connection
- Show which files would be processed
- Not actually download anything

## 4. Run Once (Test Real Download)

```bash
./file-keepalive --once --delay-seconds=60
```

This will:
- Download files once
- Use 60-second delay between files (faster for testing)
- Exit after completion

## 5. Run Continuously (Production)

```bash
./file-keepalive --interval-days=7 --delay-seconds=300
```

This will:
- Check files every 7 days
- Use 5-minute delay between downloads (safer for production)
- Run until stopped with Ctrl+C

## Common Commands

### Check only recent files (last 7 days)
```bash
./file-keepalive --max-age-days=7 --once
```

### Run with faster delays (for testing)
```bash
./file-keepalive --delay-seconds=30 --once
```

### Run in background (Linux/macOS)
```bash
nohup ./file-keepalive > keepalive.log 2>&1 &
```

### Run as Windows service
Use Task Scheduler or NSSM (Non-Sucking Service Manager)

## Docker Quick Start

### Build and run with Docker Compose
```bash
cd deploy
docker-compose up -d
```

### View logs
```bash
docker-compose logs -f file-keepalive
```

### Stop service
```bash
docker-compose down
```

## Monitoring

### Check state file
```bash
cat keepalive-state.json
```

### Watch logs in real-time (Linux/macOS)
```bash
tail -f keepalive.log
```

### Watch logs in real-time (Windows PowerShell)
```powershell
Get-Content keepalive.log -Wait -Tail 50
```

## Troubleshooting

### Connection Issues
```bash
# Test with verbose output
./file-keepalive --once --dry-run
```

### Rate Limiting
```bash
# Increase delay between files
./file-keepalive --delay-seconds=600  # 10 minutes
```

### Memory Issues
The service automatically cleans up old entries (max 10,000 processed files).
State file is saved every 30 seconds.

## Recommended Settings

### Conservative (Safe)
```bash
./file-keepalive \
  --interval-days=7 \
  --max-age-days=30 \
  --delay-seconds=300
```

### Aggressive (Faster)
```bash
./file-keepalive \
  --interval-days=3 \
  --max-age-days=14 \
  --delay-seconds=120
```

### Testing
```bash
./file-keepalive \
  --once \
  --max-age-days=7 \
  --delay-seconds=30 \
  --dry-run
```

## Next Steps

1. ✅ Test with `--dry-run --once`
2. ✅ Run once with real downloads
3. ✅ Monitor logs for errors
4. ✅ Set up continuous running
5. ✅ Configure as system service (optional)

For detailed information, see [README.md](README.md)
