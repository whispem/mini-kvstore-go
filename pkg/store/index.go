package store

import "sync"

// IndexEntry represents a location in a segment
type IndexEntry struct {
	SegmentID uint64
	Offset    uint64
}

// Index provides fast in-memory key lookups
type Index struct {
	mu   sync.RWMutex
	data map[string]*IndexEntry
}

// NewIndex creates a new empty index
func NewIndex() *Index {
	return &Index{
		data: make(map[string]*IndexEntry),
	}
}

// Insert adds or updates a key location
func (idx *Index) Insert(key string, segmentID, offset uint64) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.data[key] = &IndexEntry{
		SegmentID: segmentID,
		Offset:    offset,
	}
}

// Get retrieves the location for a key
func (idx *Index) Get(key string) (*IndexEntry, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entry, ok := idx.data[key]
	return entry, ok
}

// Remove deletes a key from the index
func (idx *Index) Remove(key string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	delete(idx.data, key)
}

// Contains checks if a key exists
func (idx *Index) Contains(key string) bool {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	_, ok := idx.data[key]
	return ok
}

// Len returns the number of keys
func (idx *Index) Len() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.data)
}

// Keys returns all keys (snapshot)
func (idx *Index) Keys() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	keys := make([]string, 0, len(idx.data))
	for k := range idx.data {
		keys = append(keys, k)
	}
	return keys
}

// IsEmpty returns true if index has no keys
func (idx *Index) IsEmpty() bool {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.data) == 0
}

// Clear removes all entries
func (idx *Index) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.data = make(map[string]*IndexEntry)
}
