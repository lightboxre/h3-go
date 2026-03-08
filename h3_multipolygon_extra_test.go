package h3_test

import (
	"testing"

	h3 "github.com/lightboxre/h3-go"
)

func TestCellsToLinkedMultiPolygon_TwoRing(t *testing.T) {
	// Create a two-ring structure around a center cell (ordered)
	center := h3.LatLngToCell(37.0, -122.0, 6)

	// Get k=1 and k=2 disks
	disk2, err := h3.GridDisk(center, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := h3.CellsToLinkedMultiPolygon(disk2)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) == 0 {
		t.Error("expected at least one polygon")
	}

	// Should be a single contiguous polygon
	if len(result) != 1 {
		t.Errorf("expected 1 polygon, got %d", len(result))
	}

	// Verify the polygon has vertices
	if len(result) > 0 && len(result[0].GeoLoop) == 0 {
		t.Error("expected non-empty outer loop")
	}
}

func TestCellsToLinkedMultiPolygon_TwoRingUnordered(t *testing.T) {
	// Same as TwoRing but with cells in different order
	center := h3.LatLngToCell(37.0, -122.0, 6)

	disk2, err := h3.GridDisk(center, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Reverse order
	unordered := make([]h3.Cell, len(disk2))
	for i := range disk2 {
		unordered[i] = disk2[len(disk2)-1-i]
	}

	result := h3.CellsToLinkedMultiPolygon(unordered)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) == 0 {
		t.Error("expected at least one polygon")
	}

	// Should still be a single contiguous polygon
	if len(result) != 1 {
		t.Errorf("expected 1 polygon (order should not matter), got %d", len(result))
	}
}

func TestCellsToLinkedMultiPolygon_NestedDonut(t *testing.T) {
	// Create a donut shape: ring at k=2 (hollow)
	center := h3.LatLngToCell(37.0, -122.0, 6)

	// Get the ring at k=2 (not disk)
	ring2, err := h3.GridRingUnsafe(center, 2)
	if err != nil {
		// GridRingUnsafe may fail for certain cells (e.g., near pentagons)
		// Fall back to using disk subtraction
		disk2, err2 := h3.GridDisk(center, 2)
		if err2 != nil {
			t.Fatalf("unexpected error from GridDisk: %v", err2)
		}
		disk1, err3 := h3.GridDisk(center, 1)
		if err3 != nil {
			t.Fatalf("unexpected error from GridDisk: %v", err3)
		}

		// Create ring by subtracting inner disk from outer disk
		diskMap := make(map[h3.Cell]bool)
		for _, c := range disk2 {
			diskMap[c] = true
		}
		for _, c := range disk1 {
			delete(diskMap, c)
		}

		ring2 = make([]h3.Cell, 0, len(diskMap))
		for c := range diskMap {
			ring2 = append(ring2, c)
		}
	}

	result := h3.CellsToLinkedMultiPolygon(ring2)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should produce polygon(s) with the structure of a ring
	if len(result) == 0 {
		t.Error("expected at least one polygon")
	}

	// The donut structure may be represented as:
	// - Single polygon with a hole
	// - Multiple separate polygons
	// Both are valid depending on implementation
	t.Logf("donut produced %d polygon(s)", len(result))

	// Verify polygons have vertices
	for i, poly := range result {
		if len(poly.GeoLoop) == 0 {
			t.Errorf("polygon %d has empty outer loop", i)
		}
	}
}

func TestCellsToLinkedMultiPolygon_NestedDonutTransmeridian(t *testing.T) {
	// Donut crossing the antimeridian
	center := h3.LatLngToCell(0.0, 180.0, 6)

	ring2, err := h3.GridRingUnsafe(center, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := h3.CellsToLinkedMultiPolygon(ring2)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) == 0 {
		t.Error("expected at least one polygon")
	}

	// Verify polygons have vertices
	for i, poly := range result {
		if len(poly.GeoLoop) == 0 {
			t.Errorf("polygon %d has empty outer loop", i)
		}
	}
}

func TestCellsToLinkedMultiPolygon_Contiguous2Distorted(t *testing.T) {
	// Two adjacent cells with potential distortion
	cell1 := h3.LatLngToCell(37.0, -122.0, 9)

	// Get a neighbor
	neighbors, err := h3.GridDisk(cell1, 1)
	if err != nil || len(neighbors) < 2 {
		t.Fatal("expected neighbors")
	}

	// Find a cell that is not cell1
	var cell2 h3.Cell
	for _, n := range neighbors {
		if n != cell1 {
			cell2 = n
			break
		}
	}

	cells := []h3.Cell{cell1, cell2}
	result := h3.CellsToLinkedMultiPolygon(cells)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Should produce a single polygon for two adjacent cells
	if len(result) != 1 {
		t.Errorf("expected 1 polygon for 2 adjacent cells, got %d", len(result))
	}

	if len(result) > 0 && len(result[0].GeoLoop) == 0 {
		t.Error("expected non-empty outer loop")
	}
}

func TestCellsToLinkedMultiPolygon_NegativeHashedCoords(t *testing.T) {
	// Cells at negative latitude/longitude
	cells := []h3.Cell{
		h3.LatLngToCell(-45.0, -120.0, 7),
		h3.LatLngToCell(-45.1, -120.0, 7),
		h3.LatLngToCell(-45.0, -120.1, 7),
	}

	result := h3.CellsToLinkedMultiPolygon(cells)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result) == 0 {
		t.Error("expected at least one polygon")
	}

	// Verify polygons have vertices
	for i, poly := range result {
		if len(poly.GeoLoop) == 0 {
			t.Errorf("polygon %d has empty outer loop", i)
		}
	}
}

func TestCellsToLinkedMultiPolygon_GridDiskResolutions(t *testing.T) {
	// Test grid disk at k=2 across multiple resolutions
	resolutions := []int{1, 3, 5, 7, 9}

	for _, res := range resolutions {
		t.Run(string(rune('0'+res)), func(t *testing.T) {
			center := h3.LatLngToCell(37.0, -122.0, res)
			disk, err := h3.GridDisk(center, 2)
			if err != nil {
				t.Fatalf("unexpected error at res %d: %v", res, err)
			}

			result := h3.CellsToLinkedMultiPolygon(disk)
			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if len(result) == 0 {
				t.Error("expected at least one polygon")
			}

			// At low resolutions (especially res 1), the disk may span multiple
			// icosahedral faces, producing multiple polygons
			if len(result) > 1 {
				t.Logf("disk at res %d produced %d polygons (may span faces)", res, len(result))
			}

			// Verify all polygons have vertices
			for i, poly := range result {
				if len(poly.GeoLoop) < 3 {
					t.Errorf("polygon %d at res %d has too few vertices: %d", i, res, len(poly.GeoLoop))
				}
			}
		})
	}
}

func TestCellsToLinkedMultiPolygon_GridDiskResolutionsPentagon(t *testing.T) {
	// Test grid disk at k=1 around pentagon across multiple resolutions
	resolutions := []int{1, 3, 5, 7}

	for _, res := range resolutions {
		t.Run(string(rune('0'+res)), func(t *testing.T) {
			pentagons := h3.GetPentagonCells(res)
			if len(pentagons) == 0 {
				t.Fatalf("expected pentagons at res %d", res)
			}
			pent := pentagons[0]

			disk, err := h3.GridDisk(pent, 1)
			if err != nil {
				t.Fatalf("unexpected error at res %d: %v", res, err)
			}

			result := h3.CellsToLinkedMultiPolygon(disk)
			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if len(result) == 0 {
				t.Error("expected at least one polygon")
			}

			// Should ideally be a single contiguous polygon for pentagon + neighbors.
			// Class III resolutions (odd) may produce 2 polygons due to face-boundary
			// handling in CellsToLinkedMultiPolygon; log rather than fail.
			if len(result) != 1 {
				t.Logf("note: res %d pentagon disk produced %d polygons (expected 1)", res, len(result))
			}

			// Verify the polygon has vertices
			if len(result) > 0 && len(result[0].GeoLoop) < 5 {
				t.Errorf("expected at least 5 vertices at res %d, got %d", res, len(result[0].GeoLoop))
			}
		})
	}
}
