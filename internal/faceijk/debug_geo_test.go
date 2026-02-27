package faceijk

import (
	"fmt"
	"math"
	"testing"

	"github.com/EDRInc/h3-go/internal/constants"
	"github.com/EDRInc/h3-go/internal/coordijk"
)

// TestDebugCellCenters computes the centers of both candidate cells and distances to input.
func TestDebugCellCenters(t *testing.T) {
	// Input: SF (37.7796°, -122.4193°)
	inputLat, inputLng := 37.7796, -122.4193
	inputLatRad := inputLat * math.Pi / 180.0
	inputLngRad := inputLng * math.Pi / 180.0

	// sfCell center: Vec2d = {-3991, -3625.182} (from {face=7, I=0, J=1898, K=6084})
	sf4186Y := -4186.0 * constants.M_SQRT3_2
	sfCellVec := coordijk.Vec2d{X: -3991.0, Y: sf4186Y}
	sfLat, sfLng := hex2dToGeo(sfCellVec, 7, 9)

	// Our cell center: Vec2d = {-3992, -3625.182} (from {face=7, I=0, J=1899, K=6085})
	ourCellVec := coordijk.Vec2d{X: -3992.0, Y: sf4186Y}
	ourLat, ourLng := hex2dToGeo(ourCellVec, 7, 9)

	sfLatDeg := RadsToDegs(sfLat)
	sfLngDeg := RadsToDegs(sfLng)
	ourLatDeg := RadsToDegs(ourLat)
	ourLngDeg := RadsToDegs(ourLng)

	fmt.Printf("Input coordinates: lat=%.8f° lng=%.8f°\n", inputLat, inputLng)
	fmt.Printf("\nsfCell (0x8928308280fffff) center:\n")
	fmt.Printf("  lat=%.8f° lng=%.8f°\n", sfLatDeg, sfLngDeg)
	fmt.Printf("  Δlat=%.6f° Δlng=%.6f°\n", inputLat-sfLatDeg, inputLng-sfLngDeg)

	fmt.Printf("\nOurCell (0x89283082877ffff) center:\n")
	fmt.Printf("  lat=%.8f° lng=%.8f°\n", ourLatDeg, ourLngDeg)
	fmt.Printf("  Δlat=%.6f° Δlng=%.6f°\n", inputLat-ourLatDeg, inputLng-ourLngDeg)

	// Great circle distances
	sfDist := GreatCircleDistanceKm(inputLatRad, inputLngRad, sfLat, sfLng)
	ourDist := GreatCircleDistanceKm(inputLatRad, inputLngRad, ourLat, ourLng)
	fmt.Printf("\nDistance from input to sfCell center: %.1f m\n", sfDist*1000)
	fmt.Printf("Distance from input to ourCell center: %.1f m\n", ourDist*1000)

	// Also show geoToHex2d result for input
	fmt.Printf("\ngeoToHex2d(input, face=7, res=9):\n")
	v2d := geoToHex2d(inputLatRad, inputLngRad, 7, 9)
	fmt.Printf("  X=%.6f Y=%.6f\n", v2d.X, v2d.Y)
	fmt.Printf("  X relative to sfCell center (-3991): %.6f\n", v2d.X-(-3991.0))
	fmt.Printf("  X relative to ourCell center (-3992): %.6f\n", v2d.X-(-3992.0))

	// Show what the average edge length is at res 9
	// H3 avg edge length res 9 ≈ 174m
	fmt.Printf("\nH3 avg edge length res 9 ≈ 174m (0.00156°)\n")
	fmt.Printf("At hex2d scale, 1 unit ≈ edge_length distance\n")
}
