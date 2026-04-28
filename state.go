package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// StateManager handles persistent state for resuming after interruptions
type StateManager struct {
	mu           sync.Mutex
	stateFile    string
	state        *State
	autoSave     bool
	saveInterval time.Duration
	stopChan     chan struct{}
}

// State represents the current progress of the keepalive service
type State struct {
	LastCheckTime     time.Time         `json:"last_check_time"`
	ProcessedFiles    map[string]bool   `json:"processed_files"`     // recordingID -> processed
	ProcessedOrder    []string          `json:"processed_order"`     // Track order for cleanup
	CurrentBatch      int               `json:"current_batch"`
	TotalBatches      int               `json:"total_batches"`
	StartTime         time.Time         `json:"start_time"`
	LastSaveTime      time.Time         `json:"last_save_time"`
	TotalDownloaded   int64             `json:"total_downloaded"`    // bytes
	FilesCompleted    int               `json:"files_completed"`
	FilesFailed       int               `json:"files_failed"`
	GofileCompleted   int               `json:"gofile_completed"`
	FilesterCompleted int               `json:"filester_completed"`
	ChunksCompleted   int               `json:"chunks_completed"`
	MaxProcessedFiles int               `json:"max_processed_files"` // Max entries to keep in memory
}

// NewStateManager creates a new state manager
func NewStateManager(stateFile string) *StateManager {
	sm := &StateManager{
		stateFile:    stateFile,
		autoSave:     true,
		saveInterval: 30 * time.Second,
		stopChan:     make(chan struct{}),
		state: &State{
			ProcessedFiles:    make(map[string]bool),
			ProcessedOrder:    make([]string, 0),
			StartTime:         time.Now(),
			MaxProcessedFiles: 10000, // Keep max 10k entries to prevent memory leak
		},
	}

	// Try to load existing state
	if err := sm.Load(); err != nil {
		log.Printf("[StateManager] No existing state found, starting fresh: %v", err)
	} else {
		log.Printf("[StateManager] Loaded existing state: %d files processed", len(sm.state.ProcessedFiles))
		
		// Validate loaded state
		if sm.state.ProcessedFiles == nil {
			sm.state.ProcessedFiles = make(map[string]bool)
		}
		if sm.state.ProcessedOrder == nil {
			sm.state.ProcessedOrder = make([]string, 0)
		}
		if sm.state.MaxProcessedFiles == 0 {
			sm.state.MaxProcessedFiles = 10000
		}
	}

	// Start auto-save goroutine
	if sm.autoSave {
		go sm.autoSaveLoop()
	}

	return sm
}

// Load loads state from disk
func (sm *StateManager) Load() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	data, err := os.ReadFile(sm.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	sm.state = &state
	return nil
}

// Save saves state to disk
func (sm *StateManager) Save() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.state.LastSaveTime = time.Now()

	data, err := json.MarshalIndent(sm.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temp file first, then rename (atomic)
	// Use 0666 for cross-platform compatibility (actual permissions depend on umask)
	tempFile := sm.stateFile + ".tmp"
	if err := os.WriteFile(tempFile, data, 0666); err != nil {
		return fmt.Errorf("failed to write temp state file: %w", err)
	}

	// Create backup of existing state file before replacing
	if _, err := os.Stat(sm.stateFile); err == nil {
		backupFile := sm.stateFile + ".backup"
		// Ignore backup errors - not critical
		_ = os.Rename(sm.stateFile, backupFile)
	}

	if err := os.Rename(tempFile, sm.stateFile); err != nil {
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// autoSaveLoop periodically saves state
func (sm *StateManager) autoSaveLoop() {
	ticker := time.NewTicker(sm.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sm.Save(); err != nil {
				log.Printf("[StateManager] Auto-save failed: %v", err)
			} else {
				log.Printf("[StateManager] State auto-saved (%d files processed)", sm.GetFilesCompleted())
			}
		case <-sm.stopChan:
			return
		}
	}
}

// Stop stops the auto-save loop and saves final state
func (sm *StateManager) Stop() {
	close(sm.stopChan)
	if err := sm.Save(); err != nil {
		log.Printf("[StateManager] Final save failed: %v", err)
	} else {
		log.Printf("[StateManager] Final state saved")
	}
}

// IsProcessed checks if a file has been processed
func (sm *StateManager) IsProcessed(recordingID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state.ProcessedFiles[recordingID]
}

// MarkProcessed marks a file as processed
func (sm *StateManager) MarkProcessed(recordingID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Add to processed files
	if !sm.state.ProcessedFiles[recordingID] {
		sm.state.ProcessedFiles[recordingID] = true
		sm.state.ProcessedOrder = append(sm.state.ProcessedOrder, recordingID)
		sm.state.FilesCompleted++
		
		// Cleanup old entries if we exceed max (prevent memory leak)
		if len(sm.state.ProcessedFiles) > sm.state.MaxProcessedFiles {
			// Remove oldest 20% of entries
			removeCount := sm.state.MaxProcessedFiles / 5
			for i := 0; i < removeCount && i < len(sm.state.ProcessedOrder); i++ {
				oldID := sm.state.ProcessedOrder[i]
				delete(sm.state.ProcessedFiles, oldID)
			}
			sm.state.ProcessedOrder = sm.state.ProcessedOrder[removeCount:]
			log.Printf("[StateManager] Cleaned up %d old entries to prevent memory leak", removeCount)
		}
	}
}

// MarkFailed marks a file as failed
func (sm *StateManager) MarkFailed(recordingID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.FilesFailed++
}

// AddDownloaded adds to total downloaded bytes
func (sm *StateManager) AddDownloaded(bytes int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.TotalDownloaded += bytes
}

// IncrementGofile increments Gofile counter
func (sm *StateManager) IncrementGofile() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.GofileCompleted++
}

// IncrementFilester increments Filester counter
func (sm *StateManager) IncrementFilester() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.FilesterCompleted++
}

// IncrementChunks increments chunks counter
func (sm *StateManager) IncrementChunks() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.ChunksCompleted++
}

// SetBatchInfo sets batch information
func (sm *StateManager) SetBatchInfo(current, total int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state.CurrentBatch = current
	sm.state.TotalBatches = total
}

// GetFilesCompleted returns number of completed files
func (sm *StateManager) GetFilesCompleted() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state.FilesCompleted
}

// GetTotalDownloaded returns total bytes downloaded
func (sm *StateManager) GetTotalDownloaded() int64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state.TotalDownloaded
}

// GetState returns a copy of the current state
func (sm *StateManager) GetState() State {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return *sm.state
}

// Reset resets the state (for new check cycle)
func (sm *StateManager) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state = &State{
		ProcessedFiles:    make(map[string]bool),
		ProcessedOrder:    make([]string, 0),
		StartTime:         time.Now(),
		MaxProcessedFiles: 10000,
	}
}

// PrintStats prints current statistics
func (sm *StateManager) PrintStats() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	uptime := time.Since(sm.state.StartTime)
	
	log.Println("\n========================================")
	log.Println("PROGRESS STATISTICS")
	log.Println("========================================")
	log.Printf("Uptime: %v", uptime.Round(time.Second))
	log.Printf("Files Completed: %d", sm.state.FilesCompleted)
	log.Printf("Files Failed: %d", sm.state.FilesFailed)
	log.Printf("Total Downloaded: %s", formatBytes(sm.state.TotalDownloaded))
	log.Printf("Gofile Downloads: %d", sm.state.GofileCompleted)
	log.Printf("Filester Downloads: %d", sm.state.FilesterCompleted)
	log.Printf("Chunks Downloaded: %d", sm.state.ChunksCompleted)
	
	if sm.state.TotalBatches > 0 {
		log.Printf("Batch Progress: %d / %d", sm.state.CurrentBatch, sm.state.TotalBatches)
	}
	
	if uptime.Seconds() > 0 {
		avgSpeed := float64(sm.state.TotalDownloaded) / uptime.Seconds()
		log.Printf("Average Speed: %.2f MB/s", avgSpeed/1024/1024)
	}
	
	log.Printf("Last Save: %v ago", time.Since(sm.state.LastSaveTime).Round(time.Second))
	log.Println("========================================")
}
