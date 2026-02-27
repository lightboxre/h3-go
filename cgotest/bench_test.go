//go:build cgo

// Package cgotest — side-by-side benchmarks: native Go H3 vs CGO Uber H3.
//
// Run with:
//
//	cd cgotest && CGO_ENABLED=1 go test -bench=. -benchmem ./...
//
// Filter to one function:
//
//	CGO_ENABLED=1 go test -bench=BenchmarkLatLngToCell -benchmem ./...
package cgotest

import (
	"testing"

	h3native "github.com/EDRInc/h3-go"
	h3cgo "github.com/uber/h3-go/v4"
)

// ============================================================================
// Core Indexing
// ============================================================================

func BenchmarkLatLngToCell_Native(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = h3native.LatLngToCell(sfLat, sfLng, 9)
	}
}

func BenchmarkLatLngToCell_CGO(b *testing.B) {
	ll := h3cgo.NewLatLng(sfLat, sfLng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.LatLngToCell(ll, 9)
	}
}

func BenchmarkCellToLatLng_Native(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToLatLng(c)
	}
}

func BenchmarkCellToLatLng_CGO(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 9)
	cgoCell := nativeToCGOCell(c)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.CellToLatLng(cgoCell)
	}
}

func BenchmarkCellToBoundary_Native(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToBoundary(c)
	}
}

func BenchmarkCellToBoundary_CGO(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 9)
	cgoCell := nativeToCGOCell(c)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.CellToBoundary(cgoCell)
	}
}

// ============================================================================
// Hierarchy
// ============================================================================

func BenchmarkCellToParent_Native(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToParent(c, 5)
	}
}

func BenchmarkCellToParent_CGO(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 9)
	cgoCell := nativeToCGOCell(c)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoCell.Parent(5)
	}
}

func BenchmarkCellToChildren_Native(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToChildren(c, 6)
	}
}

func BenchmarkCellToChildren_CGO(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 5)
	cgoCell := nativeToCGOCell(c)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoCell.Children(6)
	}
}

func BenchmarkCellToChildrenSize_Native(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToChildrenSize(c, 6)
	}
}

func BenchmarkCellToCenterChild_Native(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToCenterChild(c, 7)
	}
}

func BenchmarkCellToCenterChild_CGO(b *testing.B) {
	c := h3native.LatLngToCell(sfLat, sfLng, 5)
	cgoCell := nativeToCGOCell(c)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoCell.CenterChild(7)
	}
}

func BenchmarkCellToChildPos_Native(b *testing.B) {
	child := h3native.Cell(0x89283082877ffff)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToChildPos(child, 7)
	}
}

func BenchmarkCellToChildPos_CGO(b *testing.B) {
	child := h3native.Cell(0x89283082877ffff)
	cgoChild := nativeToCGOCell(child)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoChild.ChildPos(7)
	}
}

func BenchmarkChildPosToCell_Native(b *testing.B) {
	parent := h3native.Cell(0x85283473fffffff)
	pos := h3native.CellToChildPos(h3native.Cell(0x89283082877ffff), 7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.ChildPosToCell(pos, parent, 9)
	}
}

func BenchmarkChildPosToCell_CGO(b *testing.B) {
	parent := h3native.Cell(0x85283473fffffff)
	pos := h3native.CellToChildPos(h3native.Cell(0x89283082877ffff), 7)
	cgoParent := nativeToCGOCell(parent)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.ChildPosToCell(int(pos), cgoParent, 9)
	}
}

// ============================================================================
// Cell Information
// ============================================================================

func BenchmarkGetResolution_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.GetResolution(c)
	}
}

func BenchmarkGetResolution_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	for i := 0; i < b.N; i++ {
		_ = cgoCell.Resolution()
	}
}

func BenchmarkGetBaseCellNumber_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.GetBaseCellNumber(c)
	}
}

func BenchmarkGetBaseCellNumber_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	for i := 0; i < b.N; i++ {
		_ = cgoCell.BaseCellNumber()
	}
}

func BenchmarkIsValidCell_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.IsValidCell(c)
	}
}

func BenchmarkIsValidCell_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	for i := 0; i < b.N; i++ {
		_ = cgoCell.IsValid()
	}
}

func BenchmarkIsPentagon_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.IsPentagon(c)
	}
}

func BenchmarkIsPentagon_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	for i := 0; i < b.N; i++ {
		_ = cgoCell.IsPentagon()
	}
}

func BenchmarkIsResClassIII_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.IsResClassIII(c)
	}
}

func BenchmarkIsResClassIII_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	for i := 0; i < b.N; i++ {
		_ = cgoCell.IsResClassIII()
	}
}

// ============================================================================
// String Conversion
// ============================================================================

func BenchmarkCellToString_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToString(c)
	}
}

func BenchmarkCellToString_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	for i := 0; i < b.N; i++ {
		_ = h3cgo.CellToString(cgoCell)
	}
}

func BenchmarkStringToCell_Native(b *testing.B) {
	s := "89283082877ffff"
	for i := 0; i < b.N; i++ {
		_, _ = h3native.StringToCell(s)
	}
}

func BenchmarkStringToCell_CGO(b *testing.B) {
	s := "89283082877ffff"
	for i := 0; i < b.N; i++ {
		_ = h3cgo.CellFromString(s)
	}
}

// ============================================================================
// Enumeration
// ============================================================================

func BenchmarkGetNumCells_Native(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = h3native.GetNumCells(9)
	}
}

func BenchmarkGetNumCells_CGO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = h3cgo.NumCells(9)
	}
}

func BenchmarkGetRes0Cells_Native(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = h3native.GetRes0Cells()
	}
}

func BenchmarkGetRes0Cells_CGO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.Res0Cells()
	}
}

func BenchmarkGetPentagonCells_Native(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = h3native.GetPentagonCells(9)
	}
}

func BenchmarkGetPentagonCells_CGO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.Pentagons(9)
	}
}

// ============================================================================
// Directed Edges
// ============================================================================

func BenchmarkIsValidDirectedEdge_Native(b *testing.B) {
	nativeEdges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	e := nativeEdges[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.IsValidDirectedEdge(e)
	}
}

func BenchmarkIsValidDirectedEdge_CGO(b *testing.B) {
	nativeEdges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	cgoEdge := nativeToCGOEdge(nativeEdges[0])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cgoEdge.IsValid()
	}
}

func BenchmarkGetDirectedEdgeOrigin_Native(b *testing.B) {
	nativeEdges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	e := nativeEdges[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.GetDirectedEdgeOrigin(e)
	}
}

func BenchmarkGetDirectedEdgeOrigin_CGO(b *testing.B) {
	nativeEdges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	cgoEdge := nativeToCGOEdge(nativeEdges[0])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoEdge.Origin()
	}
}

func BenchmarkGetDirectedEdgeDestination_Native(b *testing.B) {
	nativeEdges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	e := nativeEdges[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.GetDirectedEdgeDestination(e)
	}
}

func BenchmarkGetDirectedEdgeDestination_CGO(b *testing.B) {
	nativeEdges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	cgoEdge := nativeToCGOEdge(nativeEdges[0])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoEdge.Destination()
	}
}

func BenchmarkDirectedEdgeToCells_Native(b *testing.B) {
	nativeEdges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	e := nativeEdges[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.DirectedEdgeToCells(e)
	}
}

func BenchmarkDirectedEdgeToCells_CGO(b *testing.B) {
	nativeEdges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	cgoEdge := nativeToCGOEdge(nativeEdges[0])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoEdge.Cells()
	}
}

func BenchmarkOriginToDirectedEdges_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.OriginToDirectedEdges(c)
	}
}

func BenchmarkOriginToDirectedEdges_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	for i := 0; i < b.N; i++ {
		_, _ = cgoCell.DirectedEdges()
	}
}

func BenchmarkDirectedEdgeToBoundary_Native(b *testing.B) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	e, _ := h3native.CellsToDirectedEdge(origin, disk[1])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.DirectedEdgeToBoundary(e)
	}
}

func BenchmarkDirectedEdgeToBoundary_CGO(b *testing.B) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	e, _ := h3native.CellsToDirectedEdge(origin, disk[1])
	cgoEdge := nativeToCGOEdge(e)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoEdge.Boundary()
	}
}

func BenchmarkAreNeighborCells_Native(b *testing.B) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	neighbor := disk[1]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.AreNeighborCells(origin, neighbor)
	}
}

func BenchmarkAreNeighborCells_CGO(b *testing.B) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	neighbor := disk[1]
	cgoOrigin := nativeToCGOCell(origin)
	cgoNeighbor := nativeToCGOCell(neighbor)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoOrigin.IsNeighbor(cgoNeighbor)
	}
}

func BenchmarkCellsToDirectedEdge_Native(b *testing.B) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	neighbor := disk[1]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.CellsToDirectedEdge(origin, neighbor)
	}
}

func BenchmarkCellsToDirectedEdge_CGO(b *testing.B) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	neighbor := disk[1]
	cgoOrigin := nativeToCGOCell(origin)
	cgoNeighbor := nativeToCGOCell(neighbor)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoOrigin.DirectedEdge(cgoNeighbor)
	}
}

// ============================================================================
// Grid Traversal
// ============================================================================

func BenchmarkGridDisk_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.GridDisk(c, 2)
	}
}

func BenchmarkGridDisk_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.GridDisk(cgoCell, 2)
	}
}

func BenchmarkGridDiskDistances_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.GridDiskDistances(c, 2)
	}
}

func BenchmarkGridDiskDistances_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.GridDiskDistances(cgoCell, 2)
	}
}

func BenchmarkGridDistance_Native(b *testing.B) {
	src := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(src, 2)
	dst := disk[len(disk)/2]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.GridDistance(src, dst)
	}
}

func BenchmarkGridDistance_CGO(b *testing.B) {
	src := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(src, 2)
	dst := disk[len(disk)/2]
	cgoSrc := nativeToCGOCell(src)
	cgoDst := nativeToCGOCell(dst)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.GridDistance(cgoSrc, cgoDst)
	}
}

func BenchmarkGridPathCells_Native(b *testing.B) {
	src := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(src, 2)
	dst := disk[len(disk)/2]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.GridPathCells(src, dst)
	}
}

func BenchmarkGridPathCells_CGO(b *testing.B) {
	src := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(src, 2)
	dst := disk[len(disk)/2]
	cgoSrc := nativeToCGOCell(src)
	cgoDst := nativeToCGOCell(dst)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.GridPath(cgoSrc, cgoDst)
	}
}

// ============================================================================
// Compaction
// ============================================================================

func BenchmarkCompactCells_Native(b *testing.B) {
	parent := h3native.Cell(0x85283473fffffff)
	children := h3native.CellToChildren(parent, 6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.CompactCells(children)
	}
}

func BenchmarkCompactCells_CGO(b *testing.B) {
	parent := h3native.Cell(0x85283473fffffff)
	children := h3native.CellToChildren(parent, 6)
	cgoCells := make([]h3cgo.Cell, len(children))
	for i, c := range children {
		cgoCells[i] = nativeToCGOCell(c)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.CompactCells(cgoCells)
	}
}

func BenchmarkUncompactCells_Native(b *testing.B) {
	parent := h3native.Cell(0x85283473fffffff)
	cells := []h3native.Cell{parent}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.UncompactCells(cells, 6)
	}
}

func BenchmarkUncompactCells_CGO(b *testing.B) {
	parent := h3native.Cell(0x85283473fffffff)
	cgoCells := []h3cgo.Cell{nativeToCGOCell(parent)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.UncompactCells(cgoCells, 6)
	}
}

// ============================================================================
// Region
// ============================================================================

func BenchmarkPolygonToCells_Native(b *testing.B) {
	polygon := h3native.GeoPolygon{
		GeoLoop: []h3native.LatLng{
			{Lat: 37.78, Lng: -122.42},
			{Lat: 37.78, Lng: -122.41},
			{Lat: 37.77, Lng: -122.42},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3native.PolygonToCells(polygon, 9)
	}
}

func BenchmarkPolygonToCells_CGO(b *testing.B) {
	polygon := h3native.GeoPolygon{
		GeoLoop: []h3native.LatLng{
			{Lat: 37.78, Lng: -122.42},
			{Lat: 37.78, Lng: -122.41},
			{Lat: 37.77, Lng: -122.42},
		},
	}
	cgoPoly := convertGeoPolygon(polygon)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.PolygonToCells(cgoPoly, 9)
	}
}

func BenchmarkCellsToLinkedMultiPolygon_Native(b *testing.B) {
	cells := []h3native.Cell{h3native.Cell(0x89283082877ffff)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.CellsToLinkedMultiPolygon(cells)
	}
}

func BenchmarkCellsToLinkedMultiPolygon_CGO(b *testing.B) {
	cgoCells := []h3cgo.Cell{nativeToCGOCell(h3native.Cell(0x89283082877ffff))}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.CellsToMultiPolygon(cgoCells)
	}
}

// ============================================================================
// Metrics
// ============================================================================

func BenchmarkGreatCircleDistanceKm_Native(b *testing.B) {
	a := h3native.LatLng{Lat: sfLat, Lng: sfLng}
	bll := h3native.LatLng{Lat: googLat, Lng: googLng}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.GreatCircleDistanceKm(a, bll)
	}
}

func BenchmarkGreatCircleDistanceKm_CGO(b *testing.B) {
	a := h3cgo.NewLatLng(sfLat, sfLng)
	bll := h3cgo.NewLatLng(googLat, googLng)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3cgo.GreatCircleDistanceKm(a, bll)
	}
}

func BenchmarkCellAreaKm2_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.CellAreaKm2(c)
	}
}

func BenchmarkCellAreaKm2_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.CellAreaKm2(cgoCell)
	}
}

func BenchmarkEdgeLengthKm_Native(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = h3native.EdgeLengthKm(9)
	}
}

func BenchmarkEdgeLengthKm_CGO(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.HexagonEdgeLengthAvgKm(9)
	}
}

func BenchmarkExactEdgeLengthKm_Native(b *testing.B) {
	edges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	e := edges[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.ExactEdgeLengthKm(e)
	}
}

func BenchmarkExactEdgeLengthKm_CGO(b *testing.B) {
	edges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	cgoEdge := nativeToCGOEdge(edges[0])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = h3cgo.EdgeLengthKm(cgoEdge)
	}
}

// ============================================================================
// Vertex
// ============================================================================

func BenchmarkCellToVertex_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToVertex(c, 0)
	}
}

func BenchmarkCellToVertex_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoCell.Vertex(0)
	}
}

func BenchmarkCellToVertexes_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	for i := 0; i < b.N; i++ {
		_ = h3native.CellToVertexes(c)
	}
}

func BenchmarkCellToVertexes_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoCell.Vertexes()
	}
}

func BenchmarkVertexToLatLng_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	v := h3native.CellToVertex(c, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.VertexToLatLng(v)
	}
}

func BenchmarkVertexToLatLng_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	cgoVertex, _ := cgoCell.Vertex(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cgoVertex.LatLng()
	}
}

func BenchmarkIsValidVertex_Native(b *testing.B) {
	c := h3native.Cell(0x89283082877ffff)
	v := h3native.CellToVertex(c, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3native.IsValidVertex(v)
	}
}

func BenchmarkIsValidVertex_CGO(b *testing.B) {
	cgoCell := nativeToCGOCell(h3native.Cell(0x89283082877ffff))
	cgoVertex, _ := cgoCell.Vertex(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h3cgo.IsValidVertex(cgoVertex)
	}
}
