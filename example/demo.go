package main

import (
	"fmt"
	h3 "github.com/h3-native/h3-go"
)

func main() {
	// Convert lat/lng to cell
	lat, lng := 37.7796, -122.4193
	cell := h3.LatLngToCell(lat, lng, 9)
	fmt.Printf("Cell at (%.4f, %.4f) res 9: %s\n", lat, lng, h3.CellToString(cell))

	// Get cell info
	fmt.Printf("Resolution: %d\n", h3.GetResolution(cell))
	fmt.Printf("Base Cell: %d\n", h3.GetBaseCellNumber(cell))
	fmt.Printf("Is Pentagon: %v\n", h3.IsPentagon(cell))
	fmt.Printf("Is Valid: %v\n", h3.IsValidCell(cell))

	// Get parent
	parent := h3.CellToParent(cell, 5)
	fmt.Printf("Parent at res 5: %s\n", h3.CellToString(parent))

	// Get children count
	childrenSize := h3.CellToChildrenSize(parent, 6)
	fmt.Printf("Number of children from res 5 to res 6: %d\n", childrenSize)

	// Get all res 0 cells
	res0 := h3.GetRes0Cells()
	fmt.Printf("Number of resolution 0 cells: %d\n", len(res0))
}
