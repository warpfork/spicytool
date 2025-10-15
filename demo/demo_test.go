package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"os"
	"testing"
	"time"

	"filippo.io/torchwood"
	"github.com/transparency-dev/tessera"
	"github.com/transparency-dev/tessera/storage/posix"
	"github.com/warpfork/spicytool/glue"
	"golang.org/x/mod/sumdb/note"
	"golang.org/x/mod/sumdb/tlog"
)

const (
	dummyPrivateKey = "PRIVATE+KEY+example.com/log/testdata+33d7b496+AeymY/SZAX0jZcJ8enZ5FY1Dz+wTML2yWSkK+9DSF3eg"
	dummyPublicKey  = "example.com/log/testdata+33d7b496+AeHTu4Q3hEIMHNqc6fASMsq3rKNx280NI+oO5xCFkkSx"
)

const (
	logPath            = "/tmp/tlog"
	checkpointInterval = 100 * time.Millisecond

	// Altering this should have near zero effect on outcome...
	// but does change where checkpoints occur, so,
	// while it doesn't change the eventual tree structure,
	// it *will* often change the MIPs produced, since those depend on which tree sizes have checkpoints.
	// Smaller numbers also noticeably increase the amount of disk usage you'll see from this demo,
	// since tessera creates but does not (afaik) garbage collect any of the partial tiles.
	batchSize = 100

	// entryCount         = 270 // Must be above 256 if you want to see at least one full tile!
	// targetIndex        = 260
	entryCount  = 100270 // I'm picking increasingly large numbers here to insure we see access of more than one tile.
	targetIndex = 100252
)

func Test_Hello(*testing.T) {
	ctx := context.Background()

	// Demo setup: Clean up any existing log.
	if err := os.RemoveAll(logPath); err != nil {
		slog.Warn("Failed to clean up existing log", "error", err)
	}

	// Create the tlog's signer.  This is effectively the log's identity.)
	signer, err := note.NewSigner(dummyPrivateKey)
	if err != nil {
		slog.Error("Failed to create signer", "error", err)
		return
	}

	// Initialize the storage driver.

	// Set up tessara to produce (and read back from) our tlog.
	driver, err := posix.New(ctx, posix.Config{
		Path: logPath,
	})
	if err != nil {
		slog.Error("Failed to setup storage driver", "error", err)
		return
	}
	tlogAppender, tlogShutdowner, tlogReader, err := tessera.NewAppender(ctx, driver,
		tessera.NewAppendOptions().
			WithCheckpointSigner(signer).
			WithCheckpointInterval(checkpointInterval).
			WithBatching(batchSize, time.Second))
	if err != nil {
		slog.Error("Failed to initialize tessera", "error", err)
		return
	}
	defer func() {
		if err := tlogShutdowner(ctx); err != nil {
			slog.Error("Failed to gracefully shutdown tessera", "error", err)
		}
	}()

	// Create publication tlogAwaiter
	tlogAwaiter := tessera.NewPublicationAwaiter(ctx, tlogReader.ReadCheckpoint, 100*time.Millisecond)

	// Create random entries and submit them to the log.
	// We'll keep them around, too, just for convenience in this test.
	entries := [][]byte{}
	r := rand.NewChaCha8([32]byte{})
	slog.Info("Adding dummy entries to the log...", "n", entryCount)
	var tlogIndexFutures []tessera.IndexFuture
	for i := range entryCount {
		entry := generateTestEntry(i, r)
		entries = append(entries, entry)
		indexFuture := tlogAppender.Add(ctx, tessera.NewEntry(entry))
		tlogIndexFutures = append(tlogIndexFutures, indexFuture)
	}

	// Wait for all entries to be integrated.
	slog.Info("Waiting for entries to be integrated...")
	for i, entryFuture := range tlogIndexFutures {
		_, _, err := tlogAwaiter.Await(ctx, entryFuture)
		if err != nil {
			slog.Error("Failed to integrate entry", "i", i, "error", err)
			return
		}
	}
	slog.Info("All entries integrated.")

	// Next: access information about the tree.
	// This is all a prerequisite to starting to build the MIPs.

	// Get the current tree checkpoint.
	// This is a fairly multi-step process.
	// Tessera exposes this as a byte slice, so we must parse it.  That will involve several steps.
	//
	// (Note that this is the current checkpoint -- over *all* entries -- which means that
	// it can describe a tree larger than the index of the entry we're about to build a MIP for.
	// In this demo, that will be the case because we submitted lots of entries before we switch our attention to constructing this proof,
	// but in the wild it is also quite likely to happen because of batching on the log's side.)
	checkpointRaw, err := tlogReader.ReadCheckpoint(ctx)
	if err != nil {
		slog.Error("Failed to read checkpoint", "error", err)
		return
	}

	// Create a verifier for the checkpoint.
	// This is a prerequisite for interacting with the checkpoint data,
	// because the library APIs for it are designed not to allow accessing unverified data.
	//
	// Here, we use only the log's own pubkey... but ideally we could also use a variety of witness's public keys.
	verifier, err := note.NewVerifier(dummyPublicKey)
	if err != nil {
		slog.Error("Failed to create verifier", "error", err)
		return
	}

	// First step of parsing the checkpoint: parse it as a note!
	// See c2sp.org/signed-note for more information about this format.
	n, err := note.Open(checkpointRaw, note.VerifierList(verifier))
	if err != nil {
		slog.Error("Failed to parse checkpoint note", "error", err)
		return
	}

	// At last, we can parse the checkpoint information from the body of the note.
	checkpoint, err := torchwood.ParseCheckpoint(n.Text)
	if err != nil {
		slog.Error("Failed to parse checkpoint", "error", err)
		return
	}
	slog.Info("Checkpoint info:",
		"tree name", checkpoint.Origin,
		"tree size", checkpoint.N,
		"tree root", checkpoint.Hash,
	)

	// Okay, now on to the interesting bits: making and verifying proofs of tree inclusion.

	// Compute the MIP!  The tlog library does the heavy lifting here very nicely,
	// but it needs some interesting glue code to get it to read over the tessera API.
	//
	// Note that the MIP itself does *not* include a couple of things that you might suspect it to
	// (since you'll certainly need them when verifying it):
	//   - the MIP does *not* include the tree root hash itself.
	//     (But you'll see it in the checkpoint part of a serialized spicy sig.)
	//   - the MIP does *not* include the entry hash itself.
	//     (You *won't* see this in a serialized spicy sig, either!
	//     You're expected to compute that from the thing you're verifying the sig's application to!)
	//   - the MIP does *not* include the tree height nor entry index ints.
	//     (You'll see both in a serialized spicy sig; the former in the checkpoint section, the latter at the top.)
	hashReader := torchwood.TileHashReaderWithContext(ctx, checkpoint.Tree, glue.NewTesseraTileReader(tlogReader))
	proof, err := tlog.ProveRecord(int64(checkpoint.N), int64(targetIndex), hashReader)
	if err != nil {
		slog.Error("Failed to generate proof", "error", err)
		return
	}
	slog.Info("Computed Merkle Inclusion Proof!",
		"proof length", len(proof),
		"proof", proof,
	)

}

// Produce a human-readable but "random" (deterministic, seeded) string to use as a test entry.
// The expected index number is included, for debugging ease, but the suffix is random
// and also random of length (so that any seeking code you may write against the entry files is well exercised).
//
// The ChaCha8 type is used specifically, rather than a more general interface, which may seem strange:
// the reason is that the primitive of reading bytes is only implemented on that type in the math/rand/v2 package,
// and the read bytes slice primitive in the crypto/rand package can not be used with a deterministic source or seed.
// The deprecation notes on the v1 https://pkg.go.dev/math/rand#Read method point this way as well.
func generateTestEntry(
	i int,
	r *rand.ChaCha8,
) []byte {
	n := 8 + r.Uint64()%8
	b := make([]byte, n)
	r.Read(b)
	return fmt.Appendf(nil, "entry-%03d-data-%x", i, b)
}
