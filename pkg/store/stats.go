package store

import "fmt"

// StoreStats contains statistics about the store
type StoreStats struct {
	NumKeys          int
	NumSegments      int
	TotalBytes       uint64
	ActiveSegmentID  int
	OldestSegmentID  int
}

// TotalMB returns total size in megabytes
func (s StoreStats) TotalMB() float64 {
	return float64(s.TotalBytes) / (1024.0 * 1024.0)
}

// TotalKB returns total size in kilobytes
func (s StoreStats) TotalKB() float64 {
	return float64(s.TotalBytes) / 1024.0
}

// String returns a formatted string representation
func (s StoreStats) String() string {
	return fmt.Sprintf(
		"Store Statistics:\n"+
			"  Keys: %d\n"+
			"  Segments: %d\n"+
			"  Total size: %.2f MB\n"+
			"  Active segment: %d\n"+
			"  Oldest segment: %d",
		s.NumKeys,
		s.NumSegments,
		s.TotalMB(),
		s.ActiveSegmentID,
		s.OldestSegmentID,
	)
}
