package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestStore(t *testing.T) (*KVStore, string) {
	t.Helper()
	dir := filepath.Join("testdata", t.Name())
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove test dir: %v", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	store, err := Open(dir)
	require.NoError(t, err)

	return store, dir
}

func cleanupTestStore(t *testing.T, store *KVStore, dir string) {
	t.Helper()
	if err := store.Close(); err != nil {
		t.Logf("warning: failed to close store: %v", err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Logf("warning: failed to remove test dir: %v", err)
	}
}

func TestSetAndGet(t *testing.T) {
	store, dir := setupTestStore(t)
	defer cleanupTestStore(t, store, dir)

	// Set a key
	err := store.Set("key1", []byte("value1"))
	require.NoError(t, err)

	// Get the key
	val, err := store.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, []byte("value1"), val)

	// Get non-existent key
	_, err = store.Get("nonexistent")
	assert.Equal(t, ErrNotFound, err)
}

func TestDelete(t *testing.T) {
	store, dir := setupTestStore(t)
	defer cleanupTestStore(t, store, dir)

	// Set and verify
	err := store.Set("key1", []byte("value1"))
	require.NoError(t, err)

	val, err := store.Get("key1")
	require.NoError(t, err)
	assert.Equal(t, []byte("value1"), val)

	// Delete
	err = store.Delete("key1")
	require.NoError(t, err)

	// Verify deleted
	_, err = store.Get("key1")
	assert.Equal(t, ErrNotFound, err)
}

func TestListKeys(t *testing.T) {
	store, dir := setupTestStore(t)
	defer cleanupTestStore(t, store, dir)

	// Set multiple keys
	require.NoError(t, store.Set("key1", []byte("value1")))
	require.NoError(t, store.Set("key2", []byte("value2")))
	require.NoError(t, store.Set("key3", []byte("value3")))

	keys := store.ListKeys()
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "key3")
}

func TestPersistence(t *testing.T) {
	dir := filepath.Join("testdata", t.Name())
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove test dir: %v", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("warning: failed to remove test dir: %v", err)
		}
	}()

	// First session: write data
	{
		store, err := Open(dir)
		require.NoError(t, err)

		err = store.Set("key1", []byte("value1"))
		require.NoError(t, err)

		err = store.Set("key2", []byte("value2"))
		require.NoError(t, err)

		require.NoError(t, store.Close())
	}

	// Second session: read data
	{
		store, err := Open(dir)
		require.NoError(t, err)

		val1, err := store.Get("key1")
		require.NoError(t, err)
		assert.Equal(t, []byte("value1"), val1)

		val2, err := store.Get("key2")
		require.NoError(t, err)
		assert.Equal(t, []byte("value2"), val2)

		require.NoError(t, store.Close())
	}
}

func TestCompaction(t *testing.T) {
	store, dir := setupTestStore(t)
	defer cleanupTestStore(t, store, dir)

	// Write many versions of same keys
	for round := 0; round < 5; round++ {
		for i := 0; i < 100; i++ {
			key := "key"
			val := []byte("value")
			require.NoError(t, store.Set(key, val))
		}
	}

	// Compact
	err := store.Compact()
	require.NoError(t, err)

	// Verify data still exists
	val, err := store.Get("key")
	require.NoError(t, err)
	assert.Equal(t, []byte("value"), val)

	stats := store.Stats()
	assert.Equal(t, 1, stats.NumKeys)
}

func TestStats(t *testing.T) {
	store, dir := setupTestStore(t)
	defer cleanupTestStore(t, store, dir)

	require.NoError(t, store.Set("key1", []byte("value1")))
	require.NoError(t, store.Set("key2", []byte("value2")))

	stats := store.Stats()
	assert.Equal(t, 2, stats.NumKeys)
	assert.Greater(t, stats.TotalBytes, uint64(0))
	assert.Greater(t, stats.NumSegments, 0)
}

func TestSnapshot(t *testing.T) {
	dir := filepath.Join("testdata", t.Name())
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove test dir: %v", err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("warning: failed to remove test dir: %v", err)
		}
	}()

	// Create store and write data
	{
		store, err := Open(dir)
		require.NoError(t, err)

		for i := 0; i < 100; i++ {
			key := "key"
			val := []byte("value")
			require.NoError(t, store.Set(key, val))
		}

		// Save snapshot
		err = store.SaveSnapshot()
		require.NoError(t, err)

		require.NoError(t, store.Close())
	}

	// Reopen - should load from snapshot
	{
		store, err := Open(dir)
		require.NoError(t, err)

		val, err := store.Get("key")
		require.NoError(t, err)
		assert.Equal(t, []byte("value"), val)

		require.NoError(t, store.Close())
	}
}
