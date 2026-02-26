package h3index

import (
	"testing"

	"github.com/h3-native/h3-go/internal/constants"
)

// Known-good test values from the C H3 implementation.
const (
	validH3Res5  H3Index = 0x85283473fffffff // resolution 5, base cell 20
	validH3Res9  H3Index = 0x8928308280fffff // resolution 9, base cell 20
	pentagonRes0 H3Index = 0x8009fffffffffff // base cell 4 (pentagon), res 0
	pentagonRes1 H3Index = 0x81043ffffffffff // base cell 4 (pentagon), res 1
)

func TestH3IndexAccessors(t *testing.T) {
	tests := []struct {
		name     string
		index    H3Index
		mode     int
		res      int
		baseCell int
	}{
		{"res5_bc20", validH3Res5, constants.H3_CELL_MODE, 5, 20},
		{"res9", validH3Res9, constants.H3_CELL_MODE, 9, 20},
		{"pentagon_res0", pentagonRes0, constants.H3_CELL_MODE, 0, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.index.Mode(); got != tt.mode {
				t.Errorf("Mode() = %d, want %d", got, tt.mode)
			}
			if got := tt.index.Resolution(); got != tt.res {
				t.Errorf("Resolution() = %d, want %d", got, tt.res)
			}
			if got := tt.index.BaseCell(); got != tt.baseCell {
				t.Errorf("BaseCell() = %d, want %d", got, tt.baseCell)
			}
		})
	}
}

func TestSetters(t *testing.T) {
	h := H3_INIT

	h = SetMode(h, constants.H3_CELL_MODE)
	if h.Mode() != constants.H3_CELL_MODE {
		t.Errorf("SetMode failed: got %d, want %d", h.Mode(), constants.H3_CELL_MODE)
	}

	h = SetResolution(h, 5)
	if h.Resolution() != 5 {
		t.Errorf("SetResolution failed: got %d, want 5", h.Resolution())
	}

	h = SetBaseCell(h, 42)
	if h.BaseCell() != 42 {
		t.Errorf("SetBaseCell failed: got %d, want 42", h.BaseCell())
	}

	h = SetIndexDigit(h, 0, 3)
	if h.IndexDigit(0) != 3 {
		t.Errorf("SetIndexDigit failed: got %d, want 3", h.IndexDigit(0))
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		index H3Index
		valid bool
	}{
		{"null", H3_NULL, false},
		{"valid_res5", validH3Res5, true},
		{"valid_res9", validH3Res9, true},
		{"pentagon_res0", pentagonRes0, true},
		{"init", H3_INIT, false}, // mode=0, invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.index); got != tt.valid {
				t.Errorf("IsValid(%x) = %v, want %v", tt.index, got, tt.valid)
			}
		})
	}
}

func TestIsPentagon(t *testing.T) {
	// Build a known pentagon at resolution 1 from base cell 4.
	h := H3_INIT
	h = SetMode(h, constants.H3_CELL_MODE)
	h = SetResolution(h, 1)
	h = SetBaseCell(h, 4) // Pentagon base cell
	// Set resolution 0 digit to CENTER_DIGIT
	h = SetIndexDigit(h, 0, constants.CENTER_DIGIT)

	if !IsBaseCellPentagon(4) {
		t.Error("Base cell 4 should be a pentagon")
	}

	if !IsPentagon(h) {
		t.Errorf("Index %x should be a pentagon", h)
	}

	// A hexagon cell should not be a pentagon.
	hexagon := validH3Res5
	if IsPentagon(hexagon) {
		t.Errorf("Index %x should not be a pentagon", hexagon)
	}
}

func TestIsResClassIII(t *testing.T) {
	tests := []struct {
		res       int
		classIII  bool
	}{
		{0, false},
		{1, true},
		{2, false},
		{3, true},
		{15, true},
	}

	for _, tt := range tests {
		h := H3_INIT
		h = SetMode(h, constants.H3_CELL_MODE)
		h = SetResolution(h, tt.res)

		if got := IsResClassIII(h); got != tt.classIII {
			t.Errorf("IsResClassIII(res=%d) = %v, want %v", tt.res, got, tt.classIII)
		}
	}
}

func TestToStringFromString(t *testing.T) {
	tests := []struct {
		index H3Index
		str   string
	}{
		{validH3Res5, "85283473fffffff"},
		{validH3Res9, "8928308280fffff"},
		{H3_NULL, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			str := ToString(tt.index)
			if str != tt.str {
				t.Errorf("ToString(%x) = %s, want %s", tt.index, str, tt.str)
			}

			parsed, err := FromString(str)
			if err != nil {
				t.Errorf("FromString(%s) error: %v", str, err)
			}
			if parsed != tt.index {
				t.Errorf("FromString(%s) = %x, want %x", str, parsed, tt.index)
			}
		})
	}
}

func TestFromStringInvalid(t *testing.T) {
	invalidStrings := []string{"xyz", "not_hex", ""}

	for _, s := range invalidStrings {
		_, err := FromString(s)
		if err == nil {
			t.Errorf("FromString(%s) should return error", s)
		}
	}
}

func TestGetNumCells(t *testing.T) {
	tests := []struct {
		res  int
		num  int64
	}{
		{0, 122},       // 2 + 120 * 7^0 = 2 + 120 = 122
		{1, 842},       // 2 + 120 * 7^1 = 2 + 840 = 842
		{2, 5882},      // 2 + 120 * 7^2 = 2 + 5880 = 5882
		{15, 569707381193162}, // 2 + 120 * 7^15
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := GetNumCells(tt.res); got != tt.num {
				t.Errorf("GetNumCells(%d) = %d, want %d", tt.res, got, tt.num)
			}
		})
	}

	// Invalid resolution should return 0.
	if got := GetNumCells(-1); got != 0 {
		t.Errorf("GetNumCells(-1) = %d, want 0", got)
	}
	if got := GetNumCells(16); got != 0 {
		t.Errorf("GetNumCells(16) = %d, want 0", got)
	}
}

func TestGetRes0Cells(t *testing.T) {
	cells := GetRes0Cells()

	if len(cells) != constants.NumBaseCells {
		t.Errorf("GetRes0Cells() returned %d cells, want %d", len(cells), constants.NumBaseCells)
	}

	for i, cell := range cells {
		if cell.Mode() != constants.H3_CELL_MODE {
			t.Errorf("Cell %d has mode %d, want %d", i, cell.Mode(), constants.H3_CELL_MODE)
		}
		if cell.Resolution() != 0 {
			t.Errorf("Cell %d has resolution %d, want 0", i, cell.Resolution())
		}
		if cell.BaseCell() != i {
			t.Errorf("Cell %d has base cell %d, want %d", i, cell.BaseCell(), i)
		}
	}
}

func TestGetPentagonCells(t *testing.T) {
	for res := 0; res <= constants.MaxH3Res; res++ {
		cells := GetPentagonCells(res)

		if len(cells) != constants.NumPentagons {
			t.Errorf("GetPentagonCells(%d) returned %d cells, want %d", res, len(cells), constants.NumPentagons)
		}

		for i, cell := range cells {
			if cell.Mode() != constants.H3_CELL_MODE {
				t.Errorf("Pentagon %d at res %d has mode %d, want %d", i, res, cell.Mode(), constants.H3_CELL_MODE)
			}
			if cell.Resolution() != res {
				t.Errorf("Pentagon %d at res %d has resolution %d", i, res, cell.Resolution())
			}
			if !IsBaseCellPentagon(cell.BaseCell()) {
				t.Errorf("Pentagon %d at res %d has non-pentagon base cell %d", i, res, cell.BaseCell())
			}
			if !IsPentagon(cell) {
				t.Errorf("Cell %x at res %d should be a pentagon", cell, res)
			}
		}
	}

	// Invalid resolutions.
	if cells := GetPentagonCells(-1); cells != nil {
		t.Error("GetPentagonCells(-1) should return nil")
	}
	if cells := GetPentagonCells(16); cells != nil {
		t.Error("GetPentagonCells(16) should return nil")
	}
}

func TestBaseCellNeighbor(t *testing.T) {
	// Test a known neighbor relationship: base cell 0, direction 1 -> base cell 1.
	neighbor := BaseCellNeighbor(0, 1)
	if neighbor != 1 {
		t.Errorf("BaseCellNeighbor(0, 1) = %d, want 1", neighbor)
	}

	// Pentagon base cell 4, direction 1 should be INVALID_BASE_CELL.
	neighbor = BaseCellNeighbor(4, 1)
	if neighbor != INVALID_BASE_CELL {
		t.Errorf("BaseCellNeighbor(4, 1) = %d, want %d", neighbor, INVALID_BASE_CELL)
	}

	// Out of bounds should return INVALID_BASE_CELL.
	neighbor = BaseCellNeighbor(200, 0)
	if neighbor != INVALID_BASE_CELL {
		t.Errorf("BaseCellNeighbor(200, 0) should return INVALID_BASE_CELL")
	}
}

func TestBaseCellNeighborRotations(t *testing.T) {
	// Test known rotation: base cell 0, direction 1 -> 5 CCW rotations.
	rot := BaseCellNeighborRotations(0, 1)
	if rot != 5 {
		t.Errorf("BaseCellNeighborRotations(0, 1) = %d, want 5", rot)
	}

	// Pentagon edge with -1.
	rot = BaseCellNeighborRotations(4, 1)
	if rot != -1 {
		t.Errorf("BaseCellNeighborRotations(4, 1) = %d, want -1", rot)
	}

	// Out of bounds.
	rot = BaseCellNeighborRotations(200, 0)
	if rot != -1 {
		t.Errorf("BaseCellNeighborRotations(200, 0) should return -1")
	}
}

func TestIsBaseCellPentagon(t *testing.T) {
	// All 12 pentagon base cells.
	pentagonCells := constants.PentagonBaseCells
	for _, bc := range pentagonCells {
		if !IsBaseCellPentagon(bc) {
			t.Errorf("Base cell %d should be a pentagon", bc)
		}
	}

	// Non-pentagon cells.
	nonPentagons := []int{0, 1, 2, 3, 5, 6, 7, 8, 9, 10, 100, 121}
	for _, bc := range nonPentagons {
		if IsBaseCellPentagon(bc) {
			t.Errorf("Base cell %d should not be a pentagon", bc)
		}
	}

	// Out of bounds.
	if IsBaseCellPentagon(-1) {
		t.Error("Base cell -1 should not be a pentagon")
	}
	if IsBaseCellPentagon(122) {
		t.Error("Base cell 122 should not be a pentagon")
	}
}

func TestIndexDigit(t *testing.T) {
	// Build an index with known digits.
	h := H3_INIT
	h = SetMode(h, constants.H3_CELL_MODE)
	h = SetResolution(h, 3)
	h = SetBaseCell(h, 10)

	// Set digits at resolution 0, 1, 2.
	h = SetIndexDigit(h, 0, 2)
	h = SetIndexDigit(h, 1, 5)
	h = SetIndexDigit(h, 2, 1)

	// Verify the digits.
	if got := h.IndexDigit(0); got != 2 {
		t.Errorf("IndexDigit(0) = %d, want 2", got)
	}
	if got := h.IndexDigit(1); got != 5 {
		t.Errorf("IndexDigit(1) = %d, want 5", got)
	}
	if got := h.IndexDigit(2); got != 1 {
		t.Errorf("IndexDigit(2) = %d, want 1", got)
	}

	// Digits beyond resolution should be INVALID_DIGIT.
	for r := 3; r < constants.MaxH3Res; r++ {
		if got := h.IndexDigit(r); got != constants.INVALID_DIGIT {
			t.Errorf("IndexDigit(%d) = %d, want %d (INVALID_DIGIT)", r, got, constants.INVALID_DIGIT)
		}
	}
}

func TestGetBaseCellNumber(t *testing.T) {
	// Alias for BaseCell().
	if got := GetBaseCellNumber(validH3Res5); got != 20 {
		t.Errorf("GetBaseCellNumber(validH3Res5) = %d, want 20", got)
	}
}

func BenchmarkIsValid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = IsValid(validH3Res5)
	}
}

func BenchmarkToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ToString(validH3Res5)
	}
}

func BenchmarkFromString(b *testing.B) {
	str := "85283473fffffff"
	for i := 0; i < b.N; i++ {
		_, _ = FromString(str)
	}
}
