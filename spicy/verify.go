package spicy

import (
	"io"

	"golang.org/x/mod/sumdb/tlog"
)

func (sig *SpicySigV1) Verify(body io.Reader, contextHint string) error {
	recordHash, err := recordHash(body, contextHint)
	if err != nil {
		return err
	}
	return tlog.CheckRecord(
		sig.mip,
		int64(sig.checkpoint.N),
		sig.checkpoint.Hash,
		int64(sig.entryIndex),
		recordHash,
	)
}

func recordHash(body io.Reader, contextHint string) (tlog.Hash, error) {
	if contextHint == "" { // REVIEW: if this branch should exist.
		b, err := io.ReadAll(body)
		if err != nil {
			return tlog.Hash{}, err
		}
		return tlog.RecordHash(b), nil
	} else {
		panic("nyi")
		// TODO: roughly `h("b" || len(body) || body || "c" || len(contexthint) || contexthint)`.
		// TODO: design how the lengths are encoded therein.
		// TODO: consider also doing an `h(body)` instead of using it directly... if only because `tlog.RecordHash` takes bytes and not a reader.
	}
}
