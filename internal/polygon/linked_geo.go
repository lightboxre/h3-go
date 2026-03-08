// Package polygon - LinkedGeo data structures for polygon outlines from cells.
// Ported from the C H3 reference implementation (linkedGeo.c).
package polygon

import (
	"github.com/lightboxre/h3-go/internal/bbox"
	"github.com/lightboxre/h3-go/internal/faceijk"
	"github.com/lightboxre/h3-go/internal/h3index"
)

// LinkedGeoLoop represents a loop of geographic coordinates.
// This is a sequence of lat/lng points forming a closed polygon loop.
type LinkedGeoLoop struct {
	Points []faceijk.GeoPoint // Vertices in lat/lng (radians)
}

// LinkedGeoPolygon represents a polygon with one outer loop and zero or more holes.
type LinkedGeoPolygon struct {
	Outer LinkedGeoLoop   // Outer boundary
	Holes []LinkedGeoLoop // Inner holes
}

// LinkedGeoMultiPolygon is a collection of polygons.
type LinkedGeoMultiPolygon []LinkedGeoPolygon

// directedEdge is a directed boundary edge between two lat/lng points.
type directedEdge struct {
	lat1, lng1 float64
	lat2, lng2 float64
}

// vertex is a lat/lng point used as a map key.
type vertex struct {
	lat, lng float64
}

// CellsToLinkedMultiPolygon converts a set of H3 cells into polygon outlines.
// This produces the merged boundary of all cells, with holes detected.
// From C: H3_EXPORT(cellsToLinkedMultiPolygon)
//
// Algorithm:
// 1. For each cell, collect all directed boundary edges (v[i]→v[i+1])
// 2. Interior edges: edge v→w is interior if w→v also appears (shared by two cells)
// 3. Keep only exterior directed edges (no corresponding reverse edge)
// 4. Trace directed exterior edges into closed loops
// 5. Detect holes by winding order (CCW=outer, CW=hole)
// 6. Group loops into polygons
func CellsToLinkedMultiPolygon(cells []h3index.H3Index) LinkedGeoMultiPolygon {
	if len(cells) == 0 {
		return nil
	}

	// Step 1: Collect all directed boundary edges from all cells.
	// Each cell contributes n directed edges (n = 6 for hexagons, 5 for pentagons).
	dirEdgeSet := make(map[directedEdge]bool)
	for _, h := range cells {
		if !h3index.IsValid(h) {
			continue
		}
		for _, de := range getCellDirectedEdges(h) {
			dirEdgeSet[de] = true
		}
	}

	// Step 2 & 3: Keep only exterior edges (those whose reverse is NOT present).
	// A reverse of edge (A→B) is (B→A). If both exist, the edge is shared (interior).
	var exteriorEdges []directedEdge
	for de := range dirEdgeSet {
		rev := directedEdge{de.lat2, de.lng2, de.lat1, de.lng1}
		if !dirEdgeSet[rev] {
			exteriorEdges = append(exteriorEdges, de)
		}
	}

	if len(exteriorEdges) == 0 {
		return nil
	}

	// Step 4: Build a directed adjacency map: start-vertex → outgoing edge.
	// For a valid exterior boundary, each vertex has exactly one outgoing edge.
	nextEdge := make(map[vertex]directedEdge, len(exteriorEdges))
	for _, de := range exteriorEdges {
		v := vertex{de.lat1, de.lng1}
		nextEdge[v] = de
	}

	// Trace loops by following directed edges.
	visitedVertex := make(map[vertex]bool, len(exteriorEdges))
	var loops []LinkedGeoLoop

	for _, startDE := range exteriorEdges {
		startV := vertex{startDE.lat1, startDE.lng1}
		if visitedVertex[startV] {
			continue
		}

		var loopPoints []faceijk.GeoPoint
		cur := startV

		for {
			if visitedVertex[cur] {
				break
			}
			visitedVertex[cur] = true
			loopPoints = append(loopPoints, faceijk.GeoPoint{Lat: cur.lat, Lng: cur.lng})

			de, ok := nextEdge[cur]
			if !ok {
				break
			}
			cur = vertex{de.lat2, de.lng2}
		}

		if len(loopPoints) >= 3 {
			loops = append(loops, LinkedGeoLoop{Points: loopPoints})
		}
	}

	// Step 5 & 6: Classify loops as outer or holes, group into polygons.
	var polygons []LinkedGeoPolygon
	var holes []LinkedGeoLoop

	for _, loop := range loops {
		if isCounterClockwise(loop) {
			polygons = append(polygons, LinkedGeoPolygon{Outer: loop})
		} else {
			holes = append(holes, loop)
		}
	}

	// Assign holes to their containing polygon.
	for _, hole := range holes {
		for i := range polygons {
			if loopContainsLoop(polygons[i].Outer, hole) {
				polygons[i].Holes = append(polygons[i].Holes, hole)
				break
			}
		}
	}

	return polygons
}

// getCellDirectedEdges returns the directed boundary edges of a cell.
// Edges are in order: v[0]→v[1], v[1]→v[2], ..., v[n-1]→v[0].
func getCellDirectedEdges(h h3index.H3Index) []directedEdge {
	res := h.Resolution()
	isPent := h3index.IsPentagon(h)
	fijk := faceijk.H3ToFaceIJK(h)
	boundary := faceijk.FaceIJKToGeoBoundary(fijk, res, isPent)

	n := len(boundary)
	edges := make([]directedEdge, n)
	for i := range n {
		p1 := boundary[i]
		p2 := boundary[(i+1)%n]
		edges[i] = directedEdge{p1.Lat, p1.Lng, p2.Lat, p2.Lng}
	}
	return edges
}

// isCounterClockwise determines if a loop is counter-clockwise (outer loop)
// or clockwise (hole), using the shoelace formula with (lng, lat) as (x, y).
// H3 boundary vertices are ordered CCW when viewed from above (north pole up),
// which gives positive area with x=lng, y=lat (standard math orientation).
func isCounterClockwise(loop LinkedGeoLoop) bool {
	if len(loop.Points) < 3 {
		return true // default to outer
	}

	area := 0.0
	n := len(loop.Points)
	for i := range n {
		p1 := loop.Points[i]
		p2 := loop.Points[(i+1)%n]
		area += (p1.Lng * p2.Lat) - (p2.Lng * p1.Lat)
	}
	// Positive area = CCW (outer loop) in (lng=x, lat=y) orientation.
	return area > 0
}

// loopContainsLoop checks if outer loop contains the inner loop.
func loopContainsLoop(outer, inner LinkedGeoLoop) bool {
	if len(inner.Points) == 0 {
		return false
	}
	testPoint := inner.Points[0]
	lats := make([]float64, len(outer.Points))
	lngs := make([]float64, len(outer.Points))
	for i, p := range outer.Points {
		lats[i] = p.Lat
		lngs[i] = p.Lng
	}
	bboxOuter := bboxFromGeoPoints(outer.Points)
	return PointInGeoLoop(lats, lngs, bboxOuter, testPoint.Lat, testPoint.Lng)
}

// bboxFromGeoPoints creates a bounding box from GeoPoints.
func bboxFromGeoPoints(points []faceijk.GeoPoint) BBox {
	lats := make([]float64, len(points))
	lngs := make([]float64, len(points))
	for i, p := range points {
		lats[i] = p.Lat
		lngs[i] = p.Lng
	}
	return BBoxFromGeoLoop(lats, lngs)
}

// BBoxFromGeoLoop wraps bbox package function for convenience.
func BBoxFromGeoLoop(lats, lngs []float64) BBox {
	return bbox.BBoxFromGeoLoop(lats, lngs)
}

// BBox is an alias for bbox.BBox.
type BBox = bbox.BBox

// almostEqual checks if two floats are approximately equal.
func almostEqual(a, b float64) bool {
	const epsilon = 1e-10
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}

// edge is an alias for directedEdge (backward compatibility).
type edge = directedEdge

// getCellBoundaryEdges returns directed boundary edges (kept for compatibility).
func getCellBoundaryEdges(h h3index.H3Index) []edge {
	return getCellDirectedEdges(h)
}

// normalizeEdge returns a canonical form of an edge (smaller vertex first by lat then lng).
// This is a utility used in tests; the production code uses directed edges directly.
func normalizeEdge(e edge) edge {
	if e.lat1 < e.lat2 || (e.lat1 == e.lat2 && e.lng1 < e.lng2) {
		return e
	}
	return edge{e.lat2, e.lng2, e.lat1, e.lng1}
}
