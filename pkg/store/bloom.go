package store

import (
	"crypto/sha256"

	"github.com/bits-and-blooms/bloom/v3"
)

// BloomIndex provides fast negative lookups
type BloomIndex struct {
	filter *bloom.BloomFilter
}

// NewBloomIndex creates a new bloom filter
func NewBloomIndex(expectedItems uint) *BloomIndex {
	// False positive rate of 1%
	filter := bloom.NewWithEstimates(expectedItems, 0.01)

	return &BloomIndex{
		filter: filter,
	}
}

// Insert adds a key to the bloom filter
func (b *BloomIndex) Insert(key string) {
	hash := hashKey(key)
	b.filter.Add(hash[:])
}

// MightContain checks if a key might exist
// Returns false if definitely doesn't exist
// Returns true if might exist (could be false positive)
func (b *BloomIndex) MightContain(key string) bool {
	hash := hashKey(key)
	return b.filter.Test(hash[:])
}

// hashKey computes SHA256 hash of a key
func hashKey(key string) [32]byte {
	return sha256.Sum256([]byte(key))
}
