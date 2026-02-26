# H3 Native Go — Benchmark Methodology

## Overview

This document describes the benchmarking and testing setup for comparing the native Go H3 implementation against the official CGO-based binding from Uber.

## Performance Goals

The H3 Native Go implementation aims to achieve:

- **Performance parity**: ≤ 5% difference in wall-clock time vs CGO for most operations
- **Memory efficiency**: ≤ 10% difference in heap allocations vs CGO
- **Zero CGO overhead**: No C interop costs, enabling faster cross-compilation and simpler deployment

## Running Benchmarks

### Native Go Only (No CGO)

Run benchmarks for the native Go implementation without CGO:

```bash
# Run all benchmarks once
go test -bench=. -benchmem .

# Run benchmarks 10 times for stability
go test -bench=. -benchmem -count=10 .

# Run with longer benchmark time for more accurate results
go test -bench=. -benchmem -benchtime=2s -count=10 .

# Save results to file
go test -bench=. -benchmem -count=10 . | tee bench_native.txt
```

### CGO Comparison Tests (Requires C H3 Library)

To run comparison tests against the official CGO binding:

1. **Add the CGO dependency:**
   ```bash
   go get github.com/uber/h3-go/v4
   ```

2. **Run CGO comparison tests:**
   ```bash
   CGO_ENABLED=1 go test -tags cgo ./...
   ```

3. **Run CGO benchmarks:**
   ```bash
   CGO_ENABLED=1 go test -bench=. -benchmem -tags cgo -count=10 . | tee bench_cgo.txt
   ```

### Comparing Results with benchstat

Use the `benchstat` tool to compare native Go vs CGO performance:

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Compare native vs CGO
benchstat bench_native.txt bench_cgo.txt
```

Example output:
```
name                    old time/op    new time/op    delta
LatLngToCell-8             150ns ± 2%     148ns ± 1%   -1.33%  (p=0.000 n=10+10)
CellToLatLng-8             180ns ± 1%     182ns ± 2%   +1.11%  (p=0.005 n=10+10)

name                    old alloc/op   new alloc/op   delta
LatLngToCell-8            0.00B          0.00B          ~     (all equal)
CellToLatLng-8            0.00B          0.00B          ~     (all equal)
```

## Benchmark Coverage

### Core Indexing Functions

| Benchmark | Description | Expected Performance |
|-----------|-------------|---------------------|
| `BenchmarkLatLngToCell` | Convert lat/lng to H3 cell | ~100-200ns |
| `BenchmarkCellToLatLng` | Convert H3 cell to lat/lng | ~150-250ns |
| `BenchmarkCellToBoundary` | Get cell boundary polygon | ~500-1000ns |

### Hierarchy Functions

| Benchmark | Description | Expected Performance |
|-----------|-------------|---------------------|
| `BenchmarkCellToParent` | Get parent cell | ~50-100ns |
| `BenchmarkCellToChildren` | Get all child cells (res diff = 1) | ~1-5µs |
| `BenchmarkCellToChildrenSize` | Calculate child count | ~20-50ns |
| `BenchmarkCellToCenterChild` | Get center child | ~50-100ns |
| `BenchmarkCellToChildPos` | Get child position index | Variable |
| `BenchmarkChildPosToCell` | Get child by position | Variable |

### Cell Information Functions

| Benchmark | Description | Expected Performance |
|-----------|-------------|---------------------|
| `BenchmarkGetResolution` | Extract resolution from cell | ~5-10ns |
| `BenchmarkGetBaseCellNumber` | Extract base cell number | ~5-10ns |
| `BenchmarkIsValidCell` | Validate cell index | ~20-50ns |
| `BenchmarkIsPentagon` | Check if cell is pentagon | ~20-50ns |
| `BenchmarkIsResClassIII` | Check resolution class | ~10-20ns |

### String Conversion Functions

| Benchmark | Description | Expected Performance |
|-----------|-------------|---------------------|
| `BenchmarkCellToString` | Convert cell to hex string | ~50-150ns |
| `BenchmarkStringToCell` | Parse hex string to cell | ~100-200ns |
| `BenchmarkH3ToString` | Convert uint64 to hex string | ~50-150ns |
| `BenchmarkStringToH3` | Parse hex string to uint64 | ~100-200ns |

### Enumeration Functions

| Benchmark | Description | Expected Performance |
|-----------|-------------|---------------------|
| `BenchmarkGetNumCells` | Calculate total cells at resolution | ~10-30ns |
| `BenchmarkGetRes0Cells` | Get all 122 base cells | ~1-2µs |
| `BenchmarkGetPentagonCells` | Get 12 pentagons at resolution | ~500-1000ns |

### Directed Edge Functions

| Benchmark | Description | Expected Performance |
|-----------|-------------|---------------------|
| `BenchmarkIsValidDirectedEdge` | Validate directed edge | ~30-60ns |
| `BenchmarkGetDirectedEdgeOrigin` | Get edge origin cell | ~10-20ns |
| `BenchmarkGetDirectedEdgeDestination` | Get edge destination cell | ~200-500ns |
| `BenchmarkDirectedEdgeToCells` | Get both edge cells | ~10-20ns |
| `BenchmarkOriginToDirectedEdges` | Get all edges from cell | ~100-300ns |
| `BenchmarkDirectedEdgeToBoundary` | Get edge line segment | ~1-2µs |
| `BenchmarkAreNeighborCells` | Check if cells are neighbors | ~500-1000ns |
| `BenchmarkCellsToDirectedEdge` | Find edge between cells | ~500-1000ns |

### Vertex Functions

| Benchmark | Description | Expected Performance |
|-----------|-------------|---------------------|
| `BenchmarkCellToVertex` | Get vertex at index | ~30-60ns |
| `BenchmarkCellToVertexes` | Get all vertices of cell | ~300-600ns |
| `BenchmarkVertexToLatLng` | Get vertex coordinates | ~500-1000ns |
| `BenchmarkIsValidVertex` | Validate vertex index | ~100-200ns |

## Analysis and Profiling

### CPU Profile

Generate a CPU profile to identify performance bottlenecks:

```bash
go test -bench=BenchmarkLatLngToCell -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

In pprof:
```
(pprof) top10          # Show top 10 functions by CPU time
(pprof) list FuncName  # Show source code with line-by-line timing
(pprof) web            # Generate visual call graph (requires graphviz)
```

### Memory Profile

Generate a memory allocation profile:

```bash
go test -bench=BenchmarkLatLngToCell -memprofile=mem.prof
go tool pprof mem.prof
```

### Trace Analysis

Generate an execution trace for detailed analysis:

```bash
go test -bench=BenchmarkLatLngToCell -trace=trace.out
go tool trace trace.out
```

## Continuous Benchmarking

For continuous integration, track benchmark performance over time:

```bash
# Run benchmarks and save to dated file
go test -bench=. -benchmem -count=10 . > "bench_$(date +%Y%m%d).txt"

# Compare against baseline
benchstat bench_baseline.txt "bench_$(date +%Y%m%d).txt"
```

## Benchmark Stability

To ensure stable benchmark results:

1. **Run multiple iterations**: Use `-count=10` or higher
2. **Use longer benchmark time**: `-benchtime=2s` or higher for micro-benchmarks
3. **Disable frequency scaling**: On Linux, set CPU governor to `performance`
4. **Close background applications**: Minimize system load during benchmarking
5. **Use consistent hardware**: Run comparisons on the same machine

## Interpreting Results

### Performance Deltas

- **< 5% difference**: Within noise, likely equivalent performance
- **5-15% difference**: Noticeable but acceptable difference
- **> 15% difference**: Significant performance gap requiring investigation

### Statistical Significance

`benchstat` reports p-values to indicate statistical significance:

- **p < 0.05**: Statistically significant difference
- **p ≥ 0.05**: Difference may be due to noise

### Memory Allocations

- **0 allocs/op**: Ideal for hot-path functions (indexing, validation)
- **1-2 allocs/op**: Acceptable for functions returning slices/strings
- **> 10 allocs/op**: May indicate inefficient implementation

## Known Limitations

### Current Implementation Status

The native Go implementation is under active development. Some functions may have:

- **Incomplete base cell lookup**: Simplified FaceIJK to H3Index conversion
- **Placeholder algorithms**: Some traversal functions use simplified approaches
- **Missing optimizations**: Performance will improve as the library matures

### CGO Comparison Availability

CGO comparison tests require:

1. CGO enabled (`CGO_ENABLED=1`)
2. C H3 library installed on system
3. `github.com/uber/h3-go/v4` dependency added

Without these, only native Go benchmarks will run.

## Contributing Benchmark Results

When contributing benchmark results or performance improvements:

1. **Include system info**: CPU model, RAM, OS, Go version
2. **Run multiple iterations**: `-count=10` minimum
3. **Compare before/after**: Use `benchstat` to show improvements
4. **Document changes**: Explain algorithmic improvements or optimizations

Example:

```bash
# Before optimization
go test -bench=BenchmarkLatLngToCell -benchmem -count=10 . > before.txt

# After optimization
go test -bench=BenchmarkLatLngToCell -benchmem -count=10 . > after.txt

# Show improvement
benchstat before.txt after.txt
```

## CGO Comparison Benchmarks (cgotest/)

Paired `_Native` / `_CGO` benchmarks live in `cgotest/bench_test.go`.
They cover every major H3 function group: core indexing, hierarchy, cell
inspection, string conversion, enumeration, directed edges, grid traversal,
compaction, region, metrics, and vertex operations.

Run native vs CGO side-by-side (requires H3 C library installed):

```bash
cd cgotest
CGO_ENABLED=1 go test -bench=. -benchmem ./...
```

Filter to a specific function:

```bash
CGO_ENABLED=1 go test -bench=BenchmarkLatLngToCell -benchmem ./...
```

Compare native vs CGO with benchstat:

```bash
# Save each variant separately
CGO_ENABLED=1 go test -bench='_Native' -benchmem -count=10 ./... > bench_native.txt
CGO_ENABLED=1 go test -bench='_CGO'    -benchmem -count=10 ./... > bench_cgo.txt
benchstat bench_native.txt bench_cgo.txt
```

## References

- [Go Benchmark Documentation](https://pkg.go.dev/testing#hdr-Benchmarks)
- [H3 C Library](https://github.com/uber/h3)
- [H3 Go CGO Binding](https://github.com/uber/h3-go)
- [benchstat Tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
