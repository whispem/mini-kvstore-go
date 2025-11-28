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

// StartVolumeServer starts the HTTP server
func StartVolumeServer(
	addr string,
	volumeID string,
	dataDir string,
	compactionThreshold int,
	compactionInterval int,
) error {
	fmt.Printf("Initializing storage: %s\n", dataDir)
	storage, err := NewBlobStorage(dataDir, volumeID)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Background compaction
	go func() {
		ticker := time.NewTicker(time.Duration(compactionInterval) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			stats := storage.Stats()
			if stats.NumSegments >= compactionThreshold {
				fmt.Printf("Auto-compaction triggered (%d segments >= %d threshold)\n",
					stats.NumSegments, compactionThreshold)
				start := time.Now()
				if err := storage.Compact(); err != nil {
					log.Printf("✗ Compaction error: %v\n", err)
				} else {
					fmt.Printf("✓ Compaction completed in %.2fs\n", time.Since(start).Seconds())
				}
			}
		}
	}()

	router := CreateRouter(storage)

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-done
		fmt.Println("\n🛑 Received shutdown signal, shutting down gracefully...")

		// Save snapshot before shutdown
		fmt.Println("Saving index snapshot...")
		if err := storage.SaveSnapshot(); err != nil {
			log.Printf("⚠ Failed to save snapshot: %v\n", err)
		} else {
			fmt.Println("✓ Snapshot saved")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v\n", err)
		}

		storage.Close()
	}()

	fmt.Println("✓ Volume server ready")
	fmt.Printf("  Listening: http://%s\n", addr)
	fmt.Printf("  Volume ID: %s\n", volumeID)
	fmt.Printf("  Data dir: %s\n", dataDir)
	fmt.Printf("  Compaction: %d segments, every %ds\n", compactionThreshold, compactionInterval)
	fmt.Println("\n📡 Endpoints:")
	fmt.Println("  GET    /health")
	fmt.Println("  GET    /metrics")
	fmt.Println("  GET    /blobs")
	fmt.Println("  POST   /blobs/:key")
	fmt.Println("  GET    /blobs/:key")
	fmt.Println("  DELETE /blobs/:key")
	fmt.Println("\nPress Ctrl+C to shutdown gracefully\n")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	fmt.Println("Server stopped")
	return nil
}
