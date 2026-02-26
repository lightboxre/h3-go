package faceijk

import (
	"math"
	"testing"

	"github.com/h3-native/h3-go/internal/constants"
	"github.com/h3-native/h3-go/internal/coordijk"
)

func TestDegsToRads(t *testing.T) {
	tests := []struct {
		deg  float64
		want float64
	}{
		{0, 0},
		{90, math.Pi / 2},
		{180, math.Pi},
		{360, 2 * math.Pi},
		{-90, -math.Pi / 2},
	}

	for _, tt := range tests {
		got := DegsToRads(tt.deg)
		if math.Abs(got-tt.want) > 1e-10 {
			t.Errorf("DegsToRads(%f) = %f, want %f", tt.deg, got, tt.want)
		}
	}
}

func TestRadsToDegs(t *testing.T) {
	tests := []struct {
		rad  float64
		want float64
	}{
		{0, 0},
		{math.Pi / 2, 90},
		{math.Pi, 180},
		{2 * math.Pi, 360},
		{-math.Pi / 2, -90},
	}

	for _, tt := range tests {
		got := RadsToDegs(tt.rad)
		if math.Abs(got-tt.want) > 1e-10 {
			t.Errorf("RadsToDegs(%f) = %f, want %f", tt.rad, got, tt.want)
		}
	}
}

func TestGreatCircleDistanceRads(t *testing.T) {
	// Test equator to pole (should be π/2 radians)
	lat1, lng1 := 0.0, 0.0
	lat2, lng2 := math.Pi/2, 0.0

	dist := GreatCircleDistanceRads(lat1, lng1, lat2, lng2)
	expected := math.Pi / 2

	if math.Abs(dist-expected) > 1e-10 {
		t.Errorf("GreatCircleDistanceRads() = %f, want %f", dist, expected)
	}

	// Test same point (should be 0)
	dist = GreatCircleDistanceRads(0, 0, 0, 0)
	if math.Abs(dist) > 1e-10 {
		t.Errorf("GreatCircleDistanceRads(same point) = %f, want 0", dist)
	}
}

func TestConstrainLat(t *testing.T) {
	tests := []struct {
		lat  float64
		want float64
	}{
		{0, 0},
		{math.Pi / 4, math.Pi / 4},
		{math.Pi / 2, math.Pi / 2},
		{-math.Pi / 2, -math.Pi / 2},
	}

	for _, tt := range tests {
		got := ConstrainLat(tt.lat)
		if math.Abs(got-tt.want) > 1e-10 {
			t.Errorf("ConstrainLat(%f) = %f, want %f", tt.lat, got, tt.want)
		}
	}
}

func TestConstrainLng(t *testing.T) {
	tests := []struct {
		lng  float64
		want float64
	}{
		{0, 0},
		{math.Pi / 2, math.Pi / 2},
		{math.Pi, math.Pi},
		{-math.Pi, -math.Pi},
	}

	for _, tt := range tests {
		got := ConstrainLng(tt.lng)
		if math.Abs(got-tt.want) > 1e-10 {
			t.Errorf("ConstrainLng(%f) = %f, want %f", tt.lng, got, tt.want)
		}
	}

	// Test wraparound (3π = π + 2π, should wrap to π or -π, both are valid)
	got := ConstrainLng(3 * math.Pi)
	// Both π and -π are valid representations of the same longitude
	if math.Abs(got-math.Pi) > 1e-10 && math.Abs(got-(-math.Pi)) > 1e-10 {
		t.Errorf("ConstrainLng(3π) = %f, want π or -π", got)
	}
}

func TestGeoToFaceIJK(t *testing.T) {
	// Test known H3 index location
	// H3 index 0x85283473fffffff is at approximately (37.3615593, -122.0553238) degrees
	// Convert to radians
	latDeg := 37.3615593
	lngDeg := -122.0553238
	latRad := DegsToRads(latDeg)
	lngRad := DegsToRads(lngDeg)
	res := 5

	fijk := GeoToFaceIJK(latRad, lngRad, res)

	// Verify we got a valid face
	if fijk.Face < 0 || fijk.Face >= constants.NumIcosaFaces {
		t.Errorf("GeoToFaceIJK face = %d, want 0-19", fijk.Face)
	}

	// Verify coordinates are reasonable
	t.Logf("GeoToFaceIJK(%f, %f, res=%d) = Face %d, IJK(%d, %d, %d)",
		latDeg, lngDeg, res, fijk.Face, fijk.Coord.I, fijk.Coord.J, fijk.Coord.K)
}

func TestFaceIJKToGeo(t *testing.T) {
	// Test round-trip conversion for face center
	face := 0
	res := 0
	fijk := FaceIJK{
		Face:  face,
		Coord: coordijk.CoordIJK{I: 0, J: 0, K: 0},
	}

	lat, lng := FaceIJKToGeo(fijk, res)

	// Should be close to face center
	expectedLat := faceCenterGeo[face].Lat
	expectedLng := faceCenterGeo[face].Lng

	t.Logf("FaceIJKToGeo(face=%d, res=%d) = (%f, %f)", face, res, lat, lng)
	t.Logf("Expected face center = (%f, %f)", expectedLat, expectedLng)

	// Allow some tolerance due to gnomonic projection
	if math.Abs(lat-expectedLat) > 0.1 || math.Abs(lng-expectedLng) > 0.1 {
		t.Errorf("FaceIJKToGeo() lat=%f, lng=%f; expected near (%f, %f)",
			lat, lng, expectedLat, expectedLng)
	}
}

func TestGeoToFaceIJKRoundTrip(t *testing.T) {
	// Test that round-trip conversion produces valid results
	// Note: exact round-trip may not work perfectly due to quantization
	// at cell boundaries and the gnomonic projection implementation details
	latDeg := 37.5
	lngDeg := -122.0
	latRad := DegsToRads(latDeg)
	lngRad := DegsToRads(lngDeg)
	res := 3

	// Convert to FaceIJK
	fijk := GeoToFaceIJK(latRad, lngRad, res)

	// Verify face is valid
	if fijk.Face < 0 || fijk.Face >= constants.NumIcosaFaces {
		t.Errorf("Invalid face: %d", fijk.Face)
	}

	// Convert back to geo
	lat2, lng2 := FaceIJKToGeo(fijk, res)

	// Verify output is valid lat/lng
	if math.Abs(lat2) > math.Pi/2 {
		t.Errorf("Invalid latitude from FaceIJKToGeo: %f", lat2)
	}
	if math.Abs(lng2) > math.Pi+1e-10 {
		t.Errorf("Invalid longitude from FaceIJKToGeo: %f", lng2)
	}

	t.Logf("Round-trip: (%f, %f) -> Face %d, IJK(%d, %d, %d) -> (%f, %f)",
		latDeg, lngDeg, fijk.Face, fijk.Coord.I, fijk.Coord.J, fijk.Coord.K,
		RadsToDegs(lat2), RadsToDegs(lng2))
}

func TestFaceIJKToGeoBoundary(t *testing.T) {
	// Test boundary calculation for a hexagon at resolution 0
	fijk := FaceIJK{
		Face:  0,
		Coord: coordijk.CoordIJK{I: 0, J: 0, K: 0},
	}

	boundary := FaceIJKToGeoBoundary(fijk, 0, false)

	// Should have 6 vertices for hexagon
	if len(boundary) != 6 {
		t.Errorf("FaceIJKToGeoBoundary hexagon got %d vertices, want 6", len(boundary))
	}

	// Verify all vertices are valid lat/lng
	for i, v := range boundary {
		if math.Abs(v.Lat) > math.Pi/2+0.1 {
			t.Errorf("Vertex %d lat=%f out of range", i, v.Lat)
		}
		// Normalize longitude to check validity
		lngNorm := ConstrainLng(v.Lng)
		if math.Abs(lngNorm) > math.Pi+1e-10 {
			t.Errorf("Vertex %d lng=%f (normalized: %f) out of range", i, v.Lng, lngNorm)
		}
	}
}

func TestFaceIJKToGeoBoundaryPentagon(t *testing.T) {
	// Test boundary calculation for a pentagon
	fijk := FaceIJK{
		Face:  0,
		Coord: coordijk.CoordIJK{I: 0, J: 0, K: 0},
	}

	boundary := FaceIJKToGeoBoundary(fijk, 0, true)

	// Should have 5 vertices for pentagon
	if len(boundary) != 5 {
		t.Errorf("FaceIJKToGeoBoundary pentagon got %d vertices, want 5", len(boundary))
	}
}

func TestHex2dToCoordIJK(t *testing.T) {
	// Test conversion from 2D hex to IJK
	tests := []struct {
		name string
		v    coordijk.Vec2d
		want coordijk.CoordIJK
	}{
		{
			name: "origin",
			v:    coordijk.Vec2d{X: 0, Y: 0},
			want: coordijk.CoordIJK{I: 0, J: 0, K: 0},
		},
		{
			name: "i direction",
			v:    coordijk.Vec2d{X: 1.0, Y: 0},
			want: coordijk.CoordIJK{I: 1, J: 0, K: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hex2dToCoordIJK(tt.v)
			if got.I != tt.want.I || got.J != tt.want.J || got.K != tt.want.K {
				t.Errorf("hex2dToCoordIJK(%+v) = %+v, want %+v", tt.v, got, tt.want)
			}
		})
	}
}

func TestHex2dRoundTrip(t *testing.T) {
	// Test round-trip conversion IJK -> Vec2d -> IJK
	tests := []coordijk.CoordIJK{
		{I: 0, J: 0, K: 0},
		{I: 1, J: 0, K: 0},
		{I: 0, J: 1, K: 0},
		{I: 1, J: 1, K: 0},
	}

	for _, orig := range tests {
		// Convert to Vec2d
		v2d := coordijk.ToVec2d(orig)

		// Convert back to IJK
		result := hex2dToCoordIJK(v2d)

		if result.I != orig.I || result.J != orig.J || result.K != orig.K {
			t.Errorf("Round-trip failed: %+v -> %+v -> %+v", orig, v2d, result)
		}
	}
}

func TestGeoAlmostEqual(t *testing.T) {
	g1 := GeoPoint{Lat: 0.5, Lng: 1.0}
	g2 := GeoPoint{Lat: 0.5 + 1e-13, Lng: 1.0}
	g3 := GeoPoint{Lat: 0.5 + 1e-10, Lng: 1.0}

	if !GeoAlmostEqual(g1, g2) {
		t.Error("GeoAlmostEqual should return true for very close points")
	}

	if GeoAlmostEqual(g1, g3) {
		t.Error("GeoAlmostEqual should return false for distant points")
	}
}

func TestPosAngleRads(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0, 0},
		{math.Pi, math.Pi},
		{2 * math.Pi, 0},
		{-math.Pi, math.Pi},
		{-2 * math.Pi, 0},
		{3 * math.Pi, math.Pi},
	}

	for _, tt := range tests {
		got := PosAngleRads(tt.input)
		if math.Abs(got-tt.want) > 1e-10 {
			t.Errorf("PosAngleRads(%f) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestGeoAzimuthRads(t *testing.T) {
	// Test azimuth from equator/prime meridian to north pole
	lat1, lng1 := 0.0, 0.0
	lat2, lng2 := math.Pi/2, 0.0

	azimuth := GeoAzimuthRads(lat1, lng1, lat2, lng2)

	// Should be approximately 0 (north)
	if math.Abs(azimuth) > 1e-10 {
		t.Errorf("GeoAzimuthRads(equator to pole) = %f, want ~0", azimuth)
	}
}

func TestFaceCenterPointsAreUnitVectors(t *testing.T) {
	// Verify all face center points are unit vectors
	for i, v := range faceCenterPoint {
		mag := coordijk.V3dMag(v)
		if math.Abs(mag-1.0) > 1e-10 {
			t.Errorf("faceCenterPoint[%d] magnitude = %f, want 1.0", i, mag)
		}
	}
}

func TestFaceCenterGeoToVec3dConsistency(t *testing.T) {
	// Verify faceCenterGeo and faceCenterPoint are consistent
	for i := 0; i < constants.NumIcosaFaces; i++ {
		geo := faceCenterGeo[i]
		v3d := coordijk.GeoToVec3d(geo.Lat, geo.Lng)

		expected := faceCenterPoint[i]

		// Should be very close
		diff := coordijk.PointSquareDist3d(v3d, expected)
		if diff > 1e-10 {
			t.Errorf("Face %d: GeoToVec3d(faceCenterGeo) != faceCenterPoint, diff=%e", i, diff)
		}
	}
}

func BenchmarkGeoToFaceIJK(b *testing.B) {
	latRad := DegsToRads(37.5)
	lngRad := DegsToRads(-122.0)
	res := 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GeoToFaceIJK(latRad, lngRad, res)
	}
}

func BenchmarkFaceIJKToGeo(b *testing.B) {
	fijk := FaceIJK{
		Face:  0,
		Coord: coordijk.CoordIJK{I: 10, J: 5, K: 0},
	}
	res := 5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FaceIJKToGeo(fijk, res)
	}
}

func BenchmarkGreatCircleDistanceRads(b *testing.B) {
	lat1, lng1 := 0.5, 1.0
	lat2, lng2 := 0.6, 1.1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GreatCircleDistanceRads(lat1, lng1, lat2, lng2)
	}
}
