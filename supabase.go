package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SupabaseClient handles communication with Supabase database
type SupabaseClient struct {
	url        string
	apiKey     string
	httpClient *http.Client
}

// Recording represents a recording record from Supabase
type Recording struct {
	ID             string    `json:"id"`
	Site           string    `json:"site"`
	Channel        string    `json:"channel"`
	Timestamp      time.Time `json:"timestamp"`
	Date           string    `json:"date"`
	DurationSec    int       `json:"duration_seconds"`
	FileSizeBytes  int64     `json:"file_size_bytes"`
	Quality        string    `json:"quality"`
	GofileURL      string    `json:"gofile_url"`
	FilesterURL    string    `json:"filester_url"`
	FilesterChunks []string  `json:"filester_chunks"`
	SessionID      string    `json:"session_id"`
	MatrixJob      string    `json:"matrix_job"`
	CreatedAt      time.Time `json:"created_at"`
}

// NewSupabaseClient creates a new Supabase client
func NewSupabaseClient(url, apiKey string) *SupabaseClient {
	return &SupabaseClient{
		url:    url,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetRecentRecordings retrieves recordings created after the specified date
func (sc *SupabaseClient) GetRecentRecordings(sinceDate string) ([]Recording, error) {
	// Build query URL - get recordings created after sinceDate, ordered by created_at desc
	url := fmt.Sprintf("%s/rest/v1/recordings?created_at=gte.%s&order=created_at.desc", sc.url, sinceDate)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("apikey", sc.apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sc.apiKey))
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Supabase returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON
	var recordings []Recording
	if err := json.Unmarshal(body, &recordings); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return recordings, nil
}

// UpdateLastAccessed updates the last_accessed timestamp for a recording
// Note: This requires adding a last_accessed column to your recordings table
func (sc *SupabaseClient) UpdateLastAccessed(recordingID string) error {
	url := fmt.Sprintf("%s/rest/v1/recordings?id=eq.%s", sc.url, recordingID)

	// Create update payload
	update := map[string]interface{}{
		"last_accessed": time.Now().Format(time.RFC3339),
	}
	jsonData, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("apikey", sc.apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sc.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=minimal")

	// Execute request
	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Supabase returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// TestConnection tests the connection to Supabase
func (sc *SupabaseClient) TestConnection() error {
	url := fmt.Sprintf("%s/rest/v1/recordings?limit=1", sc.url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("apikey", sc.apiKey)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sc.apiKey))

	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Supabase returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
