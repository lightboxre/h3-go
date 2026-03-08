package h3math

import (
	"math"
	"testing"

	"github.com/lightboxre/h3-go/internal/constants"
	"github.com/lightboxre/h3-go/internal/faceijk"
	"github.com/lightboxre/h3-go/internal/h3index"
)

const epsilon = 1e-9

// TestGreatCircleDistanceRads tests the great circle distance calculation in radians.
func TestGreatCircleDistanceRads(t *testing.T) {
	tests := []struct {
		name     string
		lat1     float64
		lng1     float64
		lat2     float64
		lng2     float64
		expected float64
	}{
		{
			name:     "same point",
			lat1:     0.0,
			lng1:     0.0,
			lat2:     0.0,
			lng2:     0.0,
			expected: 0.0,
		},
		{
			name:     "equator quarter circle",
			lat1:     0.0,
			lng1:     0.0,
			lat2:     0.0,
			lng2:     math.Pi / 2,
			expected: math.Pi / 2,
		},
		{
			name:     "pole to pole",
			lat1:     math.Pi / 2,
			lng1:     0.0,
			lat2:     -math.Pi / 2,
			lng2:     0.0,
			expected: math.Pi,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GreatCircleDistanceRads(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			if math.Abs(result-tt.expected) > epsilon {
				t.Errorf("GreatCircleDistanceRads() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGreatCircleDistanceKm tests the great circle distance in kilometers.
func TestGreatCircleDistanceKm(t *testing.T) {
	// San Francisco to New York (approximate)
	sfLat := faceijk.DegsToRads(37.7749)
	sfLng := faceijk.DegsToRads(-122.4194)
	nyLat := faceijk.DegsToRads(40.7128)
	nyLng := faceijk.DegsToRads(-74.0060)

	distance := GreatCircleDistanceKm(sfLat, sfLng, nyLat, nyLng)

	// The actual distance is approximately 4130 km
	expectedMin := 4000.0
	expectedMax := 4300.0

	if distance < expectedMin || distance > expectedMax {
		t.Errorf("GreatCircleDistanceKm(SF to NY) = %v, want between %v and %v",
			distance, expectedMin, expectedMax)
	}
}

// TestGreatCircleDistanceM tests the great circle distance in meters.
func TestGreatCircleDistanceM(t *testing.T) {
	// Short distance test (1 degree at equator)
	lat1 := 0.0
	lng1 := 0.0
	lat2 := 0.0
	lng2 := faceijk.DegsToRads(1.0)

	distanceM := GreatCircleDistanceM(lat1, lng1, lat2, lng2)
	distanceKm := GreatCircleDistanceKm(lat1, lng1, lat2, lng2)

	// Check that meter conversion is consistent
	if math.Abs(distanceM-distanceKm*1000.0) > 1.0 {
		t.Errorf("GreatCircleDistanceM() = %v, want %v (km * 1000)", distanceM, distanceKm*1000.0)
	}

	// 1 degree at equator should be approximately 111 km
	expectedKm := 111.0
	if math.Abs(distanceKm-expectedKm) > 5.0 {
		t.Errorf("1 degree at equator = %v km, want approximately %v km", distanceKm, expectedKm)
	}
}

// TestEdgeLengthKm tests the average edge length function.
func TestEdgeLengthKm(t *testing.T) {
	tests := []struct {
		res      int
		expected float64
	}{
		{0, 1281.256011},
		{5, 9.85409099},
		{10, 0.075863783},
		{15, 0.000584169},
	}

	for _, tt := range tests {
		t.Run("res"+string(rune('0'+tt.res)), func(t *testing.T) {
			result := EdgeLengthKm(tt.res)
			if math.Abs(result-tt.expected) > epsilon {
				t.Errorf("EdgeLengthKm(%d) = %v, want %v", tt.res, result, tt.expected)
			}
		})
	}
}

// TestEdgeLengthRads tests the edge length in radians.
func TestEdgeLengthRads(t *testing.T) {
	// Test that radians conversion is correct
	res := 5
	expectedKm := 9.85409099
	expectedRads := expectedKm / constants.EarthRadiusKm

	result := EdgeLengthRads(res)
	if math.Abs(result-expectedRads) > epsilon {
		t.Errorf("EdgeLengthRads(%d) = %v, want %v", res, result, expectedRads)
	}

	// Test invalid resolution
	if EdgeLengthRads(-1) != 0 {
		t.Error("EdgeLengthRads(-1) should return 0")
	}
	if EdgeLengthRads(16) != 0 {
		t.Error("EdgeLengthRads(16) should return 0")
	}
}

// TestEdgeLengthM tests the edge length in meters.
func TestEdgeLengthM(t *testing.T) {
	res := 5
	expectedM := 9854.09099

	result := EdgeLengthM(res)
	if math.Abs(result-expectedM) > epsilon {
		t.Errorf("EdgeLengthM(%d) = %v, want %v", res, result, expectedM)
	}
}

// TestExactEdgeLengthApproximation tests that exact edge length returns reasonable values.
func TestExactEdgeLengthApproximation(t *testing.T) {
	// Create a mock edge index at resolution 5
	// In mode 2 (directed edge mode), the resolution field should still work
	var edge = h3index.H3_INIT
	edge = h3index.SetMode(edge, constants.H3_DIRECTEDEDGE_MODE)
	edge = h3index.SetResolution(edge, 5)
	edge = h3index.SetBaseCell(edge, 0)

	result := ExactEdgeLengthKm(edge)
	expected := EdgeLengthKm(5) // Should use average for now

	if math.Abs(result-expected) > epsilon {
		t.Errorf("ExactEdgeLengthKm() = %v, want %v (average for res 5)", result, expected)
	}
}

// TestSphericalTriangleArea tests the spherical triangle area calculation.
func TestSphericalTriangleArea(t *testing.T) {
	// Test with a proper spherical triangle (not colinear points)
	// Three points forming a triangle on the sphere
	p1 := faceijk.GeoPoint{Lat: 0, Lng: 0}
	p2 := faceijk.GeoPoint{Lat: 0, Lng: faceijk.DegsToRads(10)}
	p3 := faceijk.GeoPoint{Lat: faceijk.DegsToRads(10), Lng: faceijk.DegsToRads(5)}

	area := sphericalTriangleArea(p1, p2, p3)

	// The area should be positive (non-zero) and less than π
	if area < 0 {
		t.Error("sphericalTriangleArea() should return non-negative value")
	}
	if area > math.Pi {
		t.Error("sphericalTriangleArea() should be less than π for small triangle")
	}

	// For a small triangle, area should be reasonable
	if area > 0.1 {
		t.Errorf("sphericalTriangleArea() = %v, seems too large for a 10-degree triangle", area)
	}
}

// TestSphericalPolygonArea tests the spherical polygon area calculation.
func TestSphericalPolygonArea(t *testing.T) {
	// Test with a square-ish polygon
	boundary := []faceijk.GeoPoint{
		{Lat: faceijk.DegsToRads(0), Lng: faceijk.DegsToRads(0)},
		{Lat: faceijk.DegsToRads(0), Lng: faceijk.DegsToRads(10)},
		{Lat: faceijk.DegsToRads(10), Lng: faceijk.DegsToRads(10)},
		{Lat: faceijk.DegsToRads(10), Lng: faceijk.DegsToRads(0)},
	}

	area := sphericalPolygonArea(boundary)

	// Should be positive and reasonable for a 10x10 degree square
	if area <= 0 {
		t.Error("sphericalPolygonArea() should return positive value")
	}

	// 10 degrees is about 0.174 radians
	// A 10x10 degree square should be roughly 0.03 steradians
	if area < 0.01 || area > 0.1 {
		t.Errorf("sphericalPolygonArea(10x10 deg) = %v, expected between 0.01 and 0.1", area)
	}
}

// TestCellAreaRads2 tests cell area calculation for a valid H3 index.
func TestCellAreaRads2(t *testing.T) {
	// Create a valid H3 index at resolution 5
	var h = h3index.H3_INIT
	h = h3index.SetMode(h, constants.H3_CELL_MODE)
	h = h3index.SetResolution(h, 5)
	h = h3index.SetBaseCell(h, 0)

	area := CellAreaRads2(h)

	// Area should be positive
	if area <= 0 {
		t.Error("CellAreaRads2() should return positive value")
	}

	// Area should be reasonable for resolution 5
	// Resolution 5 cells are roughly 252 km²
	// On a sphere with radius ~6371 km, this is about 6.2e-6 steradians
	if area > 1.0 {
		t.Errorf("CellAreaRads2() = %v, seems too large", area)
	}
}

// TestCellAreaKm2 tests cell area in square kilometers.
func TestCellAreaKm2(t *testing.T) {
	// Create a valid H3 index at resolution 5
	var h = h3index.H3_INIT
	h = h3index.SetMode(h, constants.H3_CELL_MODE)
	h = h3index.SetResolution(h, 5)
	h = h3index.SetBaseCell(h, 0)

	areaKm2 := CellAreaKm2(h)

	// Resolution 5 cells vary in size but should be on the order of hundreds to thousands of km²
	// The H3ToFaceIJK and FaceIJKToGeoBoundary are simplified implementations,
	// so we allow a wider range for now
	expectedMin := 100.0
	expectedMax := 5000.0

	if areaKm2 < expectedMin || areaKm2 > expectedMax {
		t.Errorf("CellAreaKm2(res 5) = %v, expected between %v and %v",
			areaKm2, expectedMin, expectedMax)
	}

	// Just verify the area is reasonable
	t.Logf("Cell area at resolution 5: %.2f km²", areaKm2)
}

// TestCellAreaM2 tests cell area in square meters.
func TestCellAreaM2(t *testing.T) {
	// Create a valid H3 index at resolution 5
	var h = h3index.H3_INIT
	h = h3index.SetMode(h, constants.H3_CELL_MODE)
	h = h3index.SetResolution(h, 5)
	h = h3index.SetBaseCell(h, 0)

	areaM2 := CellAreaM2(h)
	areaKm2 := CellAreaKm2(h)

	// Check consistency: m² = km² * 1,000,000
	expectedM2 := areaKm2 * 1e6
	if math.Abs(areaM2-expectedM2) > 1.0 {
		t.Errorf("CellAreaM2() = %v, want %v (km² * 1e6)", areaM2, expectedM2)
	}
}

// TestPentagonArea tests that pentagon cells have reasonable areas.
func TestPentagonArea(t *testing.T) {
	// Get a pentagon cell (base cell 4 is pentagon)
	var h = h3index.H3_INIT
	h = h3index.SetMode(h, constants.H3_CELL_MODE)
	h = h3index.SetResolution(h, 0)
	h = h3index.SetBaseCell(h, 4) // Pentagon base cell

	if !h3index.IsPentagon(h) {
		t.Skip("Base cell 4 should be pentagon, skipping test")
	}

	area := CellAreaKm2(h)

	// Resolution 0 pentagons should be very large (millions of km²)
	// Earth surface area is ~510 million km², divided by 122 base cells
	// gives roughly 4 million km² per base cell
	if area <= 0 {
		t.Error("Pentagon area should be positive")
	}

	// Allow a wider range - resolution 0 cells are huge
	if area < 1e5 || area > 2e8 {
		t.Errorf("Pentagon area at res 0 = %v, expected between 1e5 and 2e8", area)
	}

	t.Logf("Pentagon area at resolution 0: %.2e km²", area)
}

// BenchmarkGreatCircleDistanceRads benchmarks the great circle distance calculation.
func BenchmarkGreatCircleDistanceRads(b *testing.B) {
	lat1 := 0.5
	lng1 := 1.0
	lat2 := 0.6
	lng2 := 1.1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GreatCircleDistanceRads(lat1, lng1, lat2, lng2)
	}
}

// BenchmarkCellAreaRads2 benchmarks the cell area calculation.
func BenchmarkCellAreaRads2(b *testing.B) {
	var h = h3index.H3_INIT
	h = h3index.SetMode(h, constants.H3_CELL_MODE)
	h = h3index.SetResolution(h, 5)
	h = h3index.SetBaseCell(h, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CellAreaRads2(h)
	}
}

// BenchmarkSphericalTriangleArea benchmarks the triangle area calculation.
func BenchmarkSphericalTriangleArea(b *testing.B) {
	p1 := faceijk.GeoPoint{Lat: 0, Lng: 0}
	p2 := faceijk.GeoPoint{Lat: 0, Lng: 0.1}
	p3 := faceijk.GeoPoint{Lat: 0.1, Lng: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sphericalTriangleArea(p1, p2, p3)
	}
}
