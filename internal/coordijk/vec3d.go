package coordijk

import "math"

// Vec3d is a 3D floating-point vector (used for icosahedron face calculations).
type Vec3d struct {
	X, Y, Z float64
}

// V3dMag returns the magnitude of the vector.
func V3dMag(v Vec3d) float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// V3dDot returns the dot product of two Vec3d.
func V3dDot(v1, v2 Vec3d) float64 {
	return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}

// V3dNormalize normalizes a Vec3d to unit length.
func V3dNormalize(v Vec3d) Vec3d {
	mag := V3dMag(v)
	if mag == 0 {
		return Vec3d{0, 0, 0}
	}
	return Vec3d{
		X: v.X / mag,
		Y: v.Y / mag,
		Z: v.Z / mag,
	}
}

// PointSquareDist3d returns the square of the distance between two Vec3d.
// From C: _pointSquareDist computes sum of squared differences.
func PointSquareDist3d(v1, v2 Vec3d) float64 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	dz := v1.Z - v2.Z
	return dx*dx + dy*dy + dz*dz
}

// GeoToVec3d converts lat/lng (radians) to a 3D unit vector on the sphere.
// From C: _geoToVec3d
// v->z = sin(lat)
// r = cos(lat)
// v->x = cos(lng) * r
// v->y = sin(lng) * r
func GeoToVec3d(latRad, lngRad float64) Vec3d {
	r := math.Cos(latRad)
	return Vec3d{
		X: math.Cos(lngRad) * r,
		Y: math.Sin(lngRad) * r,
		Z: math.Sin(latRad),
	}
}

// Vec3dToGeo converts a 3D unit vector to lat/lng (radians).
// Inverse of GeoToVec3d.
func Vec3dToGeo(v Vec3d) (latRad, lngRad float64) {
	latRad = math.Asin(v.Z)
	lngRad = math.Atan2(v.Y, v.X)
	return
}
