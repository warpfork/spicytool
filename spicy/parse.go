package spicy

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"filippo.io/torchwood"
	"golang.org/x/mod/sumdb/note"
	"golang.org/x/mod/sumdb/tlog"
)

const (
	preambleV1 = "c2sp.org/spicy-signature@v1\n"
	sectionSep = "\n\n"
)

type SpicySigV1 struct {
	entryIndex int64
	mip        tlog.RecordProof

	checkpointNote note.Note // here you can see the sigs used; body is parsed in checkpint field.
	checkpoint     torchwood.Checkpoint

	contextHint string
}

func (s *SpicySigV1) Format() []byte {
	// ... it has just come to my attention that `note.Note` also cannot be created except by emitting one and parsing it again,
	// so, that's neat.
	// What, exactly, is supposed to be a sane type to use to compose in our SpicySig struct here?
	panic("nyi")
}

func ParseSpicySig(raw []byte, verifiers note.Verifiers) (*SpicySigV1, error) {
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

	// Check familiar prefix first; "this aint the right file" is the easiest thing to say.
	if !bytes.HasPrefix(raw, []byte(preambleV1)) {
		return nil, errors.New("not a spicysig -- preamble does not match")
	}

	// Split whole document into sections.
	// This is implemented by gathering indexes rather than doing a split, because
	// parts 2 and 3 will be handled together, and it makes little sense to split and then re-join them.
	part2edge := bytes.Index(raw, []byte(sectionSep))
	if part2edge < 0 {
		return nil, errors.New("malformed spicysig -- not enough sections")
	}
	part3edge := bytes.Index(raw[part2edge+2:], []byte(sectionSep))
	if part3edge < 0 {
		return nil, errors.New("malformed spicysig -- not enough sections")
	}
	part3edge += part2edge + 2
	part4edge := bytes.Index(raw[part3edge+2:], []byte(sectionSep))
	if part4edge > 0 {
		part4edge += part3edge + 2
	}

	// fmt.Printf("part 1 >>>\n%q\n<<<\n", raw[0:part2edge])
	// fmt.Printf("part 2 >>>\n%q\n<<<\n", raw[part2edge+2:part3edge])
	// fmt.Printf("part 3+ >>>\n%q\n<<<\n", raw[part3edge+2:])

	// Part 1: header and index and MIP.
	// (We've already checked the header line, so can skip over that.)
	lines := bytes.Split(raw[0:part2edge], []byte{'\n'})
	if len(lines) < 3 {
		return nil, errors.New("malformed spicysig -- too short")
	}
	rest, matches := munchPrefix(lines[1], []byte("index "))
	if !matches {
		return nil, errors.New("malformed spicysig -- expected index")
	}
	entryIndex, err := strconv.ParseInt(string(rest), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("malformed spicysig -- index must parse as int: %w", err)
	}
	mip := tlog.RecordProof{}
	for i := 2; i < len(lines); i++ {
		h := [tlog.HashSize]byte{}
		n, err := base64.StdEncoding.Decode(h[:], lines[i])
		if err != nil || n != tlog.HashSize {
			return nil, errors.New("malformed spicysig -- MIP entries must be b64")
		}
		mip = append(mip, tlog.Hash(h))
	}

	// Part 2 + 3:
	part23 := raw[part2edge+2:]
	if part4edge > 0 {
		// An interesting +1 appears here because note.Open demands a trailing linebreak.
		part23 = raw[part2edge+2 : part4edge+1]
	}
	// fmt.Printf("part 23 >>>\n%q\n<<<\n", part23)
	checkpointNote, err := note.Open(part23, verifiers)
	if err != nil {
		return nil, err
	}
	checkpoint, err := torchwood.ParseCheckpoint(checkpointNote.Text)
	if err != nil {
		return nil, err
	}

	// Part 4:
	// A contexthint is optional.  But is the only remaining possible section we accept.
	contextHint := ""
	if part4edge > 0 {
		rest, matches = munchPrefix(raw[part4edge+2:], []byte("contexthint\n"))
		if !matches {
			return nil, errors.New("malformed spicysig -- expected contexthint as last section")
		}
		contextHint = string(rest)
	}

	return &SpicySigV1{
		entryIndex:     entryIndex,
		mip:            mip,
		checkpointNote: *checkpointNote,
		checkpoint:     checkpoint,
		contextHint:    contextHint,
	}, nil
}

func munchPrefix(input []byte, expect []byte) (rest []byte, matches bool) {
	if len(input) < len(expect) {
		return input, false
	}
	if !bytes.Equal(input[0:len(expect)], expect) {
		return input, false
	}
	return input[len(expect):], true
}
