package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// GofileAPI handles Gofile API interactions
type GofileAPI struct {
	httpClient *http.Client
	antiBot    *AntiBotManager
	apiKey     string // Premium API key (optional)
}

// GofileResponse represents the API response structure
type GofileResponse struct {
	Status string `json:"status"`
	Data   struct {
		Token         string                 `json:"token"`
		Contents      map[string]interface{} `json:"contents"`
		DownloadCount int                    `json:"downloadCount"`
	} `json:"data"`
}

// NewGofileAPI creates a new Gofile API client
func NewGofileAPI(antiBot *AntiBotManager, apiKey string) *GofileAPI {
	return &GofileAPI{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		antiBot: antiBot,
		apiKey:  apiKey,
	}
}

// CreateGuestAccount creates a guest account and returns the token using the new API
func (g *GofileAPI) CreateGuestAccount(ctx context.Context) (string, error) {
	// Use POST with empty JSON body to the new endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.gofile.io/accounts", strings.NewReader("{}"))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	g.antiBot.AddRealisticHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	
	// Don't request gzip encoding to avoid decompression issues
	req.Header.Del("Accept-Encoding")
	
	// Generate website token for guest account creation (empty account token)
	websiteToken := g.generateWebsiteToken("", req.Header.Get("User-Agent"))
	req.Header.Set("X-Website-Token", websiteToken)
	req.Header.Set("X-BL", "en-US")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GofileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON (response: %s): %w", string(body[:min(len(body), 200)]), err)
	}

	if result.Status != "ok" || result.Data.Token == "" {
		return "", fmt.Errorf("failed to create guest account: %s (response: %s)", result.Status, string(body))
	}

	return result.Data.Token, nil
}

// generateWebsiteToken generates the dynamic X-Website-Token required by Gofile API
// Based on: user_agent + "::en-US::" + account_token + "::" + time_slot + "::4fd6sg89d7s6"
func (g *GofileAPI) generateWebsiteToken(accountToken, userAgent string) string {
	// Time slot is current Unix timestamp divided by 14400 (4 hours)
	timeSlot := time.Now().Unix() / 14400
	
	// Static token from Gofile's config.js
	staticToken := "4fd6sg89d7s6"
	
	// Build the raw string
	raw := fmt.Sprintf("%s::en-US::%s::%d::%s", userAgent, accountToken, timeSlot, staticToken)
	
	// Hash with SHA256
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

// GetContent retrieves content information from Gofile using the new API
func (g *GofileAPI) GetContent(ctx context.Context, contentID, token string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.gofile.io/contents/%s?cache=true", contentID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	g.antiBot.AddRealisticHeaders(req)
	
	// Don't request gzip encoding to avoid decompression issues
	req.Header.Del("Accept-Encoding")
	
	// Use Bearer token authentication
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	
	// Generate and add website token
	websiteToken := g.generateWebsiteToken(token, req.Header.Get("User-Agent"))
	req.Header.Set("X-Website-Token", websiteToken)
	req.Header.Set("X-BL", "en-US")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON (response: %s): %w", string(body[:min(len(body), 200)]), err)
	}

	status, ok := result["status"].(string)
	if !ok || status != "ok" {
		return nil, fmt.Errorf("API returned error status: %v", result)
	}

	return result, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExtractContentID extracts the content ID from a Gofile URL
func ExtractGofileContentID(url string) string {
	// Extract ID from URLs like: https://gofile.io/d/abc123
	re := regexp.MustCompile(`gofile\.io/d/([a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// DownloadGofile downloads a file from Gofile using the API
func (g *GofileAPI) DownloadGofile(ctx context.Context, url string) (int64, error) {
	log.Printf("  Processing Gofile URL...")

	// Extract content ID
	contentID := ExtractGofileContentID(url)
	if contentID == "" {
		return 0, fmt.Errorf("could not extract content ID from URL")
	}

	log.Printf("  Content ID: %s", contentID)

	// Try premium API first if we have a key
	if g.apiKey != "" {
		log.Printf("  Trying premium API...")
		written, err := g.downloadWithPremiumAPI(ctx, contentID)
		if err == nil {
			return written, nil
		}
		log.Printf("  Premium API failed: %v", err)
		log.Printf("  Falling back to guest account method...")
	}

	// Fall back to guest account method (works for public files)
	log.Printf("  Using guest account method...")
	return g.downloadWithGuestAccount(ctx, contentID, url)
}

// downloadWithPremiumAPI uses the premium API with account token
func (g *GofileAPI) downloadWithPremiumAPI(ctx context.Context, contentID string) (int64, error) {
	// Get content info using the unified GetContent method
	content, err := g.GetContent(ctx, contentID, g.apiKey)
	if err != nil {
		return 0, err
	}

	// Extract download link
	data, ok := content["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid content data structure")
	}

	children, ok := data["children"].(map[string]interface{})
	if !ok || len(children) == 0 {
		return 0, fmt.Errorf("no files found in content")
	}

	// Get the first file's download link
	var downloadURL string
	var fileName string
	for _, fileData := range children {
		fileInfo, ok := fileData.(map[string]interface{})
		if !ok {
			continue
		}
		
		link, ok := fileInfo["link"].(string)
		if ok && link != "" {
			downloadURL = link
			fileName, _ = fileInfo["name"].(string)
			log.Printf("  Found file: %s", fileName)
			break
		}
	}

	if downloadURL == "" {
		return 0, fmt.Errorf("no download link found")
	}

	log.Printf("  Starting download...")
	return g.downloadWithToken(ctx, downloadURL, g.apiKey)
}

// downloadWithGuestAccount uses the new guest account method (works for public files)
func (g *GofileAPI) downloadWithGuestAccount(ctx context.Context, contentID, pageURL string) (int64, error) {
	// Create guest account using new API
	log.Printf("  Creating guest account...")
	token, err := g.CreateGuestAccount(ctx)
	if err != nil {
		// If rate limited or API fails, try direct page scraping
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rateLimit") {
			log.Printf("  Rate limited, trying direct page access...")
			return g.downloadViaPageScraping(ctx, contentID, pageURL)
		}
		return 0, fmt.Errorf("failed to create guest account: %w", err)
	}

	log.Printf("  Guest account created successfully")

	// Get content info using new API
	log.Printf("  Fetching content info...")
	content, err := g.GetContent(ctx, contentID, token)
	if err != nil {
		// If API returns notPremium or other errors, try page scraping
		if strings.Contains(err.Error(), "notPremium") || strings.Contains(err.Error(), "401") {
			log.Printf("  API requires premium, trying direct page access...")
			return g.downloadViaPageScraping(ctx, contentID, pageURL)
		}
		return 0, fmt.Errorf("failed to get content: %w", err)
	}

	// Extract download link from the new API response structure
	data, ok := content["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("invalid content data structure")
	}

	// Check for children (new API structure)
	children, ok := data["children"].(map[string]interface{})
	if !ok || len(children) == 0 {
		return 0, fmt.Errorf("no files found in content")
	}

	// Get the first file's download link
	var downloadURL string
	var fileName string
	for _, fileData := range children {
		fileInfo, ok := fileData.(map[string]interface{})
		if !ok {
			continue
		}
		
		link, ok := fileInfo["link"].(string)
		if ok && link != "" {
			downloadURL = link
			fileName, _ = fileInfo["name"].(string)
			log.Printf("  Found file: %s", fileName)
			break
		}
	}

	if downloadURL == "" {
		return 0, fmt.Errorf("no download link found")
	}

	// Download the file with the guest token
	log.Printf("  Starting download...")
	return g.downloadWithToken(ctx, downloadURL, token)
}

// downloadViaPageScraping attempts to download by loading the page and extracting download links
// This is a fallback when API methods fail (rate limit, premium required, etc.)
func (g *GofileAPI) downloadViaPageScraping(ctx context.Context, contentID, pageURL string) (int64, error) {
	log.Printf("  Loading Gofile page to extract download link...")
	
	// Note: This method requires browser automation which is currently blocked by Windows Defender
	// For now, return an informative error
	return 0, fmt.Errorf("page scraping requires browser automation (currently blocked by Windows Defender). " +
		"Options: 1) Use a premium Gofile API key, 2) Manually whitelist the browser in Windows Defender, " +
		"3) Use a different download method")
}



// downloadWithToken downloads a file with the Gofile account token
func (g *GofileAPI) downloadWithToken(ctx context.Context, url, token string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers including the account token cookie
	g.antiBot.AddRealisticHeaders(req)
	req.Header.Set("Cookie", fmt.Sprintf("accountToken=%s", token))

	client := &http.Client{
		Timeout: 0, // No timeout for large files
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Get file size
	contentLength := resp.ContentLength
	var sizeStr string
	if contentLength > 0 {
		sizeStr = formatBytes(contentLength)
	} else {
		sizeStr = "unknown size"
	}

	log.Printf("  Downloading (%s)...", sizeStr)

	// Download to /dev/null (discard)
	written, err := io.Copy(io.Discard, &progressReader{
		reader:       resp.Body,
		total:        contentLength,
		service:      "Gofile",
		lastLogTime:  time.Now(),
		lastLogBytes: 0,
	})

	if err != nil {
		return 0, fmt.Errorf("download failed: %w", err)
	}

	return written, nil
}

// BrowserDownloader handles browser-based file downloads for sites without APIs
type BrowserDownloader struct {
	browser       *rod.Browser
	antiBot       *AntiBotManager
	gofileAPI     *GofileAPI
	filesterAPI   *FilesterAPI
}

// NewBrowserDownloader creates a new browser downloader
func NewBrowserDownloader(antiBot *AntiBotManager, gofileAPIKey, filesterToken string) (*BrowserDownloader, error) {
	// Only launch browser if we don't have API tokens (fallback mode)
	var browser *rod.Browser
	
	if gofileAPIKey == "" || filesterToken == "" {
		log.Println("  Launching browser for fallback (no API tokens provided)...")
		l := launcher.New().
			Headless(true).
			Set("disable-blink-features", "AutomationControlled").
			Set("disable-web-security").
			Delete("enable-automation")

		// Try to use system Chrome if available
		path, exists := launcher.LookPath()
		if exists {
			l.Bin(path)
		}

		url, err := l.Launch()
		if err != nil {
			log.Printf("  Warning: Failed to launch browser: %v", err)
			log.Println("  Browser automation will not be available")
		} else {
			browser = rod.New().ControlURL(url).MustConnect()
		}
	} else {
		log.Println("  Using API-based downloads (browser not needed)")
	}

	return &BrowserDownloader{
		browser:     browser,
		antiBot:     antiBot,
		gofileAPI:   NewGofileAPI(antiBot, gofileAPIKey),
		filesterAPI: NewFilesterAPI(antiBot, filesterToken),
	}, nil
}

// Close closes the browser and cleans up
func (bd *BrowserDownloader) Close() error {
	if bd.browser != nil {
		bd.browser.MustClose()
	}
	return nil
}

// DownloadFile downloads a file using the appropriate method
func (bd *BrowserDownloader) DownloadFile(ctx context.Context, url, service string) (int64, error) {
	// Use API for Gofile
	if strings.Contains(strings.ToLower(url), "gofile") {
		return bd.gofileAPI.DownloadGofile(ctx, url)
	}

	// Use API for Filester
	if strings.Contains(strings.ToLower(url), "filester") {
		return bd.filesterAPI.DownloadFilester(ctx, url)
	}

	// Use browser automation for other sites
	return bd.downloadViaBrowser(ctx, url, service)
}

// downloadViaBrowser downloads using browser automation
func (bd *BrowserDownloader) downloadViaBrowser(ctx context.Context, url, service string) (int64, error) {
	if bd.browser == nil {
		return 0, fmt.Errorf("browser not available and no API token provided for %s", service)
	}
	
	log.Printf("  Opening %s page in browser...", service)

	// Create a new page
	page, err := bd.browser.Page(proto.TargetCreateTarget{})
	if err != nil {
		return 0, fmt.Errorf("failed to create page: %w", err)
	}
	defer page.MustClose()

	// Set realistic viewport
	page.MustSetViewport(1920, 1080, 1, false)

	// Set user agent
	userAgent := bd.antiBot.GetRandomUserAgent()
	page.MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: userAgent,
	})

	// Navigate to the page
	if err := page.Navigate(url); err != nil {
		return 0, fmt.Errorf("failed to navigate to page: %w", err)
	}

	// Wait for page to load
	page.MustWaitLoad()

	// Add human-like delay
	bd.antiBot.SimulateHumanBehavior()

	// Try to find and click download button
	downloadURL, err := bd.handleGenericDownload(ctx, page)
	if err != nil {
		return 0, fmt.Errorf("failed to get download URL: %w", err)
	}

	if downloadURL == "" {
		return 0, fmt.Errorf("no download URL found")
	}

	log.Printf("  Found download URL, starting download...")

	// Download the file
	return bd.downloadViaHTTP(ctx, downloadURL, service)
}

// handleGenericDownload handles generic file hosting sites
func (bd *BrowserDownloader) handleGenericDownload(ctx context.Context, page *rod.Page) (string, error) {
	// Wait for page to fully load
	time.Sleep(3 * time.Second)

	// Try to find and click download button
	selectors := []string{
		`button:contains("Download")`,
		`a:contains("Download")`,
		`button:contains("download")`,
		`a:contains("download")`,
		`[class*="download"]`,
		`[id*="download"]`,
		`button[class*="btn"]`,
		`a[class*="btn"]`,
	}

	for _, selector := range selectors {
		elem, err := page.Timeout(3 * time.Second).Element(selector)
		if err == nil && elem != nil {
			log.Printf("  Found download button, clicking...")
			elem.MustClick()
			time.Sleep(2 * time.Second)
			break
		}
	}

	// Try to extract download URL
	return bd.extractDownloadURL(page)
}

// extractDownloadURL tries to extract the actual download URL from the page
func (bd *BrowserDownloader) extractDownloadURL(page *rod.Page) (string, error) {
	// Method 1: Check for direct download links
	links, err := page.Elements("a[href]")
	if err == nil {
		for _, link := range links {
			href, err := link.Attribute("href")
			if err != nil || href == nil {
				continue
			}

			url := *href
			// Look for direct file links
			if strings.Contains(url, ".mp4") ||
				strings.Contains(url, ".mkv") ||
				strings.Contains(url, ".avi") ||
				strings.Contains(url, ".webm") ||
				strings.Contains(url, "/download/") ||
				strings.Contains(url, "/file/") {
				return url, nil
			}
		}
	}

	// Method 2: Check current page URL
	currentURL := page.MustInfo().URL
	if strings.Contains(currentURL, "/download/") ||
		strings.Contains(currentURL, ".mp4") ||
		strings.Contains(currentURL, ".mkv") {
		return currentURL, nil
	}

	// Method 3: Try to find video elements
	videos, err := page.Elements("video[src]")
	if err == nil && len(videos) > 0 {
		src, err := videos[0].Attribute("src")
		if err == nil && src != nil {
			return *src, nil
		}
	}

	return "", fmt.Errorf("could not extract download URL from page")
}

// downloadViaHTTP downloads a file via direct HTTP request
func (bd *BrowserDownloader) downloadViaHTTP(ctx context.Context, url, service string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	bd.antiBot.AddRealisticHeaders(req)

	client := &http.Client{
		Timeout: 0, // No timeout for large files
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentLength := resp.ContentLength
	var sizeStr string
	if contentLength > 0 {
		sizeStr = formatBytes(contentLength)
	} else {
		sizeStr = "unknown size"
	}

	log.Printf("  Downloading %s (%s)...", service, sizeStr)

	written, err := io.Copy(io.Discard, &progressReader{
		reader:       resp.Body,
		total:        contentLength,
		service:      service,
		lastLogTime:  time.Now(),
		lastLogBytes: 0,
	})

	if err != nil {
		return 0, fmt.Errorf("download failed: %w", err)
	}

	return written, nil
}
