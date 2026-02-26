package faceijk

import "math"

// GreatCircleDistanceRads returns the great-circle distance in radians
// between two lat/lng points using the Haversine formula.
// From C: greatCircleDistanceRads
func GreatCircleDistanceRads(lat1, lng1, lat2, lng2 float64) float64 {
	// Haversine formula
	dlat := lat2 - lat1
	dlng := lng2 - lng1

	sinDlat := math.Sin(dlat / 2.0)
	sinDlng := math.Sin(dlng / 2.0)

	a := sinDlat*sinDlat +
		math.Cos(lat1)*math.Cos(lat2)*sinDlng*sinDlng

	return 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))
}

// GreatCircleDistanceKm returns the great-circle distance in kilometers.
func GreatCircleDistanceKm(lat1, lng1, lat2, lng2 float64) float64 {
	return GreatCircleDistanceRads(lat1, lng1, lat2, lng2) * 6371.007180918475
}

// GreatCircleDistanceM returns the great-circle distance in meters.
func GreatCircleDistanceM(lat1, lng1, lat2, lng2 float64) float64 {
	return GreatCircleDistanceKm(lat1, lng1, lat2, lng2) * 1000.0
}

// DegsToRads converts degrees to radians.
func DegsToRads(deg float64) float64 {
	return deg * math.Pi / 180.0
}

// RadsToDegs converts radians to degrees.
func RadsToDegs(rad float64) float64 {
	return rad * 180.0 / math.Pi
}

// ConstrainLat clamps latitude to [-π/2, π/2].
// From C: constrainLat
func ConstrainLat(lat float64) float64 {
	for lat > math.Pi/2.0 {
		lat -= math.Pi
	}
	for lat < -math.Pi/2.0 {
		lat += math.Pi
	}
	return lat
}

// ConstrainLng constrains longitude to [-π, π].
// From C: constrainLng
func ConstrainLng(lng float64) float64 {
	for lng > math.Pi {
		lng -= 2.0 * math.Pi
	}
	for lng < -math.Pi {
		lng += 2.0 * math.Pi
	}
	return lng
}

// GeoAlmostEqual returns true if two GeoPoints are approximately equal.
func GeoAlmostEqual(g1, g2 GeoPoint) bool {
	const epsilon = 1e-12
	return math.Abs(g1.Lat-g2.Lat) < epsilon &&
		math.Abs(g1.Lng-g2.Lng) < epsilon
}

// PosAngleRads normalizes an angle in radians to [0, 2π).
func PosAngleRads(rads float64) float64 {
	tmp := math.Mod(rads, 2.0*math.Pi)
	if tmp < 0 {
		tmp += 2.0 * math.Pi
	}
	return tmp
}

// GeoAzimuthRads calculates the azimuth (bearing) in radians from point p1 to p2.
// From C: _geoAzimuthRads
func GeoAzimuthRads(lat1, lng1, lat2, lng2 float64) float64 {
	return math.Atan2(
		math.Cos(lat2)*math.Sin(lng2-lng1),
		math.Cos(lat1)*math.Sin(lat2)-
			math.Sin(lat1)*math.Cos(lat2)*math.Cos(lng2-lng1),
	)
}
