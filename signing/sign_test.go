package signing

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-quicktest/qt"
	"github.com/transparency-dev/tessera"
	"github.com/warpfork/spicytool/spicy"
	"golang.org/x/mod/sumdb/note"
)

const (
	dummyPrivateKey = "PRIVATE+KEY+example.com/log/testdata+33d7b496+AeymY/SZAX0jZcJ8enZ5FY1Dz+wTML2yWSkK+9DSF3eg"
	dummyPublicKey  = "example.com/log/testdata+33d7b496+AeHTu4Q3hEIMHNqc6fASMsq3rKNx280NI+oO5xCFkkSx"
)

const (
	logPath = "/tmp/tlog"
)

func TestMakeSpicySig(t *testing.T) {
	// TODO sensible tempdir creation please
	qt.Assert(t, qt.IsNil(os.RemoveAll(logPath)))

	signer, err := note.NewSigner(dummyPrivateKey)
	qt.Assert(t, qt.IsNil(err))

	// Set up a log.
	lop, err := OperateLog(
		context.Background(),
		logPath,
		tessera.NewAppendOptions().
			WithCheckpointSigner(signer).
			WithCheckpointInterval(100*time.Millisecond),
	)
	qt.Assert(t, qt.IsNil(err))

	// Use the log to sign.
	sigRaw, err := lop.Sign(
		context.Background(),
		strings.NewReader("hayo"),
		"",
	)
	qt.Assert(t, qt.IsNil(err))

	// ... Funny thing, let's do it twice.  Reasons.  Shh.
	// (There's a bug in the parser for the very first entry, lol.)
	sigRaw, err = lop.Sign(
		context.Background(),
		strings.NewReader("hayo"),
		"",
	)
	qt.Assert(t, qt.IsNil(err))

	// For your eyeballing pleasure:
	t.Logf(">>>%v<<<", string(sigRaw))

	// Can we parse and verify it?
	verifier, err := note.NewVerifier(dummyPublicKey)
	qt.Assert(t, qt.IsNil(err))
	sig, err := spicy.ParseSpicySig(sigRaw, note.VerifierList(verifier))
	qt.Assert(t, qt.IsNil(err))
	err = sig.Verify(strings.NewReader("hayo"), "")
	qt.Assert(t, qt.IsNil(err))
}
