package h3

import (
	"testing"

	"github.com/h3-native/h3-go/internal/constants"
	"github.com/h3-native/h3-go/internal/h3index"
)

func TestIsValidDirectedEdge(t *testing.T) {
	// Valid hexagon cell - use a known valid H3 index
	// This is a resolution 9 hexagon
	origin, err := StringToCell("8928308280fffff")
	if err != nil {
		t.Fatalf("Failed to parse cell: %v", err)
	}

	// Get directed edges from origin
	edges := OriginToDirectedEdges(origin)

	if len(edges) != 6 {
		t.Errorf("Expected 6 directed edges for hexagon, got %d", len(edges))
	}

	// All edges should be valid
	for i, e := range edges {
		if !IsValidDirectedEdge(e) {
			t.Errorf("Edge %d should be valid", i)
		}
	}

	// Invalid edge (not a directed edge mode)
	invalidEdge := DirectedEdge(origin)
	if IsValidDirectedEdge(invalidEdge) {
		t.Error("Cell should not be valid as directed edge")
	}

	// Test pentagon
	pentagonCells := GetPentagonCells(9)
	if len(pentagonCells) > 0 {
		pentagon := pentagonCells[0]
		pentEdges := OriginToDirectedEdges(pentagon)

		// Pentagon should have 5 edges
		if len(pentEdges) != 5 {
			t.Errorf("Expected 5 directed edges for pentagon, got %d", len(pentEdges))
		}

		// All pentagon edges should be valid
		for i, e := range pentEdges {
			if !IsValidDirectedEdge(e) {
				t.Errorf("Pentagon edge %d should be valid", i)
			}

			// Check that none have K_AXES_DIGIT direction
			dir := getReservedBits(h3index.H3Index(e))
			if dir == constants.K_AXES_DIGIT {
				t.Errorf("Pentagon should not have K-axis edge")
			}
		}
	}
}

func TestGetDirectedEdgeOrigin(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	for i, e := range edges {
		recoveredOrigin := GetDirectedEdgeOrigin(e)
		if recoveredOrigin != origin {
			t.Errorf("Edge %d: expected origin %x, got %x", i, origin, recoveredOrigin)
		}
	}
}

func TestGetDirectedEdgeDestination(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	// All destinations should be valid cells
	for i, e := range edges {
		dest := GetDirectedEdgeDestination(e)
		if !IsValidCell(dest) {
			t.Errorf("Edge %d: destination should be valid", i)
		}

		// Destination should be at the same resolution
		if GetResolution(dest) != GetResolution(origin) {
			t.Errorf("Edge %d: destination resolution mismatch", i)
		}

		// Destination should be different from origin
		if dest == origin {
			t.Errorf("Edge %d: destination should differ from origin", i)
		}
	}
}

func TestDirectedEdgeToCells(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	for i, e := range edges {
		cells := DirectedEdgeToCells(e)

		if cells[0] != origin {
			t.Errorf("Edge %d: expected origin %x, got %x", i, origin, cells[0])
		}

		if !IsValidCell(cells[1]) {
			t.Errorf("Edge %d: destination should be valid", i)
		}
	}
}

func TestAreNeighborCells(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	// Get destinations
	for _, e := range edges {
		dest := GetDirectedEdgeDestination(e)

		// Origin and destination should be neighbors
		areNeighbors, err := AreNeighborCells(origin, dest)
		if err != nil {
			t.Errorf("AreNeighborCells error: %v", err)
		}
		if !areNeighbors {
			t.Error("Expected cells to be neighbors")
		}
	}

	// Test non-neighbors (use a cell from a different region)
	farAway, _ := StringToCell("89283082813ffff") // Different cell
	areNeighbors, err := AreNeighborCells(origin, farAway)
	if err != nil {
		t.Errorf("AreNeighborCells error: %v", err)
	}
	if areNeighbors {
		t.Error("Expected cells to not be neighbors")
	}

	// Test same cell
	areNeighbors, err = AreNeighborCells(origin, origin)
	if err != nil {
		t.Errorf("AreNeighborCells error: %v", err)
	}
	if areNeighbors {
		t.Error("Cell should not be its own neighbor")
	}
}

func TestCellsToDirectedEdge(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	// Test all edges
	for _, e := range edges {
		dest := GetDirectedEdgeDestination(e)

		// Should be able to recreate the edge
		recreated, err := CellsToDirectedEdge(origin, dest)
		if err != nil {
			t.Errorf("CellsToDirectedEdge error: %v", err)
		}

		if recreated != e {
			t.Errorf("Expected edge %x, got %x", e, recreated)
		}
	}

	// Test non-neighbors (use a cell from a different region)
	farAway, _ := StringToCell("89283082813ffff") // Different cell
	_, err := CellsToDirectedEdge(origin, farAway)
	if err == nil {
		t.Error("Expected error for non-neighbor cells")
	}
}

func TestOriginToDirectedEdges(t *testing.T) {
	// Test hexagon
	hexCell, _ := StringToCell("8928308280fffff")
	hexEdges := OriginToDirectedEdges(hexCell)

	if len(hexEdges) != 6 {
		t.Errorf("Hexagon should have 6 edges, got %d", len(hexEdges))
	}

	// Check that all directions are unique
	directions := make(map[int]bool)
	for _, e := range hexEdges {
		dir := getReservedBits(h3index.H3Index(e))
		if directions[dir] {
			t.Errorf("Duplicate direction %d", dir)
		}
		directions[dir] = true

		if dir < 1 || dir > 6 {
			t.Errorf("Invalid direction %d", dir)
		}
	}

	// Test pentagon
	pentagonCells := GetPentagonCells(9)
	if len(pentagonCells) > 0 {
		pentagon := pentagonCells[0]
		pentEdges := OriginToDirectedEdges(pentagon)

		if len(pentEdges) != 5 {
			t.Errorf("Pentagon should have 5 edges, got %d", len(pentEdges))
		}

		// Check that no edge has K_AXES_DIGIT direction
		directions := make(map[int]bool)
		for _, e := range pentEdges {
			dir := getReservedBits(h3index.H3Index(e))
			if directions[dir] {
				t.Errorf("Duplicate direction %d", dir)
			}
			directions[dir] = true

			if dir == constants.K_AXES_DIGIT {
				t.Error("Pentagon should not have K-axis edge")
			}
		}
	}
}

func TestDirectedEdgeToBoundary(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	for i, e := range edges {
		boundary := DirectedEdgeToBoundary(e)

		// Boundary should have 2 vertices
		if len(boundary) != 2 {
			t.Errorf("Edge %d: expected 2 vertices in boundary, got %d", i, len(boundary))
		}

		// Vertices should be valid (non-zero)
		if len(boundary) == 2 {
			if boundary[0].Lat == 0 && boundary[0].Lng == 0 {
				t.Errorf("Edge %d: first vertex is zero", i)
			}
			if boundary[1].Lat == 0 && boundary[1].Lng == 0 {
				t.Errorf("Edge %d: second vertex is zero", i)
			}

			// Vertices should be different
			if boundary[0].Lat == boundary[1].Lat && boundary[0].Lng == boundary[1].Lng {
				t.Errorf("Edge %d: vertices should be different", i)
			}
		}
	}
}

func TestDirectedEdgeSymmetry(t *testing.T) {
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	// For each edge from origin to destination,
	// there should be a reverse edge from destination to origin
	for _, e := range edges {
		dest := GetDirectedEdgeDestination(e)
		if !IsValidCell(dest) {
			continue
		}

		// Get reverse edge
		reverseEdge, err := CellsToDirectedEdge(dest, origin)
		if err != nil {
			t.Errorf("Expected reverse edge to exist: %v", err)
			continue
		}

		// Check that reverse edge's origin is our destination
		reverseOrigin := GetDirectedEdgeOrigin(reverseEdge)
		if reverseOrigin != dest {
			t.Error("Reverse edge origin mismatch")
		}

		// Check that reverse edge's destination is our origin
		reverseDest := GetDirectedEdgeDestination(reverseEdge)
		if reverseDest != origin {
			t.Error("Reverse edge destination mismatch")
		}
	}
}

func TestDirectedEdgeInvalidCases(t *testing.T) {
	// Test with invalid cell
	invalidCell := Cell(0)
	edges := OriginToDirectedEdges(invalidCell)
	if edges != nil {
		t.Error("Expected nil edges for invalid cell")
	}

	// Test invalid directed edge
	invalidEdge := DirectedEdge(0)
	if IsValidDirectedEdge(invalidEdge) {
		t.Error("Zero edge should not be valid")
	}

	// Test with cells at different resolutions
	cell1, _ := StringToCell("8928308280fffff")
	cell2, _ := StringToCell("8828308281fffff")

	_, err := CellsToDirectedEdge(cell1, cell2)
	if err == nil {
		t.Error("Expected error for cells at different resolutions")
	}

	areNeighbors, err := AreNeighborCells(cell1, cell2)
	if err == nil {
		t.Error("Expected error for cells at different resolutions")
	}
	if areNeighbors {
		t.Error("Cells at different resolutions should not be neighbors")
	}
}

func TestDirectedEdgeBidirectional(t *testing.T) {
	// Test that for every directed edge A->B, there exists B->A
	origin, _ := StringToCell("8928308280fffff")
	edges := OriginToDirectedEdges(origin)

	for _, forwardEdge := range edges {
		dest := GetDirectedEdgeDestination(forwardEdge)
		if !IsValidCell(dest) {
			continue
		}

		// Check destination has an edge back to origin
		destEdges := OriginToDirectedEdges(dest)
		foundReverse := false

		for _, reverseEdge := range destEdges {
			reverseDest := GetDirectedEdgeDestination(reverseEdge)
			if reverseDest == origin {
				foundReverse = true
				break
			}
		}

		if !foundReverse {
			t.Error("Could not find reverse edge")
		}
	}
}
