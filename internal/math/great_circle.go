// Package h3math provides mathematical functions for H3 operations.
// Functions for computing distances, areas, and edge lengths.
package h3math

import (
	"github.com/EDRInc/h3-go/internal/constants"
	"github.com/EDRInc/h3-go/internal/faceijk"
)

// GreatCircleDistanceRads returns the great-circle distance in radians
// between two points given in RADIANS.
func GreatCircleDistanceRads(lat1, lng1, lat2, lng2 float64) float64 {
	return faceijk.GreatCircleDistanceRads(lat1, lng1, lat2, lng2)
}

// GreatCircleDistanceKm returns the great-circle distance in kilometers.
func GreatCircleDistanceKm(lat1, lng1, lat2, lng2 float64) float64 {
	return GreatCircleDistanceRads(lat1, lng1, lat2, lng2) * constants.EarthRadiusKm
}

// GreatCircleDistanceM returns the great-circle distance in meters.
func GreatCircleDistanceM(lat1, lng1, lat2, lng2 float64) float64 {
	return GreatCircleDistanceRads(lat1, lng1, lat2, lng2) * constants.EarthRadiusM
}
