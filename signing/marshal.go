package signing

import (
	"bytes"
	"fmt"

	"golang.org/x/mod/sumdb/note"
	"golang.org/x/mod/sumdb/tlog"
)

type Result []byte

// marshalSpicySig marshals a spicysig together from components into a byte slice.
//
// The checkpointNote parameter is assumed to contain a correctly formatted checkpoint as a body.
// However, we currently have no type that clarifies that,
// so until that's addressed, this function is unexported, as that gap leaves it too easy to misuse.
func marshalSpicySig(
	index uint64,
	mip tlog.RecordProof,
	checkpointNote *note.Note,
	contextHint string,
) Result {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("c2sp.org/tlog-proof@v1\n")
	fmt.Fprintf(buf, "index %d\n", index)
	for _, h := range mip {
		buf.WriteString(h.String())
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')

	checkpointRaw, err := note.Sign(checkpointNote)
	if err != nil {
		panic(err)
	}
	buf.Write(checkpointRaw)

	if contextHint != "" {
		buf.WriteString("\ncontexthint\n")
		buf.Write([]byte(contextHint))
		// REVIEW: exact desired semantics for trailing linebreaks.
	}
	return buf.Bytes()
}
