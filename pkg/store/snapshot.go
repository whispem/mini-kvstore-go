package store

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

var snapshotMagic = [8]byte{'K', 'V', 'I', 'N', 'D', 'E', 'X', '1'}

// SaveSnapshot writes the index to disk for fast restarts
func SaveSnapshot(idx *Index, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	defer w.Flush()

	// Write magic
	if _, err := w.Write(snapshotMagic[:]); err != nil {
		return err
	}

	// Get all keys
	keys := idx.Keys()
	numEntries := uint64(len(keys))

	// Write number of entries
	if err := binary.Write(w, binary.LittleEndian, numEntries); err != nil {
		return err
	}

	// Write each entry
	for _, key := range keys {
		entry, ok := idx.Get(key)
		if !ok {
			continue
		}

		// Write key length and key
		keyBytes := []byte(key)
		keyLen := uint32(len(keyBytes))
		if err := binary.Write(w, binary.LittleEndian, keyLen); err != nil {
			return err
		}
		if _, err := w.Write(keyBytes); err != nil {
			return err
		}

		// Write segment ID and offset
		if err := binary.Write(w, binary.LittleEndian, entry.SegmentID); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, entry.Offset); err != nil {
			return err
		}
	}

	return nil
}

// LoadSnapshot reads the index from disk
func LoadSnapshot(path string) (*Index, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)

	// Read and verify magic
	var magic [8]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}
	if magic != snapshotMagic {
		return nil, ErrInvalidMagic
	}

	// Read number of entries
	var numEntries uint64
	if err := binary.Read(r, binary.LittleEndian, &numEntries); err != nil {
		return nil, err
	}

	idx := NewIndex()

	// Read each entry
	for i := uint64(0); i < numEntries; i++ {
		// Read key length and key
		var keyLen uint32
		if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
			return nil, err
		}

		keyBytes := make([]byte, keyLen)
		if _, err := io.ReadFull(r, keyBytes); err != nil {
			return nil, err
		}
		key := string(keyBytes)

		// Read segment ID and offset
		var segmentID, offset uint64
		if err := binary.Read(r, binary.LittleEndian, &segmentID); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.LittleEndian, &offset); err != nil {
			return nil, err
		}

		idx.Insert(key, segmentID, offset)
	}

	return idx, nil
}
