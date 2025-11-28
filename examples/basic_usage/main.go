package main

import (
	"fmt"
	"log"

	"github.com/whispem/mini-kvstore-go/pkg/store"
)

func main() {
	fmt.Println("=== Basic Usage Example ===\n")

	// Open store
	kvstore, err := store.Open("examples/basic_usage/data")
	if err != nil {
		log.Fatalf("Failed to open store: %v", err)
	}
	defer kvstore.Close()

	// Set a key
	if err := kvstore.Set("user:1:name", []byte("Alice")); err != nil {
		log.Fatalf("Failed to set key: %v", err)
	}
	fmt.Println("✓ Set user:1:name = Alice")

	// Get the key
	value, err := kvstore.Get("user:1:name")
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	fmt.Printf("✓ Get user:1:name = %s\n", string(value))

	// List keys
	keys := kvstore.ListKeys()
	fmt.Printf("\n✓ Total keys: %d\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s\n", key)
	}

	// Show stats
	stats := kvstore.Stats()
	fmt.Printf("\n%s\n", stats.String())

	fmt.Println("\n✓ Basic usage example completed!")
}
