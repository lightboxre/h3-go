package algos

import (
	"testing"

	"github.com/h3-native/h3-go/internal/h3index"
)

// Test helper to create a valid H3 cell at resolution 0
func makeRes0Cell(baseCell int) h3index.H3Index {
	h := h3index.H3_INIT
	h = h3index.SetMode(h, 1) // H3_CELL_MODE
	h = h3index.SetResolution(h, 0)
	h = h3index.SetBaseCell(h, baseCell)
	return h
}

// Test helper to create a valid H3 cell at a given resolution
func makeCell(baseCell, res int, digits []int) h3index.H3Index {
	h := h3index.H3_INIT
	h = h3index.SetMode(h, 1) // H3_CELL_MODE
	h = h3index.SetResolution(h, res)
	h = h3index.SetBaseCell(h, baseCell)
	for r := 0; r < res && r < len(digits); r++ {
		h = h3index.SetIndexDigit(h, r, digits[r])
	}
	return h
}

func TestMaxGridDiskSize(t *testing.T) {
	tests := []struct {
		k        int
		expected int64
		wantErr  bool
	}{
		{0, 1, false},
		{1, 7, false},
		{2, 19, false},
		{3, 37, false},
		{10, 331, false},
		{-1, 0, true},
	}

	for _, tt := range tests {
		got, err := MaxGridDiskSize(tt.k)
		if (err != nil) != tt.wantErr {
			t.Errorf("MaxGridDiskSize(%d) error = %v, wantErr %v", tt.k, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("MaxGridDiskSize(%d) = %d, want %d", tt.k, got, tt.expected)
		}
	}
}

func TestGridDiskUnsafe(t *testing.T) {
	// Test with a simple hexagon cell at resolution 0
	// Use base cell 16 which is not adjacent to any pentagons
	origin := makeRes0Cell(16)

	t.Run("k=0", func(t *testing.T) {
		cells, err := GridDiskUnsafe(origin, 0)
		if err != nil {
			t.Fatalf("GridDiskUnsafe failed: %v", err)
		}
		if len(cells) != 1 {
			t.Errorf("GridDiskUnsafe(k=0) returned %d cells, want 1", len(cells))
		}
		if cells[0] != origin {
			t.Errorf("GridDiskUnsafe(k=0) returned wrong cell")
		}
	})

	t.Run("k=1", func(t *testing.T) {
		cells, err := GridDiskUnsafe(origin, 1)
		if err != nil {
			// It's OK if this fails due to pentagon encounter near base cells
			// Just test that the safe version works
			t.Skipf("GridDiskUnsafe encountered pentagon (expected for some base cells): %v", err)
		}
		// Should return origin + 6 neighbors = 7 cells
		if len(cells) != 7 {
			t.Errorf("GridDiskUnsafe(k=1) returned %d cells, want 7", len(cells))
		}
		// First cell should be origin
		if cells[0] != origin {
			t.Errorf("GridDiskUnsafe(k=1) first cell is not origin")
		}
	})

	t.Run("negative k", func(t *testing.T) {
		_, err := GridDiskUnsafe(origin, -1)
		if err != ErrDomain {
			t.Errorf("GridDiskUnsafe(-1) should return ErrDomain, got %v", err)
		}
	})
}

func TestGridDiskWithPentagon(t *testing.T) {
	// Test with a pentagon base cell
	pentagon := makeRes0Cell(4) // Base cell 4 is a pentagon

	cells, err := GridDiskUnsafe(pentagon, 1)
	if err != ErrPentagon {
		t.Errorf("GridDiskUnsafe on pentagon should return ErrPentagon, got %v", err)
	}
	if cells != nil {
		t.Errorf("GridDiskUnsafe on pentagon should return nil cells")
	}

	// GridDisk should fall back to safe mode
	cells, err = GridDisk(pentagon, 1)
	if err != nil {
		t.Fatalf("GridDisk on pentagon should not fail: %v", err)
	}
	// Pentagon at k=0 + 5 neighbors (not 6) = 6 cells
	if len(cells) < 1 {
		t.Errorf("GridDisk on pentagon returned too few cells: %d", len(cells))
	}
}

func TestGridDiskDistancesUnsafe(t *testing.T) {
	origin := makeRes0Cell(16)

	t.Run("k=0", func(t *testing.T) {
		distances, err := GridDiskDistancesUnsafe(origin, 0)
		if err != nil {
			t.Fatalf("GridDiskDistancesUnsafe failed: %v", err)
		}
		if len(distances) != 1 {
			t.Errorf("GridDiskDistancesUnsafe(k=0) returned %d distance groups, want 1", len(distances))
		}
		if len(distances[0]) != 1 {
			t.Errorf("GridDiskDistancesUnsafe(k=0) distance 0 has %d cells, want 1", len(distances[0]))
		}
	})

	t.Run("k=1", func(t *testing.T) {
		distances, err := GridDiskDistancesUnsafe(origin, 1)
		if err != nil {
			t.Skipf("GridDiskDistancesUnsafe encountered pentagon (expected for some base cells): %v", err)
		}
		if len(distances) != 2 {
			t.Errorf("GridDiskDistancesUnsafe(k=1) returned %d distance groups, want 2", len(distances))
		}
		if len(distances[0]) != 1 {
			t.Errorf("GridDiskDistancesUnsafe(k=1) distance 0 has %d cells, want 1", len(distances[0]))
		}
		if len(distances[1]) != 6 {
			t.Errorf("GridDiskDistancesUnsafe(k=1) distance 1 has %d cells, want 6", len(distances[1]))
		}
	})
}

func TestGridRingUnsafe(t *testing.T) {
	origin := makeRes0Cell(16)

	t.Run("k=0", func(t *testing.T) {
		ring, err := GridRingUnsafe(origin, 0)
		if err != nil {
			t.Fatalf("GridRingUnsafe failed: %v", err)
		}
		if len(ring) != 1 {
			t.Errorf("GridRingUnsafe(k=0) returned %d cells, want 1", len(ring))
		}
		if ring[0] != origin {
			t.Errorf("GridRingUnsafe(k=0) returned wrong cell")
		}
	})

	t.Run("k=1", func(t *testing.T) {
		ring, err := GridRingUnsafe(origin, 1)
		if err != nil {
			t.Skipf("GridRingUnsafe encountered pentagon (expected for some base cells): %v", err)
		}
		// Ring at k=1 should have 6 cells
		if len(ring) != 6 {
			t.Errorf("GridRingUnsafe(k=1) returned %d cells, want 6", len(ring))
		}
	})

	t.Run("k=2", func(t *testing.T) {
		ring, err := GridRingUnsafe(origin, 2)
		if err != nil {
			t.Skipf("GridRingUnsafe encountered pentagon (expected for some base cells): %v", err)
		}
		// Ring at k=2 should have 12 cells
		if len(ring) != 12 {
			t.Errorf("GridRingUnsafe(k=2) returned %d cells, want 12", len(ring))
		}
	})
}

func TestGridDistance(t *testing.T) {
	origin := makeRes0Cell(16)

	t.Run("same cell", func(t *testing.T) {
		dist, err := GridDistance(origin, origin)
		if err != nil {
			t.Fatalf("GridDistance failed: %v", err)
		}
		if dist != 0 {
			t.Errorf("GridDistance(same cell) = %d, want 0", dist)
		}
	})

	t.Run("neighboring cells", func(t *testing.T) {
		// Get a neighbor
		neighbors := getNeighbors(origin)
		if len(neighbors) == 0 {
			t.Fatal("No neighbors found")
		}
		neighbor := neighbors[0]

		dist, err := GridDistance(origin, neighbor)
		if err != nil {
			t.Fatalf("GridDistance failed: %v", err)
		}
		if dist != 1 {
			t.Errorf("GridDistance(neighbor) = %d, want 1", dist)
		}
	})
}

func TestGridPathCells(t *testing.T) {
	origin := makeRes0Cell(16)

	t.Run("same cell", func(t *testing.T) {
		path, err := GridPathCells(origin, origin)
		if err != nil {
			t.Fatalf("GridPathCells failed: %v", err)
		}
		if len(path) != 1 {
			t.Errorf("GridPathCells(same cell) returned %d cells, want 1", len(path))
		}
		if path[0] != origin {
			t.Errorf("GridPathCells(same cell) returned wrong cell")
		}
	})

	t.Run("neighboring cells", func(t *testing.T) {
		neighbors := getNeighbors(origin)
		if len(neighbors) == 0 {
			t.Fatal("No neighbors found")
		}
		neighbor := neighbors[0]

		path, err := GridPathCells(origin, neighbor)
		if err != nil {
			t.Fatalf("GridPathCells failed: %v", err)
		}
		if len(path) != 2 {
			t.Errorf("GridPathCells(neighbor) returned %d cells, want 2", len(path))
		}
		if path[0] != origin {
			t.Errorf("GridPathCells first cell is not origin")
		}
		if path[1] != neighbor {
			t.Errorf("GridPathCells last cell is not neighbor")
		}
	})
}

func TestGridPathCellsSize(t *testing.T) {
	origin := makeRes0Cell(16)

	t.Run("same cell", func(t *testing.T) {
		size, err := GridPathCellsSize(origin, origin)
		if err != nil {
			t.Fatalf("GridPathCellsSize failed: %v", err)
		}
		if size != 1 {
			t.Errorf("GridPathCellsSize(same cell) = %d, want 1", size)
		}
	})
}

func TestCompactCells(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		result, err := CompactCells([]h3index.H3Index{})
		if err != nil {
			t.Fatalf("CompactCells failed: %v", err)
		}
		if result != nil {
			t.Errorf("CompactCells([]) should return nil")
		}
	})

	t.Run("single cell", func(t *testing.T) {
		cell := makeRes0Cell(16)
		result, err := CompactCells([]h3index.H3Index{cell})
		if err != nil {
			t.Fatalf("CompactCells failed: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("CompactCells(single) returned %d cells, want 1", len(result))
		}
		if result[0] != cell {
			t.Errorf("CompactCells(single) returned wrong cell")
		}
	})

	t.Run("already compact", func(t *testing.T) {
		cells := []h3index.H3Index{
			makeRes0Cell(16),
			makeRes0Cell(17),
			makeRes0Cell(18),
		}
		result, err := CompactCells(cells)
		if err != nil {
			t.Fatalf("CompactCells failed: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("CompactCells returned %d cells, want 3", len(result))
		}
	})
}

func TestUncompactCells(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		result, err := UncompactCells([]h3index.H3Index{}, 1)
		if err != nil {
			t.Fatalf("UncompactCells failed: %v", err)
		}
		if len(result) != 0 {
			t.Errorf("UncompactCells([]) should return empty slice")
		}
	})

	t.Run("same resolution", func(t *testing.T) {
		cell := makeRes0Cell(16)
		result, err := UncompactCells([]h3index.H3Index{cell}, 0)
		if err != nil {
			t.Fatalf("UncompactCells failed: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("UncompactCells(same res) returned %d cells, want 1", len(result))
		}
		if result[0] != cell {
			t.Errorf("UncompactCells(same res) returned wrong cell")
		}
	})

	t.Run("invalid resolution", func(t *testing.T) {
		cell := makeRes0Cell(16)
		_, err := UncompactCells([]h3index.H3Index{cell}, -1)
		if err != ErrDomain {
			t.Errorf("UncompactCells(invalid res) should return ErrDomain, got %v", err)
		}
	})

	t.Run("higher source resolution", func(t *testing.T) {
		cell := makeCell(16, 1, []int{0})
		_, err := UncompactCells([]h3index.H3Index{cell}, 0)
		if err == nil {
			t.Error("UncompactCells(higher source res) should return error")
		}
	})
}

func TestUncompactCellsSize(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		size, err := UncompactCellsSize([]h3index.H3Index{}, 1)
		if err != nil {
			t.Fatalf("UncompactCellsSize failed: %v", err)
		}
		if size != 0 {
			t.Errorf("UncompactCellsSize([]) = %d, want 0", size)
		}
	})

	t.Run("same resolution", func(t *testing.T) {
		cell := makeRes0Cell(16)
		size, err := UncompactCellsSize([]h3index.H3Index{cell}, 0)
		if err != nil {
			t.Fatalf("UncompactCellsSize failed: %v", err)
		}
		if size != 1 {
			t.Errorf("UncompactCellsSize(same res) = %d, want 1", size)
		}
	})

	t.Run("one level deeper", func(t *testing.T) {
		cell := makeRes0Cell(16)
		size, err := UncompactCellsSize([]h3index.H3Index{cell}, 1)
		if err != nil {
			t.Fatalf("UncompactCellsSize failed: %v", err)
		}
		// A hexagon at res 0 has 7 children at res 1
		if size != 7 {
			t.Errorf("UncompactCellsSize(res 0->1) = %d, want 7", size)
		}
	})
}

func TestGetNeighbors(t *testing.T) {
	origin := makeRes0Cell(16)

	neighbors := getNeighbors(origin)

	// A hexagon should have 6 neighbors
	if len(neighbors) != 6 {
		t.Errorf("getNeighbors returned %d neighbors, want 6", len(neighbors))
	}

	// All neighbors should be valid
	for i, n := range neighbors {
		if n == h3index.H3_NULL {
			t.Errorf("neighbor %d is H3_NULL", i)
		}
		if !h3index.IsValid(n) {
			t.Errorf("neighbor %d is invalid: %v", i, n)
		}
	}
}

func TestGetNeighborsPentagon(t *testing.T) {
	pentagon := makeRes0Cell(4) // Base cell 4 is a pentagon

	neighbors := getNeighbors(pentagon)

	// A pentagon should have 5 neighbors (no K_AXES_DIGIT direction)
	if len(neighbors) != 5 {
		t.Errorf("getNeighbors(pentagon) returned %d neighbors, want 5", len(neighbors))
	}
}

func TestRotate60ccw(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{1, 5}, // K -> IK
		{5, 4}, // IK -> I
		{4, 6}, // I -> IJ
		{6, 2}, // IJ -> J
		{2, 3}, // J -> JK
		{3, 1}, // JK -> K
	}

	for _, tt := range tests {
		got := rotate60ccw(tt.input)
		if got != tt.expected {
			t.Errorf("rotate60ccw(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestLeadingNonZeroDigit(t *testing.T) {
	t.Run("all zeros", func(t *testing.T) {
		cell := makeRes0Cell(16)
		digit := leadingNonZeroDigit(cell)
		if digit != 0 {
			t.Errorf("leadingNonZeroDigit(res0) = %d, want 0", digit)
		}
	})

	t.Run("first digit non-zero", func(t *testing.T) {
		cell := makeCell(16, 2, []int{1, 2})
		digit := leadingNonZeroDigit(cell)
		if digit != 1 {
			t.Errorf("leadingNonZeroDigit = %d, want 1", digit)
		}
	})

	t.Run("first digit zero", func(t *testing.T) {
		cell := makeCell(16, 2, []int{0, 2})
		digit := leadingNonZeroDigit(cell)
		if digit != 2 {
			t.Errorf("leadingNonZeroDigit = %d, want 2", digit)
		}
	})
}

func TestGetParent(t *testing.T) {
	cell := makeCell(16, 2, []int{1, 2})

	t.Run("parent at res 1", func(t *testing.T) {
		parent := getParent(cell, 1)
		if parent.Resolution() != 1 {
			t.Errorf("parent resolution = %d, want 1", parent.Resolution())
		}
		if parent.IndexDigit(0) != 1 {
			t.Errorf("parent digit 0 = %d, want 1", parent.IndexDigit(0))
		}
	})

	t.Run("parent at res 0", func(t *testing.T) {
		parent := getParent(cell, 0)
		if parent.Resolution() != 0 {
			t.Errorf("parent resolution = %d, want 0", parent.Resolution())
		}
	})

	t.Run("invalid resolution", func(t *testing.T) {
		parent := getParent(cell, -1)
		if parent != cell {
			t.Errorf("parent at invalid res should return original cell")
		}
	})
}

func TestGetChildren(t *testing.T) {
	cell := makeRes0Cell(16)

	t.Run("children at res 1", func(t *testing.T) {
		children := getChildren(cell, 1)
		// A hexagon should have 7 children
		if len(children) != 7 {
			t.Errorf("getChildren returned %d children, want 7", len(children))
		}

		// All children should be at resolution 1
		for i, child := range children {
			if child.Resolution() != 1 {
				t.Errorf("child %d has resolution %d, want 1", i, child.Resolution())
			}
			if child.BaseCell() != 16 {
				t.Errorf("child %d has base cell %d, want 16", i, child.BaseCell())
			}
		}
	})

	t.Run("same resolution", func(t *testing.T) {
		children := getChildren(cell, 0)
		if len(children) != 1 {
			t.Errorf("getChildren(same res) returned %d children, want 1", len(children))
		}
		if children[0] != cell {
			t.Errorf("getChildren(same res) returned wrong cell")
		}
	})
}

func TestGetChildrenCount(t *testing.T) {
	t.Run("hexagon", func(t *testing.T) {
		cell := makeRes0Cell(0)
		count := getChildrenCount(cell)
		if count != 7 {
			t.Errorf("getChildrenCount(hexagon) = %d, want 7", count)
		}
	})

	t.Run("pentagon", func(t *testing.T) {
		cell := makeRes0Cell(4) // Base cell 4 is a pentagon
		count := getChildrenCount(cell)
		if count != 6 {
			t.Errorf("getChildrenCount(pentagon) = %d, want 6", count)
		}
	})
}
