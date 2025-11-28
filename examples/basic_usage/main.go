package main

import (
	"fmt"
	"log"
	"os"

	"github.com/whispem/mini-kvstore-go/pkg/store"
)

func main() {
	db, err := store.Open("example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll("example.db"); err != nil {
			log.Printf("Warning: failed to remove example.db: %v", err)
		}
	}()

	if err := db.Set("foo", []byte("bar")); err != nil {
		log.Fatalf("Failed to set value: %v", err)
	}

	val, err := db.Get("foo")
	if err != nil {
		log.Fatalf("Failed to get value: %v", err)
	}

	fmt.Println("Get foo:", string(val))
}
