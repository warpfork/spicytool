package spicytool

import (
	"context"
	"fmt"
	"testing"
	"time"

	"filippo.io/torchwood"
	"filippo.io/torchwood/tesserax"
	"github.com/transparency-dev/tessera"
	"github.com/transparency-dev/tessera/storage/posix"
	"golang.org/x/mod/sumdb/note"
	"golang.org/x/mod/sumdb/tlog"
)

type LogHandle struct {
	Reader   tessera.LogReader
	Appender *tessera.Appender
	Awaiter  *tessera.PublicationAwaiter
	Shutdown func(context.Context) error
	Policy   torchwood.Policy
}

// OperateLog opens or creates an appendable Tessera log using the local filesystem via the posix driver.
// It is a bundle of defaults over `tessera.NewAppender`; you could also use that directly.
// The configuration is tuned with a log of low append frequency in mind.
//
// The log origin name is extracted from the checkpointSigner value.
// The checkpointVerifier value must be the public key corresponding to the checkpointSigner value;
// it is not directly used during most log operations, but is needed by some operations
// like GenerateProof (which end up reading back checkpoints as part of their work).
//
// Witnesses are a paramter to this function because log operation involves
// proactively reaching out to witnesses during appends.
// If you really don't want these, you can supply `tessera.NewWitnessGroup(0)`.
// Otherwise, you probably want to use `tessera.NewWitnessGroupFromPolicy`.
//
// (This function is _very_ similar to things occuring in age-keyserver;
// it could possibly be factored out to reuse.)
func OperateLog(ctx context.Context, logPath string,
	signerKeyRaw string, // Would be `note.Signer`, except we also need a verifier, and `torchwood.NewVerifierFromSigner` does that but only from strings.
	//checkpointVerifier note.Verifier, // nope!  torchwood.NewVerifierFromSigner saves us from pushing this complexity up.
	witnesses tessera.WitnessGroup,
) (*LogHandle, error) {
	// Process keys.
	// This has the least side-effects, so do it first in case of any errors.
	checkpointSigner, err := note.NewSigner(signerKeyRaw)
	if err != nil {
		return nil, err
	}
	checkpointVerifier, err := torchwood.NewVerifierFromSigner(signerKeyRaw)
	if err != nil {
		return nil, err
	}

	// Tlog setup.
	// This will begin writing to the filesystem.
	driver, err := posix.New(ctx, posix.Config{
		Path: logPath,
	})
	if err != nil {
		return nil, err
	}
	checkpointInterval := 1 * time.Second
	if testing.Testing() {
		checkpointInterval = 100 * time.Millisecond
	}
	appender, shutdown, reader, err := tessera.NewAppender(ctx, driver, tessera.NewAppendOptions().
		WithCheckpointSigner(checkpointSigner).
		WithBatching(1, tessera.DefaultBatchMaxAge).
		WithCheckpointInterval(checkpointInterval).
		//WithCheckpointRepublishInterval(1*time.Hour). // Seems unnecessary for the intended use cadence of spicytool.
		WithWitnesses(witnesses, nil))
	if err != nil {
		return nil, err
	}
	awaiter := tessera.NewPublicationAwaiter(ctx, reader.ReadCheckpoint, 25*time.Millisecond)

	// Policy setup.
	// Purely for use of reading back checkpoints in GenerateProof; otherwise would not be needed.
	policy := torchwood.ThresholdPolicy(2,
		torchwood.SingleVerifierPolicy(checkpointVerifier),
		torchwood.OriginPolicy(checkpointVerifier.Name()),
	)

	return &LogHandle{reader, appender, awaiter, shutdown, policy}, nil
}

// AppendAndAwait submits an entry to the log and returns only after the awaiter acknowledges it.
// The index in the log where the entry appears is returned.
//
// The value of 'entry' should typically be small: this is what goes directly in the log
// (e.g., there is no further hashing done by this method).
func (lh *LogHandle) AppendAndAwait(ctx context.Context, entry []byte) (uint64, error) {
	index, _, err := lh.Awaiter.Await(ctx, lh.Appender.Add(ctx, tessera.NewEntry(entry)))
	return index.Index, err
}

type CheckpointRaw []byte

// GenerateProof creates an inclusion proof for the requested index of the log.
// It uses the most current checkpoint of the top of the log to do this,
// and therefore returns that (in raw serialized form) as well as the constructed proof.
func (lh *LogHandle) GenerateProof(ctx context.Context, idx int64) (tlog.RecordProof, CheckpointRaw, error) {
	checkpointRaw, err := lh.Reader.ReadCheckpoint(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read checkpoint: %v", err)
	}

	// At this point, we must verify the checkpoint:
	// whether this is sensible is debatable, since it's been written by "us"
	// (or if we're really in a rapidly used system, some parallel processes that presumably has the same credentials),
	// but the point is forced by the APIs for this not permitting parse without verification,
	// and we need a few values out of the checkpoint to continue.
	// This is the only reason the Policy value is needed in the LogHandle:
	// it's for our own reads back of our data, during, ironically, writing.
	c, _, err := torchwood.VerifyCheckpoint(checkpointRaw, lh.Policy)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse checkpoint: %v", err)
	}

	p, err := tlog.ProveRecord(c.N, idx, torchwood.TileHashReaderWithContext(
		ctx, c.Tree, tesserax.NewTileReader(lh.Reader)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create proof: %v", err)
	}
	return p, checkpointRaw, nil
}
