package signing

import (
	"bytes"
	"fmt"

	"golang.org/x/mod/sumdb/tlog"
)

type Result []byte

// marshalSpicySig marshals a spicysig together from compoments into a byte slice.
//
// This function signature is a little strange because it takes a checkpoint as already-serialized bytes,
// assumes that's been done correctly, and just bangs it in.
// I'd like this to be a little better-typed and a little harder to misuse,
// but our underlying libraries are strangely committed to only exposing
// either {raw serialized bytes} or {parsed and verified things that can't be reserialized},
// which... leaves us... well, here.
// I'm not making this function exported because it's far too easy to misuse.
func marshalSpicySig(
	index uint64,
	mip tlog.RecordProof,
	checkpointRaw []byte,
	contextHint string,
) Result {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("c2sp.org/spicy-signature@v1\n")
	fmt.Fprintf(buf, "index %d\n", index)
	for _, h := range mip {
		buf.WriteString(h.String())
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')

	buf.Write(checkpointRaw)

	if contextHint != "" {
		buf.WriteString("\ncontexthint\n")
		buf.Write([]byte(contextHint))
		// REVIEW: exact desired semantics for trailing linebreaks.
	}
	return buf.Bytes()
}
