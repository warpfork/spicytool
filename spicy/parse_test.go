package spicy

import (
	"testing"

	"filippo.io/torchwood"
	"github.com/go-quicktest/qt"
	"golang.org/x/mod/sumdb/note"
	"golang.org/x/mod/sumdb/tlog"
)

// tl;dr of below:
//   - parse of tree checkpoints has variations on if it tolerates double-linebreak bodies.
//   - but note.Open rejects double-linebreak bodies
//   - ... so it's probably a distinction without a difference.

const cpbodyWithTrailer = `go.sum database tree
923748
nND/nri/U0xuHUrYSy0HtMeal2vzD9V4k/BO79C+QeI=
and then stuff
`

const cpbodyWithSep = `go.sum database tree
923748
nND/nri/U0xuHUrYSy0HtMeal2vzD9V4k/BO79C+QeI=

and then stuff
`

// Examine a couple of upstream libraries for consistency of interpretation
// of a text that's a bit of an edge case.
func TestCheckpointForbidsExcessSections(t *testing.T) {
	t.Run("torchwood implementation", func(t *testing.T) {
		t.Run("accepts trailer", func(t *testing.T) {
			checkp, err := torchwood.ParseCheckpoint(cpbodyWithTrailer)
			qt.Assert(t, qt.IsNil(err))
			qt.Assert(t, qt.Equals(checkp.Extension, "and then stuff\n"))
		})
		t.Run("rejects sep", func(t *testing.T) {
			_, err := torchwood.ParseCheckpoint(cpbodyWithSep)
			qt.Assert(t, qt.IsNotNil(err))
		})
	})
	t.Run("sumdb implementation", func(t *testing.T) {
		t.Run("accepts trailer", func(t *testing.T) {
			_, err := tlog.ParseTree([]byte(cpbodyWithTrailer))
			qt.Assert(t, qt.IsNil(err))
			// sumdb api does not save the trailer: so no assert is possible.
		})
		t.Run("rejects sep", func(t *testing.T) {
			t.Skip("THIS TEST FAILS.  There is a behavioral divergence between these implementations.")
			_, err := tlog.ParseTree([]byte(cpbodyWithSep))
			qt.Assert(t, qt.IsNotNil(err))
		})
	})
}

const noteWithBodySimple = `
note body

— example.com/foo EKNzoDWG8LGC0Yp9o+sv3qllpMP9uHQ9B20KNL+Q1zs=
`

const noteWithBodySep = `
note with longer body

with challenges

— example.com/foo hI2DJw[...]1roloI=
`

func TestNoteForbidsExcessSections(t *testing.T) {
	unverified := &note.UnverifiedNoteError{}
	t.Run("accepts simple note", func(t *testing.T) {
		_, err := note.Open([]byte(noteWithBodySimple), nil)
		qt.Assert(t, qt.ErrorAs(err, &unverified))
	})
	t.Run("rejects note with sep in body", func(t *testing.T) {
		_, err := note.Open([]byte(noteWithBodySep), nil)
		qt.Assert(t, qt.Not(qt.ErrorAs(err, &unverified)))
		qt.Assert(t, qt.ErrorMatches(err, "malformed note"))
	})
}
