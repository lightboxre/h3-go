// Package coordijk implements IJK+ hexagonal coordinate system operations.
// This coordinate system is used throughout H3 for grid navigation.
package coordijk

import (
	"math"

	"github.com/lightboxre/h3-go/internal/constants"
)

// mOneSeventh is 1/7, used in aperture-7 parent calculations.
const mOneSeventh = 0.14285714285714285714285714285714285

// CoordIJK represents coordinates in the IJK+ coordinate system.
// The constraint: after normalization, at least one component is 0.
// This is a 3-axis hexagonal coordinate system where each hex has 6 neighbors.
type CoordIJK struct {
	I, J, K int
}

// UNIT_VECS are direction digits corresponding to H3 digit constants.
// Index meanings: CENTER=0, K=1, J=2, JK=3, I=4, IK=5, IJ=6
var UNIT_VECS = [7]CoordIJK{ //nolint:revive // C-compatible name used in test files
	{0, 0, 0}, // 0 = CENTER
	{0, 0, 1}, // 1 = K
	{0, 1, 0}, // 2 = J
	{0, 1, 1}, // 3 = JK
	{1, 0, 0}, // 4 = I
	{1, 0, 1}, // 5 = IK
	{1, 1, 0}, // 6 = IJ
}

// IJKNormalize normalizes IJK coordinates by removing common additive factor.
// From C: if any component is negative, redistribute to make all non-negative,
// then subtract the minimum from all components to ensure at least one is 0.
func IJKNormalize(c *CoordIJK) {
	// Remove any negative components by redistributing
	if c.I < 0 {
		c.J -= c.I
		c.K -= c.I
		c.I = 0
	}
	if c.J < 0 {
		c.I -= c.J
		c.K -= c.J
		c.J = 0
	}
	if c.K < 0 {
		c.I -= c.K
		c.J -= c.K
		c.K = 0
	}

	// Find minimum and subtract from all
	minVal := min(c.I, c.J, c.K)
	if minVal > 0 {
		c.I -= minVal
		c.J -= minVal
		c.K -= minVal
	}
}

// IJKAdd adds two IJK coordinates and normalizes the result.
func IJKAdd(a, b CoordIJK) CoordIJK {
	result := CoordIJK{
		I: a.I + b.I,
		J: a.J + b.J,
		K: a.K + b.K,
	}
	IJKNormalize(&result)
	return result
}

// IJKSub subtracts b from a and normalizes.
func IJKSub(a, b CoordIJK) CoordIJK {
	result := CoordIJK{
		I: a.I - b.I,
		J: a.J - b.J,
		K: a.K - b.K,
	}
	IJKNormalize(&result)
	return result
}

// IJKScale scales an IJK coordinate by a factor.
func IJKScale(c CoordIJK, factor int) CoordIJK {
	result := CoordIJK{
		I: c.I * factor,
		J: c.J * factor,
		K: c.K * factor,
	}
	IJKNormalize(&result)
	return result
}

// IJKDistance returns the H3 "grid distance" between two IJK coordinates.
// From C: subtract, normalize, take max of absolute values.
func IJKDistance(a, b CoordIJK) int {
	diff := IJKSub(a, b)
	absI := diff.I
	if absI < 0 {
		absI = -absI
	}
	absJ := diff.J
	if absJ < 0 {
		absJ = -absJ
	}
	absK := diff.K
	if absK < 0 {
		absK = -absK
	}

	return max(absI, absJ, absK)
}

// Rotate60CCW rotates an IJK coordinate 60 degrees counter-clockwise.
// From C: scales unit vectors {1,1,0}, {0,1,1}, {1,0,1} by i,j,k components.
func Rotate60CCW(c CoordIJK) CoordIJK {
	// Rotation using unit vector scaling
	iVec := CoordIJK{1, 1, 0}
	jVec := CoordIJK{0, 1, 1}
	kVec := CoordIJK{1, 0, 1}

	iScaled := IJKScale(iVec, c.I)
	jScaled := IJKScale(jVec, c.J)
	kScaled := IJKScale(kVec, c.K)

	result := IJKAdd(IJKAdd(iScaled, jScaled), kScaled)
	return result
}

// Rotate60CW rotates an IJK coordinate 60 degrees clockwise.
// From C: scales unit vectors {1,0,1}, {1,1,0}, {0,1,1} by i,j,k components.
func Rotate60CW(c CoordIJK) CoordIJK {
	// Rotation using unit vector scaling
	iVec := CoordIJK{1, 0, 1}
	jVec := CoordIJK{1, 1, 0}
	kVec := CoordIJK{0, 1, 1}

	iScaled := IJKScale(iVec, c.I)
	jScaled := IJKScale(jVec, c.J)
	kScaled := IJKScale(kVec, c.K)

	result := IJKAdd(IJKAdd(iScaled, jScaled), kScaled)
	return result
}

// UpAp7 moves IJK coordinates one aperture-7 resolution up (parent, Class II).
// From C: converts to IJ, applies formulas with lround((3*i - j) / 7), normalizes.
func UpAp7(c CoordIJK) CoordIJK {
	i := c.I - c.K
	j := c.J - c.K

	result := CoordIJK{
		I: int(math.Round(float64(3*i-j) * mOneSeventh)),
		J: int(math.Round(float64(i+2*j) * mOneSeventh)),
		K: 0,
	}
	IJKNormalize(&result)
	return result
}

// UpAp7r moves IJK coordinates one aperture-7 resolution up (parent, Class III).
// From C: converts to IJ, applies formulas with lround((2*i + j) / 7), normalizes.
func UpAp7r(c CoordIJK) CoordIJK {
	i := c.I - c.K
	j := c.J - c.K

	result := CoordIJK{
		I: int(math.Round(float64(2*i+j) * mOneSeventh)),
		J: int(math.Round(float64(3*j-i) * mOneSeventh)),
		K: 0,
	}
	IJKNormalize(&result)
	return result
}

// DownAp7 moves IJK coordinates one aperture-7 resolution down (child, Class II).
// From C: scales unit vectors {3,0,1}, {1,3,0}, {0,1,3} by i,j,k and combines.
func DownAp7(c CoordIJK) CoordIJK {
	iVec := CoordIJK{3, 0, 1}
	jVec := CoordIJK{1, 3, 0}
	kVec := CoordIJK{0, 1, 3}

	iScaled := IJKScale(iVec, c.I)
	jScaled := IJKScale(jVec, c.J)
	kScaled := IJKScale(kVec, c.K)

	result := IJKAdd(IJKAdd(iScaled, jScaled), kScaled)
	return result
}

// DownAp7r moves IJK coordinates one aperture-7 resolution down (child, Class III).
// From C: scales unit vectors {3,1,0}, {0,3,1}, {1,0,3} by i,j,k and combines.
func DownAp7r(c CoordIJK) CoordIJK {
	iVec := CoordIJK{3, 1, 0}
	jVec := CoordIJK{0, 3, 1}
	kVec := CoordIJK{1, 0, 3}

	iScaled := IJKScale(iVec, c.I)
	jScaled := IJKScale(jVec, c.J)
	kScaled := IJKScale(kVec, c.K)

	result := IJKAdd(IJKAdd(iScaled, jScaled), kScaled)
	return result
}

// DownAp3 moves IJK coordinates one aperture-3 resolution down (CCW).
// From C: _downAp3: iVec={2,0,1}, jVec={1,2,0}, kVec={0,1,2}
// Result = {2I+J, 2J+K, I+2K} then normalize.
func DownAp3(c CoordIJK) CoordIJK {
	result := CoordIJK{
		I: 2*c.I + c.J,
		J: 2*c.J + c.K,
		K: c.I + 2*c.K,
	}
	IJKNormalize(&result)
	return result
}

// DownAp3r moves IJK coordinates one aperture-3r resolution down (CW).
// From C: _downAp3r: iVec={2,1,0}, jVec={0,2,1}, kVec={1,0,2}
// Result = {2I+K, I+2J, J+2K} then normalize.
func DownAp3r(c CoordIJK) CoordIJK {
	result := CoordIJK{
		I: 2*c.I + c.K,
		J: c.I + 2*c.J,
		K: c.J + 2*c.K,
	}
	IJKNormalize(&result)
	return result
}

// IJKToIJ converts IJK to IJ coordinates (drops K).
// From C: IJ.i = ijk.i - ijk.k; IJ.j = ijk.j - ijk.k
func IJKToIJ(c CoordIJK) (i, j int) {
	return c.I - c.K, c.J - c.K
}

// IJToIJK converts IJ to IJK coordinates.
// From C: set k=0, then normalize.
func IJToIJK(i, j int) CoordIJK {
	result := CoordIJK{I: i, J: j, K: 0}
	IJKNormalize(&result)
	return result
}

// ToVec2d converts IJK to Vec2d hex coordinates.
// From C: _ijkToHex2d: x = i - 0.5*j, y = j * M_SQRT3_2
func ToVec2d(c CoordIJK) Vec2d {
	i := c.I - c.K
	j := c.J - c.K
	return Vec2d{
		X: float64(i) - 0.5*float64(j),
		Y: float64(j) * constants.M_SQRT3_2,
	}
}
