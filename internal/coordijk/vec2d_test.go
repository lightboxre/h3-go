package coordijk

import (
	"math"
	"testing"

	"github.com/lightboxre/h3-go/internal/constants"
)

// TestV2dMag tests vector magnitude calculation.
func TestV2dMag(t *testing.T) {
	tests := []struct {
		name string
		v    Vec2d
		want float64
	}{
		{
			name: "zero vector",
			v:    Vec2d{0, 0},
			want: 0,
		},
		{
			name: "unit X",
			v:    Vec2d{1, 0},
			want: 1,
		},
		{
			name: "unit Y",
			v:    Vec2d{0, 1},
			want: 1,
		},
		{
			name: "3-4-5 triangle",
			v:    Vec2d{3, 4},
			want: 5,
		},
		{
			name: "negative components",
			v:    Vec2d{-3, -4},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := V2dMag(tt.v)
			if math.Abs(got-tt.want) > constants.EPSILON {
				t.Errorf("V2dMag(%v) = %f, want %f", tt.v, got, tt.want)
			}
		})
	}
}

// TestV2dIntersect tests line segment intersection.
func TestV2dIntersect(t *testing.T) {
	tests := []struct {
		name           string
		p0, p1, p2, p3 Vec2d
		wantX, wantY   float64
	}{
		{
			name: "perpendicular lines at origin",
			p0:   Vec2d{-1, 0}, p1: Vec2d{1, 0},
			p2: Vec2d{0, -1}, p3: Vec2d{0, 1},
			wantX: 0, wantY: 0,
		},
		{
			name: "diagonal intersection",
			p0:   Vec2d{0, 0}, p1: Vec2d{2, 2},
			p2: Vec2d{0, 2}, p3: Vec2d{2, 0},
			wantX: 1, wantY: 1,
		},
		{
			name: "intersection at (1, 0.5)",
			p0:   Vec2d{0, 0}, p1: Vec2d{2, 1},
			p2: Vec2d{1, 0}, p3: Vec2d{1, 1},
			wantX: 1, wantY: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := V2dIntersect(tt.p0, tt.p1, tt.p2, tt.p3)
			// Use looser tolerance for accumulated floating-point errors
			tolerance := 1e-14
			if math.Abs(got.X-tt.wantX) > tolerance ||
				math.Abs(got.Y-tt.wantY) > tolerance {
				t.Errorf("V2dIntersect() = (%f, %f), want (%f, %f)",
					got.X, got.Y, tt.wantX, tt.wantY)
			}
		})
	}
}

// TestV2dEquals tests vector equality comparison.
func TestV2dEquals(t *testing.T) {
	tests := []struct {
		name   string
		v1, v2 Vec2d
		want   bool
	}{
		{
			name: "identical vectors",
			v1:   Vec2d{1, 2},
			v2:   Vec2d{1, 2},
			want: true,
		},
		{
			name: "different vectors",
			v1:   Vec2d{1, 2},
			v2:   Vec2d{3, 4},
			want: false,
		},
		{
			name: "within epsilon",
			v1:   Vec2d{1.0, 2.0},
			v2:   Vec2d{1.0 + constants.EPSILON/2, 2.0},
			want: true,
		},
		{
			name: "outside epsilon",
			v1:   Vec2d{1.0, 2.0},
			v2:   Vec2d{1.0 + constants.EPSILON*2, 2.0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := V2dEquals(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("V2dEquals(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

// TestPointSquareDist tests squared distance calculation.
func TestPointSquareDist(t *testing.T) {
	tests := []struct {
		name   string
		v1, v2 Vec2d
		want   float64
	}{
		{
			name: "same point",
			v1:   Vec2d{0, 0},
			v2:   Vec2d{0, 0},
			want: 0,
		},
		{
			name: "unit distance X",
			v1:   Vec2d{0, 0},
			v2:   Vec2d{1, 0},
			want: 1,
		},
		{
			name: "unit distance Y",
			v1:   Vec2d{0, 0},
			v2:   Vec2d{0, 1},
			want: 1,
		},
		{
			name: "3-4-5 triangle (squared = 25)",
			v1:   Vec2d{0, 0},
			v2:   Vec2d{3, 4},
			want: 25,
		},
		{
			name: "negative coordinates",
			v1:   Vec2d{1, 1},
			v2:   Vec2d{-1, -1},
			want: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PointSquareDist(tt.v1, tt.v2)
			if math.Abs(got-tt.want) > constants.EPSILON {
				t.Errorf("PointSquareDist(%v, %v) = %f, want %f", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

// TestVec2dSymmetry tests that operations are symmetric where expected.
func TestVec2dSymmetry(t *testing.T) {
	v1 := Vec2d{1.5, 2.5}
	v2 := Vec2d{3.5, 4.5}

	// PointSquareDist should be symmetric
	dist1 := PointSquareDist(v1, v2)
	dist2 := PointSquareDist(v2, v1)
	if math.Abs(dist1-dist2) > constants.EPSILON {
		t.Errorf("PointSquareDist not symmetric: %f vs %f", dist1, dist2)
	}

	// V2dEquals should be symmetric
	eq1 := V2dEquals(v1, v2)
	eq2 := V2dEquals(v2, v1)
	if eq1 != eq2 {
		t.Errorf("V2dEquals not symmetric: %v vs %v", eq1, eq2)
	}
}

// TestVec2dReflexive tests reflexive property of equality.
func TestVec2dReflexive(t *testing.T) {
	vectors := []Vec2d{
		{0, 0},
		{1, 2},
		{-3.5, 4.7},
		{1e10, 1e-10},
	}

	for _, v := range vectors {
		if !V2dEquals(v, v) {
			t.Errorf("V2dEquals not reflexive for %v", v)
		}
	}
}
