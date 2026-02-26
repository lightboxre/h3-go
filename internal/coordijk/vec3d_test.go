package coordijk

import (
	"math"
	"testing"

	"github.com/h3-native/h3-go/internal/constants"
)

// TestV3dMag tests 3D vector magnitude calculation.
func TestV3dMag(t *testing.T) {
	tests := []struct {
		name string
		v    Vec3d
		want float64
	}{
		{
			name: "zero vector",
			v:    Vec3d{0, 0, 0},
			want: 0,
		},
		{
			name: "unit X",
			v:    Vec3d{1, 0, 0},
			want: 1,
		},
		{
			name: "unit Y",
			v:    Vec3d{0, 1, 0},
			want: 1,
		},
		{
			name: "unit Z",
			v:    Vec3d{0, 0, 1},
			want: 1,
		},
		{
			name: "3-4-5 in XY plane with Z=12 (total=13)",
			v:    Vec3d{3, 4, 12},
			want: 13,
		},
		{
			name: "negative components",
			v:    Vec3d{-1, -2, -2},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := V3dMag(tt.v)
			if math.Abs(got-tt.want) > constants.EPSILON {
				t.Errorf("V3dMag(%v) = %f, want %f", tt.v, got, tt.want)
			}
		})
	}
}

// TestV3dDot tests dot product calculation.
func TestV3dDot(t *testing.T) {
	tests := []struct {
		name   string
		v1, v2 Vec3d
		want   float64
	}{
		{
			name: "orthogonal X and Y",
			v1:   Vec3d{1, 0, 0},
			v2:   Vec3d{0, 1, 0},
			want: 0,
		},
		{
			name: "parallel vectors",
			v1:   Vec3d{1, 2, 3},
			v2:   Vec3d{2, 4, 6},
			want: 28, // 1*2 + 2*4 + 3*6 = 2 + 8 + 18 = 28
		},
		{
			name: "unit vectors",
			v1:   Vec3d{1, 0, 0},
			v2:   Vec3d{1, 0, 0},
			want: 1,
		},
		{
			name: "general case",
			v1:   Vec3d{1, 2, 3},
			v2:   Vec3d{4, 5, 6},
			want: 32, // 1*4 + 2*5 + 3*6 = 4 + 10 + 18 = 32
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := V3dDot(tt.v1, tt.v2)
			if math.Abs(got-tt.want) > constants.EPSILON {
				t.Errorf("V3dDot(%v, %v) = %f, want %f", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

// TestV3dNormalize tests vector normalization.
func TestV3dNormalize(t *testing.T) {
	tests := []struct {
		name string
		v    Vec3d
	}{
		{
			name: "unit X",
			v:    Vec3d{1, 0, 0},
		},
		{
			name: "arbitrary vector",
			v:    Vec3d{3, 4, 5},
		},
		{
			name: "negative components",
			v:    Vec3d{-1, -2, -3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := V3dNormalize(tt.v)
			// Check that magnitude is 1 (use looser tolerance for accumulated float errors)
			mag := V3dMag(got)
			if math.Abs(mag-1.0) > 1e-14 {
				t.Errorf("V3dNormalize(%v) magnitude = %f, want 1.0", tt.v, mag)
			}
			// Check that direction is preserved (dot product with original should be positive)
			if V3dDot(got, tt.v) < 0 {
				t.Errorf("V3dNormalize(%v) reversed direction", tt.v)
			}
		})
	}
}

// TestV3dNormalizeZero tests normalization of zero vector.
func TestV3dNormalizeZero(t *testing.T) {
	zero := Vec3d{0, 0, 0}
	got := V3dNormalize(zero)
	if got != zero {
		t.Errorf("V3dNormalize(zero) = %v, want %v", got, zero)
	}
}

// TestPointSquareDist3d tests squared distance calculation in 3D.
func TestPointSquareDist3d(t *testing.T) {
	tests := []struct {
		name   string
		v1, v2 Vec3d
		want   float64
	}{
		{
			name: "same point",
			v1:   Vec3d{0, 0, 0},
			v2:   Vec3d{0, 0, 0},
			want: 0,
		},
		{
			name: "unit distance X",
			v1:   Vec3d{0, 0, 0},
			v2:   Vec3d{1, 0, 0},
			want: 1,
		},
		{
			name: "unit distance Y",
			v1:   Vec3d{0, 0, 0},
			v2:   Vec3d{0, 1, 0},
			want: 1,
		},
		{
			name: "unit distance Z",
			v1:   Vec3d{0, 0, 0},
			v2:   Vec3d{0, 0, 1},
			want: 1,
		},
		{
			name: "3D Pythagorean",
			v1:   Vec3d{0, 0, 0},
			v2:   Vec3d{2, 3, 6},
			want: 49, // 4 + 9 + 36 = 49
		},
		{
			name: "negative coordinates",
			v1:   Vec3d{1, 1, 1},
			v2:   Vec3d{-1, -1, -1},
			want: 12, // (2^2 + 2^2 + 2^2) = 12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PointSquareDist3d(tt.v1, tt.v2)
			if math.Abs(got-tt.want) > constants.EPSILON {
				t.Errorf("PointSquareDist3d(%v, %v) = %f, want %f", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

// TestGeoToVec3d tests conversion from lat/lng to 3D unit vector.
func TestGeoToVec3d(t *testing.T) {
	tests := []struct {
		name       string
		latRad, lngRad float64
		wantX, wantY, wantZ float64
	}{
		{
			name:   "equator, prime meridian",
			latRad: 0, lngRad: 0,
			wantX: 1, wantY: 0, wantZ: 0,
		},
		{
			name:   "equator, 90° east",
			latRad: 0, lngRad: math.Pi / 2,
			wantX: 0, wantY: 1, wantZ: 0,
		},
		{
			name:   "north pole",
			latRad: math.Pi / 2, lngRad: 0,
			wantX: 0, wantY: 0, wantZ: 1,
		},
		{
			name:   "south pole",
			latRad: -math.Pi / 2, lngRad: 0,
			wantX: 0, wantY: 0, wantZ: -1,
		},
		{
			name:   "equator, 180°",
			latRad: 0, lngRad: math.Pi,
			wantX: -1, wantY: 0, wantZ: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GeoToVec3d(tt.latRad, tt.lngRad)
			// Use looser tolerance for trig functions
			tolerance := 1e-14
			if math.Abs(got.X-tt.wantX) > tolerance ||
				math.Abs(got.Y-tt.wantY) > tolerance ||
				math.Abs(got.Z-tt.wantZ) > tolerance {
				t.Errorf("GeoToVec3d(%f, %f) = (%f, %f, %f), want (%f, %f, %f)",
					tt.latRad, tt.lngRad, got.X, got.Y, got.Z, tt.wantX, tt.wantY, tt.wantZ)
			}
			// Verify it's a unit vector
			mag := V3dMag(got)
			if math.Abs(mag-1.0) > tolerance {
				t.Errorf("GeoToVec3d(%f, %f) magnitude = %f, want 1.0", tt.latRad, tt.lngRad, mag)
			}
		})
	}
}

// TestVec3dToGeo tests conversion from 3D unit vector to lat/lng.
func TestVec3dToGeo(t *testing.T) {
	tests := []struct {
		name string
		v    Vec3d
		wantLatRad, wantLngRad float64
	}{
		{
			name: "equator, prime meridian",
			v:    Vec3d{1, 0, 0},
			wantLatRad: 0, wantLngRad: 0,
		},
		{
			name: "equator, 90° east",
			v:    Vec3d{0, 1, 0},
			wantLatRad: 0, wantLngRad: math.Pi / 2,
		},
		{
			name: "north pole",
			v:    Vec3d{0, 0, 1},
			wantLatRad: math.Pi / 2, wantLngRad: 0,
		},
		{
			name: "south pole",
			v:    Vec3d{0, 0, -1},
			wantLatRad: -math.Pi / 2, wantLngRad: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLat, gotLng := Vec3dToGeo(tt.v)
			if math.Abs(gotLat-tt.wantLatRad) > constants.EPSILON ||
				math.Abs(gotLng-tt.wantLngRad) > constants.EPSILON {
				t.Errorf("Vec3dToGeo(%v) = (%f, %f), want (%f, %f)",
					tt.v, gotLat, gotLng, tt.wantLatRad, tt.wantLngRad)
			}
		})
	}
}

// TestGeoVec3dRoundtrip tests that GeoToVec3d and Vec3dToGeo are inverses.
func TestGeoVec3dRoundtrip(t *testing.T) {
	testCases := []struct {
		latRad, lngRad float64
	}{
		{0, 0},
		{math.Pi / 4, math.Pi / 4},
		{-math.Pi / 4, math.Pi},
		{math.Pi / 6, -math.Pi / 3},
	}

	for _, tc := range testCases {
		t.Run("roundtrip", func(t *testing.T) {
			vec := GeoToVec3d(tc.latRad, tc.lngRad)
			gotLat, gotLng := Vec3dToGeo(vec)
			// Use looser tolerance for accumulated trig errors
			tolerance := 1e-14
			if math.Abs(gotLat-tc.latRad) > tolerance ||
				math.Abs(gotLng-tc.lngRad) > tolerance {
				t.Errorf("Roundtrip (%f, %f) -> %v -> (%f, %f)",
					tc.latRad, tc.lngRad, vec, gotLat, gotLng)
			}
		})
	}
}

// TestVec3dDotSymmetric tests that dot product is symmetric.
func TestVec3dDotSymmetric(t *testing.T) {
	v1 := Vec3d{1.5, 2.5, 3.5}
	v2 := Vec3d{4.5, 5.5, 6.5}

	dot1 := V3dDot(v1, v2)
	dot2 := V3dDot(v2, v1)
	if math.Abs(dot1-dot2) > constants.EPSILON {
		t.Errorf("V3dDot not symmetric: %f vs %f", dot1, dot2)
	}
}

// TestPointSquareDist3dSymmetric tests distance symmetry.
func TestPointSquareDist3dSymmetric(t *testing.T) {
	v1 := Vec3d{1, 2, 3}
	v2 := Vec3d{4, 5, 6}

	dist1 := PointSquareDist3d(v1, v2)
	dist2 := PointSquareDist3d(v2, v1)
	if math.Abs(dist1-dist2) > constants.EPSILON {
		t.Errorf("PointSquareDist3d not symmetric: %f vs %f", dist1, dist2)
	}
}
