// Package grouping holds the Grouping value type: a named, bitmap-backed
// subset of a fixed entity universe. It is internal because it is re-exported
// from the top-level gg package via a type alias.
package grouping

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"

	gbinary "github.com/rschoonheim/gg/internal/binary"
)

// ErrUniverseMismatch is returned when attempting to compare or combine
// two groupings defined over universes of different sizes.
var ErrUniverseMismatch = errors.New("gg: grouping universes do not match")

// Grouping represents a named collection of entities, stored as a bitmap
// indexed by entity number. A Grouping may be "attached" (its bitmap is
// backed by a parent Groupings payload, and mutations propagate) or
// "detached" (a standalone in-memory bitmap typically produced by a
// set-theoretic operation).
type Grouping struct {
	name        string
	bitmap      []byte
	entityCount uint32
}

// New returns a Grouping bound to the provided bitmap slice. The caller is
// responsible for making sure bitmap is sized to ceil(entityCount/8) bytes.
func New(name string, bitmap []byte, entityCount uint32) *Grouping {
	return &Grouping{name: name, bitmap: bitmap, entityCount: entityCount}
}

// -----------------------------------------------------------------------------
// Extraction / queries
// -----------------------------------------------------------------------------

// Name returns the grouping's name. Detached groupings produced by set
// operations may have a synthetic name such as "a∪b".
func (g *Grouping) Name() string { return g.name }

// EntityCount returns the size of the universe the grouping lives in.
func (g *Grouping) EntityCount() uint32 { return g.entityCount }

// Contains reports whether the given entity belongs to the grouping.
func (g *Grouping) Contains(member uint32) bool {
	if member >= g.entityCount {
		return false
	}
	return gbinary.Has(g.bitmap, int(member))
}

// Cardinality returns the number of entities in the grouping (|S|).
func (g *Grouping) Cardinality() int { return gbinary.PopCount(g.bitmap) }

// IsEmpty reports whether the grouping is the empty set.
func (g *Grouping) IsEmpty() bool {
	bm := g.bitmap
	i := 0
	for ; i+8 <= len(bm); i += 8 {
		if binary.LittleEndian.Uint64(bm[i:]) != 0 {
			return false
		}
	}
	for ; i < len(bm); i++ {
		if bm[i] != 0 {
			return false
		}
	}
	return true
}

// Members returns the sorted list of entity indices in the grouping. It
// scans the bitmap in 64-bit chunks and iterates only the set bits via
// bits.TrailingZeros64, which is much faster than a byte × bit loop for
// sparse and mid-density bitmaps alike.
func (g *Grouping) Members() []uint32 {
	out := make([]uint32, 0, g.Cardinality())
	bm := g.bitmap
	ec := g.entityCount
	i := 0
	for ; i+8 <= len(bm); i += 8 {
		w := binary.LittleEndian.Uint64(bm[i:])
		base := uint32(i) * 8
		for w != 0 {
			b := bits.TrailingZeros64(w)
			m := base + uint32(b)
			if m < ec {
				out = append(out, m)
			}
			w &= w - 1
		}
	}
	for ; i < len(bm); i++ {
		bt := bm[i]
		base := uint32(i) * 8
		for bt != 0 {
			b := bits.TrailingZeros8(bt)
			m := base + uint32(b)
			if m < ec {
				out = append(out, m)
			}
			bt &= bt - 1
		}
	}
	return out
}

// String returns a compact debug representation.
func (g *Grouping) String() string {
	return fmt.Sprintf("Grouping(%q, members=%v)", g.name, g.Members())
}

// -----------------------------------------------------------------------------
// Ingestion / mutation
// -----------------------------------------------------------------------------

// Insert adds an entity to the grouping.
func (g *Grouping) Insert(member uint32) error {
	if member >= g.entityCount {
		return fmt.Errorf("gg: member %d out of range [0,%d)", member, g.entityCount)
	}
	gbinary.Set(g.bitmap, int(member))
	return nil
}

// Remove clears an entity from the grouping.
func (g *Grouping) Remove(member uint32) error {
	if member >= g.entityCount {
		return fmt.Errorf("gg: member %d out of range [0,%d)", member, g.entityCount)
	}
	gbinary.Clear(g.bitmap, int(member))
	return nil
}

// -----------------------------------------------------------------------------
// Comparison / set-theoretic operations
// -----------------------------------------------------------------------------

func sameUniverse(a, b *Grouping) bool {
	return a.entityCount == b.entityCount && len(a.bitmap) == len(b.bitmap)
}

func newDetached(name string, size int, entityCount uint32) *Grouping {
	return &Grouping{
		name:        name,
		bitmap:      make([]byte, size),
		entityCount: entityCount,
	}
}

// Union returns a new detached grouping containing every entity present in
// either g or other (g ∪ other).
func (g *Grouping) Union(other *Grouping) (*Grouping, error) {
	if !sameUniverse(g, other) {
		return nil, ErrUniverseMismatch
	}
	out := newDetached(g.name+"∪"+other.name, len(g.bitmap), g.entityCount)
	a, b, o := g.bitmap, other.bitmap, out.bitmap
	i := 0
	for ; i+8 <= len(a); i += 8 {
		binary.LittleEndian.PutUint64(o[i:],
			binary.LittleEndian.Uint64(a[i:])|binary.LittleEndian.Uint64(b[i:]))
	}
	for ; i < len(a); i++ {
		o[i] = a[i] | b[i]
	}
	return out, nil
}

// Intersection returns a new detached grouping containing only the entities
// that belong to both g and other (g ∩ other).
func (g *Grouping) Intersection(other *Grouping) (*Grouping, error) {
	if !sameUniverse(g, other) {
		return nil, ErrUniverseMismatch
	}
	out := newDetached(g.name+"∩"+other.name, len(g.bitmap), g.entityCount)
	a, b, o := g.bitmap, other.bitmap, out.bitmap
	i := 0
	for ; i+8 <= len(a); i += 8 {
		binary.LittleEndian.PutUint64(o[i:],
			binary.LittleEndian.Uint64(a[i:])&binary.LittleEndian.Uint64(b[i:]))
	}
	for ; i < len(a); i++ {
		o[i] = a[i] & b[i]
	}
	return out, nil
}

// Difference returns a new detached grouping containing the entities that
// belong to g but not to other (g \ other).
func (g *Grouping) Difference(other *Grouping) (*Grouping, error) {
	if !sameUniverse(g, other) {
		return nil, ErrUniverseMismatch
	}
	out := newDetached(g.name+`\`+other.name, len(g.bitmap), g.entityCount)
	a, b, o := g.bitmap, other.bitmap, out.bitmap
	i := 0
	for ; i+8 <= len(a); i += 8 {
		binary.LittleEndian.PutUint64(o[i:],
			binary.LittleEndian.Uint64(a[i:])&^binary.LittleEndian.Uint64(b[i:]))
	}
	for ; i < len(a); i++ {
		o[i] = a[i] &^ b[i]
	}
	return out, nil
}

// SymmetricDifference returns (g ∪ other) \ (g ∩ other).
func (g *Grouping) SymmetricDifference(other *Grouping) (*Grouping, error) {
	if !sameUniverse(g, other) {
		return nil, ErrUniverseMismatch
	}
	out := newDetached(g.name+"△"+other.name, len(g.bitmap), g.entityCount)
	a, b, o := g.bitmap, other.bitmap, out.bitmap
	i := 0
	for ; i+8 <= len(a); i += 8 {
		binary.LittleEndian.PutUint64(o[i:],
			binary.LittleEndian.Uint64(a[i:])^binary.LittleEndian.Uint64(b[i:]))
	}
	for ; i < len(a); i++ {
		o[i] = a[i] ^ b[i]
	}
	return out, nil
}

// IsSubsetOf reports whether every entity of g also belongs to other
// (g ⊆ other).
func (g *Grouping) IsSubsetOf(other *Grouping) bool {
	if !sameUniverse(g, other) {
		return false
	}
	a, b := g.bitmap, other.bitmap
	i := 0
	for ; i+8 <= len(a); i += 8 {
		if binary.LittleEndian.Uint64(a[i:])&^binary.LittleEndian.Uint64(b[i:]) != 0 {
			return false
		}
	}
	for ; i < len(a); i++ {
		if a[i]&^b[i] != 0 {
			return false
		}
	}
	return true
}

// IsSupersetOf reports whether g ⊇ other.
func (g *Grouping) IsSupersetOf(other *Grouping) bool { return other.IsSubsetOf(g) }

// Equals reports whether g and other contain exactly the same entities.
func (g *Grouping) Equals(other *Grouping) bool {
	if !sameUniverse(g, other) {
		return false
	}
	a, b := g.bitmap, other.bitmap
	i := 0
	for ; i+8 <= len(a); i += 8 {
		if binary.LittleEndian.Uint64(a[i:]) != binary.LittleEndian.Uint64(b[i:]) {
			return false
		}
	}
	for ; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Disjoint reports whether g and other share no entities (g ∩ other = ∅).
func (g *Grouping) Disjoint(other *Grouping) bool {
	if !sameUniverse(g, other) {
		return false
	}
	a, b := g.bitmap, other.bitmap
	i := 0
	for ; i+8 <= len(a); i += 8 {
		if binary.LittleEndian.Uint64(a[i:])&binary.LittleEndian.Uint64(b[i:]) != 0 {
			return false
		}
	}
	for ; i < len(a); i++ {
		if a[i]&b[i] != 0 {
			return false
		}
	}
	return true
}

