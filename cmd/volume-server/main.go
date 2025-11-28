package main

import (
	"fmt"
	"log"
	"os"

	"github.com/whispem/mini-kvstore-go/pkg/config"
	"github.com/whispem/mini-kvstore-go/pkg/volume"
)

func main() {
	cfg := config.FromEnv()

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)

	fmt.Println("Starting volume server:")
	fmt.Printf("  volume_id = %s\n", cfg.VolumeID)
	fmt.Printf("  data_dir  = %s\n", cfg.DataDir)
	fmt.Printf("  bind_addr = %s\n", addr)
	fmt.Printf("  compaction_threshold = %d\n", cfg.CompactionThreshold)
	fmt.Printf("  compaction_interval = %ds\n", cfg.CompactionIntervalSecs)
	fmt.Println()

	if err := volume.StartVolumeServer(
		addr,
		cfg.VolumeID,
		cfg.DataDir,
		cfg.CompactionThreshold,
		cfg.CompactionIntervalSecs,
	); err != nil {
		log.Fatalf("Server failed: %v\n", err)
		os.Exit(1)
	}
}
