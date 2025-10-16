package spicy

import (
	"golang.org/x/mod/sumdb/note"
	"golang.org/x/mod/sumdb/tlog"
)

const preambleV1 = "c2sp.org/spicy-signature@v1\n"

type SpicySigV1 struct {
	index uint64
	mip   tlog.RecordProof

	checkpointNote note.Note

	contextHint string
}

func ParseSpicySig(raw []byte) (*SpicySigV1, error) {
	// A spicysig comes in roughly three or four sections:
	//   1. the header + index + mip.
	//   2. the tree checkpoint.
	//   3. the signatures over the tree checkpoint.
	//   4. (optionally) the contexthint string.
	//
	// Parts 2 and 3 are handled together as a "Signed Note", which has its own parser.
	//
	// Our parse of all this is necessarily a little gnarly, because
	// the parser for parts 2+3 just takes what it is given, and does not attempt to
	// determine where the end of the content it handles is, nor report that...
	// in fact, the parser for part 3 assumes that the slice it is given is completely its own content,
	// and bases some of its parse decisions by going backwards from the end of the range.
	// So, we must have some knowledge of that format here, purely so that we can scan over it,
	// decide where it ends, and hand that subset off to the relevant parser for that section.
	//
	// Fortunately, although we must "see through" several layers of spec here,
	// they have a simple enough delimiter:
	// The parse of each these sections is delimited by "two consecutive unix linebreaks".
	// (Including that one of those delimiters occurs between parts 2 and 3,
	// and that part must be handed to the parser that consumes parts 2 and 3 together.)
	// ('\r' is not tolerated by any of the other library code I've seen so far, and is thus not tolerated here either.)
	//
	// (Looking yet deeper, to verify that that delimiter is actually true and cannot be present in a valid body:
	// It's touchy.
	// Technically, part 2 is specified as allowing unknown trailing content,
	// and in fact some implementations *do* permit the double-linebreak value in that section!
	// (The `torchwood` module's `ParseCheckpoint` rejects it; but `sumdb/tlog`'s `ParseTree` allows it.)
	// However, the parser for part 3 -- `sumdb/note.Open` -- both skips over arbitrary body to find the last double-linebreak...
	// but then *also* scans backwards over the body to ensure it doesn't contain another instance of that.
	// Thus, the parse for part 3 does forbid a variable number of double linebreaks,
	// leaving the permissiveness of some part 2 implementations as irrelevant.
	// Got a headache yet?  I know I do.
	// Composing a series of formats that lack distinctive opening and closing delimiters gets confusing, mkay?)
	panic("nyi")
}
