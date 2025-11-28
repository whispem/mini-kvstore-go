package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/whispem/mini-kvstore-go/pkg/store"
)

func main() {
	kvstore, err := store.Open("db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open store: %v\n", err)
		os.Exit(1)
	}
	defer kvstore.Close()

	fmt.Println("mini-kvstore-go (type help for instructions)")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 3)
		cmd := parts[0]

		switch cmd {
		case "set":
			if len(parts) < 3 {
				fmt.Println("Usage: set <key> <value>")
				continue
			}
			key := parts[1]
			value := parts[2]

			if err := kvstore.Set(key, []byte(value)); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("OK")
			}

		case "get":
			if len(parts) < 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			key := parts[1]

			value, err := kvstore.Get(key)
			if err == store.ErrNotFound {
				fmt.Println("Key not found")
			} else if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println(string(value))
			}

		case "delete":
			if len(parts) < 2 {
				fmt.Println("Usage: delete <key>")
				continue
			}
			key := parts[1]

			if err := kvstore.Delete(key); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Deleted")
			}

		case "list":
			keys := kvstore.ListKeys()
			for _, key := range keys {
				fmt.Printf("  %s\n", key)
			}

		case "compact":
			if err := kvstore.Compact(); err != nil {
				fmt.Printf("Compaction error: %v\n", err)
			} else {
				fmt.Println("Compaction finished")
			}

		case "stats":
			stats := kvstore.Stats()
			fmt.Println(stats)

		case "help":
			printHelp()

		case "quit", "exit":
			fmt.Println("Saving snapshot...")
			if err := kvstore.SaveSnapshot(); err != nil {
				fmt.Printf("Warning: failed to save snapshot: %v\n", err)
			}
			return

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  set <key> <value>  - Store a key-value pair")
	fmt.Println("  get <key>          - Retrieve a value")
	fmt.Println("  delete <key>       - Remove a key")
	fmt.Println("  list               - List all keys")
	fmt.Println("  compact            - Compact storage")
	fmt.Println("  stats              - Show statistics")
	fmt.Println("  help               - Show this help")
	fmt.Println("  quit / exit        - Exit the program")
}
