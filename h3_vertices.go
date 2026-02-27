package h3

import (
	"math"

	"github.com/EDRInc/h3-go/internal/constants"
	"github.com/EDRInc/h3-go/internal/faceijk"
	"github.com/EDRInc/h3-go/internal/h3index"
)

// Vertex constants.
const (
	NUM_HEX_VERTS  = 6
	NUM_PENT_VERTS = 5
)

// directionToVertexNumHex maps directions to vertex numbers for hexagonal cells.
// From C vertex.c - gives the vertex number for the given direction.
// The vertex is the one "leading" the given edge in CCW order.
var directionToVertexNumHex = [constants.NUM_DIGITS]int{
	-1, // CENTER_DIGIT (invalid)
	3,  // K_AXES_DIGIT
	1,  // J_AXES_DIGIT
	2,  // JK_AXES_DIGIT
	5,  // I_AXES_DIGIT
	4,  // IK_AXES_DIGIT
	0,  // IJ_AXES_DIGIT
}

// directionToVertexNumPent maps directions to vertex numbers for pentagonal cells.
// From C vertex.c - for pentagons, K_AXES direction is deleted.
var directionToVertexNumPent = [constants.NUM_DIGITS]int{
	-1, // CENTER_DIGIT (invalid)
	-1, // K_AXES_DIGIT (deleted for pentagons)
	1,  // J_AXES_DIGIT
	2,  // JK_AXES_DIGIT
	4,  // I_AXES_DIGIT
	3,  // IK_AXES_DIGIT
	0,  // IJ_AXES_DIGIT
}

// CellToVertex returns the H3 vertex at the given vertex number of cell c.
// For hexagons, valid vertex numbers are 0-5.
// For pentagons, valid vertex numbers are 0-4.
// Returns an invalid vertex (0) if the vertex number is out of range.
func CellToVertex(c Cell, vertexNum int) Vertex {
	h := h3index.H3Index(c)
	if !IsValidCell(c) {
		return Vertex(0)
	}

	isPent := h3index.IsPentagon(h)
	maxVertNum := NUM_HEX_VERTS - 1
	if isPent {
		maxVertNum = NUM_PENT_VERTS - 1
	}

	if vertexNum < 0 || vertexNum > maxVertNum {
		return Vertex(0)
	}

	// Simple encoding approach: Use the origin cell as the owner.
	// The C implementation finds the "lowest" of 3 adjacent cells sharing this vertex,
	// but for initial implementation we'll use a simplified approach.
	// Set mode to vertex and encode vertex number in reserved bits.

	v := uint64(h)
	// Clear mode bits (59-62) and set to H3_VERTEX_MODE
	v = (v & ^(uint64(0xF) << constants.H3_MODE_OFFSET)) | (uint64(constants.H3_VERTEX_MODE) << constants.H3_MODE_OFFSET)
	// Clear reserved bits (56-58) and set vertex number
	v = (v & ^(uint64(0x7) << constants.H3_RESERVED_OFFSET)) | (uint64(vertexNum) << constants.H3_RESERVED_OFFSET)

	return Vertex(v)
}

// CellToVertexes returns all vertices of cell c.
// Returns 6 vertices for hexagons, 5 for pentagons.
func CellToVertexes(c Cell) []Vertex {
	if !IsValidCell(c) {
		return nil
	}

	isPent := IsPentagon(c)
	count := NUM_HEX_VERTS
	if isPent {
		count = NUM_PENT_VERTS
	}

	vertexes := make([]Vertex, count)
	for i := 0; i < count; i++ {
		vertexes[i] = CellToVertex(c, i)
	}
	return vertexes
}

// VertexToLatLng returns the lat/lng (in DEGREES) of a vertex.
// Returns zero LatLng if the vertex is invalid.
func VertexToLatLng(v Vertex) LatLng {
	if !IsValidVertex(v) {
		return LatLng{}
	}

	h := h3index.H3Index(v)

	// Extract vertex number from reserved bits.
	vertexNum := int((uint64(h) >> constants.H3_RESERVED_OFFSET) & constants.H3_RESERVED_MASK)

	// Reconstruct owner cell.
	owner := uint64(h)
	owner = (owner & ^(uint64(0xF) << constants.H3_MODE_OFFSET)) | (uint64(constants.H3_CELL_MODE) << constants.H3_MODE_OFFSET)
	owner = owner & ^(uint64(0x7) << constants.H3_RESERVED_OFFSET)
	ownerH := h3index.H3Index(owner)

	// Call FaceIJKToGeoBoundary directly to avoid the double-allocation of CellToBoundary.
	fijk := faceijk.H3ToFaceIJK(ownerH)
	res := ownerH.Resolution()
	isPent := h3index.IsPentagon(ownerH)
	geoPoints := faceijk.FaceIJKToGeoBoundary(fijk, res, isPent)

	if vertexNum >= len(geoPoints) {
		return LatLng{}
	}
	p := geoPoints[vertexNum]
	return LatLng{Lat: p.Lat * (180 / math.Pi), Lng: p.Lng * (180 / math.Pi)}
}

// IsValidVertex returns true if v is a valid H3 vertex index.
func IsValidVertex(v Vertex) bool {
	h := h3index.H3Index(v)

	// Check mode is H3_VERTEX_MODE
	if h.Mode() != constants.H3_VERTEX_MODE {
		return false
	}

	// Extract vertex number
	vertexNum := int((uint64(h) >> constants.H3_RESERVED_OFFSET) & constants.H3_RESERVED_MASK)

	// Get owner cell
	owner := uint64(h)
	// Clear mode bits and set to H3_CELL_MODE
	owner = (owner & ^(uint64(0xF) << constants.H3_MODE_OFFSET)) | (uint64(constants.H3_CELL_MODE) << constants.H3_MODE_OFFSET)
	// Clear reserved bits
	owner = owner & ^(uint64(0x7) << constants.H3_RESERVED_OFFSET)

	ownerCell := Cell(owner)

	// Check owner cell is valid
	if !IsValidCell(ownerCell) {
		return false
	}

	// Check vertex number is in valid range
	isPent := IsPentagon(ownerCell)
	maxVert := NUM_HEX_VERTS - 1
	if isPent {
		maxVert = NUM_PENT_VERTS - 1
	}

	if vertexNum < 0 || vertexNum > maxVert {
		return false
	}

	// Check that recreating the vertex gives the same result
	// (ensures canonical form)
	recreated := CellToVertex(ownerCell, vertexNum)
	return recreated == v
}

// vertexToOwner extracts the owner cell from a vertex index.
// This is a helper function used internally.
func vertexToOwner(v Vertex) Cell {
	h := h3index.H3Index(v)

	// Get owner cell (change mode back to cell, clear reserved bits)
	owner := uint64(h)
	// Clear mode bits and set to H3_CELL_MODE
	owner = (owner & ^(uint64(0xF) << constants.H3_MODE_OFFSET)) | (uint64(constants.H3_CELL_MODE) << constants.H3_MODE_OFFSET)
	// Clear reserved bits
	owner = owner & ^(uint64(0x7) << constants.H3_RESERVED_OFFSET)

	return Cell(owner)
}
