// Copyright (c) 2026 rschoonheim
// SPDX-License-Identifier: MIT

package binary

import (
	"encoding/binary"
	"math/bits"
)

// Data is the binary-encoded payload of a Groupings object. It stores a
// concatenated sequence of fixed-size bitmaps, one per grouping, in the same
// order as the names table in Headers.
type Data []byte

// BitmapSize returns the number of bytes required to hold a bitmap for the
// given entity count.
func BitmapSize(entityCount uint32) int {
	return int((entityCount + 7) / 8)
}

// AppendBitmap appends a new zero-initialised bitmap of the given size.
func (d *Data) AppendBitmap(size int) {
	*d = append(*d, make([]byte, size)...)
}

// Bitmap returns the sub-slice backing the i-th bitmap. Mutations on the
// returned slice are reflected in the underlying Data.
func (d *Data) Bitmap(i, size int) []byte {
	start := i * size
	return (*d)[start : start+size]
}

// Set turns on the bit at position bit in bm.
func Set(bm []byte, bit int) { bm[bit>>3] |= 1 << (bit & 7) }

// Clear turns off the bit at position bit in bm.
func Clear(bm []byte, bit int) { bm[bit>>3] &^= 1 << (bit & 7) }

// Has reports whether the bit at position bit in bm is set.
func Has(bm []byte, bit int) bool { return bm[bit>>3]&(1<<(bit&7)) != 0 }

// PopCount returns the number of set bits in bm. It processes the bitmap in
// 64-bit chunks, falling back to byte-wise counting for the trailing tail.
func PopCount(bm []byte) int {
	n := 0
	i := 0
	for ; i+8 <= len(bm); i += 8 {
		n += bits.OnesCount64(binary.LittleEndian.Uint64(bm[i:]))
	}
	for ; i < len(bm); i++ {
		n += bits.OnesCount8(bm[i])
	}
	return n
}
