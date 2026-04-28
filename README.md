# File Keepalive Service

A service that prevents your Filester files from being deleted due to inactivity by periodically accessing them.

## How It Works

### The Problem
File hosting services like Filester automatically delete files after a period of inactivity (typically 30-45 days without downloads). This can result in losing important recordings.

### The Solution
This service periodically "downloads" your files to reset their inactivity timer, but **without actually saving them to disk**:

1. **Streams files to memory** - Downloads are sent to `io.Discard` (like `/dev/null`)
2. **Registers as activity** - Filester sees the download and resets the inactivity timer
3. **No disk usage** - Files are immediately discarded after streaming
4. **Tracks statistics** - Monitors bytes transferred and success rates

### Why Not Save Files?
- **Storage efficiency** - No need to fill up disk space with temporary files
- **Faster processing** - No disk I/O overhead
- **Same effect** - The download still counts as activity on Filester's servers
- **Automatic cleanup** - No need to delete files afterward

## Features

- ✅ **Filester support** - Works with Filester URLs and multi-part chunks
- ✅ **Supabase integration** - Reads recording URLs from your database
- ✅ **Smart scheduling** - Configurable check intervals and file age filters
- ✅ **Rate limiting** - Prevents overwhelming servers with requests
- ✅ **State management** - Tracks processed files to avoid duplicates
- ✅ **Progress tracking** - Real-time download progress and statistics
- ✅ **Retry logic** - Automatically retries failed downloads
- ✅ **Anti-bot measures** - Realistic headers and delays to avoid detection

## Configuration

### Environment Variables

Create a `.env` file or set these environment variables:

```bash
# Required: Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_KEY=your-supabase-anon-key

# Optional: Filester API Token (for faster access)
FILESTER_API_KEY=your-filester-api-token

# Optional: Override defaults
INTERVAL_DAYS=7        # Check every 7 days
MAX_AGE_DAYS=30        # Only check files from last 30 days (0 = all files)
```

### Command-Line Flags

```bash
# Run once and exit
./file-keepalive -once

# Dry run (don't actually download)
./file-keepalive -dry-run

# Custom check interval
./file-keepalive -interval-days 3

# Only check recent files
./file-keepalive -max-age-days 30

# Custom delay between files (in seconds)
./file-keepalive -delay-seconds 300
```

## Usage

### Local Development

```bash
# Install dependencies
go mod download

# Run the service
go run .

# Run once and exit
go run . -once

# Dry run to test without downloading
go run . -dry-run
```

### GitHub Actions (Recommended)

The service is designed to run as a GitHub Action:

1. Fork this repository
2. Add your secrets in GitHub Settings → Secrets:
   - `SUPABASE_URL`
   - `SUPABASE_KEY`
   - `FILESTER_API_KEY` (optional)
3. The workflow runs automatically every 7 days
4. Check the Actions tab for logs and status

### Docker

```bash
# Build the image
docker build -t file-keepalive .

# Run the container
docker run -e SUPABASE_URL=your-url -e SUPABASE_KEY=your-key file-keepalive
```

## Database Schema

Your Supabase `recordings` table should have these columns:

```sql
CREATE TABLE recordings (
  id UUID PRIMARY KEY,
  site TEXT,
  channel TEXT,
  date DATE,
  gofile_url TEXT,           -- Gofile URL (currently skipped)
  filester_url TEXT,         -- Main Filester URL
  filester_chunks TEXT[],    -- Array of chunk URLs
  last_accessed TIMESTAMPTZ  -- Optional: tracks last keepalive
);
```

## Log Output

The service provides clear logging to show what's happening:

```
[1/57] Processing: chaturbate/username (Date: 2026-04-28, ID: abc123)
  - Skipping Gofile URL (using Filester only)
  Downloading Filester file...
  File slug: XMUBWbv
  Generating download token...
  Downloading (184.24 MB)...
    Progress: 10.5% (19.35 MB / 184.24 MB)
    Progress: 25.0% (46.06 MB / 184.24 MB)
  ✓ Filester keepalive complete: 184.24 MB streamed in 45s (4.09 MB/s) - file discarded
  ✓ Filester keepalive successful

⏳ Waiting 300 seconds before next file...
```

### Key Log Messages

- **"file discarded"** - File was streamed but not saved (this is correct!)
- **"keepalive complete"** - Successfully accessed the file
- **"keepalive successful"** - Inactivity timer has been reset

## Statistics

The service tracks and reports:

```
========================================
KEEPALIVE STATISTICS
========================================
Total Recordings: 57
Filester: ✓ 54  ✗ 3
Chunks:   ✓ 162  ✗ 0
Last Check: 2026-04-28T13:45:00Z
========================================
```

## Troubleshooting

### "No files were found"
- Check your Supabase credentials
- Verify the `recordings` table exists
- Ensure there are recordings in the date range

### "Download failed"
- Check your internet connection
- Verify the Filester URL is still valid
- The file may have already been deleted

### "Rate limited"
- Increase the delay between files: `-delay-seconds 600`
- The service will automatically retry

### "API returned status 401"
- Your Filester API token may be invalid
- Try without the token (public API still works)

## FAQ

### Why does it say "file discarded"?
This is **intentional and correct**! The service streams files to memory and immediately discards them. This still counts as a download on Filester's servers, which resets the inactivity timer. We don't need to save the files to disk.

### Does this use a lot of bandwidth?
Yes, it downloads the full file size each time. However, this is necessary to register as a proper download. The service is designed to run infrequently (every 7 days by default) to minimize bandwidth usage.

### Can I use this with Gofile?
Gofile support is currently disabled because their 2025 API changes require a premium account for downloads. The implementation is in the code but not active. You can enable it by getting a Gofile premium account and uncommenting the Gofile processing code.

### How often should I run this?
- **Filester**: Files are deleted after 45 days of inactivity
- **Recommended**: Run every 7-14 days to be safe
- **Minimum**: At least once every 30 days

### Does this work with multi-part files?
Yes! The service automatically processes all chunks in the `filester_chunks` array.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

For issues or questions:
- Open an issue on GitHub
- Check existing issues for solutions
- Review the logs for error messages

---

**Note**: This service is for personal use to maintain your own files. Please respect the terms of service of file hosting providers and don't abuse their systems.
