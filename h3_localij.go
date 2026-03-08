package h3

import (
	"github.com/lightboxre/h3-go/internal/coordijk"
	"github.com/lightboxre/h3-go/internal/h3index"
)

// CellToLocalIj returns the local IJ coordinates of cell relative to origin.
// mode must be 0 (reserved for future use).
//
// Returns error if:
//   - Cells have different resolutions (ErrResMismatch)
//   - Cells are not in the same local coordinate space (ErrFailed)
//   - Either cell is invalid (ErrCellInvalid)
//   - mode != 0 (ErrOptionInvalid)
//
// The local IJ coordinate space is anchored by the origin cell and uses a
// 2D coordinate system where I and J axes are 60 degrees apart.
// This function is useful for efficient grid distance calculations and
// grid traversal within a local region.
//
// Note: This is currently a stub implementation. Full pentagon rotation
// handling needs to be ported from H3 C library localij.c.
func CellToLocalIj(origin, cell Cell, mode uint32) (CoordIJ, error) {
	if mode != 0 {
		return CoordIJ{}, coordijk.ErrOptionInvalid
	}

	ijk, err := coordijk.H3ToLocalIJK(h3index.H3Index(origin), h3index.H3Index(cell))
	if err != nil {
		return CoordIJ{}, err
	}

	i, j := coordijk.IJKToIJ(ijk)
	return CoordIJ{I: i, J: j}, nil
}

// LocalIjToCell returns the cell at local IJ coordinates relative to origin.
// mode must be 0 (reserved for future use).
//
// Returns error if:
//   - Origin cell is invalid (ErrCellInvalid)
//   - Coordinates are out of range (ErrFailed)
//   - mode != 0 (ErrOptionInvalid)
//
// This is the inverse of CellToLocalIj.
//
// Note: This is currently a stub implementation. Full pentagon rotation
// handling needs to be ported from H3 C library localij.c.
func LocalIjToCell(origin Cell, ij CoordIJ, mode uint32) (Cell, error) {
	if mode != 0 {
		return Cell(h3index.H3_NULL), coordijk.ErrOptionInvalid
	}

	ijk := coordijk.IJToIJK(ij.I, ij.J)
	h, err := coordijk.LocalIJKToH3(h3index.H3Index(origin), ijk)
	if err != nil {
		return Cell(h3index.H3_NULL), err
	}

	return Cell(h), nil
}
