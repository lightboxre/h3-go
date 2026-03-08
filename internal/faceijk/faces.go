// Package faceijk — icosahedron face computation.
// Implements getIcosahedronFaces from the H3 C reference implementation.
package faceijk

import (
	"fmt"
	"sort"

	"github.com/lightboxre/h3-go/internal/h3index"
)

// GetCellFaces returns the icosahedron face IDs (0-19) that cell h overlaps.
// Returns sorted unique face IDs. Returns error for invalid cells.
// Mirrors C getIcosahedronFaces.
func GetCellFaces(h h3index.H3Index) ([]int, error) {
	if !h3index.IsValid(h) {
		return nil, fmt.Errorf("invalid H3 cell")
	}

	res := h.Resolution()
	isPent := h3index.IsPentagon(h)

	// Class II pentagons: all vertices lie on icosahedron edges, so redirect
	// to the direct child (Class III). Mirrors C getIcosahedronFaces behavior.
	// Note: Go SetIndexDigit uses 0-based position (pos=parentRes), while C
	// H3_SET_INDEX_DIGIT uses 1-based (pos=childRes). Both compute the same bit shift.
	if isPent && res%2 == 0 {
		child := h3index.SetResolution(h, res+1)
		child = h3index.SetIndexDigit(child, res, 0) // CENTER_DIGIT = 0, Go 0-based pos
		return GetCellFaces(child)
	}

	fijk := H3ToFaceIJK(h)

	faceMap := map[int]bool{}
	faceMap[fijk.Face] = true // center face always included

	adjRes := res
	centerIJK := fijk

	if isPent {
		// faceIjkPentToVerts modifies adjRes (increments for Class III)
		fijkVerts := faceIjkPentToVerts(&centerIJK, &adjRes)
		for _, vf := range fijkVerts {
			adjustPentVertOverage(&vf, adjRes)
			faceMap[vf.Face] = true
		}
	} else {
		// faceIjkToVerts modifies adjRes (increments for Class III)
		fijkVerts := faceIjkToVerts(&centerIJK, &adjRes)
		for _, vf := range fijkVerts {
			adjustOverageClassII(&vf, adjRes, 0, 1)
			faceMap[vf.Face] = true
		}
	}

	faces := make([]int, 0, len(faceMap))
	for face := range faceMap {
		faces = append(faces, face)
	}
	sort.Ints(faces)
	return faces, nil
}
