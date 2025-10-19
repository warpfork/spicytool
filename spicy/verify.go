package spicy

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"golang.org/x/mod/sumdb/tlog"
)

func (sig *SpicySigV1) Verify(body io.Reader, contextHint string) error {
	if contextHint != "" && sig.contextHint != contextHint {
		return errors.New("contexthint expectation does not match attached contexthint")
	}
	record, err := RecordForBody(body, sig.contextHint)
	if err != nil {
		return err
	}
	recordHash := tlog.RecordHash(record)
	return tlog.CheckRecord(
		sig.mip,
		int64(sig.checkpoint.N),
		sig.checkpoint.Hash,
		int64(sig.entryIndex),
		recordHash,
	)
}

// RecordForBody calculates the log entry record that
// describes a given (possibly large) body and context hint.
//
// The result must be reasonably small:
// while transparency logs are based on merkle trees,
// they do also often keep the bodies of the records submitted to them,
// and therefore there are generally size limits to those record bodies.
//
// In SpicyTool's usage: the body is generally a file's contents,
// and the contextHint is generally the file's name.
func RecordForBody(body io.Reader, contextHint string) ([]byte, error) {
	if contextHint == "" { // REVIEW: if this branch should exist.
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, err
		}
		return b, nil
	} else {
		// Produce `"b" || h(body) || "c" || len(contexthint) || contexthint`.
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
		binary.Write(&record, binary.LittleEndian, uint16(len(contextHint)))
		record.WriteString(contextHint)
		return record.Bytes(), nil
	}
}
