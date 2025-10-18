package signing

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"filippo.io/torchwood"
	"github.com/transparency-dev/tessera"
	"github.com/warpfork/spicytool/glue"
	"github.com/warpfork/spicytool/spicy"
	"golang.org/x/mod/sumdb/note"
	"golang.org/x/mod/sumdb/tlog"
)

type Result []byte

func (lop *LogOperator) Sign(ctx context.Context, body io.Reader, contextHint string) (Result, error) {
	// Phase 1: compute our entry, and append it to the log.
	// ========

	record, err := spicy.RecordForBody(body, contextHint)
	if err != nil {
		return nil, err
	}

	// `tessera.NewEntry` can be seen as analogous to `tlog.RecordHash`,
	// and should be called with the same values.
	// (Both ultimately compute `h(0 || param)`.)
	entry := tessera.NewEntry(record)
	entryFuture := lop.appender.Add(ctx, entry)

	// Phase 2: wait.  We need a checkpoint that covers our entry in order to continue.
	// ========

	index, checkpointRaw, err := lop.publicationAwaiter.Await(ctx, entryFuture)
	if err != nil {
		return nil, err
	}

	// Phase 3: compute our spicy sig using our results from the log!
	// ========

	// Parse the checkpoint note while not verifying it.
	// This just came from... ourselves... so doing asymmetric crypto on it seems rather silly.
	// (Arguably, this could also be implemented by searching for double-linebreaks in the body, but...
	// see comments in spicy/parse.go for how much a specification/implementation conflict-hole that is already; I'd rather not engage it further.)
	n, err := note.Open(checkpointRaw, nil)
	slog.Debug("Checkpoint raw:",
		"body", checkpointRaw,
		"laments", err,
	)
	uff := &note.UnverifiedNoteError{}
	if err != nil && !errors.As(err, &uff) {
		return nil, err
	}
	n = uff.Note

	// Now parse the actual checkpoint into usable data.
	checkpoint, err := torchwood.ParseCheckpoint(n.Text)
	if err != nil {
		return nil, err
	}
	slog.Debug("Checkpoint info used for signing:",
		"tree name", checkpoint.Origin,
		"tree size", checkpoint.N,
		"tree root", checkpoint.Hash,
	)

	// Compute the MIP!  The tlog library does the heavy lifting here very nicely,
	// with just a bit of glue code to show it how to read tessera's tiles.
	mip, err := tlog.ProveRecord(
		int64(checkpoint.N), int64(index.Index),
		torchwood.TileHashReaderWithContext(ctx, checkpoint.Tree, glue.NewTesseraTileReader(lop.reader)),
	)

	// And emit.
	// TODO: extract to function.
	buf := bytes.NewBuffer(nil)
	buf.WriteString("c2sp.org/spicy-signature@v1\n")
	fmt.Fprintf(buf, "index %d\n", index.Index)
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
	return buf.Bytes(), nil
}
