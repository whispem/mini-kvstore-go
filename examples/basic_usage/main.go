package main

import (
	"fmt"
	"log"
	"os"

	"mini-kvstore-go/pkg/store"
)

func main() {
	db, err := store.NewStore("example.db")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove("example.db")

	_ = db.Set("foo", []byte("bar"))   // Set a value
	val, _ := db.Get("foo")            // Get a value
	fmt.Println("Get foo:", string(val)) 
}
