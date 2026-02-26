package h3math

import (
	"math"

	"github.com/h3-native/h3-go/internal/constants"
	"github.com/h3-native/h3-go/internal/faceijk"
	"github.com/h3-native/h3-go/internal/h3index"
)

// CellAreaRads2 returns the area of the H3 cell in steradians (square radians).
// The area is computed as a spherical polygon using the cell's boundary vertices.
func CellAreaRads2(h h3index.H3Index) float64 {
	// Get cell boundary polygon
	fijk := faceijk.H3ToFaceIJK(h)
	res := h.Resolution()
	isPent := h3index.IsPentagon(h)
	boundary := faceijk.FaceIJKToGeoBoundary(fijk, res, isPent)

	// Compute spherical polygon area using the spherical excess formula
	return sphericalPolygonArea(boundary)
}

// CellAreaKm2 returns the cell area in square kilometers.
func CellAreaKm2(h h3index.H3Index) float64 {
	r := constants.EarthRadiusKm
	return CellAreaRads2(h) * r * r
}

// CellAreaM2 returns the cell area in square meters.
func CellAreaM2(h h3index.H3Index) float64 {
	r := constants.EarthRadiusM
	return CellAreaRads2(h) * r * r
}

// sphericalPolygonArea computes the area of a spherical polygon in steradians.
// Uses the method of summing spherical triangle areas.
// From C source: triangulates the polygon from the first vertex and sums the areas.
func sphericalPolygonArea(boundary []faceijk.GeoPoint) float64 {
	n := len(boundary)
	if n < 3 {
		return 0
	}

	// Triangulate from the first vertex and sum the areas
	area := 0.0
	for i := 1; i < n-1; i++ {
		area += sphericalTriangleArea(boundary[0], boundary[i], boundary[i+1])
	}

	return math.Abs(area)
}

// sphericalTriangleArea computes the area of a spherical triangle in steradians
// using L'Huilier's theorem.
//
// L'Huilier's theorem states that for a spherical triangle with sides a, b, c
// (measured as angles in radians), the spherical excess E (which equals the area
// on a unit sphere) is:
//
//	E = 4 * atan(sqrt(tan(s/2) * tan((s-a)/2) * tan((s-b)/2) * tan((s-c)/2)))
//
// where s = (a + b + c) / 2 is the semiperimeter.
//
// This formula avoids the numerical instability of Girard's theorem for small triangles.
// Reference: https://en.wikipedia.org/wiki/L'Huilier's_theorem
func sphericalTriangleArea(a, b, c faceijk.GeoPoint) float64 {
	// Calculate the great-circle distances (arc lengths) for each side
	ab := faceijk.GreatCircleDistanceRads(a.Lat, a.Lng, b.Lat, b.Lng)
	bc := faceijk.GreatCircleDistanceRads(b.Lat, b.Lng, c.Lat, c.Lng)
	ca := faceijk.GreatCircleDistanceRads(c.Lat, c.Lng, a.Lat, a.Lng)

	// L'Huilier's theorem
	s := (ab + bc + ca) / 2.0

	// Compute the product inside the square root
	// Protection against numerical errors: if any term is negative, return 0
	tanS2 := math.Tan(s / 2.0)
	tanSA2 := math.Tan((s - ab) / 2.0)
	tanSB2 := math.Tan((s - bc) / 2.0)
	tanSC2 := math.Tan((s - ca) / 2.0)

	t := tanS2 * tanSA2 * tanSB2 * tanSC2

	// Numerical protection: ensure t is non-negative
	if t < 0 {
		t = 0
	}

	// L'Huilier's formula for spherical excess (area on unit sphere)
	return 4.0 * math.Atan(math.Sqrt(t))
}
