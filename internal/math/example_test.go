package h3math_test

import (
	"fmt"

	"github.com/h3-native/h3-go/internal/constants"
	"github.com/h3-native/h3-go/internal/h3index"
	"github.com/h3-native/h3-go/internal/math"
)

// ExampleGreatCircleDistanceKm demonstrates calculating the distance between two points.
func ExampleGreatCircleDistanceKm() {
	// San Francisco coordinates (in radians)
	sfLat := 0.6588 // ~37.7749 degrees
	sfLng := -2.1364 // ~-122.4194 degrees

	// Los Angeles coordinates (in radians)
	laLat := 0.5924 // ~33.9416 degrees
	laLng := -2.0594 // ~-118.4085 degrees

	distance := h3math.GreatCircleDistanceKm(sfLat, sfLng, laLat, laLng)
	fmt.Printf("Distance: %.0f km\n", distance)
	// Output: Distance: 580 km
}

// ExampleEdgeLengthKm demonstrates getting average edge lengths for different resolutions.
func ExampleEdgeLengthKm() {
	// Get average edge length at resolution 5
	edgeLength := h3math.EdgeLengthKm(5)
	fmt.Printf("Resolution 5 average edge length: %.2f km\n", edgeLength)
	// Output: Resolution 5 average edge length: 9.85 km
}

// ExampleCellAreaKm2 demonstrates calculating the area of an H3 cell.
func ExampleCellAreaKm2() {
	// Create an H3 cell at resolution 5
	var h h3index.H3Index = h3index.H3_INIT
	h = h3index.SetMode(h, constants.H3_CELL_MODE)
	h = h3index.SetResolution(h, 5)
	h = h3index.SetBaseCell(h, 0)

	area := h3math.CellAreaKm2(h)
	fmt.Printf("Cell area: %.0f km²\n", area)
	// Output: Cell area: 249 km²
}
