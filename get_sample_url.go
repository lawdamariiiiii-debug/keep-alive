package main

import (
	"log"
	"os"
)

func main() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	
	if supabaseURL == "" || supabaseKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_KEY required")
	}
	
	supabase := NewSupabaseClient(supabaseURL, supabaseKey)
	
	// Get one recent recording
	recordings, err := supabase.GetRecentRecordings("2026-04-01")
	if err != nil {
		log.Fatalf("Failed to get recordings: %v", err)
	}
	
	if len(recordings) == 0 {
		log.Println("No recordings found")
		return
	}
	
	log.Println("Sample URLs from your database:")
	log.Println("================================")
	
	for i, rec := range recordings {
		if i >= 3 {
			break
		}
		log.Printf("\nRecording %d: %s/%s", i+1, rec.Site, rec.Channel)
		log.Printf("  Date: %s", rec.Date)
		if rec.GofileURL != "" {
			log.Printf("  Gofile: %s", rec.GofileURL)
		}
		if rec.FilesterURL != "" {
			log.Printf("  Filester: %s", rec.FilesterURL)
		}
	}
}
