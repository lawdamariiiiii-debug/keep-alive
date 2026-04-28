package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// KeepaliveService handles periodic file access to prevent deletion
type KeepaliveService struct {
	supabase     *SupabaseClient
	httpClient   *http.Client
	dryRun       bool
	maxAgeDays   int
	delaySeconds int
	stats        *Stats
	stateManager *StateManager
	antiBot      *AntiBotManager
	rateLimiter  *RateLimiter
}

// Stats tracks keepalive statistics
type Stats struct {
	mu                sync.Mutex
	TotalFiles        int
	GofileSuccess     int
	GofileFailed      int
	FilesterSuccess   int
	FilesterFailed    int
	ChunksSuccess     int
	ChunksFailed      int
	TotalDownloaded   int64
	StartTime         time.Time
	LastCheckTime     time.Time
}

// NewKeepaliveService creates a new keepalive service
func NewKeepaliveService(supabase *SupabaseClient, dryRun bool, maxAgeDays int, delaySeconds int) *KeepaliveService {
	return &KeepaliveService{
		supabase: supabase,
		httpClient: &http.Client{
			Timeout: 0, // No timeout for large downloads (intentional for large files)
		},
		dryRun:       dryRun,
		maxAgeDays:   maxAgeDays,
		delaySeconds: delaySeconds,
		stateManager: NewStateManager("keepalive-state.json"),
		antiBot:      NewAntiBotManager(),
		rateLimiter:  NewRateLimiter(100, 5), // Max 100/hour, 5/minute
		stats: &Stats{
			StartTime: time.Now(),
		},
	}
}

// StartLoop runs the keepalive check in a loop
func (ks *KeepaliveService) StartLoop(ctx context.Context, interval time.Duration) {
	// Run immediately on start
	if err := ks.CheckAllFiles(ctx); err != nil {
		log.Printf("Error during initial keepalive check: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping keepalive loop")
			ks.PrintFinalStats()
			return
		case <-ticker.C:
			if err := ks.CheckAllFiles(ctx); err != nil {
				log.Printf("Error during keepalive check: %v", err)
			}
		}
	}
}

// CheckAllFiles retrieves and accesses all recent files
func (ks *KeepaliveService) CheckAllFiles(ctx context.Context) error {
	log.Println("========================================")
	log.Println("Starting keepalive check...")
	log.Println("========================================")

	// Calculate cutoff date
	var cutoffDate string
	if ks.maxAgeDays > 0 {
		cutoffDate = time.Now().AddDate(0, 0, -ks.maxAgeDays).Format("2006-01-02")
		log.Printf("Checking files uploaded since: %s (%d days)", cutoffDate, ks.maxAgeDays)
	} else {
		cutoffDate = "1970-01-01" // Unix epoch - effectively all files
		log.Printf("Checking ALL files (no date limit)")
	}

	// Test Supabase connection first
	if err := ks.supabase.TestConnection(); err != nil {
		return fmt.Errorf("Supabase connection test failed: %w", err)
	}
	log.Println("✓ Supabase connection OK")

	// Get recent recordings
	recordings, err := ks.supabase.GetRecentRecordings(cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to get recordings: %w", err)
	}

	log.Printf("Found %d recordings to check", len(recordings))

	if len(recordings) == 0 {
		log.Println("No recordings found, nothing to do")
		return nil
	}

	// Reset stats for this check
	ks.stats.mu.Lock()
	ks.stats.TotalFiles = len(recordings)
	ks.stats.GofileSuccess = 0
	ks.stats.GofileFailed = 0
	ks.stats.FilesterSuccess = 0
	ks.stats.FilesterFailed = 0
	ks.stats.ChunksSuccess = 0
	ks.stats.ChunksFailed = 0
	ks.stats.LastCheckTime = time.Now()
	ks.stats.mu.Unlock()

	// Process each recording
	for i, rec := range recordings {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping file checks")
			ks.stateManager.Stop()
			return ctx.Err()
		default:
		}

		// Skip if already processed
		if ks.stateManager.IsProcessed(rec.ID) {
			log.Printf("\n[%d/%d] SKIPPING (already processed): %s/%s (ID: %s)",
				i+1, len(recordings), rec.Site, rec.Channel, rec.ID)
			continue
		}

		log.Printf("\n[%d/%d] Processing: %s/%s (Date: %s, ID: %s)",
			i+1, len(recordings), rec.Site, rec.Channel, rec.Date, rec.ID)

		fileSuccess := true

		// Download Gofile URL
		if rec.GofileURL != "" {
			if err := ks.accessFile(ctx, rec.GofileURL, "Gofile", rec.ID); err != nil {
				log.Printf("  ✗ Gofile download failed: %v", err)
				ks.stats.mu.Lock()
				ks.stats.GofileFailed++
				ks.stats.mu.Unlock()
				fileSuccess = false
			} else {
				log.Printf("  ✓ Gofile downloaded successfully")
				ks.stats.mu.Lock()
				ks.stats.GofileSuccess++
				ks.stats.mu.Unlock()
				ks.stateManager.IncrementGofile()
			}
		} else {
			log.Printf("  - Gofile URL not available")
		}

		// Download Filester URL
		if rec.FilesterURL != "" {
			if err := ks.accessFile(ctx, rec.FilesterURL, "Filester", rec.ID); err != nil {
				log.Printf("  ✗ Filester download failed: %v", err)
				ks.stats.mu.Lock()
				ks.stats.FilesterFailed++
				ks.stats.mu.Unlock()
				fileSuccess = false
			} else {
				log.Printf("  ✓ Filester downloaded successfully")
				ks.stats.mu.Lock()
				ks.stats.FilesterSuccess++
				ks.stats.mu.Unlock()
				ks.stateManager.IncrementFilester()
			}
		} else {
			log.Printf("  - Filester URL not available")
		}

		// Download Filester chunks if present
		if len(rec.FilesterChunks) > 0 {
			log.Printf("  Downloading %d Filester chunks...", len(rec.FilesterChunks))
			for j, chunkURL := range rec.FilesterChunks {
				if err := ks.accessFile(ctx, chunkURL, fmt.Sprintf("Filester-Chunk-%d", j+1), rec.ID); err != nil {
					log.Printf("    ✗ Chunk %d failed: %v", j+1, err)
					ks.stats.mu.Lock()
					ks.stats.ChunksFailed++
					ks.stats.mu.Unlock()
					fileSuccess = false
				} else {
					log.Printf("    ✓ Chunk %d downloaded", j+1)
					ks.stats.mu.Lock()
					ks.stats.ChunksSuccess++
					ks.stats.mu.Unlock()
					ks.stateManager.IncrementChunks()
				}
			}
		}

		// Mark as processed or failed
		if fileSuccess {
			ks.stateManager.MarkProcessed(rec.ID)
		} else {
			ks.stateManager.MarkFailed(rec.ID)
		}

		// Print progress stats every 10 files
		if (i+1)%10 == 0 {
			ks.stateManager.PrintStats()
		}

		// Small delay between recordings to avoid rate limiting and anti-bot detection
		log.Printf("\n⏳ Waiting %d seconds before next file...", ks.delaySeconds)
		
		// Sleep in smaller chunks to allow for graceful cancellation
		for i := 0; i < ks.delaySeconds; i += 10 {
			select {
			case <-ctx.Done():
				log.Println("Context cancelled during delay")
				ks.stateManager.Stop()
				return ctx.Err()
			default:
				time.Sleep(10 * time.Second)
				if i%60 == 0 && i > 0 {
					log.Printf("  ⏳ %d seconds remaining...", ks.delaySeconds-i)
				}
			}
		}
	}

	// Print summary
	ks.PrintStats()

	log.Println("========================================")
	log.Println("Keepalive check completed")
	log.Println("========================================")

	return nil
}

// accessFile fully downloads the file to ensure it counts as a complete download,
// then immediately deletes it. This guarantees the file is marked as "downloaded"
// on Gofile/Filester, resetting the inactivity timer.
func (ks *KeepaliveService) accessFile(ctx context.Context, url, service, recordingID string) error {
	if ks.dryRun {
		log.Printf("  [DRY RUN] Would download and delete %s: %s", service, url)
		return nil
	}

	// Check rate limiter
	ks.rateLimiter.WaitIfNeeded()
	
	// Simulate human behavior (random micro-delay)
	ks.antiBot.SimulateHumanBehavior()

	log.Printf("  Downloading %s file...", service)
	startTime := time.Now()

	// Create GET request to download the full file
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add realistic anti-bot headers
	ks.antiBot.AddRealisticHeaders(req)
	ks.antiBot.LogRequest(service, url, req.Header.Get("User-Agent"))

	// Record request for rate limiting
	ks.rateLimiter.RecordRequest()

	resp, err := ks.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code and handle retries
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		shouldRetry, retryDelay := ks.antiBot.ShouldRetry(resp.StatusCode, 0)
		if shouldRetry {
			log.Printf("  ⚠️  Status %d, retrying after %v...", resp.StatusCode, retryDelay)
			
			// Check if context is cancelled before sleeping
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Check context again before retry
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return ks.accessFile(ctx, url, service, recordingID) // Retry
			}
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Get file size from Content-Length header
	contentLength := resp.ContentLength
	var sizeStr string
	if contentLength > 0 {
		sizeStr = formatBytes(contentLength)
	} else {
		sizeStr = "unknown size"
	}

	log.Printf("  Downloading %s (%s)...", service, sizeStr)

	// Download the entire file to /dev/null (discard)
	// Use a progress tracker to show download progress
	written, err := io.Copy(io.Discard, &progressReader{
		reader:        resp.Body,
		total:         contentLength,
		service:       service,
		lastLogTime:   time.Now(),
		lastLogBytes:  0,
		stateManager:  ks.stateManager,
	})
	
	if err != nil {
		return fmt.Errorf("download failed after %s: %w", formatBytes(written), err)
	}

	// Track total downloaded
	ks.stats.mu.Lock()
	ks.stats.TotalDownloaded += written
	ks.stats.mu.Unlock()
	ks.stateManager.AddDownloaded(written)

	duration := time.Since(startTime)
	speed := float64(written) / duration.Seconds()

	log.Printf("  ✓ %s downloaded: %s in %v (%.2f MB/s)", 
		service, formatBytes(written), duration.Round(time.Second), speed/1024/1024)

	// Update last_accessed in Supabase
	if err := ks.supabase.UpdateLastAccessed(recordingID); err != nil {
		log.Printf("  ⚠️  Warning: Failed to update last_accessed timestamp: %v", err)
	}

	return nil
}

// progressReader wraps an io.Reader to track download progress
type progressReader struct {
	reader        io.Reader
	total         int64
	downloaded    int64
	service       string
	lastLogTime   time.Time
	lastLogBytes  int64
	stateManager  *StateManager
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.downloaded += int64(n)

	// Log progress every 5 seconds or every 100MB
	now := time.Now()
	bytesSinceLastLog := pr.downloaded - pr.lastLogBytes
	
	if now.Sub(pr.lastLogTime) >= 5*time.Second || bytesSinceLastLog >= 100*1024*1024 {
		if pr.total > 0 {
			percent := float64(pr.downloaded) / float64(pr.total) * 100
			log.Printf("    Progress: %.1f%% (%s / %s)", 
				percent, formatBytes(pr.downloaded), formatBytes(pr.total))
		} else {
			log.Printf("    Progress: %s downloaded", formatBytes(pr.downloaded))
		}
		pr.lastLogTime = now
		pr.lastLogBytes = pr.downloaded
	}

	return n, err
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// PrintStats prints current statistics
func (ks *KeepaliveService) PrintStats() {
	ks.stats.mu.Lock()
	defer ks.stats.mu.Unlock()

	log.Println("\n========================================")
	log.Println("KEEPALIVE STATISTICS")
	log.Println("========================================")
	log.Printf("Total Recordings: %d", ks.stats.TotalFiles)
	log.Printf("Gofile:   ✓ %d  ✗ %d", ks.stats.GofileSuccess, ks.stats.GofileFailed)
	log.Printf("Filester: ✓ %d  ✗ %d", ks.stats.FilesterSuccess, ks.stats.FilesterFailed)
	if ks.stats.ChunksSuccess > 0 || ks.stats.ChunksFailed > 0 {
		log.Printf("Chunks:   ✓ %d  ✗ %d", ks.stats.ChunksSuccess, ks.stats.ChunksFailed)
	}
	log.Printf("Last Check: %s", ks.stats.LastCheckTime.Format(time.RFC3339))
	log.Println("========================================")
}

// PrintFinalStats prints final statistics on shutdown
func (ks *KeepaliveService) PrintFinalStats() {
	ks.stats.mu.Lock()
	defer ks.stats.mu.Unlock()

	uptime := time.Since(ks.stats.StartTime)

	log.Println("\n========================================")
	log.Println("FINAL STATISTICS")
	log.Println("========================================")
	log.Printf("Service Uptime: %s", uptime.Round(time.Second))
	log.Printf("Total Recordings Processed: %d", ks.stats.TotalFiles)
	log.Printf("Gofile:   ✓ %d  ✗ %d", ks.stats.GofileSuccess, ks.stats.GofileFailed)
	log.Printf("Filester: ✓ %d  ✗ %d", ks.stats.FilesterSuccess, ks.stats.FilesterFailed)
	if ks.stats.ChunksSuccess > 0 || ks.stats.ChunksFailed > 0 {
		log.Printf("Chunks:   ✓ %d  ✗ %d", ks.stats.ChunksSuccess, ks.stats.ChunksFailed)
	}
	log.Println("========================================")
}
