package spicy

import (
	"io"

	"golang.org/x/mod/sumdb/tlog"
)

func (sig *SpicySigV1) Verify(body io.Reader, contextHint string) error {
	record, err := RecordForBody(body, contextHint)
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
		panic("nyi")
		// TODO: roughly `h("b" || len(body) || body || "c" || len(contexthint) || contexthint)`.
		// TODO: design how the lengths are encoded therein.
		// TODO: consider also doing an `h(body)` instead of using it directly... if only because `tlog.RecordHash` takes bytes and not a reader.
		// REVIEW: do we want a more humanely legible entry format than this?  One might argue that a purpose of a transparency log is to have entries that can be reasonably examined.
	}
}
