package spicy

import (
	"strings"
	"testing"

	"github.com/go-quicktest/qt"
	"golang.org/x/mod/sumdb/note"
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
