# Go - Groupings (GG)

[![CI](https://github.com/rschoonheim/gg/actions/workflows/ci.yml/badge.svg)](https://github.com/rschoonheim/gg/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/rschoonheim/gg.svg)](https://pkg.go.dev/github.com/rschoonheim/gg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Go groupings (GG) is a Go package that provides a high-performance data structure for managing and organizing groupings
of entities. It operates at the binary level using bitwise operations, making membership tests, set operations, and
grouping queries extremely fast. The package allows you to create, manipulate, and query groupings of entities with
minimal overhead, regardless of the scale of your data.

## Using GG
To use the `gg` package in your Go project, you can install it using the following command:
```bash
go get github.com/rschoonheim/gg
```

## Entities

### Groupings

`Groupings` is the root entity of the package. It acts as a container that holds a collection of `Grouping` instances.
It provides the top-level API for creating, retrieving, and managing groupings.

### Grouping

A `Grouping` represents a named collection of entities. Each grouping holds an arbitrary number of members and can be
queried or manipulated independently. Groupings within the same `Groupings` container are uniquely identified by their
name.

## Link to set theory

The concepts in this package are closely related to **set theory** in mathematics.

- A **Grouping** corresponds to a **set**: an unordered collection of distinct entities (members).
- A **Groupings** corresponds to a **family of sets** (or a set of sets): a collection of sets, each of which can be
  independently defined and manipulated.

### Set operations

Because each `Grouping` models a set, the package can naturally support classical set-theoretic operations:

| Operation        | Description                                            |
|------------------|--------------------------------------------------------|
| **Union**        | All members that belong to either grouping             |
| **Intersection** | Members common to both groupings                       |
| **Difference**   | Members in one grouping but not the other              |
| **Subset**       | Check whether one grouping is contained within another |
| **Membership**   | Check whether an entity belongs to a grouping          |

### Formal analogy

| Set theory concept  | GG concept                         |
|---------------------|------------------------------------|
| Universal set       | `Groupings` (root)                 |
| Set                 | `Grouping`                         |
| Element / member    | Entity within a `Grouping`         |
| Empty set (∅)       | An empty `Grouping`                |
| Cardinality (\|S\|) | Number of entities in a `Grouping` |

## How it works

Every `Groupings` container defines a fixed **universe** of `N` entities,
identified by integer indices in the range `[0, N)`. Each `Grouping` stored
inside the container is represented by a **bitmap** of exactly
`ceil(N / 8)` bytes: bit `i` is set if and only if entity `i` belongs to
that grouping.

Because every grouping in the same container shares the same bitmap layout,
operations between two groupings reduce to a single linear pass over their
backing byte slices using hardware bitwise instructions:

| Operation              | Bitwise expression |
|------------------------|--------------------|
| Union (`a ∪ b`)        | `a \| b`           |
| Intersection (`a ∩ b`) | `a & b`            |
| Difference (`a \ b`)   | `a &^ b`           |
| Symmetric difference   | `a ^ b`            |
| Subset (`a ⊆ b`)       | `a &^ b == 0`      |
| Membership             | `a[i>>3] & (1<<(i&7))` |
| Cardinality            | `popcount(a)`      |

Attached groupings returned by `Add`, `Get` and `All` share storage with
the parent container, so `Insert` / `Remove` propagate directly to the
encoded payload. Set-theoretic operations (`Union`, `Intersection`, …)
return detached groupings with a freshly allocated bitmap and a synthetic
name such as `"a∪b"`.

### Usage

```go
package main

import (
    "fmt"

    "github.com/rschoonheim/gg"
)

func main() {
    // Create a container over a universe of 1024 entities.
    gs := gg.New(1024)

    // Define two groupings with some initial members.
    a, _ := gs.Add("a", 1, 2, 3, 10, 50)
    b, _ := gs.Add("b", 2, 3, 4, 50, 999)

    // Mutate an attached grouping.
    _ = a.Insert(7)
    _ = a.Remove(1)

    // Query.
    fmt.Println(a.Contains(7))   // true
    fmt.Println(a.Cardinality()) // 5
    fmt.Println(a.Members())     // [2 3 7 10 50]

    // Set-theoretic operations return detached groupings.
    u, _ := a.Union(b)
    i, _ := a.Intersection(b)
    d, _ := a.Difference(b)
    fmt.Println(u.Members(), i.Members(), d.Members())
    fmt.Println(a.IsSubsetOf(b)) // false

    // Persist and reload via an in-memory byte slice…
    raw, _ := gs.Encode()
    restored, _ := gg.Decode(raw)
    a2, _ := restored.Get("a")
    fmt.Println(a.Equals(a2)) // true

    // …or directly to/from a `.bin` file.
    _ = gs.SaveFile("groupings.bin")
    loaded, _ := gg.LoadFile("groupings.bin")

    // Search for every grouping that contains a particular entity.
    for _, g := range loaded.Find(50) {
        fmt.Println("50 ∈", g.Name())
    }
    fmt.Println(loaded.FindNames(2))     // ["a", "b"]
    fmt.Println(loaded.FindAll(2, 50))   // groupings containing both
    fmt.Println(loaded.FindAny(1, 999))  // groupings containing either
}
```

### Persistence

`Groupings` containers can be serialised either as a byte slice or as a
`.bin` file on disk.

| Function                        | Description                                         |
|---------------------------------|-----------------------------------------------------|
| `(*Groupings).Encode()`         | Return the binary payload as `[]byte`.              |
| `gg.Decode([]byte)`             | Parse a byte slice back into a `*Groupings`.        |
| `(*Groupings).SaveFile(path)`   | Atomically write the payload to `path` (via temp + rename). |
| `gg.LoadFile(path)`             | Read a `.bin` file from `path` and decode it.       |

`SaveFile` writes to a temporary sibling file and renames it into place,
so readers never observe a half-written payload.

### Searching

The container exposes reverse-lookup helpers to find every grouping that
references a particular entity (or a set of entities):

| Method                             | Returns                                                  |
|------------------------------------|----------------------------------------------------------|
| `Groupings.Find(m)`                | All groupings containing entity `m`.                     |
| `Groupings.FindAll(m1, m2, …)`     | All groupings containing **every** listed entity (⊇).    |
| `Groupings.FindAny(m1, m2, …)`     | All groupings containing **at least one** listed entity. |
| `Groupings.FindNames(m)`           | Only the names of groupings containing entity `m`.       |

All search methods scan bitmaps linearly at bitwise speed and return
attached groupings, so mutations on the results propagate to the
container.

## Binary format

A `Groupings` payload is a single contiguous byte slice made of two
blocks: a variable-length **header** followed by the fixed-stride
**data** block. All multi-byte integers are **little-endian**.

```
┌────────────────────── Headers ──────────────────────┬──── Data ────┐
│ magic │ entityCount │ groupingCount │ names table   │ bitmaps …    │
│ 4 B   │ 4 B  uint32 │ 4 B    uint32 │ variable      │ G × S bytes  │
└─────────────────────────────────────────────────────┴──────────────┘
```

### Headers

| Offset | Size | Field           | Description                                 |
|--------|------|-----------------|---------------------------------------------|
| `0`    | 4    | `magic`         | Magic number, always the ASCII bytes `GGGG` |
| `4`    | 4    | `entityCount`   | `N` — size of the universe (`uint32` LE)    |
| `8`    | 4    | `groupingCount` | `G` — number of groupings (`uint32` LE)     |
| `12`   | …    | `names`         | Names table, `G` entries, see below         |

Each entry in the names table is a length-prefixed UTF-8 string:

| Size         | Field     | Description                             |
|--------------|-----------|-----------------------------------------|
| 2 B `uint16` | `nameLen` | Byte length of the grouping name (LE)   |
| `nameLen` B  | `name`    | UTF-8 encoded grouping name             |

Names appear in insertion order; the i-th name corresponds to the i-th
bitmap in the data block. Names are unique within a container.

### Data

The data block is the concatenation of `G` bitmaps, each of exactly

```
S = ceil(entityCount / 8)
```

bytes. The bitmap for the i-th grouping lives at byte offset `i * S`.
Within a bitmap, entity `k` is encoded in bit `k & 7` of byte `k >> 3`
(little-endian bit order).

Unused trailing bits in the last byte (when `entityCount` is not a
multiple of 8) are always zero and are ignored by every operation.

### Example

Universe of 10 entities, two groupings `"a" = {1, 2, 3}` and
`"b" = {2, 3, 9}`; `S = ceil(10/8) = 2`:

```
Headers:
  47 47 47 47                       "GGGG"
  0A 00 00 00                       entityCount = 10
  02 00 00 00                       groupingCount = 2
  01 00 61                          len=1, "a"
  01 00 62                          len=1, "b"

Data (4 bytes, two bitmaps of 2 bytes):
  0E 00                             bitmap "a": bits 1,2,3 set
  0C 02                             bitmap "b": bits 2,3,9 set
```

## Package layout

```
gg/
├── ingestion.go          // Groupings constructor, Add, Encode/Decode, LoadFile
├── extraction.go         // Names, Get, All, SaveFile, Find / FindAll / FindAny / FindNames
├── comparison.go         // Re-exports ErrUniverseMismatch
└── internal/
    ├── binary/           // Low-level byte layout
    │   ├── headers.go    //   magic, entityCount, groupingCount, names table
    │   └── data.go       //   packed bitmaps + bitwise primitives
    ├── groupings/        // Internal container binding Headers + Data
    │   └── groupings.go
    └── grouping/         // Grouping value type and all its methods
        └── grouping.go   //   Insert/Remove/Contains/Members/Union/Intersection/…
```

The public `gg.Grouping` type is a Go **type alias** for
`internal/grouping.Grouping`, so every method documented above is defined
in `internal/grouping` but reachable directly through the top-level API.
The `gg` package itself only exposes the `Groupings` container and a
handful of package-level functions (`New`, `Decode`, `LoadFile`,
`MagicNumber`, `ErrUniverseMismatch`).

## Benchmarks

Micro-benchmarks for every public operation live in `bench_test.go` and
can be executed with:

```sh
go test -run=^$ -bench=. -benchmem ./...
```

## Examples

Runnable end-to-end examples illustrating real-world use cases live under
[`.examples/`](./.examples/):

| Example                                  | What it demonstrates                                              |
|------------------------------------------|-------------------------------------------------------------------|
| [`access_control`](./.examples/access_control/) | Role-based access control (intersection, membership, invariants)  |
| [`inverted_index`](./.examples/inverted_index/) | Boolean keyword search over documents via bitmap set operations   |
| [`feature_flags`](./.examples/feature_flags/)   | Feature-flag cohorts / A/B audiences, persisted to a `.bin` file  |

Run any of them directly:

```sh
go run ./.examples/access_control
go run ./.examples/inverted_index
go run ./.examples/feature_flags
```

The directory is prefixed with `.` so Go's `./...` wildcard skips it and
examples never interfere with `go build` / `go test`.

## License

This project is released under the [MIT License](LICENSE).

```
Copyright (c) 2026 rschoonheim
```

You are free to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the software. See the [LICENSE](LICENSE) file for the
full text.
