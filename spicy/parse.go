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
	// and so we must have some knowledge of that format here, purely so that we can scan over it,
	// decide where it ends, and hand that subset off to the relevant parser for that section.
	//
	// (Looking yet deeper: technically part 2 is specified as allowing unknown trailing content,
	// but part 3 explicitly forbids any trailers: the signatures must come last.
	// Since the spicysig-v1 format places the contexthint section *after* the checkpoint+signatures,
	// that means yes, we still have to figure out how to leap over both those elements.
	// Composing a series of formats that lack opening and closing delimiters gets confusing, mkay?)
	//
	// Fortunately, although we must "see through" several layers of spec here,
	// they have a simple enough delimiter:
	// The parse of each these sections is delimited by "two consecutive unix linebreaks".
	// ('\r' is not tolerated by any of the other library code I've seen so far, and is thus not tolerated here either.)
	//
	// (Can the unspecified forward-compatability trailer section of part 2 include double linebreaks,
	// thus making a fool of all of this?  *Yes*, although it depends on who you ask:
	//   - `sumdb/note.Open` defines the signature section as the *last* occurence of "\n\n",
	//     which means that part 3's parse does not forbid that sequence in the section occupied by part 2;
	//   - `sumdb/tlog.ParseTree` very explicitly ignores *all* trailing content, regardless of content,
	//     so by that interpretation then one can indeed still have more double linebreaks in part 2;
	//   - but `torchwood.ParseCheckpoint` *does* reject a checkpoint as malformed if it contains double linebreaks in the trailer.
	// We're going with the torchwood interpretation, here; the alternative is unworkable
	// if one wants to have a format that both keeps concatenating to the end and also refuses to use any other bounding
	// such as either opening and closing delimiters for sections or length prefixes.)
	panic("nyi")
}
