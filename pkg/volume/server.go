package volume

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// StartVolumeServer starts the HTTP server with graceful shutdown
func StartVolumeServer(addr, volumeID, dataDir string, compactionThreshold, compactionIntervalSecs int) error {
	// Create data directory
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}

	// Initialize blob storage
	storage, err := NewBlobStorage(dataDir, volumeID)
	if err != nil {
		return fmt.Errorf("failed to create blob storage: %w", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			log.Printf("Error closing storage: %v", err)
		}
	}()

	// Create HTTP router
	router := CreateRouter(storage)

	// Create HTTP server
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start compaction goroutine
	stopCompaction := make(chan struct{})
	compactionDone := make(chan struct{})
	
	if compactionIntervalSecs > 0 {
		go func() {
			defer close(compactionDone)
			ticker := time.NewTicker(time.Duration(compactionIntervalSecs) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					stats := storage.Stats()
					if stats.NumSegments >= compactionThreshold {
						log.Printf("[%s] Running compaction (segments=%d, threshold=%d)...", 
							volumeID, stats.NumSegments, compactionThreshold)
						if err := storage.Compact(); err != nil {
							log.Printf("[%s] Compaction error: %v", volumeID, err)
						} else {
							log.Printf("[%s] Compaction completed", volumeID)
						}
					}
				case <-stopCompaction:
					return
				}
			}
		}()
	}

	// Setup graceful shutdown
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server listening on %s (volume=%s)", addr, volumeID)
		serverErrors <- server.ListenAndServe()
	}()

	// Wait for interrupt signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		log.Printf("Received signal %v, starting graceful shutdown...", sig)

		// Stop compaction
		close(stopCompaction)
		<-compactionDone

		// Save snapshot before shutdown
		log.Printf("Saving snapshot...")
		if err := storage.SaveSnapshot(); err != nil {
			log.Printf("Warning: failed to save snapshot: %v", err)
		}

		// Shutdown HTTP server with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}

		log.Printf("Server stopped gracefully")
	}

	return nil
}
