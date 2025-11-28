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
	if _, err := w.Write(Magic[:]); err != nil {
		return err
	}
	if _, err := w.Write([]byte{rec.Op}); err != nil {
		return err
	}

	keyBytes := []byte(rec.Key)
	keyLen := uint32(len(keyBytes))
	if err := binary.Write(w, binary.LittleEndian, keyLen); err != nil {
		return err
	}

	valLen := uint32(len(rec.Value))
	if err := binary.Write(w, binary.LittleEndian, valLen); err != nil {
		return err
	}

	if _, err := w.Write(keyBytes); err != nil {
		return err
	}

	if rec.Op == OpSet && len(rec.Value) > 0 {
		if _, err := w.Write(rec.Value); err != nil {
			return err
		}
	}

	checksum, err := computeChecksum(rec)
	if err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, checksum); err != nil {
		return err
	}

	return nil
}

// ReadRecord reads a record from a reader
func ReadRecord(r io.Reader) (*Record, error) {
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

	var op [1]byte
	if _, err := io.ReadFull(r, op[:]); err != nil {
		return nil, err
	}

	var keyLen uint32
	if err := binary.Read(r, binary.LittleEndian, &keyLen); err != nil {
		return nil, err
	}

	var valLen uint32
	if err := binary.Read(r, binary.LittleEndian, &valLen); err != nil {
		return nil, err
	}

	keyBytes := make([]byte, keyLen)
	if _, err := io.ReadFull(r, keyBytes); err != nil {
		return nil, err
	}
	key := string(keyBytes)

	var value []byte
	if op[0] == OpSet && valLen > 0 {
		value = make([]byte, valLen)
		if _, err := io.ReadFull(r, value); err != nil {
			return nil, err
		}
	}

	var checksumStored uint32
	if err := binary.Read(r, binary.LittleEndian, &checksumStored); err != nil {
		return nil, err
	}

	rec := &Record{
		Op:    op[0],
		Key:   key,
		Value: value,
	}

	checksumCalc, err := computeChecksum(rec)
	if err != nil {
		return nil, err
	}

	if checksumCalc != checksumStored {
		return nil, ErrChecksumMismatch
	}

	return rec, nil
}

// computeChecksum calculates CRC32 for a record and returns error
func computeChecksum(rec *Record) (uint32, error) {
	h := crc32.NewIEEE()

	if _, err := h.Write([]byte{rec.Op}); err != nil {
		return 0, err
	}

	keyBytes := []byte(rec.Key)
	keyLen := uint32(len(keyBytes))
	if err := binary.Write(h, binary.LittleEndian, keyLen); err != nil {
		return 0, err
	}

	valLen := uint32(len(rec.Value))
	if err := binary.Write(h, binary.LittleEndian, valLen); err != nil {
		return 0, err
	}

	if _, err := h.Write(keyBytes); err != nil {
		return 0, err
	}

	if rec.Op == OpSet && len(rec.Value) > 0 {
		if _, err := h.Write(rec.Value); err != nil {
			return 0, err
		}
	}

	return h.Sum32(), nil
}
