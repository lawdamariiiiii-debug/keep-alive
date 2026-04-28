package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"
)

// FilesterAPI handles Filester API interactions
type FilesterAPI struct {
	httpClient *http.Client
	antiBot    *AntiBotManager
	apiToken   string // Bearer token for API authentication
}

// FilesterFileResponse represents the file details response
type FilesterFileResponse struct {
	Success bool `json:"success"`
	Data    struct {
		ID           int    `json:"id"`
		UUID         string `json:"uuid"`
		Slug         string `json:"slug"`
		Name         string `json:"name"`
		Size         int64  `json:"size"`
		Type         string `json:"type"`
		URL          string `json:"url"`
		HasThumbnail bool   `json:"has_thumbnail"`
		CreatedAt    string `json:"created_at"`
		Downloads    int    `json:"downloads"`
		Views        int    `json:"views"`
	} `json:"data"`
	Message string `json:"message"`
}

// NewFilesterAPI creates a new Filester API client
func NewFilesterAPI(antiBot *AntiBotManager, apiToken string) *FilesterAPI {
	return &FilesterAPI{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		antiBot:  antiBot,
		apiToken: apiToken,
	}
}

// ExtractFilesterSlug extracts the slug from a Filester URL
func ExtractFilesterSlug(url string) string {
	// Extract slug from URLs like: https://filester.me/d/aBcDeFg
	re := regexp.MustCompile(`filester\.me/d/([a-zA-Z0-9]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// GetFileDetails retrieves file details from Filester API
func (f *FilesterAPI) GetFileDetails(ctx context.Context, slug string) (*FilesterFileResponse, error) {
	url := fmt.Sprintf("https://u1.filester.me/api/v1/file/%s", slug)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token is available
	if f.apiToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.apiToken))
	}

	f.antiBot.AddRealisticHeaders(req)

	resp, err := f.httpClient.Do(req)
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

	var result FilesterFileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

// DownloadFilester downloads a file from Filester
func (f *FilesterAPI) DownloadFilester(ctx context.Context, url string) (int64, error) {
	log.Printf("  Processing Filester URL...")

	// Extract slug from URL
	slug := ExtractFilesterSlug(url)
	if slug == "" {
		return 0, fmt.Errorf("could not extract slug from URL")
	}

	log.Printf("  File slug: %s", slug)

	// If we have an API token, try to get file details first
	if f.apiToken != "" {
		log.Printf("  Fetching file details via API...")
		fileInfo, err := f.GetFileDetails(ctx, slug)
		if err != nil {
			log.Printf("  Warning: Failed to get file details via API: %v", err)
			log.Printf("  File may be guest-uploaded or not owned by this account")
		} else {
			log.Printf("  File: %s (%s)", fileInfo.Data.Name, formatBytes(fileInfo.Data.Size))
		}
	}

	// Generate download token using the public API
	log.Printf("  Generating download token...")
	downloadURL, err := f.generateDownloadToken(ctx, slug)
	if err != nil {
		return 0, fmt.Errorf("failed to generate download token: %w", err)
	}

	log.Printf("  Starting download...")
	return f.downloadFile(ctx, downloadURL)
}

// generateDownloadToken calls the public download API to get a download URL
func (f *FilesterAPI) generateDownloadToken(ctx context.Context, slug string) (string, error) {
	apiURL := "https://filester.me/api/public/download"
	
	// Create JSON payload
	payload := map[string]string{
		"file_slug": slug,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	f.antiBot.AddRealisticHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Del("Accept-Encoding") // Don't accept gzip

	resp, err := f.httpClient.Do(req)
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

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract download_url
	downloadPath, ok := result["download_url"].(string)
	if !ok || downloadPath == "" {
		return "", fmt.Errorf("no download_url in response: %s", string(body))
	}

	// Construct full download URL with CDN
	// Use cache1.filester.me as the CDN
	downloadURL := fmt.Sprintf("https://cache1.filester.me%s?download=true", downloadPath)
	
	return downloadURL, nil
}

// downloadFile downloads a file from Filester
func (f *FilesterAPI) downloadFile(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token is available
	if f.apiToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.apiToken))
	}

	f.antiBot.AddRealisticHeaders(req)

	client := &http.Client{
		Timeout: 0, // No timeout for large files
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Follow redirects but preserve auth header
			if f.apiToken != "" && len(via) > 0 {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.apiToken))
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check for various success status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
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

	// Download to memory and discard (keepalive - no disk usage)
	written, err := io.Copy(io.Discard, &progressReader{
		reader:       resp.Body,
		total:        contentLength,
		service:      "Filester",
		lastLogTime:  time.Now(),
		lastLogBytes: 0,
	})

	if err != nil {
		return 0, fmt.Errorf("download failed: %w", err)
	}

	return written, nil
}
