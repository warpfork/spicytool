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
	// ('\r' is not tolerated by any of the other library code I've seen so far,
	// and the spec is also explicit about using exactly U+000A,
	// thus we are similarly strict here.)
	//
	// ... Sort of.
	// In fact, whether this is unambiguously parsable at all somewhat depends on who you ask.
	// There is variation amongst implementations of parsers for part 2.
	//   - `sumdb/note.Open` defines the signature section as the *last* occurence of "\n\n",
	//     which means that part 3's parse does not forbid that sequence in the section occupied by part 2;
	//   - `sumdb/tlog.ParseTree` very explicitly ignores *all* trailing content, regardless of content,
	//     so by that interpretation then one can indeed still have more double linebreaks in part 2;
	//   - but `torchwood.ParseCheckpoint` *does* reject a checkpoint as malformed if it contains double linebreaks in the trailer.
	// In specs?
	//   - https://c2sp.org/tlog-checkpoint declares that all lines in the extension section must be non-empty.
	//     So, that's on the same page as torchwood's implementation.
	//   - https://c2sp.org/signed-note does explicitly discuss double-linebreak (aka empty lines) as permissable in the note body.
	//     So, everyone is in concordance with what `sumdb/note.Open` does (which is a relief).
	//
	// We're going with the torchwood interpretation and the spec, here, for the checkpoint body;
	// if we tolerated a less strict interpretation, the result is simply unworkable
	// unless we also change the spicysig format to use one of: opening and closing delimiters for sections, or length prefixes, or some other more serious grammar.
	// But, it does mean we're relying on the constraints on part 2 to be able to
	// select the byte range that has to be handed to the parser for part 3, which is a bit exotic.
	// Got a headache yet?  I know I do.
	// Composing a series of formats that lack distinctive opening and closing delimiters gets confusing, mkay?
	//
	// But here we are.  Allons-y!
	panic("nyi")
}
