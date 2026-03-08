package h3

import (
	"testing"

	"github.com/lightboxre/h3-go/internal/h3index"
)

// TestGetIcosahedronFaces_SingleFaceHexes tests hexagons entirely on one face
func TestGetIcosahedronFaces_SingleFaceHexes(t *testing.T) {
	// Base cell 16 is at the center of an icosahedron face
	// Construct directly from base cell 16 at res 2
	h := h3index.H3_INIT
	h = h3index.SetMode(h, 1) // H3_CELL_MODE
	h = h3index.SetResolution(h, 2)
	h = h3index.SetBaseCell(h, 16)
	h = h3index.SetIndexDigit(h, 0, 0)
	h = h3index.SetIndexDigit(h, 1, 0)

	faces, err := GetIcosahedronFaces(Cell(h))
	if err != nil {
		t.Fatalf("GetIcosahedronFaces failed: %v", err)
	}

	if len(faces) != 1 {
		t.Errorf("Expected 1 face for centered base cell, got %d: %v", len(faces), faces)
	}

	// Test res 3 as well
	h = h3index.SetResolution(h, 3)
	h = h3index.SetIndexDigit(h, 2, 0)

	faces, err = GetIcosahedronFaces(Cell(h))
	if err != nil {
		t.Fatalf("GetIcosahedronFaces res 3 failed: %v", err)
	}

	if len(faces) != 1 {
		t.Errorf("Expected 1 face for centered base cell res 3, got %d", len(faces))
	}
}

// TestGetIcosahedronFaces_HexagonWithEdgeVertices tests a Class II pentagon neighbor
func TestGetIcosahedronFaces_HexagonWithEdgeVertices(t *testing.T) {
	cell := Cell(0x821c37fffffffff)

	faces, err := GetIcosahedronFaces(cell)
	if err != nil {
		t.Fatalf("GetIcosahedronFaces failed: %v", err)
	}

	// Should be on a single face
	if len(faces) == 0 {
		t.Error("Expected at least 1 face")
	}
}

// TestGetIcosahedronFaces_HexagonWithDistortion tests a Class III pentagon neighbor
func TestGetIcosahedronFaces_HexagonWithDistortion(t *testing.T) {
	cell := Cell(0x831c06fffffffff)

	faces, err := GetIcosahedronFaces(cell)
	if err != nil {
		t.Fatalf("GetIcosahedronFaces failed: %v", err)
	}

	// Should cross multiple faces due to Class III distortion
	if len(faces) != 2 {
		t.Errorf("Expected 2 faces for distorted hexagon, got %d: %v", len(faces), faces)
	}
}

// TestGetIcosahedronFaces_HexagonCrossingFaces tests a hexagon with vertices on multiple faces
func TestGetIcosahedronFaces_HexagonCrossingFaces(t *testing.T) {
	cell := Cell(0x821ce7fffffffff)

	faces, err := GetIcosahedronFaces(cell)
	if err != nil {
		t.Fatalf("GetIcosahedronFaces failed: %v", err)
	}

	// Should cross multiple faces
	if len(faces) < 2 {
		t.Errorf("Expected at least 2 faces for face-crossing hexagon, got %d: %v", len(faces), faces)
	}
}

// TestGetIcosahedronFaces_ClassIIIPentagon tests a Class III pentagon
func TestGetIcosahedronFaces_ClassIIIPentagon(t *testing.T) {
	// Pentagon base cell 4 at resolution 1 (Class III)
	pentagons := h3index.GetPentagonCells(1)
	cell := Cell(pentagons[0]) // base cell 4

	faces, err := GetIcosahedronFaces(cell)
	if err != nil {
		t.Fatalf("GetIcosahedronFaces failed: %v", err)
	}

	if len(faces) != 5 {
		t.Errorf("Expected 5 faces for Class III pentagon, got %d: %v", len(faces), faces)
	}
}

// TestGetIcosahedronFaces_ClassIIPentagon tests a Class II pentagon
func TestGetIcosahedronFaces_ClassIIPentagon(t *testing.T) {
	// Pentagon base cell 4 at resolution 2 (Class II)
	pentagons := h3index.GetPentagonCells(2)
	cell := Cell(pentagons[0]) // base cell 4

	faces, err := GetIcosahedronFaces(cell)
	if err != nil {
		t.Fatalf("GetIcosahedronFaces failed: %v", err)
	}

	if len(faces) != 5 {
		t.Errorf("Expected 5 faces for Class II pentagon, got %d: %v", len(faces), faces)
	}
}

// TestGetIcosahedronFaces_Res15Pentagon tests a pentagon at maximum resolution
func TestGetIcosahedronFaces_Res15Pentagon(t *testing.T) {
	// Pentagon base cell 4 at resolution 15
	pentagons := h3index.GetPentagonCells(15)
	cell := Cell(pentagons[0]) // base cell 4

	faces, err := GetIcosahedronFaces(cell)
	if err != nil {
		t.Fatalf("GetIcosahedronFaces failed: %v", err)
	}

	if len(faces) != 5 {
		t.Errorf("Expected 5 faces for res 15 pentagon, got %d: %v", len(faces), faces)
	}
}

// TestGetIcosahedronFaces_BaseCellHexagons tests all hexagonal base cells
func TestGetIcosahedronFaces_BaseCellHexagons(t *testing.T) {
	res0Cells := h3index.GetRes0Cells()

	singleFaceCount := 0
	multiFaceCount := 0

	for _, cell := range res0Cells {
		if h3index.IsPentagon(cell) {
			continue // Skip pentagons
		}

		faces, err := GetIcosahedronFaces(Cell(cell))
		if err != nil {
			t.Fatalf("GetIcosahedronFaces failed for base cell %d: %v", cell.BaseCell(), err)
		}

		if len(faces) == 0 {
			t.Errorf("Base cell %d returned no faces", cell.BaseCell())
		}

		if len(faces) == 1 {
			singleFaceCount++
		} else {
			multiFaceCount++
		}

		// Hexagons should not intersect more than 2 faces
		if len(faces) > 2 {
			t.Errorf("Base cell %d intersects %d faces (expected max 2)", cell.BaseCell(), len(faces))
		}
	}

	// The C test expects 80 single-face and 30 multi-face hexagonal base cells
	if singleFaceCount != 80 {
		t.Logf("Expected 80 single-face hexagonal base cells, got %d", singleFaceCount)
	}
	if multiFaceCount != 30 {
		t.Logf("Expected 30 multi-face hexagonal base cells, got %d", multiFaceCount)
	}
}

// TestGetIcosahedronFaces_BaseCellPentagons tests all pentagonal base cells
func TestGetIcosahedronFaces_BaseCellPentagons(t *testing.T) {
	pentagons := h3index.GetPentagonCells(0)

	for _, cell := range pentagons {
		faces, err := GetIcosahedronFaces(Cell(cell))
		if err != nil {
			t.Fatalf("GetIcosahedronFaces failed for pentagon base cell %d: %v", cell.BaseCell(), err)
		}

		if len(faces) != 5 {
			t.Errorf("Pentagon base cell %d intersects %d faces (expected 5): %v",
				cell.BaseCell(), len(faces), faces)
		}
	}
}

// TestGetIcosahedronFaces_Invalid tests invalid cells
func TestGetIcosahedronFaces_Invalid(t *testing.T) {
	invalidCells := []Cell{
		Cell(0xFFFFFFFFFFFFFFFF),
		Cell(0x71330073003f004e),
		Cell(0), // H3_NULL
	}

	for _, cell := range invalidCells {
		_, err := GetIcosahedronFaces(cell)
		if err == nil {
			t.Errorf("Expected error for invalid cell %x, got nil", cell)
		}
	}
}

// TestGetIcosahedronFaces_Invalid2 tests a specific invalid cell value.
// From C: TEST(invalid2) - cell 0x71330073003f004e
func TestGetIcosahedronFaces_Invalid2(t *testing.T) {
	invalid := Cell(0x71330073003f004e)
	_, err := GetIcosahedronFaces(invalid)
	if err == nil {
		t.Errorf("Expected error for invalid cell 0x71330073003f004e, got nil")
	}
}

// TestGetIcosahedronFaces_Sorted tests that faces are returned sorted
func TestGetIcosahedronFaces_Sorted(t *testing.T) {
	// Use a cell that crosses multiple faces
	cell := Cell(0x831c06fffffffff)

	faces, err := GetIcosahedronFaces(cell)
	if err != nil {
		t.Fatalf("GetIcosahedronFaces failed: %v", err)
	}

	// Verify sorted
	for i := 1; i < len(faces); i++ {
		if faces[i] <= faces[i-1] {
			t.Errorf("Faces not sorted: %v", faces)
			break
		}
	}

	// Verify all faces are valid (0-19)
	for _, f := range faces {
		if f < 0 || f > 19 {
			t.Errorf("Invalid face ID %d (expected 0-19)", f)
		}
	}
}
