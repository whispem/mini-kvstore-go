package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// SetupTestDir creates a clean test directory
func SetupTestDir(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join("testdata", name)
	CleanupTestDir(t, dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	return dir
}

// CleanupTestDir removes a test directory
func CleanupTestDir(t *testing.T, dir string) {
	t.Helper()
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		t.Logf("warning: failed to cleanup test dir: %v", err)
	}
}
