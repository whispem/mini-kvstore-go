package store

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	segmentPrefix = "segment-"
	segmentSuffix = ".dat"
	snapshotFile  = "index.snapshot"
)

// KVStore is the main storage engine
type KVStore struct {
	mu sync.RWMutex

	baseDir          string
	values           map[string][]byte
	index            *Index
	bloom            *BloomIndex
	activeSegmentID  uint64
	activeWriter     *bufio.Writer
	activeFile       *os.File
	maxSegmentSize   uint64
}

// Open opens or creates a KVStore at the given directory
func Open(dir string) (*KVStore, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	store := &KVStore{
		baseDir:        dir,
		values:         make(map[string][]byte),
		index:          NewIndex(),
		bloom:          NewBloomIndex(50000),
		maxSegmentSize: 16 * 1024 * 1024, // 16 MB
	}

	snapshotPath := filepath.Join(dir, snapshotFile)
	if _, err := os.Stat(snapshotPath); err == nil {
		if idx, err := LoadSnapshot(snapshotPath); err == nil {
			store.index = idx
			fmt.Printf("✓ Loaded index from snapshot (%d keys)\n", idx.Len())
		} else {
			fmt.Printf("⚠ Failed to load snapshot: %v, rebuilding from segments\n", err)
		}
	}

	segments, err := findSegments(dir)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	for _, segID := range segments {
		path := segmentPath(dir, segID)
		if err := store.replaySegment(path, segID); err != nil {
			return nil, fmt.Errorf("replay segment %d: %w", segID, err)
		}
	}

	if len(segments) > 0 && store.index.IsEmpty() {
		fmt.Printf("✓ Rebuilt index from segments in %.2fs\n", time.Since(start).Seconds())
	}

	lastID := uint64(0)
	if len(segments) > 0 {
		lastID = segments[len(segments)-1]
	}

	if err := store.resetActiveSegment(lastID + 1); err != nil {
		return nil, err
	}

	return store, nil
}

// Set stores or updates a key-value pair
func (s *KVStore) Set(key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec := &Record{Op: OpSet, Key: key, Value: value}

	if err := WriteRecord(s.activeWriter, rec); err != nil {
		return err
	}
	if err := s.activeWriter.Flush(); err != nil {
		return err
	}
	if err := s.activeFile.Sync(); err != nil {
		return err
	}

	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	s.values[key] = valueCopy
	s.index.Insert(key, s.activeSegmentID, 0)
	s.bloom.Insert(key)

	info, err := s.activeFile.Stat()
	if err != nil {
		return err
	}
	if uint64(info.Size()) >= s.maxSegmentSize {
		return s.rotateSegment()
	}

	return nil
}

// Get retrieves a value by key
func (s *KVStore) Get(key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, ok := s.values[key]; ok {
		result := make([]byte, len(val))
		copy(result, val)
		return result, nil
	}

	if !s.bloom.MightContain(key) {
		return nil, ErrNotFound
	}

	return nil, ErrNotFound
}

// Delete removes a key
func (s *KVStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec := &Record{Op: OpDelete, Key: key}

	if err := WriteRecord(s.activeWriter, rec); err != nil {
		return err
	}
	if err := s.activeWriter.Flush(); err != nil {
		return err
	}
	if err := s.activeFile.Sync(); err != nil {
		return err
	}

	delete(s.values, key)
	s.index.Remove(key)

	return nil
}

// ListKeys returns all keys
func (s *KVStore) ListKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.values))
	for k := range s.values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Stats returns storage statistics
func (s *KVStore) Stats() StoreStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	segments, _ := findSegments(s.baseDir)
	totalBytes := uint64(0)
	for _, val := range s.values {
		totalBytes += uint64(len(val))
	}

	oldestID := 0
	if len(segments) > 0 {
		oldestID = int(segments[0])
	}

	return StoreStats{
		NumKeys:         len(s.values),
		NumSegments:     len(segments),
		TotalBytes:      totalBytes,
		ActiveSegmentID: int(s.activeSegmentID),
		OldestSegmentID: oldestID,
	}
}

// SaveSnapshot saves the index to disk
func (s *KVStore) SaveSnapshot() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.baseDir, snapshotFile)
	return SaveSnapshot(s.index, path)
}

// Close closes the store
func (s *KVStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeWriter != nil {
		if err := s.activeWriter.Flush(); err != nil {
			return err
		}
	}
	if s.activeFile != nil {
		if err := s.activeFile.Close(); err != nil {
			return err
		}
	}
	return nil
}

// replaySegment replays all records in a segment
func (s *KVStore) replaySegment(path string, segID uint64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		rec, err := ReadRecord(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch rec.Op {
		case OpSet:
			s.values[rec.Key] = rec.Value
			s.index.Insert(rec.Key, segID, 0)
			s.bloom.Insert(rec.Key)
		case OpDelete:
			delete(s.values, rec.Key)
			s.index.Remove(rec.Key)
		default:
			return fmt.Errorf("unknown opcode: %d", rec.Op)
		}
	}

	return nil
}

// resetActiveSegment creates a new active segment
func (s *KVStore) resetActiveSegment(newID uint64) error {
	if s.activeWriter != nil {
		if err := s.activeWriter.Flush(); err != nil {
			return err
		}
	}
	if s.activeFile != nil {
		if err := s.activeFile.Close(); err != nil {
			return err
		}
	}

	path := segmentPath(s.baseDir, newID)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	s.activeSegmentID = newID
	s.activeFile = file
	s.activeWriter = bufio.NewWriter(file)

	return nil
}

// rotateSegment creates a new active segment
func (s *KVStore) rotateSegment() error {
	return s.resetActiveSegment(s.activeSegmentID + 1)
}

// Helper functions

func segmentPath(dir string, id uint64) string {
	return filepath.Join(dir, fmt.Sprintf("%s%d%s", segmentPrefix, id, segmentSuffix))
}

func findSegments(dir string) ([]uint64, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var segments []uint64
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, segmentPrefix) && strings.HasSuffix(name, segmentSuffix) {
			idStr := name[len(segmentPrefix) : len(name)-len(segmentSuffix)]
			if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
				segments = append(segments, id)
			}
		}
	}

	sort.Slice(segments, func(i, j int) bool {
		return segments[i] < segments[j]
	})

	return segments, nil
}
