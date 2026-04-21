// Copyright (c) 2026 rschoonheim
// SPDX-License-Identifier: MIT

package binary

import (
	"encoding/binary"
	"errors"
)

// Headers is the binary-encoded header block of a Groupings payload.
//
// Layout (little-endian):
//
//	[0..4]   magic number (4 bytes)
//	[4..8]   entity count (uint32)
//	[8..12]  grouping count (uint32)
//	[12..]   names table: for each grouping, uint16 length + utf-8 name bytes
type Headers []byte

const (
	magicOffset      = 0
	entityCountOff   = 4
	groupingCountOff = 8
	namesOff         = 12
)

// NewHeaders returns a new Headers initialised with the given entity count
// and zero groupings.
func NewHeaders(entityCount uint32) *Headers {
	h := make(Headers, namesOff)
	binary.LittleEndian.PutUint32(h[entityCountOff:], entityCount)
	return &h
}

// ParseHeaders parses a Headers block from raw bytes and returns the parsed
// headers together with the number of bytes consumed.
func ParseHeaders(raw []byte) (Headers, int, error) {
	if len(raw) < namesOff {
		return nil, 0, errors.New("binary: truncated headers")
	}
	count := binary.LittleEndian.Uint32(raw[groupingCountOff:])
	offset := namesOff
	for i := uint32(0); i < count; i++ {
		if offset+2 > len(raw) {
			return nil, 0, errors.New("binary: truncated name length")
		}
		l := int(binary.LittleEndian.Uint16(raw[offset:]))
		offset += 2 + l
		if offset > len(raw) {
			return nil, 0, errors.New("binary: truncated name")
		}
	}
	out := make(Headers, offset)
	copy(out, raw[:offset])
	return out, offset, nil
}

func (h *Headers) SetMagicNumber(magicNumber [4]byte) {
	(*h)[0] = magicNumber[0]
	(*h)[1] = magicNumber[1]
	(*h)[2] = magicNumber[2]
	(*h)[3] = magicNumber[3]
}

func (h *Headers) MagicNumber() [4]byte {
	return [4]byte{(*h)[0], (*h)[1], (*h)[2], (*h)[3]}
}

// EntityCount returns the number of entities addressable by every grouping.
func (h *Headers) EntityCount() uint32 {
	return binary.LittleEndian.Uint32((*h)[entityCountOff:])
}

// GroupingCount returns the number of groupings.
func (h *Headers) GroupingCount() uint32 {
	return binary.LittleEndian.Uint32((*h)[groupingCountOff:])
}

func (h *Headers) setGroupingCount(n uint32) {
	binary.LittleEndian.PutUint32((*h)[groupingCountOff:], n)
}

// AppendName appends a grouping name to the names table and returns its index.
func (h *Headers) AppendName(name string) uint32 {
	idx := h.GroupingCount()
	nameBytes := []byte(name)
	var lenBuf [2]byte
	binary.LittleEndian.PutUint16(lenBuf[:], uint16(len(nameBytes)))
	*h = append(*h, lenBuf[:]...)
	*h = append(*h, nameBytes...)
	h.setGroupingCount(idx + 1)
	return idx
}

// Names returns all grouping names in insertion order.
func (h *Headers) Names() []string {
	count := h.GroupingCount()
	names := make([]string, 0, count)
	offset := namesOff
	for i := uint32(0); i < count; i++ {
		l := int(binary.LittleEndian.Uint16((*h)[offset:]))
		offset += 2
		names = append(names, string((*h)[offset:offset+l]))
		offset += l
	}
	return names
}

// IndexOf returns the index of the grouping with the given name, or -1.
func (h *Headers) IndexOf(name string) int {
	count := h.GroupingCount()
	offset := namesOff
	for i := uint32(0); i < count; i++ {
		l := int(binary.LittleEndian.Uint16((*h)[offset:]))
		offset += 2
		if string((*h)[offset:offset+l]) == name {
			return int(i)
		}
		offset += l
	}
	return -1
}
