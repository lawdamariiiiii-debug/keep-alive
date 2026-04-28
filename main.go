package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	// Command-line flags
	supabaseURL := flag.String("supabase-url", os.Getenv("SUPABASE_URL"), "Supabase project URL")
	supabaseKey := flag.String("supabase-key", os.Getenv("SUPABASE_KEY"), "Supabase API key")
	intervalDays := flag.Int("interval-days", 7, "Check interval in days")
	runOnce := flag.Bool("once", false, "Run once and exit (don't loop)")
	dryRun := flag.Bool("dry-run", false, "Dry run mode (don't actually access files)")
	maxAgeDays := flag.Int("max-age-days", 0, "Only check files uploaded in the last N days (0 = all files)")
	delaySeconds := flag.Int("delay-seconds", 300, "Delay between file downloads in seconds (default: 300 = 5 minutes)")
	
	flag.Parse()

	// Validate required parameters
	if *supabaseURL == "" || *supabaseKey == "" {
		log.Fatal("Error: SUPABASE_URL and SUPABASE_KEY are required (via flags or environment variables)")
	}
	
	// Validate URL format
	if !strings.HasPrefix(*supabaseURL, "http://") && !strings.HasPrefix(*supabaseURL, "https://") {
		log.Fatal("Error: SUPABASE_URL must start with http:// or https://")
	}
	
	// Validate API key is not empty
	if len(*supabaseKey) < 10 {
		log.Fatal("Error: SUPABASE_KEY appears to be invalid (too short)")
	}
	
	// Validate delay is non-negative
	if *delaySeconds < 0 {
		log.Fatal("Error: delay-seconds must be >= 0")
	}
	
	// Validate max-age-days is non-negative
	if *maxAgeDays < 0 {
		log.Fatal("Error: max-age-days must be >= 0")
	}

	log.Println("=== File Keepalive Service ===")
	log.Printf("Supabase URL: %s", *supabaseURL)
	log.Printf("Check Interval: %d days", *intervalDays)
	if *maxAgeDays > 0 {
		log.Printf("Max File Age: %d days", *maxAgeDays)
	} else {
		log.Printf("Max File Age: ALL files (no limit)")
	}
	log.Printf("Delay Between Files: %d seconds", *delaySeconds)
	log.Printf("Run Once: %v", *runOnce)
	log.Printf("Dry Run: %v", *dryRun)
	log.Println("==============================")

	// Create Supabase client
	supabase := NewSupabaseClient(*supabaseURL, *supabaseKey)

	// Create keepalive service
	keepalive := NewKeepaliveService(supabase, *dryRun, *maxAgeDays, *delaySeconds)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("\nReceived shutdown signal, stopping...")
		cancel()
	}()

	// Ensure state manager is stopped on exit
	defer keepalive.stateManager.Stop()

	// Run once or loop
	if *runOnce {
		log.Println("Running single keepalive check...")
		if err := keepalive.CheckAllFiles(ctx); err != nil {
			log.Fatalf("Error during keepalive check: %v", err)
		}
		log.Println("Keepalive check completed successfully")
	} else {
		log.Println("Starting keepalive service (press Ctrl+C to stop)...")
		keepalive.StartLoop(ctx, time.Duration(*intervalDays)*24*time.Hour)
	}
}
