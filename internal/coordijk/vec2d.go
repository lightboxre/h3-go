package coordijk

import (
	"math"

	"github.com/EDRInc/h3-go/internal/constants"
)

// Vec2d is a 2D floating-point vector.
type Vec2d struct {
	X, Y float64
}

// V2dMag returns the magnitude (length) of the vector.
// From C: sqrt(v->x * v->x + v->y * v->y)
func V2dMag(v Vec2d) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// V2dIntersect finds the intersection point of two line segments.
// From C: uses parametric line intersection formulas.
// Line 1: p0 to p1
// Line 2: p2 to p3
func V2dIntersect(p0, p1, p2, p3 Vec2d) Vec2d {
	// Direction vectors
	s1 := Vec2d{X: p1.X - p0.X, Y: p1.Y - p0.Y}
	s2 := Vec2d{X: p3.X - p2.X, Y: p3.Y - p2.Y}

	// Calculate parameter t using cross products
	// From C: t = (s2.x * (p0.y - p2.y) - s2.y * (p0.x - p2.x)) / (-s2.x * s1.y + s1.x * s2.y)
	denominator := -s2.X*s1.Y + s1.X*s2.Y
	t := (s2.X*(p0.Y-p2.Y) - s2.Y*(p0.X-p2.X)) / denominator

	// Compute intersection point
	return Vec2d{
		X: p0.X + t*s1.X,
		Y: p0.Y + t*s1.Y,
	}
}

// V2dEquals returns true if two Vec2d are equal within EPSILON.
// From C: _v2dAlmostEquals uses FLT_EPSILON, we use EPSILON for float64.
func V2dEquals(v1, v2 Vec2d) bool {
	return math.Abs(v1.X-v2.X) < constants.EPSILON &&
		math.Abs(v1.Y-v2.Y) < constants.EPSILON
}

// PointSquareDist returns the square of the distance between two Vec2d.
func PointSquareDist(v1, v2 Vec2d) float64 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	return dx*dx + dy*dy
}
