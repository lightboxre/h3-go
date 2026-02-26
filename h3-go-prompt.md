# Task: Native Go Re-implementation of Uber's H3 Library

## Overview

You are tasked with producing a **complete, production-quality, pure-Go re-implementation** of Uber's H3 geospatial indexing library. The goal is a zero-CGO, zero-external-dependency Go module whose public API is **100% compatible** with the official Go binding at:

> https://pkg.go.dev/github.com/uber/h3-go/v4

and whose output is **bit-for-bit identical** to the original C library at:

> https://github.com/uber/h3

---

## Reference Material

| Resource | Purpose |
|---|---|
| `https://github.com/uber/h3` | Authoritative C source and golden test fixtures |
| `https://github.com/uber/h3-go` | Official CGO Go binding — use as API contract and test suite |
| `https://github.com/ZipRecruiter/h3-go` | Human-written partial Go port — use as a starting reference, not a final answer |
| `https://pkg.go.dev/github.com/uber/h3-go/v4` | Full public API documentation you must match exactly |

---

## Deliverables

### 1. Go Module

- Module path: `github.com/<your-org>/h3-go` (or as instructed)
- Go version: `1.22+`
- **No CGO**. No `import "C"`. No `syscall` wrappers around the C library.
- No third-party runtime dependencies (only `stdlib`). Test/bench dependencies are fine.

### 2. Public API

Implement every exported symbol listed on the pkg.go.dev page above, including but not limited to:

**Core types**
- `Cell` (H3Index), `DirectedEdge`, `Vertex`, `LatLng`, `CellBoundary`, `GeoLoop`, `GeoPolygon`, `GeoMultiPolygon`, `CoordIJ`

**Indexing**
- `LatLngToCell`, `CellToLatLng`, `CellToBoundary`

**Grid traversal**
- `GridDisk`, `GridDiskDistances`, `GridDiskUnsafe`, `GridDiskDistancesUnsafe`, `GridRingUnsafe`, `GridPathCells`, `GridDistance`

**Hierarchical operations**
- `CellToParent`, `CellToChildren`, `CellToCenterChild`, `CellToChildrenSize`, `CellToChildPos`, `ChildPosToCell`, `CompactCells`, `UncompactCells`, `UncompactCellsSize`

**Region operations**
- `PolygonToCells`, `CellsToLinkedMultiPolygon`

**Directed edges**
- `AreNeighborCells`, `CellsToDirectedEdge`, `IsValidDirectedEdge`, `GetDirectedEdgeOrigin`, `GetDirectedEdgeDestination`, `DirectedEdgeToCells`, `OriginToDirectedEdges`, `DirectedEdgeToBoundary`

**Vertex functions**
- `CellToVertex`, `CellToVertexes`, `VertexToLatLng`, `IsValidVertex`

**Miscellaneous**
- `IsValidCell`, `IsResClassIII`, `IsPentagon`, `GetResolution`, `GetBaseCellNumber`, `StringToH3`, `H3ToString`, `H3IndexesAreNeighbors`, `GetNumCells`, `GetRes0Cells`, `GetPentagonCells`, `GreatCircleDistanceRads`, `GreatCircleDistanceKm`, `GreatCircleDistanceM`, `CellAreaRads2`, `CellAreaKm2`, `CellAreaM2`, `EdgeLengthRads`, `EdgeLengthKm`, `EdgeLengthM`, `ExactEdgeLengthRads`, `ExactEdgeLengthKm`, `ExactEdgeLengthM`

The function signatures, return types, and error semantics **must match** the official binding exactly.

---

## Testing Requirements

### Correctness Tests (`CGO_ENABLED=1` required)

Create a test file (e.g. `h3_cgo_test.go`) guarded with a build tag:

```go
//go:build cgo
// +build cgo
```

This file must:
1. Import the **official CGO-based binding** (`github.com/uber/h3-go/v3` or the C binding via CGO directly).
2. For **every public function**, generate a representative set of inputs (drawn from the golden fixtures in `https://github.com/uber/h3` under `tests/inputfiles/` and `tests/`) and assert that the native Go output equals the C output exactly.
3. Cover all 16 H3 resolutions (0–15).
4. Cover known edge cases: pentagons, poles, antimeridian-crossing polygons, invalid inputs.
5. Use `t.Fatal` / `t.Errorf` with descriptive messages showing both expected (C) and actual (Go) values.

Example pattern:
```go
func TestLatLngToCell_MatchesC(t *testing.T) {
    cases := goldenLatLngToCellCases() // loaded from C golden fixtures
    for _, tc := range cases {
        got := h3native.LatLngToCell(tc.lat, tc.lng, tc.res)
        want := h3c.LatLngToCell(tc.lat, tc.lng, tc.res) // CGO call
        if got != want {
            t.Errorf("LatLngToCell(%v,%v,%d) = %v, want %v", tc.lat, tc.lng, tc.res, got, want)
        }
    }
}
```

### Unit Tests (`CGO_ENABLED=0` compatible)

Standard table-driven tests for every function using hardcoded known-good values (taken from the golden fixtures). These must pass with `CGO_ENABLED=0`.

---

## Benchmarking Requirements

Create `h3_bench_test.go` with a benchmark for **every public function**:

```go
func BenchmarkLatLngToCell(b *testing.B) { ... }
func BenchmarkCellToLatLng(b *testing.B) { ... }
// ... one per public function
```

Each benchmark must:
1. Use `b.ReportAllocs()`.
2. Run with a representative, non-trivial input (not a trivial/zero value).
3. Be structured so it can be compared directly against an equivalent `BenchmarkLatLngToCell_C` that calls the CGO binding (in a `//go:build cgo` file).

**Performance target:** The native Go implementation must be within **1% wall-clock time** and **1% heap allocations** of the CGO C implementation as measured by `go test -bench=. -benchmem -count=10` averaged across runs.

Document the benchmark methodology in `BENCHMARKS.md`.

---

## Implementation Guidance

### Algorithm Fidelity

- Port the C algorithms from `https://github.com/uber/h3` directly — do not guess or approximate.
- Pay close attention to:
  - The face-IJK coordinate system and its transformations
  - Pentagon distortion handling in grid traversal
  - The specific bit layout of H3Index (mode bits, resolution nibble, base cell, digit sequence)
  - Floating-point operations — match C `double` semantics using Go `float64` throughout
  - Exact integer overflow behavior where C uses `uint64`

### H3Index Bit Layout (critical)

```
Bit 63     : reserved (0)
Bits 59-62 : mode (1 = hexagon, 2 = directed edge, 3 = vertex, ...)
Bits 52-58 : reserved
Bits 56-58 : resolution (0–15)
Bits 49-55 : base cell (0–121)
Bits 0-48  : digit sequence (3 bits per resolution level × 15 levels)
```

Verify this against the C source `h3api.h` and `h3Index.c`.

### File Structure (suggested)

```
h3-go/
├── go.mod
├── go.sum
├── h3.go                  # public API surface (thin wrappers)
├── internal/
│   ├── h3index/           # H3Index bit manipulation
│   ├── faceijk/           # face-IJK coordinate math
│   ├── algos/             # grid traversal, compaction, polygon fill
│   ├── bbox/              # bounding box utilities
│   ├── polygon/           # GeoPolygon / containment
│   ├── math/              # great circle, cell area, edge length
│   └── constants/         # resolution tables, pentagon list, etc.
├── h3_test.go             # CGO=0 unit tests
├── h3_cgo_test.go         # CGO=1 comparison tests (build tag: cgo)
└── h3_bench_test.go       # benchmarks
```

### Golden Fixture Loading

The C repo's test inputs live under `tests/inputfiles/*.txt` and follow a consistent CSV-like format. Write a helper package `internal/testutil` that can parse these files and return typed test cases. Check these files into the repo under `testdata/`.

---

## Quality Bar

- `go vet ./...` — zero warnings
- `staticcheck ./...` — zero warnings  
- `golangci-lint run` — zero issues (configure a `.golangci.yml`)
- 100% of C golden tests passing
- 100% of official `uber/h3-go` test suite passing (with CGO comparison)
- All benchmarks within performance targets
- Full `godoc` comments on every exported symbol

---

## Execution Strategy

Work through the implementation in this order to enable early test feedback:

1. **H3Index type + bit manipulation** — implement and test `IsValidCell`, `GetResolution`, `GetBaseCellNumber`, `IsPentagon`, `IsResClassIII` first. These are pure bit operations and validate your index layout.
2. **Coordinate conversions** — `LatLngToCell` / `CellToLatLng` / `CellToBoundary`. These exercise the core face-IJK math.
3. **Grid traversal** — `GridDisk`, `GridDiskDistances`, `GridDistance`, `GridPathCells`.
4. **Hierarchy** — `CellToParent`, `CellToChildren`, `CompactCells`, `UncompactCells`.
5. **Region** — `PolygonToCells`, `CellsToLinkedMultiPolygon`.
6. **Edges & Vertices** — directed edge and vertex functions.
7. **Metrics** — area, distance, edge length functions.
8. **Final pass** — run full golden test suite, fix any discrepancies, run benchmarks.

For each step, write the CGO comparison test first (red), then implement until green.

---

## Important Constraints

- **Do not** wrap the C library via CGO in the production code path — only in test files.
- **Do not** use `unsafe` unless absolutely necessary for performance-critical inner loops, and document every usage.
- **Do not** silently swallow errors — match the error return semantics of the official Go binding exactly.
- **Do not** introduce goroutines or concurrency into the core algorithms; H3 functions are stateless and the caller manages parallelism.
- If you find a discrepancy between the ZipRecruiter port and the C source, **always trust the C source**.
