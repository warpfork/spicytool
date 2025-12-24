package contexts

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// RecordForBody calculates the log entry record format that spicytool uses
// to describe a given (possibly large) body and context hint.
//
// This value is:
// `"b" || h(body) || "c" || len(contexthint) || contexthint`,
// where `||` denotes concatenation, `h` is a sha256 hash,
// and length is encoded as a big endian uint16.
//
// This value is designed to be reasonably small,
// such that it is reasonable to be retained in a transparency log that keeps entry bodies.
//
// If the contextHint string is empty, the entire format is still used:
// the length of context hint will simply be zeros, and no further data follows.
//
// In SpicyTool's usage: the body is generally a file's contents,
// and the contextHint is generally the file's name.
func RecordForBody(body io.Reader, contextHint string) ([]byte, error) {
	if len(contextHint) > math.MaxUint16 {
		return nil, errors.New("contexthint too long")
	}
	h := sha256.New()
	if _, err := io.Copy(h, body); err != nil {
		return nil, err
	}
	var record bytes.Buffer
	record.WriteByte('b')
	record.Write(h.Sum(nil))
	record.WriteByte('c')
	binary.Write(&record, binary.BigEndian, uint16(len(contextHint)))
	record.WriteString(contextHint)
	return record.Bytes(), nil
}
