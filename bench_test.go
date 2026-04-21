// Copyright (c) 2026 rschoonheim
// SPDX-License-Identifier: MIT

package gg_test

import (
	"math/rand/v2"
	"testing"

	"github.com/rschoonheim/gg"
)

// buildGroupings creates a Groupings container of the given universe size
// with two groupings ("a" and "b"), each populated with density fraction
// of the universe drawn from a deterministic PRNG.
func buildGroupings(b *testing.B, universe uint32, density float64) (*gg.Groupings, *gg.Grouping, *gg.Grouping) {
	b.Helper()
	gs := gg.New(universe)
	r := rand.New(rand.NewPCG(1, 2))
	pick := func() []uint32 {
		out := make([]uint32, 0, int(float64(universe)*density))
		for i := uint32(0); i < universe; i++ {
			if r.Float64() < density {
				out = append(out, i)
			}
		}
		return out
	}
	ga, err := gs.Add("a", pick()...)
	if err != nil {
		b.Fatal(err)
	}
	gb, err := gs.Add("b", pick()...)
	if err != nil {
		b.Fatal(err)
	}
	return gs, ga, gb
}

var universes = []uint32{1 << 10, 1 << 14, 1 << 18, 1 << 22}

func benchName(u uint32) string {
	switch u {
	case 1 << 10:
		return "1Ki"
	case 1 << 14:
		return "16Ki"
	case 1 << 18:
		return "256Ki"
	case 1 << 22:
		return "4Mi"
	}
	return "?"
}

func BenchmarkAdd(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			// Pre-compute members to isolate the Add cost.
			r := rand.New(rand.NewPCG(1, 2))
			members := make([]uint32, 0, u/2)
			for i := uint32(0); i < u; i++ {
				if r.Float64() < 0.5 {
					members = append(members, i)
				}
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				gs := gg.New(u)
				if _, err := gs.Add("a", members...); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkContains(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, _ := buildGroupings(b, u, 0.5)
			r := rand.New(rand.NewPCG(3, 4))
			b.ResetTimer()
			b.ReportAllocs()
			var sink bool
			for i := 0; i < b.N; i++ {
				sink = ga.Contains(r.Uint32N(u))
			}
			_ = sink
		})
	}
}

func BenchmarkUnion(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, gb2 := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := ga.Union(gb2); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkIntersection(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, gb2 := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := ga.Intersection(gb2); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDifference(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, gb2 := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := ga.Difference(gb2); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkSymmetricDifference(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, gb2 := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := ga.SymmetricDifference(gb2); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkIsSubsetOf(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, gb2 := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			var sink bool
			for i := 0; i < b.N; i++ {
				sink = ga.IsSubsetOf(gb2)
			}
			_ = sink
		})
	}
}

func BenchmarkEquals(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, gb2 := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			var sink bool
			for i := 0; i < b.N; i++ {
				sink = ga.Equals(gb2)
			}
			_ = sink
		})
	}
}

func BenchmarkCardinality(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, _ := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			var sink int
			for i := 0; i < b.N; i++ {
				sink = ga.Cardinality()
			}
			_ = sink
		})
	}
}

func BenchmarkMembers(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			_, ga, _ := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ga.Members()
			}
		})
	}
}

func BenchmarkEncode(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			gs, _, _ := buildGroupings(b, u, 0.5)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := gs.Encode(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecode(b *testing.B) {
	for _, u := range universes {
		b.Run(benchName(u), func(b *testing.B) {
			gs, _, _ := buildGroupings(b, u, 0.5)
			raw, err := gs.Encode()
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := gg.Decode(raw); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
