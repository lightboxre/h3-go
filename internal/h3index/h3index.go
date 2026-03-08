// Package h3index provides core H3Index encoding/decoding and validation.
// H3Index is a 64-bit packed integer representing a hierarchical geospatial index.
package h3index

import (
	"fmt"
	"strconv"

	"github.com/lightboxre/h3-go/internal/constants"
)

// H3Index is a 64-bit packed geospatial index.
//
// Bit layout (from MSB to LSB):
//   - Bit 63:    reserved (always 0)
//   - Bits 59-62: mode (1=cell, 2=directed edge, 3=vertex)
//   - Bits 56-58: reserved (always 0)
//   - Bits 52-55: resolution (0-15)
//   - Bits 45-51: base cell (0-121)
//   - Bits 0-44:  digit sequence (3 bits per digit, 15 digits)
type H3Index uint64

// Constants for H3Index.
const (
	// H3_NULL is the invalid/null H3 index.
	H3_NULL H3Index = 0

	// H3_INIT is the base initializer with all digits set to invalid (7).
	// C value: UINT64_C(35184372088831) = 0x1FFFFFFFFFFF
	H3_INIT H3Index = 35184372088831
)

// Mode returns the mode field of the H3 index.
// Valid modes: 1=cell, 2=directed edge, 3=vertex.
func (h H3Index) Mode() int {
	return int((uint64(h) >> constants.H3_MODE_OFFSET) & constants.H3_MODE_MASK)
}

// Resolution returns the resolution field (0-15).
func (h H3Index) Resolution() int {
	return int((uint64(h) >> constants.H3_RES_OFFSET) & constants.H3_RES_MASK)
}

// BaseCell returns the base cell number (0-121).
func (h H3Index) BaseCell() int {
	return int((uint64(h) >> constants.H3_BC_OFFSET) & constants.H3_BC_MASK)
}

// IndexDigit returns the digit at the given resolution level (0-14).
// Digit encoding: bits (15-1-res)*3 to (15-1-res)*3+2.
func (h H3Index) IndexDigit(res int) int {
	return int((uint64(h) >> ((constants.MaxH3Res - 1 - res) * constants.H3_PER_DIGIT_OFFSET)) & constants.H3_DIGIT_MASK)
}

// SetMode returns a new H3Index with the mode field set.
func SetMode(h H3Index, mode int) H3Index {
	return H3Index((uint64(h) & ^(constants.H3_MODE_MASK << constants.H3_MODE_OFFSET)) |
		(uint64(mode&int(constants.H3_MODE_MASK)) << constants.H3_MODE_OFFSET))
}

// SetResolution returns a new H3Index with the resolution field set.
func SetResolution(h H3Index, res int) H3Index {
	return H3Index((uint64(h) & ^(constants.H3_RES_MASK << constants.H3_RES_OFFSET)) |
		(uint64(res&int(constants.H3_RES_MASK)) << constants.H3_RES_OFFSET))
}

// SetBaseCell returns a new H3Index with the base cell field set.
func SetBaseCell(h H3Index, bc int) H3Index {
	return H3Index((uint64(h) & ^(constants.H3_BC_MASK << constants.H3_BC_OFFSET)) |
		(uint64(bc&int(constants.H3_BC_MASK)) << constants.H3_BC_OFFSET))
}

// SetIndexDigit returns a new H3Index with the digit at resolution res set.
func SetIndexDigit(h H3Index, res, digit int) H3Index {
	shift := (constants.MaxH3Res - 1 - res) * constants.H3_PER_DIGIT_OFFSET
	return H3Index((uint64(h) & ^(constants.H3_DIGIT_MASK << shift)) |
		(uint64(digit&int(constants.H3_DIGIT_MASK)) << shift))
}

// IsValid checks if the H3Index is a valid cell index.
//
// Validation rules:
//   - Mode must be H3_CELL_MODE (1)
//   - Resolution must be 0-15
//   - Base cell must be 0-121
//   - All digits at resolution levels 0 to res-1 must be valid (0-6)
//   - All digits at resolution levels res to 14 must be INVALID_DIGIT (7)
//   - Pentagon cells must not have K-axes digits (1) except when valid
func IsValid(h H3Index) bool {
	if h == H3_NULL {
		return false
	}

	mode := h.Mode()
	if mode != constants.H3_CELL_MODE {
		return false
	}

	res := h.Resolution()
	if res < 0 || res > constants.MaxH3Res {
		return false
	}

	bc := h.BaseCell()
	if bc < 0 || bc >= constants.NumBaseCells {
		return false
	}

	// Check digit validity.
	foundFirstNonZeroDigit := false
	for r := 0; r < res; r++ {
		digit := h.IndexDigit(r)
		if digit < 0 || digit >= constants.NUM_DIGITS {
			return false
		}

		// For pentagon cells, check for deleted subsequence.
		// The first nonzero digit cannot be K_AXES_DIGIT (1).
		if IsBaseCellPentagon(bc) && !foundFirstNonZeroDigit {
			if digit != constants.CENTER_DIGIT {
				foundFirstNonZeroDigit = true
				if digit == constants.K_AXES_DIGIT {
					return false
				}
			}
		}
	}

	// All digits beyond resolution must be INVALID_DIGIT (7).
	for r := res; r < constants.MaxH3Res; r++ {
		digit := h.IndexDigit(r)
		if digit != constants.INVALID_DIGIT {
			return false
		}
	}

	return true
}

// IsPentagon returns true if the cell is a pentagon.
// A cell is a pentagon if its base cell is pentagonal and all
// resolution digits are CENTER_DIGIT (0).
func IsPentagon(h H3Index) bool {
	if !IsBaseCellPentagon(h.BaseCell()) {
		return false
	}

	// Check if there is any non-zero leading digit.
	res := h.Resolution()
	for r := 0; r < res; r++ {
		if h.IndexDigit(r) != constants.CENTER_DIGIT {
			return false
		}
	}

	return true
}

// IsResClassIII returns true if the resolution is Class III (odd resolution).
// Class III resolutions have rotated hex grids; Class II (even) are aligned.
func IsResClassIII(h H3Index) bool {
	return (h.Resolution() % 2) == 1
}

// ToString converts an H3Index to a hexadecimal string.
func ToString(h H3Index) string {
	return fmt.Sprintf("%x", uint64(h))
}

// FromString parses a hexadecimal string into an H3Index.
// Returns an error if the string is invalid.
func FromString(s string) (H3Index, error) {
	val, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return H3_NULL, fmt.Errorf("invalid H3 string: %w", err)
	}
	return H3Index(val), nil
}

// GetNumCells returns the total number of cells at the given resolution.
// Formula: 2 + 120 * 7^res
func GetNumCells(res int) int64 {
	if res < 0 || res > constants.MaxH3Res {
		return 0
	}
	return 2 + 120*ipow7(res)
}

// ipow7 computes 7^n efficiently.
func ipow7(n int) int64 {
	if n == 0 {
		return 1
	}
	result := int64(1)
	base := int64(7)
	exp := n
	for exp > 0 {
		if exp&1 == 1 {
			result *= base
		}
		base *= base
		exp >>= 1
	}
	return result
}

// GetRes0Cells returns all 122 resolution-0 base cells as H3 indices.
func GetRes0Cells() []H3Index {
	cells := make([]H3Index, constants.NumBaseCells)
	for bc := 0; bc < constants.NumBaseCells; bc++ {
		h := H3_INIT
		h = SetMode(h, constants.H3_CELL_MODE)
		h = SetResolution(h, 0)
		h = SetBaseCell(h, bc)
		cells[bc] = h
	}
	return cells
}

// GetPentagonCells returns all 12 pentagon cells at the given resolution.
func GetPentagonCells(res int) []H3Index {
	if res < 0 || res > constants.MaxH3Res {
		return nil
	}

	cells := make([]H3Index, constants.NumPentagons)
	for i, bc := range constants.PentagonBaseCells {
		h := H3_INIT
		h = SetMode(h, constants.H3_CELL_MODE)
		h = SetResolution(h, res)
		h = SetBaseCell(h, bc)
		// All digits are already set to INVALID_DIGIT by H3_INIT.
		// Set resolution digits to CENTER_DIGIT (0).
		for r := 0; r < res; r++ {
			h = SetIndexDigit(h, r, constants.CENTER_DIGIT)
		}
		cells[i] = h
	}
	return cells
}

// GetBaseCellNumber returns the base cell number for the H3 index.
// Alias for BaseCell() method.
func GetBaseCellNumber(h H3Index) int {
	return h.BaseCell()
}
