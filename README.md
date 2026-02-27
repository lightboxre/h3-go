# h3-go

[![CI](https://github.com/EDRInc/h3-go/actions/workflows/ci.yml/badge.svg)](https://github.com/EDRInc/h3-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/EDRInc/h3-go.svg)](https://pkg.go.dev/github.com/EDRInc/h3-go)
[![Go 1.22+](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

A pure Go implementation of Uber's [H3 geospatial indexing system](https://h3geo.org/). Zero CGO. Zero external dependencies.

H3 is a hierarchical hexagonal grid system that partitions the Earth into cells at 16 resolutions (0–15). It is used for spatial indexing, proximity analysis, geographic aggregation, and routing.

## Why this library?

The official Go binding ([uber/h3-go](https://github.com/uber/h3-go)) requires CGO and a compiled C library. This library is a complete native Go port — no C toolchain, no `CGO_ENABLED=1`, no shared library installation.

| | This library | uber/h3-go |
|---|---|---|
| CGO required | No | Yes |
| External dependencies | None | H3 C library |
| API compatibility | uber/h3-go v4 | — |
| Correctness validated by | 50-test CGO oracle | — |

## Installation

```bash
go get github.com/EDRInc/h3-go
```

Requires Go 1.22 or later.

## Quick start

```go
import h3 "github.com/EDRInc/h3-go"

// Convert a lat/lng coordinate to an H3 cell at resolution 9
cell := h3.LatLngToCell(37.7796, -122.4193, 9)
fmt.Println(h3.CellToString(cell)) // 89283082877ffff

// Get the cell's center coordinate
center := h3.CellToLatLng(cell)
fmt.Printf("%.4f, %.4f\n", center.Lat, center.Lng)

// Walk up the hierarchy
parent := h3.CellToParent(cell, 5)

// Find all cells within 2 hops
neighbors, err := h3.GridDisk(cell, 2)

// Get the hexagonal boundary polygon
boundary := h3.CellToBoundary(cell)
```

## API reference

The API mirrors [uber/h3-go v4](https://pkg.go.dev/github.com/uber/h3-go/v4). Drop-in replacement for most use cases.

### Types

```go
type Cell         uint64         // An H3 cell index
type DirectedEdge uint64         // A directed edge between two cells
type Vertex       uint64         // A vertex of an H3 cell
type LatLng       struct { Lat, Lng float64 }
type CellBoundary []LatLng       // Polygon vertices of a cell
type GeoLoop      []LatLng       // A closed ring of coordinates
type GeoPolygon   struct {
    GeoLoop GeoLoop              // Outer boundary
    Holes   []GeoLoop            // Interior holes
}
type GeoMultiPolygon []GeoPolygon
```

### Indexing

```go
// Convert between coordinates and cells
LatLngToCell(lat, lng float64, res int) Cell
CellToLatLng(c Cell) LatLng
CellToBoundary(c Cell) CellBoundary
```

### Cell hierarchy

```go
CellToParent(c Cell, parentRes int) Cell
CellToChildren(c Cell, childRes int) []Cell
CellToChildrenSize(c Cell, childRes int) int64
CellToCenterChild(c Cell, childRes int) Cell
CellToChildPos(c Cell, parentRes int) int64
ChildPosToCell(childPos int64, parent Cell, childRes int) Cell
```

### Cell inspection

```go
GetResolution(c Cell) int
GetBaseCellNumber(c Cell) int
IsValidCell(c Cell) bool
IsPentagon(c Cell) bool
IsResClassIII(c Cell) bool
```

### String conversion

```go
CellToString(c Cell) string
StringToCell(s string) (Cell, error)
```

### Enumeration

```go
GetNumCells(res int) int64
GetRes0Cells() []Cell
GetPentagonCells(res int) []Cell
```

### Grid traversal

```go
GridDisk(origin Cell, k int) ([]Cell, error)
GridDiskDistances(origin Cell, k int) ([][]Cell, error)
GridDiskUnsafe(origin Cell, k int) ([]Cell, error)
GridDiskDistancesUnsafe(origin Cell, k int) ([][]Cell, error)
GridRingUnsafe(origin Cell, k int) ([]Cell, error)
GridPathCells(src, dst Cell) ([]Cell, error)
GridDistance(src, dst Cell) (int64, error)
```

### Compaction

```go
CompactCells(cells []Cell) ([]Cell, error)
UncompactCells(cells []Cell, res int) ([]Cell, error)
UncompactCellsSize(cells []Cell, res int) (int64, error)
```

### Region

```go
// Fill a polygon with cells at the given resolution
PolygonToCells(polygon GeoPolygon, res int) ([]Cell, error)

// Reconstruct polygon boundaries from a set of cells
CellsToLinkedMultiPolygon(cells []Cell) GeoMultiPolygon
```

### Directed edges

```go
AreNeighborCells(origin, destination Cell) (bool, error)
CellsToDirectedEdge(origin, destination Cell) (DirectedEdge, error)
IsValidDirectedEdge(e DirectedEdge) bool
GetDirectedEdgeOrigin(e DirectedEdge) Cell
GetDirectedEdgeDestination(e DirectedEdge) Cell
DirectedEdgeToCells(e DirectedEdge) [2]Cell
OriginToDirectedEdges(origin Cell) []DirectedEdge
DirectedEdgeToBoundary(e DirectedEdge) CellBoundary
```

### Vertices

```go
CellToVertex(c Cell, vertexNum int) Vertex
CellToVertexes(c Cell) []Vertex
VertexToLatLng(v Vertex) LatLng
IsValidVertex(v Vertex) bool
```

### Metrics

```go
GreatCircleDistanceRads(a, b LatLng) float64
GreatCircleDistanceKm(a, b LatLng) float64
GreatCircleDistanceM(a, b LatLng) float64

CellAreaRads2(c Cell) float64
CellAreaKm2(c Cell) float64
CellAreaM2(c Cell) float64

EdgeLengthRads(res int) float64    // Average edge length at resolution
EdgeLengthKm(res int) float64
EdgeLengthM(res int) float64

ExactEdgeLengthRads(e DirectedEdge) float64   // Geodesic length of a specific edge
ExactEdgeLengthKm(e DirectedEdge) float64
ExactEdgeLengthM(e DirectedEdge) float64
```

## Repository layout

```
h3-go/
├── h3.go                  # Public API — indexing, hierarchy, grid, region, metrics
├── h3_edges.go            # Directed edge functions
├── h3_vertices.go         # Vertex functions
├── internal/
│   ├── algos/             # Grid traversal algorithms (disk, ring, path, compact)
│   ├── bbox/              # Bounding box helpers for polygon pre-filtering
│   ├── constants/         # H3 constants, digit tables, mode values
│   ├── coordijk/          # IJK hexagonal coordinate arithmetic and Vec2d/Vec3d
│   ├── faceijk/           # Face-IJK ↔ geo projection, cell boundary computation
│   ├── h3index/           # H3Index bit manipulation, validation, base cell tables
│   ├── math/              # Metrics: great circle distance, cell area, edge length
│   ├── polygon/           # Polygon fill and cell-to-multipolygon reconstruction
│   └── testutil/          # Fixture file parser for table-driven tests
├── testdata/              # Golden fixture files (CSV-like, verified against C library)
│   ├── cell_to_parent.txt
│   ├── grid_distance.txt
│   ├── latlng_to_cell.txt
│   └── latlng_to_cell_res5.txt
├── cgotest/               # Separate module: CGO oracle + benchmarks (requires H3 C library)
│   ├── compare_test.go    # 50 oracle tests comparing native vs CGO output
│   └── bench_test.go      # Paired _Native / _CGO benchmarks for every major function
└── example/
    └── demo.go
```

## Contributing

Contributions are welcome. Please read this section before opening a pull request.

### Prerequisites

- Go 1.22 or later
- For CGO oracle tests only: CMake, a C compiler, and the [H3 C library](https://github.com/uber/h3) installed

### Running tests

```bash
# All pure Go tests (no C dependency required)
go test ./...

# With race detector
go test -race ./...

# A single test function
go test -run TestGridDisk -v .

# CGO oracle (requires H3 C library installed — see below)
cd cgotest && CGO_ENABLED=1 go test -v -count=1 ./...
```

### Installing the H3 C library (for oracle tests only)

The CGO oracle is optional for local development but runs in CI on every pull request.

```bash
git clone --depth 1 https://github.com/uber/h3.git /tmp/h3
cd /tmp/h3
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build --parallel $(nproc)
sudo cmake --install build
sudo ldconfig          # Linux only
```

### Running benchmarks

```bash
# Paired native/CGO benchmarks (requires H3 C library)
cd cgotest && CGO_ENABLED=1 go test -bench=. -benchmem ./...

# Benchmark a specific function
cd cgotest && CGO_ENABLED=1 go test -bench=BenchmarkLatLngToCell -benchmem ./...
```

See [`cgotest/BENCHMARKS.md`](cgotest/BENCHMARKS.md) for methodology and results.

### Linting

```bash
golangci-lint run
```

Configuration is in `.golangci.yml`. The enabled linters are `errcheck`, `govet`, `staticcheck`, `unused`, `gofmt`, `goimports`, `revive`, and `misspell`.

### Pull request checklist

- [ ] `go test ./...` passes
- [ ] `go vet ./...` is clean
- [ ] New code follows the patterns in the package it touches (see below)
- [ ] If touching a correctness-sensitive path, add or update a test in the appropriate `_test.go` file
- [ ] If adding a new exported function, add a paired benchmark in `cgotest/bench_test.go` and an oracle comparison in `cgotest/compare_test.go`

### Correctness standard

The ground truth for this library is the [H3 C reference implementation](https://github.com/uber/h3). When in doubt, the C code wins. Key invariants:

- **Bit layout**: H3Index is a 64-bit integer. Bits 59–62 = mode, 58–56 = reserved/vertex number, 55–52 = resolution, 51–45 = base cell, lower bits = digit sequence. See `internal/h3index/` for accessors.
- **Floating-point**: `float64` throughout. Match C `double` arithmetic. Epsilon = `1e-16`.
- **Pentagon base cells**: `{4, 14, 24, 38, 49, 58, 63, 72, 83, 97, 107, 117}` — 12 cells at every resolution.
- **Class II vs Class III**: Even resolutions are Class II; odd are Class III. Class III applies an additional `M_AP7_ROT_RADS` rotation in face-IJK conversions.
- **Face boundary crossings**: `FaceIJKToGeoBoundary` must add interpolated vertices when a cell boundary crosses an icosahedral face edge.

### Package responsibilities

Each internal package has a single owner to prevent circular imports:

| Package | Responsibility |
|---|---|
| `internal/constants` | Compile-time constants and lookup tables. No logic. |
| `internal/h3index` | H3Index bit manipulation, validation, base cell neighbour tables. |
| `internal/coordijk` | IJK hexagonal arithmetic, Vec2d/Vec3d, aperture-7 transforms. |
| `internal/bbox` | Bounding box construction and containment tests. |
| `internal/faceijk` | Gnomonic projection between geo coordinates and face-IJK. Core of `LatLngToCell` / `CellToLatLng` / `CellToBoundary`. |
| `internal/algos` | Grid traversal: disk, ring, path, compact/uncompact. |
| `internal/polygon` | Polygon fill (`PolygonToCells`) and outline reconstruction (`CellsToLinkedMultiPolygon`). |
| `internal/math` | Metrics: great circle distance, cell area, average and exact edge lengths. |
| `internal/testutil` | Fixture file parser. Test code only. |

### Adding test data

Golden fixture files live in `testdata/`. The format is whitespace-separated values with `#` line comments. Parsers are in `internal/testutil/fixtures.go`. All values must be verified against the C reference library or the CGO oracle before committing.

### Known implementation gaps

These are tracked but not yet fixed. PRs welcome:

- **Class III cell boundaries**: Cells near icosahedral face edges should return 7 boundary vertices. Our implementation returns 6 for a subset of such cells (the face-boundary intersection vertex is not always inserted).
- **Class III pentagon edge boundaries**: Directed edges of odd-resolution pentagon cells should have 3 boundary vertices. Our implementation returns 2 (the extra face-crossing vertex is omitted).

## License

Apache 2.0 — see [LICENSE](LICENSE).

H3 is developed by Uber Technologies and released under the Apache 2.0 license. This library is an independent implementation and is not affiliated with Uber.
