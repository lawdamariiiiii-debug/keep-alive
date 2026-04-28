# File Keepalive Service

A Go-based service that automatically downloads files from Gofile and Filester to prevent them from being deleted due to inactivity. The service periodically accesses files stored in a Supabase database to reset their inactivity timers.

## Features

- ✅ **Automatic File Access**: Downloads files completely to ensure they count as "accessed"
- ✅ **Anti-Bot Protection**: Randomized user agents, headers, and delays to avoid detection
- ✅ **Rate Limiting**: Built-in rate limiter to prevent triggering service limits
- ✅ **State Management**: Persistent state with auto-save to resume after interruptions
- ✅ **Graceful Shutdown**: Handles SIGTERM/SIGINT with proper cleanup
- ✅ **Configurable Delays**: Customizable delay between downloads
- ✅ **Memory Leak Prevention**: Automatic cleanup of old processed files
- ✅ **Thread-Safe**: All concurrent operations properly synchronized
- ✅ **Cross-Platform**: Works on Windows, Linux, and macOS

## Recent Bug Fixes (v1.1)

All critical bugs have been fixed:
- ✅ Fixed unused import compilation error
- ✅ Fixed race condition in anti-bot manager (added mutex protection)
- ✅ Fixed context cancellation in retry logic
- ✅ Fixed memory leak in state manager (bounded cache with cleanup)
- ✅ Fixed silent error handling (now logs all errors)
- ✅ Replaced inefficient string functions with stdlib
- ✅ Fixed cross-platform file permissions
- ✅ Added URL and API key validation
- ✅ Made delay configurable via command-line flag

See [BUG_REPORT.md](BUG_REPORT.md) for detailed information.

## Project Structure

```
file-keepalive/
├── deploy/                    # Deployment configurations
│   ├── Dockerfile            # Docker container definition
│   ├── docker-compose.yml    # Docker Compose setup
│   └── file-keepalive.yml    # Kubernetes deployment (if applicable)
├── scripts/                   # Setup and utility scripts
│   ├── setup.sh              # Linux/macOS setup script
│   └── setup.ps1             # Windows PowerShell setup script
├── antibot.go                # Anti-bot detection avoidance
├── keepalive.go              # Main keepalive service logic
├── main.go                   # Application entry point
├── state.go                  # Persistent state management
├── supabase.go               # Supabase database client
├── go.mod                    # Go module definition
├── .env.example              # Environment variables template
└── .gitignore                # Git ignore rules
```

## Prerequisites

- Go 1.21 or higher
- Supabase account with a `recordings` table
- Access to Gofile and/or Filester URLs

## Supabase Schema

Your Supabase `recordings` table should have the following columns:

```sql
CREATE TABLE recordings (
  id TEXT PRIMARY KEY,
  site TEXT,
  channel TEXT,
  timestamp TIMESTAMPTZ,
  date TEXT,
  duration_seconds INTEGER,
  file_size_bytes BIGINT,
  quality TEXT,
  gofile_url TEXT,
  filester_url TEXT,
  filester_chunks TEXT[],
  session_id TEXT,
  matrix_job TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  last_accessed TIMESTAMPTZ  -- Add this column for tracking
);
```

## Installation

### Option 1: Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd file-keepalive

# Build the binary
go build -o file-keepalive .

# Run
./file-keepalive --supabase-url="https://xxx.supabase.co" --supabase-key="your-key"
```

### Option 2: Using Docker

```bash
# Build the Docker image
cd deploy
docker build -t file-keepalive .

# Run with Docker Compose
docker-compose up -d
```

### Option 3: Quick Setup Scripts

**Linux/macOS:**
```bash
chmod +x scripts/setup.sh
./scripts/setup.sh
```

**Windows:**
```powershell
.\scripts\setup.ps1
```

## Configuration

### Environment Variables

Create a `.env` file based on `.env.example`:

```bash
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-api-key-here
```

### Command-Line Flags

```bash
./file-keepalive [options]

Options:
  --supabase-url string       Supabase project URL (or set SUPABASE_URL env var)
  --supabase-key string       Supabase API key (or set SUPABASE_KEY env var)
  --interval-days int         Check interval in days (default: 7)
  --max-age-days int          Only check files uploaded in last N days (default: 30)
  --delay-seconds int         Delay between downloads in seconds (default: 300)
  --once                      Run once and exit (don't loop)
  --dry-run                   Dry run mode (don't actually download files)
```

### Examples

**Run once with custom delay:**
```bash
./file-keepalive --once --delay-seconds=60
```

**Run continuously, checking every 3 days:**
```bash
./file-keepalive --interval-days=3
```

**Dry run to test configuration:**
```bash
./file-keepalive --dry-run --once
```

**Check only recent files (last 7 days):**
```bash
./file-keepalive --max-age-days=7 --delay-seconds=120
```

## How It Works

1. **Fetch Recent Files**: Queries Supabase for recordings created within the specified time window
2. **Skip Processed**: Checks persistent state to skip already-processed files
3. **Download Files**: Fully downloads each file (Gofile, Filester, and chunks) to reset inactivity timer
4. **Anti-Bot Measures**: 
   - Randomized user agents (Chrome, Firefox, Safari, Edge)
   - Realistic browser headers
   - Random delays between requests
   - Rate limiting (max 100/hour, 5/minute)
5. **Update Database**: Updates `last_accessed` timestamp in Supabase
6. **Save State**: Auto-saves progress every 30 seconds
7. **Repeat**: Waits for the specified interval before next check

## State Management

The service maintains a persistent state file (`keepalive-state.json`) that tracks:
- Processed files (with automatic cleanup to prevent memory leaks)
- Download statistics
- Last check time
- Total bytes downloaded

This allows the service to resume after interruptions without re-processing files.

## Anti-Bot Features

- **Randomized User Agents**: Rotates between realistic browser profiles
- **Realistic Headers**: Accept, Accept-Language, Referer, Sec-Fetch-* headers
- **Human-like Delays**: Random micro-delays and configurable delays between files
- **Rate Limiting**: Prevents triggering service rate limits
- **Retry Logic**: Exponential backoff for 429, 503, 504 errors

## Monitoring

The service logs detailed information:
- Connection status
- Files processed/failed
- Download progress and speed
- Rate limiting events
- State saves
- Errors and warnings

Example output:
```
[1/50] Processing: youtube/channel-name (Date: 2024-01-15, ID: abc123)
  🌐 Using Chrome browser profile
  Downloading Gofile file...
  Downloading Gofile (1.2 GB)...
    Progress: 25.0% (300 MB / 1.2 GB)
  ✓ Gofile downloaded: 1.2 GB in 2m30s (8.00 MB/s)
⏳ Waiting 300 seconds before next file...
```

## Troubleshooting

### "Supabase connection test failed"
- Verify your `SUPABASE_URL` and `SUPABASE_KEY` are correct
- Check that the `recordings` table exists
- Ensure your API key has read/write permissions

### "Rate limit reached"
- Increase `--delay-seconds` value
- The service will automatically wait when rate limits are hit

### "Context cancelled during delay"
- This is normal during graceful shutdown (Ctrl+C)
- State is saved automatically

### Memory usage growing
- The service now automatically cleans up old entries (max 10,000 processed files)
- State file is periodically compacted

## Development

### Running Tests
```bash
go test ./...
```

### Building for Different Platforms
```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o file-keepalive-linux

# Windows
GOOS=windows GOARCH=amd64 go build -o file-keepalive.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o file-keepalive-mac
```

## License

[Your License Here]

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues and questions, please open an issue on GitHub.
