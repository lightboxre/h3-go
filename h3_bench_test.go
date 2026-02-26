// Benchmarks for H3 Native Go implementation.
// These benchmarks measure the performance of the native Go implementation
// without requiring CGO or the C H3 library.
package h3

import (
	"testing"
)

// Benchmark data - using San Francisco City Hall and Google Googleplex
const (
	benchLat  = 37.7796
	benchLng  = -122.4193
	benchRes  = 9
	benchCell = Cell(0x8928308280fffff)
)

// ============================================================================
// Core Indexing Benchmarks
// ============================================================================

func BenchmarkLatLngToCell(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = LatLngToCell(benchLat, benchLng, benchRes)
	}
}

func BenchmarkCellToLatLng(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellToLatLng(benchCell)
	}
}

func BenchmarkCellToBoundary(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellToBoundary(benchCell)
	}
}

// ============================================================================
// Hierarchy Benchmarks
// ============================================================================

func BenchmarkCellToParent(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellToParent(benchCell, benchRes-1)
	}
}

func BenchmarkCellToChildren(b *testing.B) {
	// Use a lower resolution cell for reasonable child count
	cell := LatLngToCell(benchLat, benchLng, 5)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellToChildren(cell, 6)
	}
}

func BenchmarkCellToChildrenSize(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellToChildrenSize(benchCell, benchRes+1)
	}
}

func BenchmarkCellToCenterChild(b *testing.B) {
	cell := LatLngToCell(benchLat, benchLng, 5)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellToCenterChild(cell, 7)
	}
}

func BenchmarkCellToChildPos(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellToChildPos(benchCell, 5)
	}
}

func BenchmarkChildPosToCell(b *testing.B) {
	cell := LatLngToCell(benchLat, benchLng, 5)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ChildPosToCell(0, cell, 7)
	}
}

// ============================================================================
// Cell Information Benchmarks
// ============================================================================

func BenchmarkGetResolution(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetResolution(benchCell)
	}
}

func BenchmarkGetBaseCellNumber(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetBaseCellNumber(benchCell)
	}
}

func BenchmarkIsValidCell(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsValidCell(benchCell)
	}
}

func BenchmarkIsPentagon(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsPentagon(benchCell)
	}
}

func BenchmarkIsResClassIII(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsResClassIII(benchCell)
	}
}

// ============================================================================
// String Conversion Benchmarks
// ============================================================================

func BenchmarkCellToString(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellToString(benchCell)
	}
}

func BenchmarkStringToCell(b *testing.B) {
	str := "8928308280fffff"
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = StringToCell(str)
	}
}

func BenchmarkH3ToString(b *testing.B) {
	h := uint64(benchCell)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = H3ToString(h)
	}
}

func BenchmarkStringToH3(b *testing.B) {
	str := "8928308280fffff"
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = StringToH3(str)
	}
}

// ============================================================================
// Enumeration Benchmarks
// ============================================================================

func BenchmarkGetNumCells(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetNumCells(9)
	}
}

func BenchmarkGetRes0Cells(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetRes0Cells()
	}
}

func BenchmarkGetPentagonCells(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetPentagonCells(9)
	}
}

// ============================================================================
// Directed Edge Benchmarks
// ============================================================================

func BenchmarkIsValidDirectedEdge(b *testing.B) {
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	edge := edges[0]
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IsValidDirectedEdge(edge)
	}
}

func BenchmarkGetDirectedEdgeOrigin(b *testing.B) {
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	edge := edges[0]
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetDirectedEdgeOrigin(edge)
	}
}

func BenchmarkGetDirectedEdgeDestination(b *testing.B) {
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	edge := edges[0]
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GetDirectedEdgeDestination(edge)
	}
}

func BenchmarkDirectedEdgeToCells(b *testing.B) {
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	edge := edges[0]
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = DirectedEdgeToCells(edge)
	}
}

func BenchmarkOriginToDirectedEdges(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = OriginToDirectedEdges(benchCell)
	}
}

func BenchmarkDirectedEdgeToBoundary(b *testing.B) {
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	edge := edges[0]
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = DirectedEdgeToBoundary(edge)
	}
}

func BenchmarkAreNeighborCells(b *testing.B) {
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	dest := GetDirectedEdgeDestination(edges[0])
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = AreNeighborCells(benchCell, dest)
	}
}

func BenchmarkCellsToDirectedEdge(b *testing.B) {
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	dest := GetDirectedEdgeDestination(edges[0])
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = CellsToDirectedEdge(benchCell, dest)
	}
}

// ============================================================================
// Grid Traversal Benchmarks
// ============================================================================

func BenchmarkGridDisk_k1(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GridDisk(benchCell, 1)
	}
}

func BenchmarkGridDisk_k5(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GridDisk(benchCell, 5)
	}
}

func BenchmarkGridDisk_k10(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GridDisk(benchCell, 10)
	}
}

func BenchmarkGridDiskDistances(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GridDiskDistances(benchCell, 5)
	}
}

func BenchmarkGridDistance(b *testing.B) {
	// Get a neighbor cell to measure distance
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	dest := GetDirectedEdgeDestination(edges[0])
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = GridDistance(benchCell, dest)
	}
}

func BenchmarkCompactCells(b *testing.B) {
	// Use GridDisk result as input
	cells, err := GridDisk(benchCell, 5)
	if err != nil || len(cells) == 0 {
		b.Skip("GridDisk failed")
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = CompactCells(cells)
	}
}

func BenchmarkUncompactCells(b *testing.B) {
	// Get a compacted set first
	cells, err := GridDisk(benchCell, 5)
	if err != nil || len(cells) == 0 {
		b.Skip("GridDisk failed")
	}
	compacted, err := CompactCells(cells)
	if err != nil || len(compacted) == 0 {
		b.Skip("CompactCells failed")
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = UncompactCells(compacted, benchRes)
	}
}

// ============================================================================
// Region Benchmarks
// ============================================================================

func BenchmarkPolygonToCells(b *testing.B) {
	// Create a small triangle polygon around the benchmark location
	polygon := GeoPolygon{
		GeoLoop: GeoLoop{
			{Lat: benchLat, Lng: benchLng},
			{Lat: benchLat + 0.01, Lng: benchLng},
			{Lat: benchLat, Lng: benchLng + 0.01},
			{Lat: benchLat, Lng: benchLng}, // close the loop
		},
		Holes: nil,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = PolygonToCells(polygon, 9)
	}
}

func BenchmarkCellsToLinkedMultiPolygon(b *testing.B) {
	// Use a small grid of cells
	cells, err := GridDisk(benchCell, 3)
	if err != nil || len(cells) == 0 {
		b.Skip("GridDisk failed")
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellsToLinkedMultiPolygon(cells)
	}
}

// ============================================================================
// Metrics Benchmarks
// ============================================================================

func BenchmarkGreatCircleDistanceKm(b *testing.B) {
	a := LatLng{Lat: benchLat, Lng: benchLng}
	bLL := LatLng{Lat: benchLat + 0.1, Lng: benchLng + 0.1}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = GreatCircleDistanceKm(a, bLL)
	}
}

func BenchmarkCellAreaKm2(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = CellAreaKm2(benchCell)
	}
}

func BenchmarkEdgeLengthKm(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = EdgeLengthKm(benchRes)
	}
}

func BenchmarkExactEdgeLengthKm(b *testing.B) {
	edges := OriginToDirectedEdges(benchCell)
	if len(edges) == 0 {
		b.Skip("No edges available")
	}
	edge := edges[0]
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ExactEdgeLengthKm(edge)
	}
}

// ============================================================================
// Vertex Benchmarks
// ============================================================================

// Note: Vertex benchmarks are in h3_vertices_test.go to avoid duplication
