package spicy

import (
	"testing"

	"filippo.io/torchwood"
	"github.com/go-quicktest/qt"
	"golang.org/x/mod/sumdb/tlog"
)

const cpnoteWithTrailer = `go.sum database tree
923748
nND/nri/U0xuHUrYSy0HtMeal2vzD9V4k/BO79C+QeI=
and then stuff
`

const cpnoteWithSep = `go.sum database tree
923748
nND/nri/U0xuHUrYSy0HtMeal2vzD9V4k/BO79C+QeI=

and then stuff
`

// Examine a couple of upstream libraries for consistency of interpretation
// of a text that's a bit of an edge case.
func TestCheckpointForbidsExcessSections(t *testing.T) {
	t.Run("torchwood implementation", func(t *testing.T) {
		t.Run("accepts trailer", func(t *testing.T) {
			checkp, err := torchwood.ParseCheckpoint(cpnoteWithTrailer)
			qt.Assert(t, qt.IsNil(err))
			qt.Assert(t, qt.Equals(checkp.Extension, "and then stuff\n"))
		})
		t.Run("rejects sep", func(t *testing.T) {
			_, err := torchwood.ParseCheckpoint(cpnoteWithSep)
			qt.Assert(t, qt.IsNotNil(err))
		})
	})
	t.Run("sumdb implementation", func(t *testing.T) {
		t.Run("accepts trailer", func(t *testing.T) {
			_, err := tlog.ParseTree([]byte(cpnoteWithTrailer))
			qt.Assert(t, qt.IsNil(err))
			// sumdb api does not save the trailer: so no assert is possible.
		})
		t.Run("rejects sep", func(t *testing.T) {
			t.Skip("THIS TEST FAILS.  There is a behavioral divergence between these implementations.")
			_, err := tlog.ParseTree([]byte(cpnoteWithSep))
			qt.Assert(t, qt.IsNotNil(err))
		})
	})
}
