package coordijk

import (
	"math"
	"testing"

	"github.com/EDRInc/h3-go/internal/constants"
)

// TestIJKNormalize tests the normalization function.
func TestIJKNormalize(t *testing.T) {
	tests := []struct {
		name  string
		input CoordIJK
		want  CoordIJK
	}{
		{
			name:  "already normalized",
			input: CoordIJK{1, 0, 0},
			want:  CoordIJK{1, 0, 0},
		},
		{
			name:  "all equal - normalize to zero",
			input: CoordIJK{2, 2, 2},
			want:  CoordIJK{0, 0, 0},
		},
		{
			name:  "subtract minimum",
			input: CoordIJK{2, 1, 1},
			want:  CoordIJK{1, 0, 0},
		},
		{
			name:  "subtract minimum from all",
			input: CoordIJK{5, 3, 4},
			want:  CoordIJK{2, 0, 1},
		},
		{
			name:  "negative components",
			input: CoordIJK{-1, 2, 1},
			want:  CoordIJK{0, 3, 2},
		},
		{
			name:  "all negative",
			input: CoordIJK{-2, -1, -3},
			want:  CoordIJK{1, 2, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input
			IJKNormalize(&got)
			if got != tt.want {
				t.Errorf("IJKNormalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIJKAdd tests addition of IJK coordinates.
func TestIJKAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b CoordIJK
		want CoordIJK
	}{
		{
			name: "center + center",
			a:    CoordIJK{0, 0, 0},
			b:    CoordIJK{0, 0, 0},
			want: CoordIJK{0, 0, 0},
		},
		{
			name: "unit vectors",
			a:    CoordIJK{1, 0, 0},
			b:    CoordIJK{0, 1, 0},
			want: CoordIJK{1, 1, 0},
		},
		{
			name: "add and normalize",
			a:    CoordIJK{2, 1, 1},
			b:    CoordIJK{1, 2, 1},
			want: CoordIJK{1, 1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IJKAdd(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("IJKAdd(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestIJKSub tests subtraction of IJK coordinates.
func TestIJKSub(t *testing.T) {
	tests := []struct {
		name string
		a, b CoordIJK
		want CoordIJK
	}{
		{
			name: "same coordinates",
			a:    CoordIJK{1, 0, 0},
			b:    CoordIJK{1, 0, 0},
			want: CoordIJK{0, 0, 0},
		},
		{
			name: "simple subtraction",
			a:    CoordIJK{2, 1, 0},
			b:    CoordIJK{1, 0, 0},
			want: CoordIJK{1, 1, 0},
		},
		{
			name: "subtract larger from smaller",
			a:    CoordIJK{0, 1, 0},
			b:    CoordIJK{1, 0, 0},
			want: CoordIJK{0, 2, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IJKSub(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("IJKSub(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestIJKScale tests scaling of IJK coordinates.
func TestIJKScale(t *testing.T) {
	tests := []struct {
		name   string
		input  CoordIJK
		factor int
		want   CoordIJK
	}{
		{
			name:   "scale by zero",
			input:  CoordIJK{1, 2, 3},
			factor: 0,
			want:   CoordIJK{0, 0, 0},
		},
		{
			name:   "scale by one",
			input:  CoordIJK{1, 0, 0},
			factor: 1,
			want:   CoordIJK{1, 0, 0},
		},
		{
			name:   "scale by positive",
			input:  CoordIJK{1, 0, 0},
			factor: 3,
			want:   CoordIJK{3, 0, 0},
		},
		{
			name:   "scale all components",
			input:  CoordIJK{1, 1, 0},
			factor: 2,
			want:   CoordIJK{2, 2, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IJKScale(tt.input, tt.factor)
			if got != tt.want {
				t.Errorf("IJKScale(%v, %d) = %v, want %v", tt.input, tt.factor, got, tt.want)
			}
		})
	}
}

// TestIJKDistance tests distance calculation.
func TestIJKDistance(t *testing.T) {
	tests := []struct {
		name string
		a, b CoordIJK
		want int
	}{
		{
			name: "same point",
			a:    CoordIJK{0, 0, 0},
			b:    CoordIJK{0, 0, 0},
			want: 0,
		},
		{
			name: "adjacent cells - I direction",
			a:    CoordIJK{0, 0, 0},
			b:    CoordIJK{1, 0, 0},
			want: 1,
		},
		{
			name: "adjacent cells - J direction",
			a:    CoordIJK{0, 0, 0},
			b:    CoordIJK{0, 1, 0},
			want: 1,
		},
		{
			name: "adjacent cells - K direction",
			a:    CoordIJK{0, 0, 0},
			b:    CoordIJK{0, 0, 1},
			want: 1,
		},
		{
			name: "distance 2",
			a:    CoordIJK{0, 0, 0},
			b:    CoordIJK{2, 0, 0},
			want: 2,
		},
		{
			name: "diagonal distance",
			a:    CoordIJK{0, 0, 0},
			b:    CoordIJK{1, 1, 0},
			want: 1, // After normalization, max component is 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IJKDistance(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("IJKDistance(%v, %v) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestRotate60CCW tests counter-clockwise rotation.
func TestRotate60CCW(t *testing.T) {
	tests := []struct {
		name  string
		input CoordIJK
		want  CoordIJK
	}{
		{
			name:  "center stays center",
			input: CoordIJK{0, 0, 0},
			want:  CoordIJK{0, 0, 0},
		},
		{
			name:  "rotate I axis",
			input: CoordIJK{1, 0, 0},
			want:  CoordIJK{1, 1, 0},
		},
		{
			name:  "rotate J axis",
			input: CoordIJK{0, 1, 0},
			want:  CoordIJK{0, 1, 1},
		},
		{
			name:  "rotate K axis",
			input: CoordIJK{0, 0, 1},
			want:  CoordIJK{1, 0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Rotate60CCW(tt.input)
			if got != tt.want {
				t.Errorf("Rotate60CCW(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestRotate60CW tests clockwise rotation.
func TestRotate60CW(t *testing.T) {
	tests := []struct {
		name  string
		input CoordIJK
		want  CoordIJK
	}{
		{
			name:  "center stays center",
			input: CoordIJK{0, 0, 0},
			want:  CoordIJK{0, 0, 0},
		},
		{
			name:  "rotate I axis",
			input: CoordIJK{1, 0, 0},
			want:  CoordIJK{1, 0, 1},
		},
		{
			name:  "rotate J axis",
			input: CoordIJK{0, 1, 0},
			want:  CoordIJK{1, 1, 0},
		},
		{
			name:  "rotate K axis",
			input: CoordIJK{0, 0, 1},
			want:  CoordIJK{0, 1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Rotate60CW(tt.input)
			if got != tt.want {
				t.Errorf("Rotate60CW(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestRotationInverse tests that CW and CCW rotations are inverses.
func TestRotationInverse(t *testing.T) {
	testCases := []CoordIJK{
		{0, 0, 0},
		{1, 0, 0},
		{0, 1, 0},
		{0, 0, 1},
		{1, 1, 0},
		{2, 1, 0},
	}

	for _, tc := range testCases {
		// 6 CCW rotations should return to original
		current := tc
		for i := 0; i < 6; i++ {
			current = Rotate60CCW(current)
		}
		if current != tc {
			t.Errorf("6x Rotate60CCW(%v) = %v, want %v", tc, current, tc)
		}

		// 6 CW rotations should return to original
		current = tc
		for i := 0; i < 6; i++ {
			current = Rotate60CW(current)
		}
		if current != tc {
			t.Errorf("6x Rotate60CW(%v) = %v, want %v", tc, current, tc)
		}

		// CW then CCW should return to original
		rotated := Rotate60CW(tc)
		restored := Rotate60CCW(rotated)
		if restored != tc {
			t.Errorf("Rotate60CW then CCW(%v) = %v, want %v", tc, restored, tc)
		}
	}
}

// TestUpAp7 tests aperture-7 parent operation for Class II.
func TestUpAp7(t *testing.T) {
	tests := []struct {
		name  string
		input CoordIJK
		want  CoordIJK
	}{
		{
			name:  "center",
			input: CoordIJK{0, 0, 0},
			want:  CoordIJK{0, 0, 0},
		},
		{
			name:  "basic case",
			input: CoordIJK{1, 0, 0},
			want:  CoordIJK{0, 0, 0},
		},
		{
			name:  "aperture-7 cell",
			input: CoordIJK{7, 0, 0},
			want:  CoordIJK{3, 1, 0}, // lround((21 - 0) / 7) = 3, lround((7 + 0) / 7) = 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpAp7(tt.input)
			if got != tt.want {
				t.Errorf("UpAp7(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestUpAp7r tests aperture-7 parent operation for Class III.
func TestUpAp7r(t *testing.T) {
	tests := []struct {
		name  string
		input CoordIJK
		want  CoordIJK
	}{
		{
			name:  "center",
			input: CoordIJK{0, 0, 0},
			want:  CoordIJK{0, 0, 0},
		},
		{
			name:  "basic case",
			input: CoordIJK{1, 0, 0},
			want:  CoordIJK{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpAp7r(tt.input)
			if got != tt.want {
				t.Errorf("UpAp7r(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestDownAp7 tests aperture-7 child operation for Class II.
func TestDownAp7(t *testing.T) {
	tests := []struct {
		name  string
		input CoordIJK
		want  CoordIJK
	}{
		{
			name:  "center",
			input: CoordIJK{0, 0, 0},
			want:  CoordIJK{0, 0, 0},
		},
		{
			name:  "expand I",
			input: CoordIJK{1, 0, 0},
			want:  CoordIJK{3, 0, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DownAp7(tt.input)
			if got != tt.want {
				t.Errorf("DownAp7(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestDownAp7r tests aperture-7 child operation for Class III.
func TestDownAp7r(t *testing.T) {
	tests := []struct {
		name  string
		input CoordIJK
		want  CoordIJK
	}{
		{
			name:  "center",
			input: CoordIJK{0, 0, 0},
			want:  CoordIJK{0, 0, 0},
		},
		{
			name:  "expand I",
			input: CoordIJK{1, 0, 0},
			want:  CoordIJK{3, 1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DownAp7r(tt.input)
			if got != tt.want {
				t.Errorf("DownAp7r(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestUpDownInverse tests that UpAp7 and DownAp7 are related.
func TestUpDownInverse(t *testing.T) {
	// DownAp7 followed by UpAp7 should return to origin for center
	start := CoordIJK{0, 0, 0}
	down := DownAp7(start)
	up := UpAp7(down)
	if up != start {
		t.Errorf("DownAp7 then UpAp7 on center = %v, want %v", up, start)
	}

	// Same for Class III
	down = DownAp7r(start)
	up = UpAp7r(down)
	if up != start {
		t.Errorf("DownAp7r then UpAp7r on center = %v, want %v", up, start)
	}
}

// TestIJKToIJ tests conversion from IJK to IJ.
func TestIJKToIJ(t *testing.T) {
	tests := []struct {
		name         string
		input        CoordIJK
		wantI, wantJ int
	}{
		{
			name:  "center",
			input: CoordIJK{0, 0, 0},
			wantI: 0, wantJ: 0,
		},
		{
			name:  "I axis",
			input: CoordIJK{1, 0, 0},
			wantI: 1, wantJ: 0,
		},
		{
			name:  "J axis",
			input: CoordIJK{0, 1, 0},
			wantI: 0, wantJ: 1,
		},
		{
			name:  "K axis",
			input: CoordIJK{0, 0, 1},
			wantI: -1, wantJ: -1,
		},
		{
			name:  "with K component",
			input: CoordIJK{2, 3, 1},
			wantI: 1, wantJ: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotI, gotJ := IJKToIJ(tt.input)
			if gotI != tt.wantI || gotJ != tt.wantJ {
				t.Errorf("IJKToIJ(%v) = (%d, %d), want (%d, %d)",
					tt.input, gotI, gotJ, tt.wantI, tt.wantJ)
			}
		})
	}
}

// TestIJToIJK tests conversion from IJ to IJK.
func TestIJToIJK(t *testing.T) {
	tests := []struct {
		name string
		i, j int
		want CoordIJK
	}{
		{
			name: "origin",
			i:    0, j: 0,
			want: CoordIJK{0, 0, 0},
		},
		{
			name: "positive I",
			i:    1, j: 0,
			want: CoordIJK{1, 0, 0},
		},
		{
			name: "positive J",
			i:    0, j: 1,
			want: CoordIJK{0, 1, 0},
		},
		{
			name: "both positive",
			i:    1, j: 1,
			want: CoordIJK{1, 1, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IJToIJK(tt.i, tt.j)
			if got != tt.want {
				t.Errorf("IJToIJK(%d, %d) = %v, want %v", tt.i, tt.j, got, tt.want)
			}
		})
	}
}

// TestToVec2d tests conversion from IJK to Vec2d.
func TestToVec2d(t *testing.T) {
	tests := []struct {
		name         string
		input        CoordIJK
		wantX, wantY float64
	}{
		{
			name:  "center",
			input: CoordIJK{0, 0, 0},
			wantX: 0, wantY: 0,
		},
		{
			name:  "I axis",
			input: CoordIJK{1, 0, 0},
			wantX: 1, wantY: 0,
		},
		{
			name:  "J axis",
			input: CoordIJK{0, 1, 0},
			wantX: -0.5, wantY: constants.M_SQRT3_2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToVec2d(tt.input)
			if math.Abs(got.X-tt.wantX) > constants.EPSILON ||
				math.Abs(got.Y-tt.wantY) > constants.EPSILON {
				t.Errorf("ToVec2d(%v) = (%f, %f), want (%f, %f)",
					tt.input, got.X, got.Y, tt.wantX, tt.wantY)
			}
		})
	}
}

// TestUNIT_VECS verifies the unit vector array.
func TestUNIT_VECS(t *testing.T) {
	expectedVecs := [7]CoordIJK{
		{0, 0, 0}, // CENTER
		{0, 0, 1}, // K
		{0, 1, 0}, // J
		{0, 1, 1}, // JK
		{1, 0, 0}, // I
		{1, 0, 1}, // IK
		{1, 1, 0}, // IJ
	}

	for i, expected := range expectedVecs {
		if UNIT_VECS[i] != expected {
			t.Errorf("UNIT_VECS[%d] = %v, want %v", i, UNIT_VECS[i], expected)
		}
	}
}
