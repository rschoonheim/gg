package gg

import (
	"fmt"
	"os"

	"gg/internal/binary"
	"gg/internal/grouping"
)

// Names returns the names of every grouping in insertion order.
func (gs *Groupings) Names() []string { return gs.inner.Headers.Names() }

// Get returns the grouping with the given name, or false if none exists.
// The returned *Grouping is attached: mutations on it propagate to the
// parent Groupings payload.
func (gs *Groupings) Get(name string) (*Grouping, bool) {
	idx := gs.inner.Headers.IndexOf(name)
	if idx < 0 {
		return nil, false
	}
	return grouping.New(name, gs.inner.Bitmap(idx), gs.EntityCount()), true
}

// All returns every grouping in the container.
func (gs *Groupings) All() []*Grouping {
	names := gs.Names()
	ec := gs.EntityCount()
	out := make([]*Grouping, len(names))
	for i, n := range names {
		out[i] = grouping.New(n, gs.inner.Bitmap(i), ec)
	}
	return out
}

// SaveFile encodes the container and writes it to path (conventionally with
// a `.bin` extension). The file is written atomically through a temporary
// sibling file and a rename.
func (gs *Groupings) SaveFile(path string) error {
	raw, err := gs.Encode()
	if err != nil {
		return fmt.Errorf("gg: save %q: %w", path, err)
	}
	tmp, err := os.CreateTemp(filepathDir(path), ".gg-*.tmp")
	if err != nil {
		return fmt.Errorf("gg: save %q: %w", path, err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("gg: save %q: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("gg: save %q: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("gg: save %q: %w", path, err)
	}
	return nil
}

// Find returns every grouping that contains the given entity.
func (gs *Groupings) Find(member uint32) []*Grouping {
	ec := gs.EntityCount()
	if member >= ec {
		return nil
	}
	names := gs.inner.Headers.Names()
	out := make([]*Grouping, 0)
	for i, n := range names {
		bm := gs.inner.Bitmap(i)
		if binary.Has(bm, int(member)) {
			out = append(out, grouping.New(n, bm, ec))
		}
	}
	return out
}

// FindAll returns every grouping that contains all of the given entities
// (i.e. groupings G such that {members...} ⊆ G). With no members it returns
// every grouping.
func (gs *Groupings) FindAll(members ...uint32) []*Grouping {
	ec := gs.EntityCount()
	for _, m := range members {
		if m >= ec {
			return nil
		}
	}
	names := gs.inner.Headers.Names()
	out := make([]*Grouping, 0)
	for i, n := range names {
		bm := gs.inner.Bitmap(i)
		ok := true
		for _, m := range members {
			if !binary.Has(bm, int(m)) {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, grouping.New(n, bm, ec))
		}
	}
	return out
}

// FindAny returns every grouping that contains at least one of the given
// entities. With no members it returns no groupings.
func (gs *Groupings) FindAny(members ...uint32) []*Grouping {
	ec := gs.EntityCount()
	names := gs.inner.Headers.Names()
	out := make([]*Grouping, 0)
	for i, n := range names {
		bm := gs.inner.Bitmap(i)
		for _, m := range members {
			if m >= ec {
				continue
			}
			if binary.Has(bm, int(m)) {
				out = append(out, grouping.New(n, bm, ec))
				break
			}
		}
	}
	return out
}

// FindNames is a convenience wrapper around Find that returns only the
// matching grouping names, in insertion order.
func (gs *Groupings) FindNames(member uint32) []string {
	gr := gs.Find(member)
	out := make([]string, len(gr))
	for i, g := range gr {
		out[i] = g.Name()
	}
	return out
}

// filepathDir returns the directory of path without importing path/filepath
// at the top of the file. Kept local to avoid churn in existing imports.
func filepathDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
