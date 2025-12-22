package spicy

import (
	"strings"
	"testing"

	"filippo.io/torchwood"
	"github.com/go-quicktest/qt"
	"golang.org/x/mod/sumdb/note"
	"golang.org/x/mod/sumdb/tlog"
)

func TestVerifySpicySig(t *testing.T) {
	verifier, err := note.NewVerifier(dummyPublicKey)
	if err != nil {
		panic(err)
	}

	t.Run("spicy1", func(t *testing.T) {
		s, err := ParseSpicySig([]byte(spicy1), note.VerifierList(verifier))
		qt.Assert(t, qt.IsNil(err))

		t.Run("verifies successfully", func(t *testing.T) {
			qt.Assert(t, qt.IsNil(s.Verify(strings.NewReader("entry-100252-data-f8e847551052084e98"), "")))
		})
		t.Run("rejects other content", func(t *testing.T) {
			qt.Assert(t, qt.ErrorMatches(
				s.Verify(strings.NewReader("not it"), ""),
				"invalid transparency proof",
			))
		})
		t.Run("rejects other contextHint", func(t *testing.T) {
			t.Skip("nyi")
			qt.Assert(t, qt.ErrorMatches(
				s.Verify(strings.NewReader("entry-100252-data-f8e847551052084e98"), "also not it"),
				"invalid transparency proof",
			))
		})
	})

	// TODO need to finish signature minting before we can test something with contexthint.
	// that'll need a new fixture we haven't made yet.
	// because it requires a new entry format we haven't made yet.
}

// TestTorchwoodVerifySpicySig covers roughly similar behavior to TestVerifySpicySig,
// but now with most of the implemention coming from the upstreamed torchwood functions.
func TestTorchwoodVerifySpicySig(t *testing.T) {
	verifier, err := note.NewVerifier(dummyPublicKey)
	if err != nil {
		panic(err)
	}
	// This API for policy composition is a little wild.
	// The `torchwood.VerifyProof` function is going to check that the aggregate policy *rejects* an origin of "check.invalid"...
	// so, you *must* compose an `OriginPolicy` in here to have things work even in test.
	// Which, okay, fine.  Sure.
	//
	// But, the only composition mechanism that's exported is this ThresholdPolicy thing.
	// And within that, if you use any threshold other than "all of them", the origin policy can get ignored,
	// which means the "check.invalid" probe won't get flunked,
	// which means `torchwood.VerifyProof` will reject the whole operation.
	//
	// So there's approximately one correct way to strap all this together,
	// and that way doesn't actually permit using... most of its parameters.
	//
	// Methinks another version on this API will be required.
	policy := torchwood.ThresholdPolicy(2,
		torchwood.SingleVerifierPolicy(verifier),
		torchwood.OriginPolicy("example.com/log/testdata"),
	)

	record, err := RecordForBody(strings.NewReader("entry-100252-data-f8e847551052084e98"), "")
	if err != nil {
		panic(err)
	}
	recordHash := tlog.RecordHash(record)

	qt.Assert(t, qt.IsNil(torchwood.VerifyProof(policy, recordHash, []byte(spicy1))))
}
