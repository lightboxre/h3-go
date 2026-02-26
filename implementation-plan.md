# Implementation Plan: H3 Native Go Library

## Context

Re-implement Uber's H3 geospatial indexing library as a zero-CGO, zero-external-dependency pure-Go module. The public API must match `github.com/uber/h3-go/v4` exactly. Output must be bit-for-bit identical to the C library. The ZipRecruiter partial port (~30–40% complete) provides a useful starting reference for the coordinate system and grid disk algorithm, but the full H3 API is not yet ported.

---

## Critical Files to Create

```
h3-go/
├── go.mod                          # module github.com/<org>/h3-go
├── go.sum
├── h3.go                           # thin public API wrappers
├── h3_test.go                      # CGO=0 unit tests (table-driven, golden values)
├── h3_cgo_test.go                  # //go:build cgo comparison tests
├── h3_bench_test.go                # benchmarks (one per public function)
├── BENCHMARKS.md                   # benchmark methodology doc
├── .golangci.yml                   # lint config
├── testdata/                       # C golden fixtures (checked in)
│   └── *.txt                       # from uber/h3/tests/inputfiles/
└── internal/
    ├── constants/      constants.go
    ├── h3index/        h3index.go, base_cells.go
    ├── coordijk/       coordijk.go, vec2d.go, vec3d.go
    ├── faceijk/        faceijk.go
    ├── bbox/           bbox.go
    ├── polygon/        polygon.go, linked_geo.go
    ├── algos/          algos.go (grid disk, path, compact)
    ├── math/           great_circle.go, cell_area.go, edge_length.go
    └── testutil/       fixtures.go (golden fixture parser)
```

---

## Implementation Phases & Sub-agent Strategy

### Phase 1 — Foundation *(1 agent, sequential)*

**Agent: Foundation Setup**

- Create `go.mod` (module path, Go 1.22+)
- Scaffold all directories
- Write `internal/constants/constants.go`:
  - `MAX_H3_RES = 15`, `NUM_BASE_CELLS = 122`, `NUM_ICOSA_FACES = 20`, `NUM_PENTAGONS = 12`
  - `M_AP7_ROT_RADS`, `M_SIN60`, `M_SQRT3_2`, `RES0_U_GNOMONIC`, `EPSILON = 1e-16`
  - `maxDimByCIIres[16]`, `unitScaleByCIIres[16]` lookup tables
- Transcribe all 12 pentagon base cell indices
- Initialize `go vet`, `staticcheck`, `.golangci.yml`

**Deliverable:** Compilable module skeleton with constants.

---

### Phase 2 — Parallel: Low-level Math & Bit Ops *(3 agents in parallel)*

#### Agent A: `internal/h3index` — H3Index Bit Manipulation

Ports: `h3Index.c`, `baseCells.c`

**Bit layout to implement:**

| Bits | Field |
|------|-------|
| 63 | Reserved (0) |
| 59–62 | Mode (1=cell, 2=directed edge, 3=vertex) |
| 56–58 | Reserved |
| 52–55 | Resolution (0–15) |
| 45–51 | Base cell (0–121) |
| 0–44 | Digit sequence (3 bits × 15 levels) |

**Functions:**
- `H3Index` type (`uint64`)
- `GetMode`, `SetMode`, `GetResolution`, `SetResolution`, `GetBaseCell`, `SetBaseCell`
- `GetIndexDigit`, `SetIndexDigit`
- `IsValid`, `IsPentagon`, `IsResClassIII`, `IsBaseCellPentagon`
- `H3ToString`, `StringToH3`
- `GetBaseCellNumber`, `GetNumCells`, `GetRes0Cells`, `GetPentagonCells`

**Large data:** `baseCellNeighbors[122][7]`, `baseCellNeighbor60CCWRots[122][7]` (transcribed from C source).

**Tests:** Table-driven tests with hardcoded known-good values for every bit operation.

---

#### Agent B: `internal/coordijk` — Hexagonal Coordinate Math

Ports: `coordijk.c`, `vec2d.c`, `vec3d.c`

**Data structures:**
- `CoordIJK {I, J, K int}` — constraint: i+j+k always reduced to canonical form
- `Vec2d {X, Y float64}`
- `Vec3d {X, Y, Z float64}`

**Functions:**
- `IJKNormalize`, `IJKAdd`, `IJKSub`, `IJKScale`, `IJKDistance`
- `Rotate60CW`, `Rotate60CCW`
- `UpAp7`, `UpAp7r`, `DownAp7`, `DownAp7r` (aperture-7 parent/child navigation)
- `IJKToIJ`, `IJToIJK`
- `Vec2d`/`Vec3d` arithmetic, `PointSquareDist`, unit vector ops

---

#### Agent C: `internal/bbox` — Bounding Box

Ports: `bbox.c`

- `BBox {North, South, East, West float64}`
- `BBoxContains`, `BBoxOverlap`, `BBoxCenter`, `BBoxFromGeoLoop`
- Antimeridian-aware longitude normalization

---

### Phase 3 — Core Coordinate Conversion *(1 agent, after Phase 2)*

**Agent: `internal/faceijk` — Face-IJK Transformations**

Ports: `faceijk.c`, `latLng.c`  
Depends on: `coordijk`, `constants`

**Critical tables to transcribe:**
- `faceNeighbors[20][3]` — face boundary crossing transformations
- `adjacentFaceDir[20][20]` — neighboring face relationship matrix
- `maxDimByCIIres[16]`, `unitScaleByCIIres[16]` (already in constants)
- `faceCenterGeo[20]` — icosahedron face center lat/lngs
- `faceCenterPoint[20]` — 3D Cartesian face centers

**Functions:**
- `GeoToFaceIJK(lat, lng float64, res int) FaceIJK`
- `FaceIJKToGeo(fijk FaceIJK, res int) (lat, lng float64)`
- `FaceIJKToH3(fijk FaceIJK, res int) H3Index`
- `H3ToFaceIJK(h H3Index) FaceIJK`
- `FaceIJKToGeoBoundary(fijk FaceIJK, res int, isPentagon bool) CellBoundary`
- Internal: gnomonic projection, Class II/III rotation handling

> **Note:** Match double semantics exactly — use `math.Tan`, `math.Atan`, `math.Sqrt` throughout. Pentagon distortion at face boundaries requires special-casing (follow C exactly).

---

### Phase 4 — Public Indexing API + Grid Traversal *(2 agents in parallel, after Phase 3)*

#### Agent D: `h3.go` Indexing + Hierarchy

Ports: `h3Index.c` higher-level functions

**Functions:**
- `LatLngToCell(lat, lng float64, res int) Cell`
- `CellToLatLng(c Cell) LatLng`
- `CellToBoundary(c Cell) CellBoundary`
- `CellToParent(c Cell, res int) Cell`
- `CellToChildren(c Cell, res int) []Cell`
- `CellToCenterChild(c Cell, res int) Cell`
- `CellToChildrenSize(c Cell, res int) int`
- `CellToChildPos`, `ChildPosToCell`
- `GetResolution`, `GetBaseCellNumber`, `IsValidCell`, `IsPentagon`, `IsResClassIII`
- `GetNumCells`, `GetRes0Cells`, `GetPentagonCells`

---

#### Agent E: `internal/algos` — Grid Traversal

Ports: `algos.c`

**Key lookup tables:**
- `DIRECTIONS[6]` — CCW traversal directions
- `NEW_DIGIT_II[7][7]`, `NEW_DIGIT_III[7][7]` — digit transition matrices
- `NEW_ADJUSTMENT_II[7][7]`, `NEW_ADJUSTMENT_III[7][7]` — adjustment vectors

**Functions:**
- `GridDisk(origin H3Index, k int) []H3Index`
- `GridDiskDistances(origin H3Index, k int) [][]H3Index`
- `GridDiskUnsafe`, `GridDiskDistancesUnsafe` — fast path (fails on pentagons)
- `GridRingUnsafe` — hollow ring at exactly k
- `GridPathCells(src, dst H3Index) []H3Index`
- `GridDistance(src, dst H3Index) int`
- `CompactCells(cells []H3Index) []H3Index`
- `UncompactCells(cells []H3Index, res int) []H3Index`, `UncompactCellsSize`

**Pentagon strategy:** Attempt unsafe fast-path first, fall back to safe recursive hash-based approach on `E_PENTAGON` error.

---

### Phase 5 — Region, Edges, Vertices, Metrics *(4 agents in parallel, after Phase 4)*

#### Agent F: `internal/polygon` — Region Operations

Ports: `polygon.c`, `linkedGeo.c`  
Depends on: `bbox`, `faceijk`, `algos`

- `PolygonToCells(polygon GeoPolygon, res int) []Cell`
- `CellsToLinkedMultiPolygon(cells []Cell) GeoMultiPolygon`
- Internal: `PointInGeoLoop`, `PointInGeoPolygon`, vertex graph construction
- Antimeridian handling, bounding box pre-filtering

---

#### Agent G: Directed Edges

Ports: `directedEdge.c`  
Depends on: `h3index`, `faceijk`

- `AreNeighborCells`, `CellsToDirectedEdge`
- `IsValidDirectedEdge`, `GetDirectedEdgeOrigin`, `GetDirectedEdgeDestination`
- `DirectedEdgeToCells`, `OriginToDirectedEdges`
- `DirectedEdgeToBoundary`
- Pentagon constraint: only 5 valid directions (no K-axis)

---

#### Agent H: Vertex Functions

Ports: `vertex.c`  
Depends on: `h3index`, `faceijk`, directed edges

- `CellToVertex(c Cell, vertexNum int) Vertex`
- `CellToVertexes(c Cell) []Vertex`
- `VertexToLatLng(v Vertex) LatLng`
- `IsValidVertex`
- Lookup tables: `directionToVertexNumHex`, `pentagonDirectionFaces`
- Vertex ownership: lowest-cell-number of 3 adjacent cells owns the vertex

---

#### Agent I: `internal/math` — Metrics

Depends on: `faceijk`, `h3index`

- `GreatCircleDistanceRads`, `GreatCircleDistanceKm`, `GreatCircleDistanceM`
- `CellAreaRads2`, `CellAreaKm2`, `CellAreaM2`
- `EdgeLengthRads`, `EdgeLengthKm`, `EdgeLengthM`
- `ExactEdgeLengthRads`, `ExactEdgeLengthKm`, `ExactEdgeLengthM`
- WGS84 authalic sphere radius constants

---

### Phase 6 — Testing Infrastructure *(2 agents in parallel, after Phase 5)*

#### Agent J: `internal/testutil` + `testdata/` + CGO=0 Tests

- Write golden fixture parser (`internal/testutil/fixtures.go`) — CSV-like format from `tests/inputfiles/`
- Download and check in golden fixture files from `uber/h3/tests/inputfiles/` to `testdata/`
- Write `h3_test.go`: table-driven tests for all 60+ public functions
- Coverage: all 16 resolutions (0–15), 12 pentagons, poles, antimeridian, invalid inputs

---

#### Agent K: CGO Comparison Tests + Benchmarks

- Write `h3_cgo_test.go` (build tag: `//go:build cgo`):
  - Import `github.com/uber/h3-go/v3` as CGO oracle
  - For each public function: generate inputs from golden fixtures, assert native Go == C output
  - Pattern: `got := h3native.Fn(args); want := h3cgo.Fn(args); if got != want { t.Errorf(...) }`
- Write `h3_bench_test.go`:
  - One `BenchmarkXxx` per public function
  - `b.ReportAllocs()` on each
  - Parallel `BenchmarkXxx_C` in `//go:build cgo` file for comparison
- Write `BENCHMARKS.md`: methodology, how to run, target: ≤1% wall-clock and ≤1% heap vs CGO

---

## Type Definitions *(public surface in `h3.go`)*

```go
type Cell             uint64
type DirectedEdge     uint64
type Vertex           uint64
type LatLng           struct{ Lat, Lng float64 }
type CellBoundary     []LatLng          // max 10 points (6 hex + 5 pent + crossing)
type GeoLoop          []LatLng
type GeoPolygon       struct{ GeoLoop GeoLoop; Holes []GeoLoop }
type GeoMultiPolygon  []GeoPolygon
type CoordIJ          struct{ I, J int }
```

---

## Key Correctness Constraints

1. **H3Index bit layout** — verify offset constants against C `h3Index.h` before any encoding
2. **Pentagon base cells** — indices `{4, 14, 24, 38, 49, 58, 63, 72, 83, 97, 107, 117}`; polar pentagons at `4` and `117`
3. **Floating-point** — use `float64` everywhere, match C double arithmetic; `EPSILON = 1e-16`
4. **Face boundary crossing** — apply distortion correction in `FaceIJKToGeoBoundary`; critical for `CellToBoundary`
5. **Class II vs Class III** — odd resolutions apply `M_AP7_ROT_RADS` rotation; wrong class detection breaks all coord math
6. **Trust C over ZipRecruiter** — where ZipRecruiter diverges from C source, follow C

---

## Execution Order Summary

| Phase | Agents | Work | Mode |
|-------|--------|------|------|
| 1 | 1 | Foundation, constants, module setup | Sequential |
| 2 | 3 | h3index bits, coordijk+vecs, bbox | Parallel |
| 3 | 1 | faceijk + latLng (core coordinate math) | Sequential |
| 4 | 2 | Indexing API (`LatLngToCell` etc), algos | Parallel |
| 5 | 4 | Polygon, directed edges, vertices, metrics | Parallel |
| 6 | 2 | testutil+golden tests, CGO tests+benchmarks | Parallel |

**Total:** ~6 sequential waves, up to 4 agents running in parallel per wave.

---

## Verification

```sh
go build ./...                                                    # no errors
go vet ./...                                                      # zero warnings
staticcheck ./...                                                 # zero issues
golangci-lint run                                                 # zero issues
go test ./... -count=1                                            # all CGO=0 unit tests pass
go test -tags cgo ./... -count=1                                  # all CGO comparison tests pass
go test -bench=. -benchmem -count=10 -tags cgo                   # native ≤1% of CGO wall-clock and allocs
```
