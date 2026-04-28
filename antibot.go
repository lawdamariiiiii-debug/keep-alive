package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AntiBotManager handles anti-bot detection avoidance
type AntiBotManager struct {
	userAgents []string
	referers   []string
	rand       *rand.Rand
	mu         sync.Mutex // Protects rand for thread-safety
}

// NewAntiBotManager creates a new anti-bot manager
func NewAntiBotManager() *AntiBotManager {
	return &AntiBotManager{
		userAgents: []string{
			// Chrome on Windows
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
			
			// Chrome on macOS
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
			
			// Firefox on Windows
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
			
			// Firefox on macOS
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:121.0) Gecko/20100101 Firefox/121.0",
			
			// Safari on macOS
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
			"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
			
			// Edge on Windows
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
		},
		referers: []string{
			"https://www.google.com/",
			"https://www.bing.com/",
			"https://duckduckgo.com/",
			"https://www.reddit.com/",
			"https://twitter.com/",
			"", // Sometimes no referer
		},
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetRandomUserAgent returns a random realistic user agent
func (abm *AntiBotManager) GetRandomUserAgent() string {
	abm.mu.Lock()
	defer abm.mu.Unlock()
	return abm.userAgents[abm.rand.Intn(len(abm.userAgents))]
}

// GetRandomReferer returns a random referer (or empty)
func (abm *AntiBotManager) GetRandomReferer() string {
	abm.mu.Lock()
	defer abm.mu.Unlock()
	return abm.referers[abm.rand.Intn(len(abm.referers))]
}

// AddRealisticHeaders adds realistic browser headers to avoid bot detection
func (abm *AntiBotManager) AddRealisticHeaders(req *http.Request) {
	// User-Agent (random)
	req.Header.Set("User-Agent", abm.GetRandomUserAgent())
	
	// Accept headers (realistic browser values)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	
	// Referer (random, sometimes empty)
	referer := abm.GetRandomReferer()
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	
	// Connection
	req.Header.Set("Connection", "keep-alive")
	
	// Cache control
	req.Header.Set("Cache-Control", "max-age=0")
	
	// Upgrade insecure requests
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	
	// Sec-Fetch headers (modern browsers)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	
	// DNT (Do Not Track)
	abm.mu.Lock()
	shouldSetDNT := abm.rand.Float32() < 0.3 // 30% of users have DNT enabled
	abm.mu.Unlock()
	if shouldSetDNT {
		req.Header.Set("DNT", "1")
	}
}

// GetRandomDelay returns a random delay between min and max seconds
func (abm *AntiBotManager) GetRandomDelay(minSeconds, maxSeconds int) time.Duration {
	abm.mu.Lock()
	defer abm.mu.Unlock()
	delay := minSeconds + abm.rand.Intn(maxSeconds-minSeconds+1)
	return time.Duration(delay) * time.Second
}

// GetHumanizedDelay returns a delay that mimics human behavior
// Base delay with random variation (±20%)
func (abm *AntiBotManager) GetHumanizedDelay(baseSeconds int) time.Duration {
	abm.mu.Lock()
	defer abm.mu.Unlock()
	variation := float64(baseSeconds) * 0.2 // ±20%
	randomVariation := (abm.rand.Float64()*2 - 1) * variation // -20% to +20%
	finalDelay := float64(baseSeconds) + randomVariation
	
	if finalDelay < 1 {
		finalDelay = 1
	}
	
	return time.Duration(finalDelay) * time.Second
}

// SimulateHumanBehavior adds random micro-delays to simulate human interaction
func (abm *AntiBotManager) SimulateHumanBehavior() {
	abm.mu.Lock()
	microDelay := abm.rand.Intn(2000)
	abm.mu.Unlock()
	// Random micro-delay (0-2 seconds) before request
	time.Sleep(time.Duration(microDelay) * time.Millisecond)
}

// GetDownloadChunkSize returns a realistic chunk size for downloads
// Varies to avoid pattern detection
func (abm *AntiBotManager) GetDownloadChunkSize() int {
	abm.mu.Lock()
	defer abm.mu.Unlock()
	// Common buffer sizes: 32KB, 64KB, 128KB
	sizes := []int{32 * 1024, 64 * 1024, 128 * 1024}
	return sizes[abm.rand.Intn(len(sizes))]
}

// ShouldRetry determines if a request should be retried based on status code
func (abm *AntiBotManager) ShouldRetry(statusCode int, attempt int) (bool, time.Duration) {
	maxAttempts := 3
	
	if attempt >= maxAttempts {
		return false, 0
	}
	
	switch statusCode {
	case 429: // Too Many Requests
		// Exponential backoff: 1min, 2min, 4min
		delay := time.Duration(1<<uint(attempt)) * time.Minute
		return true, delay
		
	case 503, 504: // Service Unavailable, Gateway Timeout
		// Linear backoff: 30s, 60s, 90s
		delay := time.Duration(30*(attempt+1)) * time.Second
		return true, delay
		
	case 500, 502: // Internal Server Error, Bad Gateway
		// Short retry: 10s, 20s, 30s
		delay := time.Duration(10*(attempt+1)) * time.Second
		return true, delay
		
	default:
		return false, 0
	}
}

// LogRequest logs request details in a human-readable format
func (abm *AntiBotManager) LogRequest(service, url, userAgent string) {
	// Extract browser from user agent
	browser := "Unknown"
	if strings.Contains(userAgent, "Chrome") && !strings.Contains(userAgent, "Edg") {
		browser = "Chrome"
	} else if strings.Contains(userAgent, "Firefox") {
		browser = "Firefox"
	} else if strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome") {
		browser = "Safari"
	} else if strings.Contains(userAgent, "Edg") {
		browser = "Edge"
	}
	
	fmt.Printf("  🌐 Using %s browser profile\n", browser)
}

// RateLimiter tracks request rates to avoid triggering rate limits
type RateLimiter struct {
	requests      []time.Time
	maxPerHour    int
	maxPerMinute  int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxPerHour, maxPerMinute int) *RateLimiter {
	return &RateLimiter{
		requests:     make([]time.Time, 0),
		maxPerHour:   maxPerHour,
		maxPerMinute: maxPerMinute,
	}
}

// CanMakeRequest checks if a request can be made without exceeding limits
func (rl *RateLimiter) CanMakeRequest() bool {
	now := time.Now()
	
	// Clean old requests (older than 1 hour)
	cutoff := now.Add(-1 * time.Hour)
	newRequests := make([]time.Time, 0)
	for _, t := range rl.requests {
		if t.After(cutoff) {
			newRequests = append(newRequests, t)
		}
	}
	rl.requests = newRequests
	
	// Check hourly limit
	if len(rl.requests) >= rl.maxPerHour {
		return false
	}
	
	// Check per-minute limit
	minuteCutoff := now.Add(-1 * time.Minute)
	recentRequests := 0
	for _, t := range rl.requests {
		if t.After(minuteCutoff) {
			recentRequests++
		}
	}
	
	if recentRequests >= rl.maxPerMinute {
		return false
	}
	
	return true
}

// RecordRequest records a new request
func (rl *RateLimiter) RecordRequest() {
	rl.requests = append(rl.requests, time.Now())
}

// WaitIfNeeded waits if rate limit would be exceeded
func (rl *RateLimiter) WaitIfNeeded() {
	for !rl.CanMakeRequest() {
		fmt.Println("  ⏳ Rate limit reached, waiting 60 seconds...")
		time.Sleep(60 * time.Second)
	}
}
