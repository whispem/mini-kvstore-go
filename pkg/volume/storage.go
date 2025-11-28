package volume

import (
	"fmt"
	"hash/crc32"

	"github.com/whispem/mini-kvstore-go/pkg/store"
)

// BlobMeta contains metadata about a stored blob
type BlobMeta struct {
	Key      string `json:"key"`
	ETag     string `json:"etag"`
	Size     uint64 `json:"size"`
	VolumeID string `json:"volume_id"`
}

// BlobStorage provides high-level blob operations
type BlobStorage struct {
	store    *store.KVStore
	volumeID string
}

// NewBlobStorage creates a new blob storage instance
func NewBlobStorage(dataDir, volumeID string) (*BlobStorage, error) {
	kvstore, err := store.Open(dataDir)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	return &BlobStorage{
		store:    kvstore,
		volumeID: volumeID,
	}, nil
}

// Put stores a blob and returns metadata
func (b *BlobStorage) Put(key string, data []byte) (*BlobMeta, error) {
	etag := fmt.Sprintf("%08x", crc32.ChecksumIEEE(data))

	if err := b.store.Set(key, data); err != nil {
		return nil, err
	}

	return &BlobMeta{
		Key:      key,
		ETag:     etag,
		Size:     uint64(len(data)),
		VolumeID: b.volumeID,
	}, nil
}

// Get retrieves a blob by key
func (b *BlobStorage) Get(key string) ([]byte, error) {
	return b.store.Get(key)
}

// Delete removes a blob
func (b *BlobStorage) Delete(key string) error {
	return b.store.Delete(key)
}

// ListKeys returns all blob keys
func (b *BlobStorage) ListKeys() []string {
	return b.store.ListKeys()
}

// VolumeID returns the volume identifier
func (b *BlobStorage) VolumeID() string {
	return b.volumeID
}

// Stats returns storage statistics
func (b *BlobStorage) Stats() store.StoreStats {
	return b.store.Stats()
}

// Compact performs compaction
func (b *BlobStorage) Compact() error {
	return b.store.Compact()
}

// SaveSnapshot saves index snapshot
func (b *BlobStorage) SaveSnapshot() error {
	return b.store.SaveSnapshot()
}

// Close closes the storage
func (b *BlobStorage) Close() error {
	return b.store.Close()
}
