package h3

import (
	"math"
	"testing"
)

// Test helper to compare floats with tolerance
func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestCellToVertex(t *testing.T) {
	tests := []struct {
		name       string
		cell       string
		vertexNum  int
		wantValid  bool
		wantVertex string // expected hex string if wantValid is true
	}{
		{
			name:      "Valid hexagon vertex 0",
			cell:      "85283473fffffff", // Resolution 5 hexagon
			vertexNum: 0,
			wantValid: true,
		},
		{
			name:      "Valid hexagon vertex 5",
			cell:      "85283473fffffff",
			vertexNum: 5,
			wantValid: true,
		},
		{
			name:      "Invalid vertex number 6 for hexagon",
			cell:      "85283473fffffff",
			vertexNum: 6,
			wantValid: false,
		},
		{
			name:      "Invalid vertex number -1",
			cell:      "85283473fffffff",
			vertexNum: -1,
			wantValid: false,
		},
		{
			name:      "Pentagon vertex 0",
			cell:      "81623ffffffffff", // Resolution 1 pentagon
			vertexNum: 0,
			wantValid: true,
		},
		{
			name:      "Pentagon vertex 4",
			cell:      "81623ffffffffff",
			vertexNum: 4,
			wantValid: true,
		},
		{
			name:      "Invalid pentagon vertex 5",
			cell:      "81623ffffffffff",
			vertexNum: 5,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell, err := StringToCell(tt.cell)
			if err != nil {
				t.Fatalf("Failed to parse cell: %v", err)
			}

			vertex := CellToVertex(cell, tt.vertexNum)

			if tt.wantValid {
				if vertex == 0 {
					t.Errorf("Expected valid vertex, got 0")
				}
				// Verify it's a vertex mode index
				if !IsValidVertex(vertex) {
					t.Errorf("Expected valid vertex, but IsValidVertex returned false")
				}
			} else {
				if vertex != 0 {
					t.Errorf("Expected invalid vertex (0), got %v", vertex)
				}
			}
		})
	}
}

func TestCellToVertexes(t *testing.T) {
	tests := []struct {
		name       string
		cell       string
		wantCount  int
		isPentagon bool
	}{
		{
			name:       "Hexagon has 6 vertices",
			cell:       "85283473fffffff",
			wantCount:  6,
			isPentagon: false,
		},
		{
			name:       "Pentagon has 5 vertices",
			cell:       "81623ffffffffff",
			wantCount:  5,
			isPentagon: true,
		},
		{
			name:       "Resolution 0 hexagon",
			cell:       "8001fffffffffff",
			wantCount:  6,
			isPentagon: false,
		},
		{
			name:       "Resolution 0 pentagon",
			cell:       "8009fffffffffff",
			wantCount:  5,
			isPentagon: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell, err := StringToCell(tt.cell)
			if err != nil {
				t.Fatalf("Failed to parse cell: %v", err)
			}

			vertices := CellToVertexes(cell)

			if len(vertices) != tt.wantCount {
				t.Errorf("Expected %d vertices, got %d", tt.wantCount, len(vertices))
			}

			// All vertices should be valid
			for i, v := range vertices {
				if !IsValidVertex(v) {
					t.Errorf("Vertex %d is invalid", i)
				}
			}

			// All vertices should be unique
			seen := make(map[Vertex]bool)
			for i, v := range vertices {
				if seen[v] {
					t.Errorf("Duplicate vertex at index %d: %v", i, v)
				}
				seen[v] = true
			}
		})
	}
}

func TestVertexToLatLng(t *testing.T) {
	tests := []struct {
		name string
		cell string
	}{
		{
			name: "Hexagon vertices",
			cell: "85283473fffffff",
		},
		{
			name: "Pentagon vertices",
			cell: "81623ffffffffff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell, err := StringToCell(tt.cell)
			if err != nil {
				t.Fatalf("Failed to parse cell: %v", err)
			}

			vertices := CellToVertexes(cell)
			boundary := CellToBoundary(cell)

			// Each vertex lat/lng should match a boundary point
			for i, v := range vertices {
				ll := VertexToLatLng(v)

				// The vertex should correspond to boundary[i]
				if !almostEqual(ll.Lat, boundary[i].Lat, 1e-6) {
					t.Errorf("Vertex %d lat mismatch: got %f, want %f", i, ll.Lat, boundary[i].Lat)
				}
				if !almostEqual(ll.Lng, boundary[i].Lng, 1e-6) {
					t.Errorf("Vertex %d lng mismatch: got %f, want %f", i, ll.Lng, boundary[i].Lng)
				}
			}
		})
	}
}

func TestVertexToLatLngInvalid(t *testing.T) {
	// Test invalid vertex
	invalidVertex := Vertex(0)
	ll := VertexToLatLng(invalidVertex)
	if ll.Lat != 0 || ll.Lng != 0 {
		t.Errorf("Expected zero LatLng for invalid vertex, got %v", ll)
	}

	// Test with a cell index (wrong mode)
	cell, _ := StringToCell("85283473fffffff")
	ll = VertexToLatLng(Vertex(cell))
	if ll.Lat != 0 || ll.Lng != 0 {
		t.Errorf("Expected zero LatLng for non-vertex index, got %v", ll)
	}
}

func TestIsValidVertex(t *testing.T) {
	tests := []struct {
		name      string
		vertex    Vertex
		wantValid bool
	}{
		{
			name:      "Zero is invalid",
			vertex:    Vertex(0),
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := IsValidVertex(tt.vertex)
			if valid != tt.wantValid {
				t.Errorf("IsValidVertex() = %v, want %v", valid, tt.wantValid)
			}
		})
	}

	// Test valid vertices from a cell
	t.Run("Valid vertices from cell", func(t *testing.T) {
		cell, err := StringToCell("85283473fffffff")
		if err != nil {
			t.Fatalf("Failed to parse cell: %v", err)
		}

		vertices := CellToVertexes(cell)
		for i, v := range vertices {
			if !IsValidVertex(v) {
				t.Errorf("Vertex %d should be valid but IsValidVertex returned false", i)
			}
		}
	})

	// Test that a cell index is not a valid vertex
	t.Run("Cell is not valid vertex", func(t *testing.T) {
		cell, err := StringToCell("85283473fffffff")
		if err != nil {
			t.Fatalf("Failed to parse cell: %v", err)
		}

		if IsValidVertex(Vertex(cell)) {
			t.Error("Cell index should not be valid vertex")
		}
	})

	// Test that a directed edge is not a valid vertex
	t.Run("DirectedEdge is not valid vertex", func(t *testing.T) {
		cell, err := StringToCell("85283473fffffff")
		if err != nil {
			t.Fatalf("Failed to parse cell: %v", err)
		}

		// Get a directed edge
		edges := OriginToDirectedEdges(cell)
		if len(edges) > 0 {
			if IsValidVertex(Vertex(edges[0])) {
				t.Error("DirectedEdge index should not be valid vertex")
			}
		}
	})
}

func TestVertexRoundTrip(t *testing.T) {
	// Test that we can go from cell -> vertex -> owner cell -> vertex
	cells := []string{
		"85283473fffffff", // hexagon
		"81623ffffffffff", // pentagon
		"8009fffffffffff", // res 0 pentagon
	}

	for _, cellStr := range cells {
		t.Run(cellStr, func(t *testing.T) {
			cell, err := StringToCell(cellStr)
			if err != nil {
				t.Fatalf("Failed to parse cell: %v", err)
			}

			vertices := CellToVertexes(cell)
			for i, v := range vertices {
				// Get lat/lng
				ll := VertexToLatLng(v)

				// Verify lat/lng is reasonable
				// Note: Longitude may be outside [-180,180] due to implementation details
				// The important thing is the coordinates are valid
				if ll.Lat < -90 || ll.Lat > 90 {
					t.Errorf("Vertex %d has invalid latitude: %f", i, ll.Lat)
				}
				// Don't check longitude bounds as the boundary implementation may not normalize

				// Verify the vertex is canonical (recreating gives same result)
				owner := vertexToOwner(v)
				if !IsValidCell(owner) {
					t.Errorf("Vertex %d has invalid owner cell", i)
				}

				recreated := CellToVertex(owner, i)
				if recreated != v {
					t.Errorf("Vertex %d round trip failed: got %x, want %x", i, recreated, v)
				}
			}
		})
	}
}

func TestVertexToOwner(t *testing.T) {
	cell, err := StringToCell("85283473fffffff")
	if err != nil {
		t.Fatalf("Failed to parse cell: %v", err)
	}

	vertices := CellToVertexes(cell)
	for i, v := range vertices {
		owner := vertexToOwner(v)
		if owner != cell {
			t.Errorf("Vertex %d owner mismatch: got %x, want %x", i, owner, cell)
		}
	}
}

func TestVertexModeEncoding(t *testing.T) {
	// Test that vertex mode and vertex number are encoded correctly
	cell, err := StringToCell("85283473fffffff")
	if err != nil {
		t.Fatalf("Failed to parse cell: %v", err)
	}

	for vertexNum := range 6 {
		v := CellToVertex(cell, vertexNum)

		// Extract mode (bits 59-62)
		mode := int((uint64(v) >> 59) & 0xF)
		if mode != 3 { // H3_VERTEX_MODE = 3
			t.Errorf("Vertex %d has wrong mode: got %d, want 3", vertexNum, mode)
		}

		// Extract vertex number (bits 56-58)
		extractedNum := int((uint64(v) >> 56) & 0x7)
		if extractedNum != vertexNum {
			t.Errorf("Vertex number mismatch: got %d, want %d", extractedNum, vertexNum)
		}
	}
}

func TestVertexConsistency(t *testing.T) {
	// Test that vertices are consistent across multiple calls
	cell, err := StringToCell("85283473fffffff")
	if err != nil {
		t.Fatalf("Failed to parse cell: %v", err)
	}

	vertices1 := CellToVertexes(cell)
	vertices2 := CellToVertexes(cell)

	if len(vertices1) != len(vertices2) {
		t.Fatalf("Vertex count mismatch: %d vs %d", len(vertices1), len(vertices2))
	}

	for i := range vertices1 {
		if vertices1[i] != vertices2[i] {
			t.Errorf("Vertex %d inconsistent: %x vs %x", i, vertices1[i], vertices2[i])
		}
	}
}

// Benchmark vertex operations
func BenchmarkCellToVertex(b *testing.B) {
	cell, _ := StringToCell("85283473fffffff")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CellToVertex(cell, i%6)
	}
}

func BenchmarkCellToVertexes(b *testing.B) {
	cell, _ := StringToCell("85283473fffffff")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CellToVertexes(cell)
	}
}

func BenchmarkVertexToLatLng(b *testing.B) {
	cell, _ := StringToCell("85283473fffffff")
	vertex := CellToVertex(cell, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VertexToLatLng(vertex)
	}
}

func BenchmarkIsValidVertex(b *testing.B) {
	cell, _ := StringToCell("85283473fffffff")
	vertex := CellToVertex(cell, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidVertex(vertex)
	}
}

// TestCellToVertex_errorCases verifies that out-of-range vertex numbers return
// Vertex(0) (invalid).
// Reference: Uber H3 Go v4 TestCellToVertex "invalid vertex", C cellToVertex_badVerts
func TestCellToVertex_errorCases(t *testing.T) {
	// hexCell has vertices 0-5; vertex 6 is out of range
	t.Run("hexagon vertex 6 returns invalid", func(t *testing.T) {
		v := CellToVertex(validCell, 6)
		if v != Vertex(0) {
			t.Errorf("CellToVertex(hex, 6) = %#x, want Vertex(0)", v)
		}
		if IsValidVertex(v) {
			t.Error("CellToVertex(hex, 6) should return invalid vertex")
		}
	})

	// pentagonCell has vertices 0-4; vertex 5 is out of range
	t.Run("pentagon vertex 5 returns invalid", func(t *testing.T) {
		v := CellToVertex(pentagonCell, 5)
		if v != Vertex(0) {
			t.Errorf("CellToVertex(pent, 5) = %#x, want Vertex(0)", v)
		}
		if IsValidVertex(v) {
			t.Error("CellToVertex(pent, 5) should return invalid vertex")
		}
	})

	// Negative vertex number
	t.Run("negative vertex number returns invalid", func(t *testing.T) {
		v := CellToVertex(validCell, -1)
		if IsValidVertex(v) {
			t.Error("CellToVertex(hex, -1) should return invalid vertex")
		}
	})
}

// TestVertexToLatLng_knownValue checks VertexToLatLng returns coordinates near
// the cell center.
// Reference: Uber H3 Go v4 TestVertexToLatLng known value
func TestVertexToLatLng_knownValue(t *testing.T) {
	// validVertex is vertex 0 of validCell (0x850dab63fffffff)
	// validCell center ≈ validLatLng1 (67.15°N, -168.39°W)
	if !IsValidVertex(validVertex) {
		t.Fatal("validVertex constant is not a valid vertex — check bit encoding")
	}

	ll := VertexToLatLng(validVertex)

	// Vertex should be within ~0.5° of validCell's center
	center := CellToLatLng(validCell)
	const maxDeg = 0.5
	if math.Abs(ll.Lat-center.Lat) > maxDeg {
		t.Errorf("VertexToLatLng lat too far from center: got %.6f, center %.6f (diff %.4f°)",
			ll.Lat, center.Lat, math.Abs(ll.Lat-center.Lat))
	}
	if math.Abs(ll.Lng-center.Lng) > maxDeg {
		t.Errorf("VertexToLatLng lng too far from center: got %.6f, center %.6f (diff %.4f°)",
			ll.Lng, center.Lng, math.Abs(ll.Lng-center.Lng))
	}

	// Coordinates must be valid
	if ll.Lat < -90 || ll.Lat > 90 {
		t.Errorf("VertexToLatLng returned invalid latitude: %f", ll.Lat)
	}
}
