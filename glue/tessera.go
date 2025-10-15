package glue

import (
	"context"
	"fmt"
	"log/slog"

	"filippo.io/torchwood"
	"github.com/transparency-dev/tessera"
	"golang.org/x/mod/sumdb/tlog"
)

var _ torchwood.TileReaderWithContext = (*tesseraTileReader)(nil)

// tesseraTileReader implements torchwood.TileReaderWithContext
type tesseraTileReader struct {
	reader tessera.LogReader
}

// NewTesseraTileReader produces a TileReader which works over the reader interface
// that you get from using tessera as an appender.
//
// For other many situations, consider using [torchwood.TileFetcher]
// (which operates over http).
func NewTesseraTileReader(tesseraReader tessera.LogReader) torchwood.TileReaderWithContext {
	return &tesseraTileReader{tesseraReader}
}

// ReadTiles implements torchwood.TileReaderWithContext.
// It transforms the `sumdb/tlog`/`torchwood` calling conventions (which use structs, and return values in batches)
// to the tessera convention (which uses positional arguments, and reads one tile at a time).
func (tr *tesseraTileReader) ReadTiles(ctx context.Context, tiles []tlog.Tile) ([][]byte, error) {
	data := make([][]byte, len(tiles))
	for i, tile := range tiles {
		// TODO:REVIEW: this bit of bodging appears to be operationally correct but I'm not confident it's exactly correct or unfragile.
		// Everyone seems to tacitly agree that "H = 8" right now, and tessera doesn't even appear to have a parameter for that, but I don't know what breaks if that changes.
		// Is there some existing home of a better form of this code?  It feels weird that something this fiddly wouldn't already be exported somewhere.
		level := uint64(tile.L)
		index := uint64(tile.N)
		partial := uint8(tile.W)
		slog.Debug("readtiles", "tile", tile)
		tileData, err := tr.reader.ReadTile(ctx, level, index, partial)
		if err != nil {
			return nil, fmt.Errorf("failed to read tessera tile level=%d, index=%d, partial=%d: %w", level, index, partial, err)
		}
		data[i] = tileData
	}
	return data, nil
}

// SaveTiles is required to implement torchwood.TileReaderWithContext (but does nothing in this implementation).
func (tr *tesseraTileReader) SaveTiles(tiles []tlog.Tile, data [][]byte) {
	// No-op.
}
