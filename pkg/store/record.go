package store

import (
	"encoding/binary"
	"hash/crc32"
	"io"
)

// Record opcodes
const (
	OpSet    byte = 1
	OpDelete byte = 2
)

// Magic bytes for record framing
var Magic = [2]byte{0xF0, 0xF1}

// Record represents a single key-value operation
type Record struct {
	Op    byte
	Key   string
	Value []byte
}

// WriteRecord writes a record to a writer
func WriteRecord(w io.Writer, rec *Record) error {
	// Write magic
	if _, err := w.Write(Magic[:]); err != nil {
		return err
	}

	// Write opcode
	if _, err := w.Write([]byte{rec.Op}); err != nil {
		return err
	}

	// Write key length and key
	keyBytes := []byte(rec.Key)
	keyLen := uint32(len(keyBytes))
	if err := binary.Write(w, binary.LittleEndian, keyLen); err != nil {
		return err
	}

	// Write value length
	valLen := uint32(len(rec.Value))
	if err := binary.Write(w, binary.LittleEndian, valLen); err != nil {
		return err
	}

	// Write key
	if _, err := w.Write(keyBytes); err != nil {
		return err
	}

	// Write value (only for SET)
	if rec.Op == OpSet && len(rec.Value) > 0 {
		if _, err := w.Write(rec.Value); err != nil {
			return err
		}
	}

	// Compute and write checksum
	checksum := computeChecksum(rec)
	if err := binary.Write(w, binary.LittleEndian, checksum); err != nil {
		return err
	}

	return nil
}

// ReadRecord reads a record from a reader
func ReadRecord(r io.Reader) (*Record, error) {
	// Read magic
	var magic [2]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	if magic != Magic {
		return nil, ErrInvalidMagic
	}

	// Read opcode
	var op [1]byte
	if _, err := io.ReadFull(r, op[:]); err != nil {
		return nil, err
	}

	// Read key length
	var keyLen uint32
	if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
		return nil, err
	}

	// Read value length
	var valLen uint32
	if err := binary.Read(r, binary.LittleEndian, &valLen); err != nil {
		return nil, err
	}

	// Read key
	keyBytes := make([]byte, keyLen)
	if _, err := io.ReadFull(r, keyBytes); err != nil {
		return nil, err
	}
	key := string(keyBytes)

	// Read value (only for SET)
	var value []byte
	if op[0] == OpSet && valLen > 0 {
		value = make([]byte, valLen)
		if _, err := io.ReadFull(r, value); err != nil {
			return nil, err
		}
	}

	// Read and verify checksum
	var checksumStored uint32
	if err := binary.Read(r, binary.LittleEndian, &checksumStored); err != nil {
		return nil, err
	}

	rec := &Record{
		Op:    op[0],
		Key:   key,
		Value: value,
	}

	checksumCalc := computeChecksum(rec)
	if checksumCalc != checksumStored {
		return nil, ErrChecksumMismatch
	}

	return rec, nil
}

// computeChecksum calculates CRC32 for a record
func computeChecksum(rec *Record) uint32 {
	h := crc32.NewIEEE()

	_, _ = h.Write([]byte{rec.Op}) // Ignore error for hash.Write

	keyBytes := []byte(rec.Key)
	keyLen := uint32(len(keyBytes))
	_ = binary.Write(h, binary.LittleEndian, keyLen)

	valLen := uint32(len(rec.Value))
	_ = binary.Write(h, binary.LittleEndian, valLen)

	_, _ = h.Write(keyBytes)

	if rec.Op == OpSet && len(rec.Value) > 0 {
		_, _ = h.Write(rec.Value)
	}

	return h.Sum32()
}
