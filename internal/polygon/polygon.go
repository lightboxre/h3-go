// Package polygon implements polygon operations for H3 geospatial indexing.
// This includes point-in-polygon testing and polygon-to-cells conversion.
// Ported from the C H3 reference implementation (polygon.c).
package polygon

import (
	"math"

	"github.com/lightboxre/h3-go/internal/algos"
	"github.com/lightboxre/h3-go/internal/bbox"
	"github.com/lightboxre/h3-go/internal/constants"
	"github.com/lightboxre/h3-go/internal/faceijk"
	"github.com/lightboxre/h3-go/internal/h3index"
)

// PointInGeoLoop tests if a point (lat, lng in radians) is inside a polygon loop.
// Uses the ray-casting algorithm with antimeridian support.
// From C: pointInsideGeoloop
func PointInGeoLoop(lats, lngs []float64, bboxLoop bbox.BBox, lat, lng float64) bool {
	if len(lats) < 3 || len(lngs) < 3 || len(lats) != len(lngs) {
		return false
	}

	// Quick rejection: check if point is in bounding box
	if !bbox.BBoxContains(bboxLoop, lat, lng) {
		return false
	}

	isTransmeridian := bbox.BBoxIsTransmeridian(bboxLoop)
	contains := false

	// Adjust point longitude if transmeridian
	testLng := lng
	if isTransmeridian && lng < 0 {
		testLng += 2 * math.Pi
	}

	// Ray casting algorithm: cast a ray from point to infinity and count intersections
	numVerts := len(lats)
	for i := range numVerts {
		lat1 := lats[i]
		lng1 := lngs[i]
		lat2 := lats[(i+1)%numVerts]
		lng2 := lngs[(i+1)%numVerts]

		// Adjust longitudes for transmeridian case
		if isTransmeridian {
			if lng1 < 0 {
				lng1 += 2 * math.Pi
			}
			if lng2 < 0 {
				lng2 += 2 * math.Pi
			}
		}

		// Ray casting test: does horizontal ray from (lat, lng) intersect edge?
		// The ray goes to the right (increasing longitude)
		if ((lng1 >= testLng) != (lng2 >= testLng)) &&
			(lat <= (lat2-lat1)*(testLng-lng1)/(lng2-lng1)+lat1) {
			contains = !contains
		}
	}

	return contains
}

// PointInGeoPolygon tests if a point is inside a polygon (considering holes).
// outerLats/outerLngs define the outer boundary (radians).
// holesLats/holesLngs are arrays of hole loops (radians).
// Returns true if point is inside outer loop and outside all holes.
// From C: pointInsidePolygon
func PointInGeoPolygon(outerLats, outerLngs []float64, holesLats, holesLngs [][]float64, bboxPoly bbox.BBox, lat, lng float64) bool {
	// Check if point is inside outer loop
	if !PointInGeoLoop(outerLats, outerLngs, bboxPoly, lat, lng) {
		return false
	}

	// Check if point is inside any hole (if so, it's outside the polygon)
	for i := range holesLats {
		if len(holesLats[i]) < 3 {
			continue
		}
		holeBBox := bbox.BBoxFromGeoLoop(holesLats[i], holesLngs[i])
		if PointInGeoLoop(holesLats[i], holesLngs[i], holeBBox, lat, lng) {
			return false
		}
	}

	return true
}

// PolygonToCells returns all H3 cells at the given resolution that are contained
// within or overlap with the polygon.
// outerLats/outerLngs are in RADIANS.
// holesLats/holesLngs are arrays of hole loops in RADIANS.
// flags is reserved for future use (e.g., containment mode vs. overlapping mode).
// From C: H3_EXPORT(polygonToCells)
func PolygonToCells(outerLats, outerLngs []float64, holesLats, holesLngs [][]float64, res int, _ uint32) ([]h3index.H3Index, error) {
	if len(outerLats) < 3 || len(outerLngs) < 3 {
		return nil, nil
	}

	if res < 0 || res > constants.MaxH3Res {
		return nil, algos.ErrDomain
	}

	bboxPoly := bbox.BBoxFromGeoLoop(outerLats, outerLngs)
	var result []h3index.H3Index
	for _, root := range h3index.GetRes0Cells() {
		collectPolygonCells(root, 0, res, outerLats, outerLngs, holesLats, holesLngs, bboxPoly, &result)
	}

	return result, nil
}

func collectPolygonCells(h h3index.H3Index, currentRes, targetRes int, outerLats, outerLngs []float64, holesLats, holesLngs [][]float64, bboxPoly bbox.BBox, out *[]h3index.H3Index) {
	// Coarse parent boundaries near the antimeridian are still imprecise enough
	// to produce false-negative bboxes at the top of the hierarchy. Delay bbox
	// pruning until res 2, then use exact boundary-derived bboxes below that.
	if currentRes >= 2 && !cellBBoxOverlapsPolygonBBox(h, currentRes, bboxPoly) {
		return
	}

	if currentRes == targetRes {
		if cellCenterInPolygon(h, currentRes, outerLats, outerLngs, holesLats, holesLngs, bboxPoly) {
			*out = append(*out, h)
		}
		return
	}

	for _, child := range cellChildren(h, currentRes) {
		collectPolygonCells(child, currentRes+1, targetRes, outerLats, outerLngs, holesLats, holesLngs, bboxPoly, out)
	}
}

func cellBBoxOverlapsPolygonBBox(h h3index.H3Index, res int, bboxPoly bbox.BBox) bool {
	fijk := faceijk.H3ToFaceIJK(h)
	boundary := faceijk.FaceIJKToGeoBoundary(fijk, res, h3index.IsPentagon(h))
	if len(boundary) == 0 {
		return false
	}
	return bbox.BBoxOverlap(cellToBBox(h, res, boundary), bboxPoly)
}

func cellChildren(h h3index.H3Index, currentRes int) []h3index.H3Index {
	digits := []int{
		constants.CENTER_DIGIT,
		constants.K_AXES_DIGIT,
		constants.J_AXES_DIGIT,
		constants.JK_AXES_DIGIT,
		constants.I_AXES_DIGIT,
		constants.IK_AXES_DIGIT,
		constants.IJ_AXES_DIGIT,
	}
	if h3index.IsPentagon(h) {
		digits = []int{
			constants.CENTER_DIGIT,
			constants.J_AXES_DIGIT,
			constants.JK_AXES_DIGIT,
			constants.I_AXES_DIGIT,
			constants.IK_AXES_DIGIT,
			constants.IJ_AXES_DIGIT,
		}
	}

	children := make([]h3index.H3Index, 0, len(digits))
	for _, digit := range digits {
		child := h3index.SetResolution(h, currentRes+1)
		child = h3index.SetIndexDigit(child, currentRes, digit)
		for r := currentRes + 1; r < constants.MaxH3Res; r++ {
			child = h3index.SetIndexDigit(child, r, constants.INVALID_DIGIT)
		}
		children = append(children, child)
	}
	return children
}

// cellCenterInPolygon returns true if the cell's center point is inside the polygon.
// This is the default containment mode (CONTAINMENT_CENTER / flags=0) used by H3.
func cellCenterInPolygon(h h3index.H3Index, res int, outerLats, outerLngs []float64, holesLats, holesLngs [][]float64, bboxPoly bbox.BBox) bool {
	fijk := faceijk.H3ToFaceIJK(h)
	centerLat, centerLng := faceijk.FaceIJKToGeo(fijk, res)
	return PointInGeoPolygon(outerLats, outerLngs, holesLats, holesLngs, bboxPoly, centerLat, centerLng)
}

// cellOverlapsPolygon checks if a cell overlaps with the polygon.
// A cell overlaps if:
// - Its center is inside the polygon, OR
// - Any of its boundary vertices are inside the polygon, OR
// - Any polygon vertices are inside the cell
func cellOverlapsPolygon(h h3index.H3Index, res int, outerLats, outerLngs []float64, holesLats, holesLngs [][]float64, bboxPoly bbox.BBox) bool {
	// Get cell center
	fijk := faceijk.H3ToFaceIJK(h)
	centerLat, centerLng := faceijk.FaceIJKToGeo(fijk, res)

	// Check if center is in polygon
	if PointInGeoPolygon(outerLats, outerLngs, holesLats, holesLngs, bboxPoly, centerLat, centerLng) {
		return true
	}

	// Check if any boundary vertex is inside the polygon
	isPent := h3index.IsPentagon(h)
	boundary := faceijk.FaceIJKToGeoBoundary(fijk, res, isPent)

	for _, vertex := range boundary {
		if PointInGeoPolygon(outerLats, outerLngs, holesLats, holesLngs, bboxPoly, vertex.Lat, vertex.Lng) {
			return true
		}
	}

	// Check if any polygon vertex is inside the cell (for large cells)
	// This is a simplified check - for production, would also check edge intersections
	cellBBox := cellToBBox(h, res, boundary)
	for i := range outerLats {
		if bbox.BBoxContains(cellBBox, outerLats[i], outerLngs[i]) {
			// Polygon vertex is within cell's bbox - consider it overlapping
			return true
		}
	}

	return false
}

// cellToBBox creates a bounding box from a cell's boundary
func cellToBBox(_ h3index.H3Index, _ int, boundary []faceijk.GeoPoint) bbox.BBox {
	if len(boundary) == 0 {
		return bbox.BBox{}
	}

	lats := make([]float64, len(boundary))
	lngs := make([]float64, len(boundary))
	for i, pt := range boundary {
		lats[i] = pt.Lat
		lngs[i] = pt.Lng
	}

	return bbox.BBoxFromGeoLoop(lats, lngs)
}

// getAvgHexagonEdgeLengthRads returns approximate average edge length in radians
// for hexagons at the given resolution.
// From C: getHexagonEdgeLengthAvg
func getAvgHexagonEdgeLengthRads(res int) float64 {
	// Average edge lengths in kilometers for each resolution (from H3 spec)
	avgEdgeLenKm := []float64{
		1107.712591, // res 0
		418.676005,  // res 1
		158.244655,  // res 2
		59.810857,   // res 3
		22.606379,   // res 4
		8.544408,    // res 5
		3.229482,    // res 6
		1.220629,    // res 7
		0.461354,    // res 8
		0.174375,    // res 9
		0.065907,    // res 10
		0.024910,    // res 11
		0.009415,    // res 12
		0.003559,    // res 13
		0.001348,    // res 14
		0.000509,    // res 15
	}

	if res < 0 || res >= len(avgEdgeLenKm) {
		return 0.1 // fallback
	}

	// Convert km to radians using Earth radius
	return avgEdgeLenKm[res] / constants.EarthRadiusKm
}

// faceIJKToH3 converts FaceIJK coordinates to an H3Index at the given resolution.
// Delegates to the canonical implementation in the faceijk package.
func faceIJKToH3(fijk faceijk.FaceIJK, res int) h3index.H3Index {
	return faceijk.FaceIJKToH3(fijk, res)
}
