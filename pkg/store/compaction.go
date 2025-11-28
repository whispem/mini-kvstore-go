package store

import (
	"fmt"
	"os"
	"path/filepath"
)

// Compact performs manual compaction
func (s *KVStore) Compact() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find all segments
	segments, err := findSegments(s.baseDir)
	if err != nil {
		return fmt.Errorf("find segments: %w", err)
	}

	// Remove all old segment files
	for _, segID := range segments {
		path := segmentPath(s.baseDir, segID)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove segment %d: %w", segID, err)
		}
	}

	// Create new active segment starting from 0
	if err := s.resetActiveSegment(0); err != nil {
		return fmt.Errorf("reset active segment: %w", err)
	}

	// Rewrite all live keys to the new segment
	for key, value := range s.values {
		rec := &Record{
			Op:    OpSet,
			Key:   key,
			Value: value,
		}

		if err := WriteRecord(s.activeWriter, rec); err != nil {
			return fmt.Errorf("write record: %w", err)
		}
	}

	if err := s.activeWriter.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	if err := s.activeFile.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	// Save snapshot after compaction
	snapshotPath := filepath.Join(s.baseDir, snapshotFile)
	if err := SaveSnapshot(s.index, snapshotPath); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	return nil
}
