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
func PolygonToCells(outerLats, outerLngs []float64, holesLats, holesLngs [][]float64, res int, flags uint32) ([]h3index.H3Index, error) {
	if len(outerLats) < 3 || len(outerLngs) < 3 {
		return nil, nil
	}

	if res < 0 || res > constants.MaxH3Res {
		return nil, algos.ErrDomain
	}

	// Compute bounding box of the polygon
	bboxPoly := bbox.BBoxFromGeoLoop(outerLats, outerLngs)

	// Find the center of the bounding box to seed the search
	centerLat, centerLng := bbox.BBoxCenter(bboxPoly)

	// Get the H3 cell at this center point
	seedFijk := faceijk.GeoToFaceIJK(centerLat, centerLng, res)
	seed := faceIJKToH3(seedFijk, res)

	// Estimate the grid disk size needed to cover the polygon
	// Use the bbox dimensions to estimate the k value
	bboxHeight := bbox.BBoxHeight(bboxPoly)
	bboxWidth := bbox.BBoxWidth(bboxPoly)
	bboxDiameter := math.Sqrt(bboxHeight*bboxHeight + bboxWidth*bboxWidth)

	// Approximate average edge length at this resolution (in radians)
	avgEdgeLenRads := getAvgHexagonEdgeLengthRads(res)

	// Calculate k (number of rings) to cover the bbox
	k := max(int(math.Ceil(bboxDiameter/(2*avgEdgeLenRads))), 1)
	// Add a safety margin
	k += 2

	// Get all candidate cells within k steps of the seed
	candidates, err := algos.GridDisk(seed, k)
	if err != nil {
		// Fallback to a larger k if pentagon encountered
		k *= 2
		candidates, err = algos.GridDisk(seed, k)
		if err != nil {
			return nil, err
		}
	}

	// Filter candidates: include only cells whose center is inside the polygon.
	// This matches H3 C library default (CONTAINMENT_CENTER / flags=0).
	var result []h3index.H3Index
	for _, h := range candidates {
		if cellCenterInPolygon(h, res, outerLats, outerLngs, holesLats, holesLngs, bboxPoly) {
			result = append(result, h)
		}
	}

	return result, nil
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
func cellToBBox(h h3index.H3Index, res int, boundary []faceijk.GeoPoint) bbox.BBox {
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
