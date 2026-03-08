// Package faceijk — cell boundary computation.
// Implements _faceIjkToCellBoundary and _faceIjkPentToCellBoundary from the
// H3 C reference implementation (src/h3lib/lib/faceijk.c).
package faceijk

import (
	"math"

	"github.com/lightboxre/h3-go/internal/constants"
	"github.com/lightboxre/h3-go/internal/coordijk"
)

// faceOrientIJK holds adjacent-face orientation info for _adjustOverageClassII.
// Corresponds to C: FaceOrientIJK {int face; CoordIJK translate; int ccwRot60}
type faceOrientIJK struct {
	face      int
	translate coordijk.CoordIJK
	ccwRot60  int
}

// faceNeighbors maps (face, quadrant) → adjacent face orientation.
// Indexed as faceNeighbors[face][0=central, 1=IJ, 2=KI, 3=JK].
// Transcribed from C: faceNeighbors[NUM_ICOSA_FACES][4] in faceijk.c.
var faceNeighbors = [20][4]faceOrientIJK{
	{ // face 0
		{0, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{4, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{1, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{5, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 1
		{1, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{0, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{2, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{6, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 2
		{2, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{1, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{3, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{7, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 3
		{3, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{2, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{4, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{8, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 4
		{4, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{3, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{0, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{9, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 5
		{5, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{10, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{14, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{0, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 6
		{6, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{11, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{10, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{1, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 7
		{7, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{12, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{11, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{2, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 8
		{8, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{13, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{12, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{3, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 9
		{9, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{14, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{13, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{4, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 10
		{10, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{5, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{6, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{15, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 11
		{11, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{6, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{7, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{16, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 12
		{12, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{7, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{8, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{17, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 13
		{13, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{8, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{9, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{18, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 14
		{14, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{9, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 3},
		{5, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 3},
		{19, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 15
		{15, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{16, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{19, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{10, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 16
		{16, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{17, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{15, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{11, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 17
		{17, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{18, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{16, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{12, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 18
		{18, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{19, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{17, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{13, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
	{ // face 19
		{19, coordijk.CoordIJK{I: 0, J: 0, K: 0}, 0},
		{15, coordijk.CoordIJK{I: 2, J: 0, K: 2}, 1},
		{18, coordijk.CoordIJK{I: 2, J: 2, K: 0}, 5},
		{14, coordijk.CoordIJK{I: 0, J: 2, K: 2}, 3},
	},
}

// adjacentFaceDir maps pairs of faces to the quadrant direction (IJ/KI/JK/-1).
// adjacentFaceDir[face1][face2] = direction of face2 as seen from face1.
// -1 means not adjacent. Transcribed from C faceijk.c.
var adjacentFaceDir = [20][20]int{
	{0, KI, -1, -1, IJ, JK, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // 0
	{IJ, 0, KI, -1, -1, -1, JK, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // 1
	{-1, IJ, 0, KI, -1, -1, -1, JK, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // 2
	{-1, -1, IJ, 0, KI, -1, -1, -1, JK, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // 3
	{KI, -1, -1, IJ, 0, -1, -1, -1, -1, JK, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // 4
	{JK, -1, -1, -1, -1, 0, -1, -1, -1, -1, IJ, -1, -1, -1, KI, -1, -1, -1, -1, -1}, // 5
	{-1, JK, -1, -1, -1, -1, 0, -1, -1, -1, KI, IJ, -1, -1, -1, -1, -1, -1, -1, -1}, // 6
	{-1, -1, JK, -1, -1, -1, -1, 0, -1, -1, -1, KI, IJ, -1, -1, -1, -1, -1, -1, -1}, // 7
	{-1, -1, -1, JK, -1, -1, -1, -1, 0, -1, -1, -1, KI, IJ, -1, -1, -1, -1, -1, -1}, // 8
	{-1, -1, -1, -1, JK, -1, -1, -1, -1, 0, -1, -1, -1, KI, IJ, -1, -1, -1, -1, -1}, // 9
	{-1, -1, -1, -1, -1, IJ, KI, -1, -1, -1, 0, -1, -1, -1, -1, JK, -1, -1, -1, -1}, // 10
	{-1, -1, -1, -1, -1, -1, IJ, KI, -1, -1, -1, 0, -1, -1, -1, -1, JK, -1, -1, -1}, // 11
	{-1, -1, -1, -1, -1, -1, -1, IJ, KI, -1, -1, -1, 0, -1, -1, -1, -1, JK, -1, -1}, // 12
	{-1, -1, -1, -1, -1, -1, -1, -1, IJ, KI, -1, -1, -1, 0, -1, -1, -1, -1, JK, -1}, // 13
	{-1, -1, -1, -1, -1, KI, -1, -1, -1, IJ, -1, -1, -1, -1, 0, -1, -1, -1, -1, JK}, // 14
	{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, JK, -1, -1, -1, -1, 0, IJ, -1, -1, KI}, // 15
	{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, JK, -1, -1, -1, KI, 0, IJ, -1, -1}, // 16
	{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, JK, -1, -1, -1, KI, 0, IJ, -1}, // 17
	{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, JK, -1, -1, -1, KI, 0, IJ}, // 18
	{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, JK, IJ, -1, -1, KI, 0}, // 19
}

// adjustOverageClassII adjusts a FaceIJK address for overage in a substrate
// grid (Class II resolution). Returns the overage type.
// From C: _adjustOverageClassII(fijk, res, pentLeading4, substrate)
// Called with substrate=1 for boundary vertex computation.
func adjustOverageClassII(fijk *FaceIJK, res, pentLeading4, substrate int) int {
	maxDim := constants.MaxDimByCIIres[res]
	if substrate != 0 {
		maxDim *= 3
	}

	sum := fijk.Coord.I + fijk.Coord.J + fijk.Coord.K
	if substrate != 0 && sum == maxDim {
		return FACE_EDGE
	}
	if sum <= maxDim {
		return NO_OVERAGE
	}

	// Overage: determine which face quadrant we're in.
	var orient faceOrientIJK
	if fijk.Coord.K > 0 {
		if fijk.Coord.J > 0 {
			orient = faceNeighbors[fijk.Face][JK]
		} else {
			orient = faceNeighbors[fijk.Face][KI]
			if pentLeading4 != 0 {
				// Translate origin to center of pentagon, rotate CW, translate back.
				origin := coordijk.CoordIJK{I: maxDim, J: 0, K: 0}
				tmp := coordijk.CoordIJK{
					I: fijk.Coord.I - origin.I,
					J: fijk.Coord.J - origin.J,
					K: fijk.Coord.K - origin.K,
				}
				tmp = coordijk.Rotate60CW(tmp)
				fijk.Coord.I = tmp.I + origin.I
				fijk.Coord.J = tmp.J + origin.J
				fijk.Coord.K = tmp.K + origin.K
			}
		}
	} else {
		orient = faceNeighbors[fijk.Face][IJ]
	}

	fijk.Face = orient.face

	// Apply CCW rotations.
	for i := 0; i < orient.ccwRot60; i++ {
		fijk.Coord = coordijk.Rotate60CCW(fijk.Coord)
	}

	// Scale and apply translation.
	unitScale := int(constants.UnitScaleByCIIres[res])
	if substrate != 0 {
		unitScale *= 3
	}
	// Raw scale: no normalize until after add (matches C _ijkScale + _ijkAdd then _ijkNormalize).
	fijk.Coord.I += orient.translate.I * unitScale
	fijk.Coord.J += orient.translate.J * unitScale
	fijk.Coord.K += orient.translate.K * unitScale
	coordijk.IJKNormalize(&fijk.Coord)

	// After correction, check if now exactly on edge.
	if substrate != 0 && fijk.Coord.I+fijk.Coord.J+fijk.Coord.K == maxDim {
		return FACE_EDGE
	}
	return NEW_FACE
}

// adjustPentVertOverage loops adjustOverageClassII until not on a new face.
// From C: _adjustPentVertOverage
func adjustPentVertOverage(fijk *FaceIJK, res int) int {
	var overage int
	for {
		overage = adjustOverageClassII(fijk, res, 0, 1)
		if overage != NEW_FACE {
			break
		}
	}
	return overage
}

// faceIjkToVerts computes the 6 substrate-grid vertex FaceIJK coordinates
// for a hexagon cell. Modifies fijk in place; updates adjRes.
// From C: _faceIjkToVerts
func faceIjkToVerts(fijk *FaceIJK, adjRes *int) [6]FaceIJK {
	// Class II substrate vertex offsets (aperture 33r)
	vertsCII := [6]coordijk.CoordIJK{
		{I: 2, J: 1, K: 0},
		{I: 1, J: 2, K: 0},
		{I: 0, J: 2, K: 1},
		{I: 0, J: 1, K: 2},
		{I: 1, J: 0, K: 2},
		{I: 2, J: 0, K: 1},
	}
	// Class III substrate vertex offsets (aperture 33r7r)
	vertsCIII := [6]coordijk.CoordIJK{
		{I: 5, J: 4, K: 0},
		{I: 1, J: 5, K: 0},
		{I: 0, J: 5, K: 4},
		{I: 0, J: 1, K: 5},
		{I: 4, J: 0, K: 5},
		{I: 5, J: 0, K: 1},
	}

	var verts [6]coordijk.CoordIJK
	if *adjRes%2 == 1 {
		verts = vertsCIII
	} else {
		verts = vertsCII
	}

	// Adjust center to aperture 33r substrate grid.
	fijk.Coord = coordijk.DownAp3(fijk.Coord)
	fijk.Coord = coordijk.DownAp3r(fijk.Coord)

	// For Class III, add aperture 7r to reach icosahedral Class II.
	if *adjRes%2 == 1 {
		fijk.Coord = coordijk.DownAp7r(fijk.Coord)
		*adjRes++
	}

	// Compute each vertex: center + offset, then normalize.
	var fijkVerts [6]FaceIJK
	for v := 0; v < 6; v++ {
		fijkVerts[v].Face = fijk.Face
		fijkVerts[v].Coord = coordijk.CoordIJK{
			I: fijk.Coord.I + verts[v].I,
			J: fijk.Coord.J + verts[v].J,
			K: fijk.Coord.K + verts[v].K,
		}
		coordijk.IJKNormalize(&fijkVerts[v].Coord)
	}
	return fijkVerts
}

// faceIjkPentToVerts computes the 5 substrate-grid vertex FaceIJK coordinates
// for a pentagon cell. Modifies fijk in place; updates adjRes.
// From C: _faceIjkPentToVerts
func faceIjkPentToVerts(fijk *FaceIJK, adjRes *int) [5]FaceIJK {
	// Class II substrate vertex offsets (aperture 33r)
	vertsCII := [5]coordijk.CoordIJK{
		{I: 2, J: 1, K: 0},
		{I: 1, J: 2, K: 0},
		{I: 0, J: 2, K: 1},
		{I: 0, J: 1, K: 2},
		{I: 1, J: 0, K: 2},
	}
	// Class III substrate vertex offsets (aperture 33r7r)
	vertsCIII := [5]coordijk.CoordIJK{
		{I: 5, J: 4, K: 0},
		{I: 1, J: 5, K: 0},
		{I: 0, J: 5, K: 4},
		{I: 0, J: 1, K: 5},
		{I: 4, J: 0, K: 5},
	}

	var verts [5]coordijk.CoordIJK
	if *adjRes%2 == 1 {
		verts = vertsCIII
	} else {
		verts = vertsCII
	}

	// Adjust center to aperture 33r substrate grid.
	fijk.Coord = coordijk.DownAp3(fijk.Coord)
	fijk.Coord = coordijk.DownAp3r(fijk.Coord)

	// For Class III, add aperture 7r to reach icosahedral Class II.
	if *adjRes%2 == 1 {
		fijk.Coord = coordijk.DownAp7r(fijk.Coord)
		*adjRes++
	}

	// Compute each vertex: center + offset, then normalize.
	var fijkVerts [5]FaceIJK
	for v := 0; v < 5; v++ {
		fijkVerts[v].Face = fijk.Face
		fijkVerts[v].Coord = coordijk.CoordIJK{
			I: fijk.Coord.I + verts[v].I,
			J: fijk.Coord.J + verts[v].J,
			K: fijk.Coord.K + verts[v].K,
		}
		coordijk.IJKNormalize(&fijkVerts[v].Coord)
	}
	return fijkVerts
}

// hex2dToGeoSubstrate is _hex2dToGeo with substrate=1.
// Converts a substrate-grid Vec2d to lat/lng in radians.
// When substrate=1 the scaling includes an extra /3 factor and the
// Class III rotation is NOT applied (it was already handled by DownAp7r).
func hex2dToGeoSubstrate(v coordijk.Vec2d, face, res int) (latRad, lngRad float64) {
	center := faceCenterGeo[face]
	r := math.Sqrt(v.X*v.X + v.Y*v.Y)
	if r < constants.EPSILON {
		return center.Lat, center.Lng
	}

	theta := math.Atan2(v.Y, v.X)

	// Reverse resolution scaling.
	for i := 0; i < res; i++ {
		r *= M_RSQRT7
	}

	// Substrate scale: divide by 3; if ClassIII also divide by sqrt(7).
	r /= 3.0
	if res%2 == 1 {
		r *= M_RSQRT7
	}

	r *= constants.RES0_U_GNOMONIC
	r = math.Atan(r)

	// No Class III theta rotation when substrate=1 (already handled).
	theta = PosAngleRads(faceAxesAzRadsCII[face][0] - theta)

	return geoAzDistanceRads(center.Lat, center.Lng, theta, r)
}

// v2dIntersect returns the intersection of line segments (p0,p1) and (p2,p3).
// From C: _v2dIntersect
func v2dIntersect(p0, p1, p2, p3 coordijk.Vec2d) coordijk.Vec2d {
	s1x := p1.X - p0.X
	s1y := p1.Y - p0.Y
	s2x := p3.X - p2.X
	s2y := p3.Y - p2.Y

	t := (s2x*(p0.Y-p2.Y) - s2y*(p0.X-p2.X)) / (-s2x*s1y + s1x*s2y)

	return coordijk.Vec2d{
		X: p0.X + t*s1x,
		Y: p0.Y + t*s1y,
	}
}

// v2dAlmostEquals returns true if two Vec2d are very close.
// From C: _v2dAlmostEquals (uses FLT_EPSILON ≈ 1.19e-7)
func v2dAlmostEquals(v1, v2 coordijk.Vec2d) bool {
	const fltEpsilon = 1.1920928955078125e-7
	return math.Abs(v1.X-v2.X) < fltEpsilon && math.Abs(v1.Y-v2.Y) < fltEpsilon
}

// FaceIJKToGeoBoundary computes the cell boundary polygon for a face-IJK cell.
// For hexagons (isPentagon=false): returns 6 vertices (may have extra edge vertices).
// For pentagons (isPentagon=true): returns 5 vertices.
// Implements _faceIjkToCellBoundary and _faceIjkPentToCellBoundary from C.
func FaceIJKToGeoBoundary(fijk FaceIJK, res int, isPentagon bool) []GeoPoint {
	if isPentagon {
		return faceIjkPentToCellBoundary(fijk, res)
	}
	return faceIjkToCellBoundary(fijk, res)
}

// faceIjkToCellBoundary computes the hexagon cell boundary.
// From C: _faceIjkToCellBoundary(h, res, 0, NUM_HEX_VERTS, g)
func faceIjkToCellBoundary(h FaceIJK, res int) []GeoPoint {
	const numHexVerts = 6

	adjRes := res
	centerIJK := h
	fijkVerts := faceIjkToVerts(&centerIJK, &adjRes)

	maxDim := constants.MaxDimByCIIres[adjRes]

	// For a full boundary (length == numHexVerts), we do one extra iteration
	// to detect face crossings for Class III cells.
	var boundary []GeoPoint
	lastFace := -1
	lastOverage := NO_OVERAGE

	for vert := 0; vert < numHexVerts+1; vert++ {
		v := vert % numHexVerts

		fv := fijkVerts[v]
		overage := adjustOverageClassII(&fv, adjRes, 0, 1)

		// For Class III cells crossing face boundaries: insert edge intersection.
		if res%2 == 1 && vert > 0 && fv.Face != lastFace && lastOverage != FACE_EDGE {
			lastV := (v + numHexVerts - 1) % numHexVerts

			orig2d0 := coordijk.ToVec2d(fijkVerts[lastV].Coord)
			orig2d1 := coordijk.ToVec2d(fijkVerts[v].Coord)

			v0 := coordijk.Vec2d{X: 3.0 * float64(maxDim), Y: 0.0}
			v1 := coordijk.Vec2d{X: -1.5 * float64(maxDim), Y: 3.0 * constants.M_SQRT3_2 * float64(maxDim)}
			v2 := coordijk.Vec2d{X: -1.5 * float64(maxDim), Y: -3.0 * constants.M_SQRT3_2 * float64(maxDim)}

			face2 := fv.Face
			if lastFace == centerIJK.Face {
				face2 = fv.Face
			} else {
				face2 = lastFace
			}

			var edge0, edge1 coordijk.Vec2d
			switch adjacentFaceDir[centerIJK.Face][face2] {
			case IJ:
				edge0, edge1 = v0, v1
			case JK:
				edge0, edge1 = v1, v2
			default: // KI
				edge0, edge1 = v2, v0
			}

			inter := v2dIntersect(orig2d0, orig2d1, edge0, edge1)
			if !v2dAlmostEquals(orig2d0, inter) && !v2dAlmostEquals(orig2d1, inter) {
				lat, lng := hex2dToGeoSubstrate(inter, centerIJK.Face, adjRes)
				boundary = append(boundary, GeoPoint{Lat: lat, Lng: lng})
			}
		}

		if vert < numHexVerts {
			vec := coordijk.ToVec2d(fv.Coord)
			lat, lng := hex2dToGeoSubstrate(vec, fv.Face, adjRes)
			boundary = append(boundary, GeoPoint{Lat: lat, Lng: lng})
		}

		lastFace = fv.Face
		lastOverage = overage
	}

	return boundary
}

// faceIjkPentToCellBoundary computes the pentagon cell boundary.
// From C: _faceIjkPentToCellBoundary(h, res, 0, NUM_PENT_VERTS, g)
func faceIjkPentToCellBoundary(h FaceIJK, res int) []GeoPoint {
	const numPentVerts = 5

	adjRes := res
	centerIJK := h
	fijkVerts := faceIjkPentToVerts(&centerIJK, &adjRes)

	maxDim := constants.MaxDimByCIIres[adjRes]

	var boundary []GeoPoint
	var lastFijk FaceIJK

	for vert := 0; vert < numPentVerts+1; vert++ {
		v := vert % numPentVerts

		fv := fijkVerts[v]
		adjustPentVertOverage(&fv, adjRes)

		// For Class III cells crossing face boundaries: insert edge intersection.
		// Only insert when consecutive vertices are on different faces.
		if res%2 == 1 && vert > 0 && fv.Face != lastFijk.Face {
			tmpFijk := fv

			orig2d0 := coordijk.ToVec2d(lastFijk.Coord)

			currentToLastDir := adjacentFaceDir[tmpFijk.Face][lastFijk.Face]
			if currentToLastDir >= 1 && currentToLastDir <= 3 {
				orient := faceNeighbors[tmpFijk.Face][currentToLastDir]
				tmpFijk.Face = orient.face
				for i := 0; i < orient.ccwRot60; i++ {
					tmpFijk.Coord = coordijk.Rotate60CCW(tmpFijk.Coord)
				}
				unitScale := int(constants.UnitScaleByCIIres[adjRes]) * 3
				tmpFijk.Coord.I += orient.translate.I * unitScale
				tmpFijk.Coord.J += orient.translate.J * unitScale
				tmpFijk.Coord.K += orient.translate.K * unitScale
				coordijk.IJKNormalize(&tmpFijk.Coord)
			}

			orig2d1 := coordijk.ToVec2d(tmpFijk.Coord)

			v0 := coordijk.Vec2d{X: 3.0 * float64(maxDim), Y: 0.0}
			v1 := coordijk.Vec2d{X: -1.5 * float64(maxDim), Y: 3.0 * constants.M_SQRT3_2 * float64(maxDim)}
			v2 := coordijk.Vec2d{X: -1.5 * float64(maxDim), Y: -3.0 * constants.M_SQRT3_2 * float64(maxDim)}

			var edge0, edge1 coordijk.Vec2d
			switch adjacentFaceDir[tmpFijk.Face][fv.Face] {
			case IJ:
				edge0, edge1 = v0, v1
			case JK:
				edge0, edge1 = v1, v2
			default: // KI
				edge0, edge1 = v2, v0
			}

			inter := v2dIntersect(orig2d0, orig2d1, edge0, edge1)
			lat, lng := hex2dToGeoSubstrate(inter, tmpFijk.Face, adjRes)
			boundary = append(boundary, GeoPoint{Lat: lat, Lng: lng})
		}

		if vert < numPentVerts {
			vec := coordijk.ToVec2d(fv.Coord)
			lat, lng := hex2dToGeoSubstrate(vec, fv.Face, adjRes)
			boundary = append(boundary, GeoPoint{Lat: lat, Lng: lng})
		}

		lastFijk = fv
	}

	return boundary
}
