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

	// Set up a log.
	signer, err := note.NewSigner(dummyPrivateKey)
	qt.Assert(t, qt.IsNil(err))
	lop, err := OperateLog(
		context.Background(),
		logPath,
		tessera.NewAppendOptions().
			WithCheckpointSigner(signer).
			WithCheckpointInterval(100*time.Millisecond),
	)
	qt.Assert(t, qt.IsNil(err))

	// Parse the pubkey we'll use during verify checks.
	verifier, err := note.NewVerifier(dummyPublicKey)
	qt.Assert(t, qt.IsNil(err))

	t.Run("entry 1", func(t *testing.T) {
		fixtureBody, fixtureContextHint := "hayo", ""

		// Use the log to sign.
		sigRaw, err := lop.Sign(
			context.Background(),
			strings.NewReader(fixtureBody),
			fixtureContextHint,
		)
		qt.Assert(t, qt.IsNil(err))

		// For your eyeballing pleasure:
		t.Logf(">>>%v<<<", string(sigRaw))

		t.Run("result verifies", func(t *testing.T) {
			sig, err := spicy.ParseSpicySig(sigRaw, note.VerifierList(verifier))
			qt.Assert(t, qt.IsNil(err))
			err = sig.Verify(
				strings.NewReader(fixtureBody),
				fixtureContextHint,
			)
			qt.Assert(t, qt.IsNil(err))
		})

		t.Run("entry 2", func(t *testing.T) {
			fixtureBody, fixtureContextHint := "heckie", ""

			// Use the log to sign.
			sigRaw, err := lop.Sign(
				context.Background(),
				strings.NewReader(fixtureBody),
				fixtureContextHint,
			)
			qt.Assert(t, qt.IsNil(err))

			// For your eyeballing pleasure:
			t.Logf(">>>%v<<<", string(sigRaw))

			t.Run("result verifies", func(t *testing.T) {
				sig, err := spicy.ParseSpicySig(sigRaw, note.VerifierList(verifier))
				qt.Assert(t, qt.IsNil(err))
				err = sig.Verify(
					strings.NewReader(fixtureBody),
					fixtureContextHint,
				)
				qt.Assert(t, qt.IsNil(err))
			})
		})
	})
}
