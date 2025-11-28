package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Port                    int
	VolumeID                string
	DataDir                 string
	CompactionThreshold     int
	CompactionIntervalSecs  int
	MaxRequestSizeMB        int
}

// FromEnv creates config from environment variables
func FromEnv() *Config {
	return &Config{
		Port:                    getEnvInt("PORT", 9002),
		VolumeID:                getEnvString("VOLUME_ID", "vol-1"),
		DataDir:                 getEnvString("DATA_DIR", "data"),
		CompactionThreshold:     getEnvInt("COMPACTION_THRESHOLD", 5),
		CompactionIntervalSecs:  getEnvInt("COMPACTION_INTERVAL_SECS", 60),
		MaxRequestSizeMB:        getEnvInt("MAX_REQUEST_SIZE_MB", 100),
	}
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		Port:                    9002,
		VolumeID:                "vol-1",
		DataDir:                 "data",
		CompactionThreshold:     5,
		CompactionIntervalSecs:  60,
		MaxRequestSizeMB:        100,
	}
}

func getEnvString(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}
