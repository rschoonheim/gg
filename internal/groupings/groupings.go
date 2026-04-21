package groupings

import "gg/internal/binary"

// Groupings is the internal binary-backed representation of a family of
// groupings. It owns the Headers (meta-data) and Data (concatenated bitmaps)
// blocks of the binary payload.
type Groupings struct {
	Headers *binary.Headers
	Data    *binary.Data
}

// New returns a new instance of Groupings with the provided headers and data.
func New(headers *binary.Headers, data *binary.Data) *Groupings {
	return &Groupings{
		Headers: headers,
		Data:    data,
	}
}

// Encode encodes the Groupings into a contiguous byte slice suitable for
// persistence or transport.
func (g *Groupings) Encode() ([]byte, error) {
	out := make([]byte, 0, len(*g.Headers)+len(*g.Data))
	out = append(out, *g.Headers...)
	out = append(out, *g.Data...)
	return out, nil
}

// BitmapSize returns the per-grouping bitmap size in bytes.
func (g *Groupings) BitmapSize() int {
	return binary.BitmapSize(g.Headers.EntityCount())
}

// Bitmap returns the bitmap slice of the i-th grouping.
func (g *Groupings) Bitmap(i int) []byte {
	return g.Data.Bitmap(i, g.BitmapSize())
}

// AppendGrouping registers a new grouping name and allocates its bitmap.
// It returns the index of the newly-created grouping.
func (g *Groupings) AppendGrouping(name string) int {
	idx := g.Headers.AppendName(name)
	g.Data.AppendBitmap(g.BitmapSize())
	return int(idx)
}
