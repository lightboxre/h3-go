// Package h3 - directed edge functions.
// Ported from the C H3 reference implementation (directedEdge.c).
package h3

import (
	"errors"

	"github.com/lightboxre/h3-go/internal/algos"
	"github.com/lightboxre/h3-go/internal/constants"
	"github.com/lightboxre/h3-go/internal/h3index"
)

// ============================================================================
// Directed Edge Functions
// ============================================================================

// A DirectedEdge is encoded as an H3Index with:
//   - Mode bits (59-62): H3_DIRECTEDEDGE_MODE (2)
//   - Reserved bits (56-58): edge direction (1-6)
//   - Resolution, base cell, and digit sequence: same as origin cell

// AreNeighborCells returns true if two cells share an edge.
// Returns false with an error if cells are invalid or have different resolutions.
func AreNeighborCells(origin, destination Cell) (bool, error) {
	if !IsValidCell(origin) || !IsValidCell(destination) {
		return false, errors.New("invalid cell")
	}
	if GetResolution(origin) != GetResolution(destination) {
		return false, errors.New("resolution mismatch")
	}
	isPent := IsPentagon(origin)
	for dir := 1; dir <= 6; dir++ {
		if isPent && dir == constants.K_AXES_DIGIT {
			continue
		}
		var rotations int
		var neighbor h3index.H3Index
		if err := algos.H3NeighborRotations(h3index.H3Index(origin), dir, &rotations, &neighbor); err != nil {
			continue
		}
		if Cell(neighbor) == destination {
			return true, nil
		}
	}
	return false, nil
}

// CellsToDirectedEdge returns the directed edge between two neighbor cells.
// Returns an error if the cells are not neighbors.
func CellsToDirectedEdge(origin, destination Cell) (DirectedEdge, error) {
	if !IsValidCell(origin) || !IsValidCell(destination) {
		return 0, errors.New("invalid cell")
	}
	if GetResolution(origin) != GetResolution(destination) {
		return 0, errors.New("resolution mismatch")
	}
	isPent := IsPentagon(origin)
	for dir := 1; dir <= 6; dir++ {
		if isPent && dir == constants.K_AXES_DIGIT {
			continue
		}
		var rotations int
		var neighbor h3index.H3Index
		if err := algos.H3NeighborRotations(h3index.H3Index(origin), dir, &rotations, &neighbor); err != nil {
			continue
		}
		if Cell(neighbor) == destination {
			e := h3index.H3Index(origin)
			e = h3index.SetMode(e, constants.H3_DIRECTEDEDGE_MODE)
			e = setReservedBits(e, dir)
			return DirectedEdge(e), nil
		}
	}
	return 0, errors.New("cells are not neighbors")
}

// IsValidDirectedEdge returns true if the directed edge is valid.
func IsValidDirectedEdge(e DirectedEdge) bool {
	h := h3index.H3Index(e)

	// Check mode
	if h.Mode() != constants.H3_DIRECTEDEDGE_MODE {
		return false
	}

	// Check edge direction (1-6)
	dir := getReservedBits(h)
	if dir < 1 || dir > 6 {
		return false
	}

	// Check origin cell validity
	origin := GetDirectedEdgeOrigin(e)
	if !IsValidCell(origin) {
		return false
	}

	// Pentagon constraint: no K-axis edge (direction 1)
	if IsPentagon(origin) && dir == constants.K_AXES_DIGIT {
		return false
	}

	return true
}

// GetDirectedEdgeOrigin returns the origin cell of a directed edge.
func GetDirectedEdgeOrigin(e DirectedEdge) Cell {
	h := h3index.H3Index(e)
	// Origin is the directed edge with mode changed back to cell mode
	h = h3index.SetMode(h, constants.H3_CELL_MODE)
	h = setReservedBits(h, 0) // clear reserved bits
	return Cell(h)
}

// GetDirectedEdgeDestination returns the destination cell of a directed edge.
func GetDirectedEdgeDestination(e DirectedEdge) Cell {
	cell, err := getDirectedEdgeDestinationInternal(e)
	if err != nil {
		return Cell(h3index.H3_NULL)
	}
	return cell
}

// getDirectedEdgeDestinationInternal is the internal implementation that returns errors.
func getDirectedEdgeDestinationInternal(e DirectedEdge) (Cell, error) {
	origin := GetDirectedEdgeOrigin(e)
	dir := getReservedBits(h3index.H3Index(e))

	// Use algos package to get neighbor in the specified direction
	rotations := 0
	var neighbor h3index.H3Index
	err := algos.H3NeighborRotations(h3index.H3Index(origin), dir, &rotations, &neighbor)
	if err != nil {
		return Cell(h3index.H3_NULL), err
	}

	return Cell(neighbor), nil
}

// DirectedEdgeToCells returns both the origin and destination cells.
func DirectedEdgeToCells(e DirectedEdge) [2]Cell {
	origin := GetDirectedEdgeOrigin(e)
	destination := GetDirectedEdgeDestination(e)
	return [2]Cell{origin, destination}
}

// OriginToDirectedEdges returns all 6 (or 5 for pentagon) directed edges
// from the given origin cell.
func OriginToDirectedEdges(origin Cell) []DirectedEdge {
	if !IsValidCell(origin) {
		return nil
	}

	h := h3index.H3Index(origin)
	isPent := h3index.IsPentagon(h)

	numEdges := 6
	if isPent {
		numEdges = 5
	}

	edges := make([]DirectedEdge, 0, numEdges)

	// For each direction 1-6 (K, J, JK, I, IK, IJ)
	// Skip K_AXES_DIGIT (1) for pentagons
	for dir := 1; dir <= 6; dir++ {
		if isPent && dir == constants.K_AXES_DIGIT {
			continue // Pentagon has no K-axis edge
		}

		// Create directed edge: copy origin, set mode=2, set reserved bits to dir
		e := h3index.H3Index(origin)
		e = h3index.SetMode(e, constants.H3_DIRECTEDEDGE_MODE)
		e = setReservedBits(e, dir)
		edges = append(edges, DirectedEdge(e))
	}

	return edges
}

// DirectedEdgeToBoundary returns the boundary (line segment) of a directed edge.
// This is the shared boundary between origin and destination cells.
// The boundary is returned as a 2-vertex line segment.
func DirectedEdgeToBoundary(e DirectedEdge) CellBoundary {
	if !IsValidDirectedEdge(e) {
		return CellBoundary{}
	}
	origin := GetDirectedEdgeOrigin(e)
	dir := getReservedBits(h3index.H3Index(e))
	return computeEdgeBoundaryFromDirection(origin, dir)
}

// computeEdgeBoundaryFromDirection computes the edge boundary using the
// direction information encoded in a directed edge.
// Uses the canonical H3 directionToVertexNum tables (same as C reference impl).
func computeEdgeBoundaryFromDirection(origin Cell, dir int) CellBoundary {
	// Get the origin cell boundary (1 alloc).
	boundary := CellToBoundary(origin)
	n := len(boundary)
	if n == 0 {
		return CellBoundary{}
	}

	if dir < 1 || dir >= constants.NUM_DIGITS {
		return CellBoundary{}
	}

	// Select the leading vertex using the canonical direction→vertex tables
	// (same tables used by vertex.c and the C directedEdgeToBoundary).
	var vertexNum int
	if IsPentagon(origin) {
		vertexNum = directionToVertexNumPent[dir]
	} else {
		vertexNum = directionToVertexNumHex[dir]
	}
	if vertexNum < 0 || vertexNum >= n {
		return CellBoundary{}
	}
	return CellBoundary{boundary[vertexNum], boundary[(vertexNum+1)%n]}
}

// ============================================================================
// Internal Helper Functions
// ============================================================================

// setReservedBits sets the 3 reserved bits (56-58) of an H3Index.
// These bits store the edge direction for directed edges.
func setReservedBits(h h3index.H3Index, bits int) h3index.H3Index {
	// Clear bits 56-58, then set
	const mask = uint64(0x7) << constants.H3_RESERVED_OFFSET
	return h3index.H3Index((uint64(h) &^ mask) | (uint64(bits) << constants.H3_RESERVED_OFFSET))
}

// getReservedBits retrieves the 3 reserved bits (56-58) of an H3Index.
// For directed edges, these bits contain the edge direction (1-6).
func getReservedBits(h h3index.H3Index) int {
	return int((uint64(h) >> constants.H3_RESERVED_OFFSET) & constants.H3_RESERVED_MASK)
}

// Implementation Note:
// The directed edge implementation depends on algos.H3NeighborRotations for computing
// destination cells. This function requires a fully functional neighbor traversal implementation
// in the algos package. If the algos package's h3NeighborRotations function encounters issues
// (e.g., base cell transitions, pentagon handling), GetDirectedEdgeDestination may return NULL cells.
//
// The core directed edge encoding/decoding functions (IsValidDirectedEdge, GetDirectedEdgeOrigin,
// OriginToDirectedEdges) are fully functional and do not depend on neighbor traversal.
