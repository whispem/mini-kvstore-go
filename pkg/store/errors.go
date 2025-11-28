package store

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound indicates a key was not found
	ErrNotFound = errors.New("key not found")

	// ErrCorrupted indicates corrupted data
	ErrCorrupted = errors.New("corrupted data")

	// ErrInvalidMagic indicates invalid magic bytes in record
	ErrInvalidMagic = errors.New("invalid magic bytes")

	// ErrChecksumMismatch indicates checksum validation failed
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrInvalidOpcode indicates an unknown operation code
	ErrInvalidOpcode = errors.New("invalid operation code")

	// ErrNoActiveSegment indicates no active segment is available
	ErrNoActiveSegment = errors.New("no active segment")
)

// StoreError wraps errors with context
type StoreError struct {
	Op  string // Operation that failed
	Err error  // Original error
}

func (e *StoreError) Error() string {
	if e.Op == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *StoreError) Unwrap() error {
	return e.Err
}

// NewStoreError creates a new StoreError
func NewStoreError(op string, err error) *StoreError {
	return &StoreError{
		Op:  op,
		Err: err,
	}
}
