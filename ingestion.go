// Package gg provides a high-performance data structure for managing and
// organising groupings of entities. Groupings are stored as packed bitmaps
// over a fixed universe of entity indices, which makes membership tests and
// set-theoretic operations (union, intersection, difference, subset) run at
// bitwise speed regardless of scale.
package gg

import (
	"errors"
	"fmt"
	"os"

	"gg/internal/binary"
	"gg/internal/grouping"
	"gg/internal/groupings"
)

// MagicNumber is written at the start of every encoded Groupings payload.
var MagicNumber = [4]byte{'G', 'G', 'G', 'G'}

// Groupings is the root entity: a container holding any number of named
// Grouping instances defined over the same universe of entity indices.
type Groupings struct {
	inner *groupings.Groupings
}

// Grouping is re-exported from the internal grouping package so that callers
// can refer to it as gg.Grouping while the type itself — along with its
// methods — lives in internal/grouping.
type Grouping = grouping.Grouping

// New creates a new, empty Groupings container over a universe of
// entityCount entity indices (valid members are 0 .. entityCount-1).
func New(entityCount uint32) *Groupings {
	h := binary.NewHeaders(entityCount)
	h.SetMagicNumber(MagicNumber)
	d := binary.Data{}
	return &Groupings{inner: groupings.New(h, &d)}
}

// EntityCount returns the size of the universe.
func (gs *Groupings) EntityCount() uint32 { return gs.inner.Headers.EntityCount() }

// Len returns the number of groupings currently stored.
func (gs *Groupings) Len() int { return int(gs.inner.Headers.GroupingCount()) }

// Add creates a new grouping with the given name and optional initial
// members. It returns an error if a grouping with the same name already
// exists or if any member is out of range.
func (gs *Groupings) Add(name string, members ...uint32) (*Grouping, error) {
	if gs.inner.Headers.IndexOf(name) >= 0 {
		return nil, fmt.Errorf("gg: grouping %q already exists", name)
	}
	for _, m := range members {
		if m >= gs.EntityCount() {
			return nil, fmt.Errorf("gg: member %d out of range [0,%d)", m, gs.EntityCount())
		}
	}
	idx := gs.inner.AppendGrouping(name)
	bm := gs.inner.Bitmap(idx)
	for _, m := range members {
		binary.Set(bm, int(m))
	}
	return grouping.New(name, bm, gs.EntityCount()), nil
}

// Encode serialises the Groupings container to a single byte slice.
func (gs *Groupings) Encode() ([]byte, error) { return gs.inner.Encode() }

// Decode parses a byte slice previously produced by (*Groupings).Encode
// back into a Groupings container.
func Decode(raw []byte) (*Groupings, error) {
	headers, consumed, err := binary.ParseHeaders(raw)
	if err != nil {
		return nil, err
	}
	if headers.MagicNumber() != MagicNumber {
		return nil, errors.New("gg: invalid magic number")
	}
	size := binary.BitmapSize(headers.EntityCount())
	expected := size * int(headers.GroupingCount())
	if len(raw)-consumed < expected {
		return nil, errors.New("gg: truncated data payload")
	}
	data := make(binary.Data, expected)
	copy(data, raw[consumed:consumed+expected])
	return &Groupings{inner: groupings.New(&headers, &data)}, nil
}

// LoadFile reads a Groupings payload from the `.bin` file at path and
// decodes it. The file is expected to be the output of (*Groupings).SaveFile
// or an equivalent byte stream produced by (*Groupings).Encode.
func LoadFile(path string) (*Groupings, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("gg: load %q: %w", path, err)
	}
	gs, err := Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("gg: load %q: %w", path, err)
	}
	return gs, nil
}
