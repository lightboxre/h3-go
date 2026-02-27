// Package faceijk implements face-IJK coordinate system conversions.
// This system maps lat/lng to/from icosahedron faces with IJK coordinates.
package faceijk

import (
	"math"

	"github.com/EDRInc/h3-go/internal/constants"
	"github.com/EDRInc/h3-go/internal/coordijk"
	"github.com/EDRInc/h3-go/internal/h3index"
)

// FaceIJK represents an H3 index on a specific icosahedron face.
type FaceIJK struct {
	Face  int               // icosahedron face number (0-19)
	Coord coordijk.CoordIJK // IJK coordinates on that face
}

// GeoPoint is a lat/lng pair in radians.
type GeoPoint struct {
	Lat, Lng float64
}

// Quadrant direction constants for adjacent face relationships.
const (
	IJ           = 1
	KI           = 2
	JK           = 3
	INVALID_FACE = -1
)

// Overage classification for boundary conditions.
const (
	NO_OVERAGE = 0
	FACE_EDGE  = 1
	NEW_FACE   = 2
)

// M_SQRT7 and M_RSQRT7 constants from C source.
const (
	M_SQRT7  = 2.6457513110645905905016157536392604257102
	M_RSQRT7 = 0.37796447300922722721451653623418006081576
)

// faceCenterGeo contains lat/lng (radians) of icosahedron face centers.
// Transcribed from C: src/h3lib/lib/faceijk.c
var faceCenterGeo = [20]GeoPoint{
	{0.803582649718989942, 1.248397419617396099},   // face  0
	{1.307747883455638156, 2.536945009877921159},   // face  1
	{1.054751253523952054, -1.347517358900396623},  // face  2
	{0.600191595538186799, -0.450603909469755746},  // face  3
	{0.491715428198773866, 0.401988202911306943},   // face  4
	{0.172745327415618701, 1.678146885280433686},   // face  5
	{0.605929321571350690, 2.953923329812411617},   // face  6
	{0.427370518328979641, -1.888876200336285401},  // face  7
	{-0.079066118549212831, -0.733429513380867741}, // face  8
	{-0.230961644455383637, 0.506495587332349035},  // face  9
	{0.079066118549212831, 2.408163140208925497},   // face 10
	{0.230961644455383637, -2.635097066257444203},  // face 11
	{-0.172745327415618701, -1.463445768309359553}, // face 12
	{-0.605929321571350690, -0.187669323777381622}, // face 13
	{-0.427370518328979641, 1.252716453253507838},  // face 14
	{-0.600191595538186799, 2.690988744120037492},  // face 15
	{-0.491715428198773866, -2.739604450678486295}, // face 16
	{-0.803582649718989942, -1.893195233972397139}, // face 17
	{-1.307747883455638156, -0.604647643711872080}, // face 18
	{-1.054751253523952054, 1.794075294689396615},  // face 19
}

// faceCenterPoint contains 3D unit vectors of face centers.
// Transcribed from C: src/h3lib/lib/faceijk.c
var faceCenterPoint = [20]coordijk.Vec3d{
	{X: 0.2199307791404606, Y: 0.6583691780274996, Z: 0.7198475378926182},    // face  0
	{X: -0.2139234834501421, Y: 0.1478171829550703, Z: 0.9656017935214205},   // face  1
	{X: 0.1092625278784797, Y: -0.4811951572873210, Z: 0.8697775121287253},   // face  2
	{X: 0.7428567301586791, Y: -0.3593941678278028, Z: 0.5648005936517033},   // face  3
	{X: 0.8112534709140969, Y: 0.3448953237639384, Z: 0.4721387736413930},    // face  4
	{X: -0.1055498149613921, Y: 0.9794457296411413, Z: 0.1718874610009365},   // face  5
	{X: -0.8075407579970092, Y: 0.1533552485898818, Z: 0.5695261994882688},   // face  6
	{X: -0.2846148069787907, Y: -0.8644080972654206, Z: 0.4144792552473539},  // face  7
	{X: 0.7405621473854482, Y: -0.6673299564565524, Z: -0.0789837646326737},  // face  8
	{X: 0.8512303986474293, Y: 0.4722343788582681, Z: -0.2289137388687808},   // face  9
	{X: -0.7405621473854481, Y: 0.6673299564565524, Z: 0.0789837646326737},   // face 10
	{X: -0.8512303986474292, Y: -0.4722343788582682, Z: 0.2289137388687808},  // face 11
	{X: 0.1055498149613919, Y: -0.9794457296411413, Z: -0.1718874610009365},  // face 12
	{X: 0.8075407579970092, Y: -0.1533552485898819, Z: -0.5695261994882688},  // face 13
	{X: 0.2846148069787908, Y: 0.8644080972654204, Z: -0.4144792552473539},   // face 14
	{X: -0.7428567301586791, Y: 0.3593941678278027, Z: -0.5648005936517033},  // face 15
	{X: -0.8112534709140971, Y: -0.3448953237639382, Z: -0.4721387736413930}, // face 16
	{X: -0.2199307791404607, Y: -0.6583691780274996, Z: -0.7198475378926182}, // face 17
	{X: 0.2139234834501420, Y: -0.1478171829550704, Z: -0.9656017935214205},  // face 18
	{X: -0.1092625278784796, Y: 0.4811951572873210, Z: -0.8697775121287253},  // face 19
}

// faceAxesAzRadsCII contains azimuth angles in radians for face axes (i,j,k)
// at Class II resolution 0. Shape: [face][axis].
// Transcribed from C: src/h3lib/lib/faceijk.c
var faceAxesAzRadsCII = [20][3]float64{
	{5.619958268523939882, 3.525563166130744542, 1.431168063737548730}, // face  0
	{5.760339081714187279, 3.665943979320991689, 1.571548876927796127}, // face  1
	{0.780213654393430055, 4.969003859179821079, 2.874608756786625655}, // face  2
	{0.430469363979999913, 4.619259568766391033, 2.524864466373195467}, // face  3
	{6.130269123335111400, 4.035874020941915804, 1.941478918548720291}, // face  4
	{2.692877706530642877, 0.598482604137447119, 4.787272808923838195}, // face  5
	{2.982963003477243874, 0.888567901084048369, 5.077358105870439581}, // face  6
	{3.532912002790141181, 1.438516900396945656, 5.627307105183336758}, // face  7
	{3.494305004259568154, 1.399909901866372864, 5.588700106652763840}, // face  8
	{3.003214169499538391, 0.908819067106342928, 5.097609271892733906}, // face  9
	{5.930472956509811562, 3.836077854116615875, 1.741682751723420374}, // face 10
	{0.138378484090254847, 4.327168688876645809, 2.232773586483450311}, // face 11
	{0.448714947059150361, 4.637505151845541521, 2.543110049452346120}, // face 12
	{0.158629650112549365, 4.347419854898940135, 2.253024752505744869}, // face 13
	{5.891865957979238535, 3.797470855586042958, 1.703075753192847583}, // face 14
	{2.711123289609793325, 0.616728187216597771, 4.805518392002988683}, // face 15
	{3.294508837434268316, 1.200113735041072948, 5.388903939827463911}, // face 16
	{3.804819692245439833, 1.710424589852244509, 5.899214794638635174}, // face 17
	{3.664438879055192436, 1.570043776661997111, 5.758833981448388027}, // face 18
	{2.361378999196363184, 0.266983896803167583, 4.455774101589558636}, // face 19
}

// GeoToFaceIJK converts lat/lng (radians) to face-IJK coordinates at given resolution.
// From C: _geoToFaceIjk
func GeoToFaceIJK(latRad, lngRad float64, res int) FaceIJK {
	// Convert to 3D unit vector
	v3d := coordijk.GeoToVec3d(latRad, lngRad)

	// Find the nearest face by squared distance
	face := 0
	sqd := coordijk.PointSquareDist3d(faceCenterPoint[0], v3d)
	for i := 1; i < constants.NumIcosaFaces; i++ {
		tmp := coordijk.PointSquareDist3d(faceCenterPoint[i], v3d)
		if tmp < sqd {
			sqd = tmp
			face = i
		}
	}

	// Gnomonic projection from the face center onto tangent plane
	v2d := geoToHex2d(latRad, lngRad, face, res)

	// Convert 2D hex coordinates to IJK
	coord := hex2dToCoordIJK(v2d)

	return FaceIJK{Face: face, Coord: coord}
}

// FaceIJKToGeo converts face-IJK coordinates to lat/lng (radians) at given resolution.
// From C: _faceIjkToGeo
func FaceIJKToGeo(fijk FaceIJK, res int) (latRad, lngRad float64) {
	// Convert IJK to 2D hex coordinates
	v2d := coordijk.ToVec2d(fijk.Coord)

	// Reverse gnomonic projection
	return hex2dToGeo(v2d, fijk.Face, res)
}

// geoToHex2d performs gnomonic projection from lat/lng to 2D hex coordinates.
// From C: _geoToHex2d
// The returned Vec2d is in the hex2d coordinate system where one unit is one
// hex-grid step at the given resolution.
func geoToHex2d(latRad, lngRad float64, face, res int) coordijk.Vec2d {
	v3d := coordijk.GeoToVec3d(latRad, lngRad)
	sqd := coordijk.PointSquareDist3d(faceCenterPoint[face], v3d)

	// Angular distance from face center to point
	r := math.Acos(1.0 - sqd*0.5)
	if r < constants.EPSILON {
		return coordijk.Vec2d{}
	}

	// Azimuth from face center to the point (CCW from geographic north)
	center := faceCenterGeo[face]
	az := GeoAzimuthRads(center.Lat, center.Lng, latRad, lngRad)

	// theta = CCW angle from the face i-axis to the point
	theta := PosAngleRads(faceAxesAzRadsCII[face][0] - PosAngleRads(az))

	// Class III resolutions require an additional rotation
	if res%2 == 1 {
		theta = PosAngleRads(theta - constants.M_AP7_ROT_RADS)
	}

	// Gnomonic projection: angular distance → planar distance
	r = math.Tan(r)
	r /= constants.RES0_U_GNOMONIC

	// Scale by sqrt(7) for each resolution level (aperture-7 grid)
	for i := 0; i < res; i++ {
		r *= M_SQRT7
	}

	return coordijk.Vec2d{
		X: r * math.Cos(theta),
		Y: r * math.Sin(theta),
	}
}

// hex2dToGeo performs reverse gnomonic projection from 2D hex coordinates to lat/lng.
// From C: _hex2dToGeo
func hex2dToGeo(v coordijk.Vec2d, face, res int) (latRad, lngRad float64) {
	center := faceCenterGeo[face]

	r := math.Sqrt(v.X*v.X + v.Y*v.Y)
	if r < constants.EPSILON {
		return center.Lat, center.Lng
	}

	// Angle from hex2d x-axis (i-axis direction)
	theta := math.Atan2(v.Y, v.X)

	// Reverse the resolution scaling: divide by sqrt(7) for each level
	for i := 0; i < res; i++ {
		r *= M_RSQRT7
	}

	// Reverse the RES0_U_GNOMONIC scaling
	r *= constants.RES0_U_GNOMONIC

	// Inverse gnomonic: planar distance → angular distance
	r = math.Atan(r)

	// Reverse Class III rotation
	if res%2 == 1 {
		theta = PosAngleRads(theta + constants.M_AP7_ROT_RADS)
	}

	// Convert hex2d angle back to geographic azimuth
	theta = PosAngleRads(faceAxesAzRadsCII[face][0] - theta)

	return geoAzDistanceRads(center.Lat, center.Lng, theta, r)
}

// geoAzDistanceRads computes the destination point given a start point,
// azimuth (radians from north, clockwise), and angular distance (radians).
// From C: _geoAzDistanceRads
func geoAzDistanceRads(lat1, lng1, az, dist float64) (lat2, lng2 float64) {
	if dist < constants.EPSILON {
		return lat1, lng1
	}

	sinDist := math.Sin(dist)
	cosDist := math.Cos(dist)
	sinLat1 := math.Sin(lat1)
	cosLat1 := math.Cos(lat1)

	sinLat2 := sinLat1*cosDist + cosLat1*sinDist*math.Cos(az)
	lat2 = math.Asin(sinLat2)

	if math.Abs(math.Cos(lat2)) < constants.EPSILON {
		lng2 = 0.0
		return
	}

	lng2 = ConstrainLng(lng1 + math.Atan2(
		math.Sin(az)*sinDist*cosLat1,
		cosDist-sinLat1*sinLat2,
	))
	return
}

// hex2dToCoordIJK converts 2D hex coordinates to IJK.
// From C: _hex2dToCoordIJK
// Uses the full quantization algorithm to find the nearest hexagonal grid point.
func hex2dToCoordIJK(v coordijk.Vec2d) coordijk.CoordIJK {
	var h coordijk.CoordIJK
	h.K = 0

	// Work in the positive quadrant using absolute values
	a1 := math.Abs(v.X)
	a2 := math.Abs(v.Y)

	// Convert to natural hex grid coordinates
	// From _ijkToHex2d: x = i - 0.5*j, y = j * M_SQRT3_2
	// So inverse: j = y / M_SQRT3_2, i = x + j/2
	x2 := a2 / constants.M_SQRT3_2
	x1 := a1 + x2*0.5

	// Integer parts (floor for non-negative values)
	m1 := int(x1)
	m2 := int(x2)

	// Fractional parts
	r1 := x1 - float64(m1)
	r2 := x2 - float64(m2)

	// Determine which hex cell we're in using the fractional parts
	// This is a 6-region decision tree
	if r1 < 0.5 {
		if r1 < 1.0/3.0 {
			if r2 < (1.0+r1)*0.5 {
				h.I = m1
				h.J = m2
			} else {
				h.I = m1
				h.J = m2 + 1
			}
		} else {
			if r2 < (1.0 - r1) {
				h.J = m2
			} else {
				h.J = m2 + 1
			}
			if (1.0-r1) <= r2 && r2 < (2.0*r1) {
				h.I = m1 + 1
			} else {
				h.I = m1
			}
		}
	} else {
		if r1 < 2.0/3.0 {
			if r2 < (1.0 - r1) {
				h.J = m2
			} else {
				h.J = m2 + 1
			}
			if (2.0*r1-1.0) < r2 && r2 < (1.0-r1) {
				h.I = m1
			} else {
				h.I = m1 + 1
			}
		} else {
			if r2 < (r1 * 0.5) {
				h.I = m1 + 1
				h.J = m2
			} else {
				h.I = m1 + 1
				h.J = m2 + 1
			}
		}
	}

	// Correct for negative x (reflection across y-axis).
	// The mirror of the hex at (i, j) across x=0 has coordinates (j-i, j).
	if v.X < 0.0 {
		h.I = h.J - h.I
		// h.J is unchanged
	}

	// Correct for negative y (reflection across x-axis)
	if v.Y < 0.0 {
		h.I = h.I - (2*h.J+1)/2
		h.J = -h.J
	}

	coordijk.IJKNormalize(&h)
	return h
}

// H3ToFaceIJK converts an H3Index to face-IJK coordinates.
// From C: _h3ToFaceIjkWithInitializedFijk (called from _h3ToFaceIjk)
// Starts from the base cell's home face-IJK and walks down the digit sequence.
func H3ToFaceIJK(h h3index.H3Index) FaceIJK {
	bc := h.BaseCell()
	res := h.Resolution()

	d := baseCellHomeFIJK[bc]
	fijk := FaceIJK{
		Face:  d.Face,
		Coord: coordijk.CoordIJK{I: d.I, J: d.J, K: d.K},
	}

	if res == 0 {
		return fijk
	}

	// Walk down the digit sequence, applying aperture-7 scaling and direction offsets.
	// From C _h3ToFaceIjkWithInitializedFijk:
	//   if (isResolutionClassIII(r)) _downAp7(ijk); else _downAp7r(ijk);
	// Note: Class III (odd r) → DownAp7, Class II (even r) → DownAp7r.
	for r := 1; r <= res; r++ {
		if r%2 == 1 { // Class III resolution (odd r) → DownAp7
			fijk.Coord = coordijk.DownAp7(fijk.Coord)
		} else { // Class II resolution (even r) → DownAp7r
			fijk.Coord = coordijk.DownAp7r(fijk.Coord)
		}
		// IndexDigit(r-1) in Go = H3_GET_INDEX_DIGIT(h, r) in C (1-indexed).
		// Guard against INVALID_DIGIT (7): H3_INIT cells have uninitialized digit bits=7.
		digit := h.IndexDigit(r - 1)
		if digit < constants.NUM_DIGITS {
			fijk.Coord = coordijk.IJKAdd(fijk.Coord, coordijk.UNIT_VECS[digit])
		}
	}

	return fijk
}

// FaceIJKToH3 converts face-IJK coordinates to an H3Index at the given resolution.
// From C: _faceIjkToH3
// Walks backward through resolutions to extract digits, then looks up the base cell.
func FaceIJKToH3(fijk FaceIJK, res int) h3index.H3Index {
	h := h3index.H3_INIT
	h = h3index.SetMode(h, constants.H3_CELL_MODE)
	h = h3index.SetResolution(h, res)

	if res == 0 {
		bc := faceIJKToBaseCell(fijk)
		h = h3index.SetBaseCell(h, bc)
		return h
	}

	fijkBC := fijk // working copy

	// Walk backward through resolutions, extracting the digit at each level.
	// From C _faceIjkToH3:
	//   if (isResolutionClassIII(r+1)) { _upAp7; _downAp7 } else { _upAp7r; _downAp7r }
	// i.e. Class III (r+1 odd) → UpAp7/DownAp7, Class II (r+1 even) → UpAp7r/DownAp7r.
	for r := res - 1; r >= 0; r-- {
		lastIJK := fijkBC.Coord
		var lastCenter coordijk.CoordIJK

		if (r+1)%2 == 1 { // Class III resolution (r+1 is odd) → UpAp7
			fijkBC.Coord = coordijk.UpAp7(fijkBC.Coord)
			lastCenter = coordijk.DownAp7(fijkBC.Coord)
		} else { // Class II resolution (r+1 is even) → UpAp7r
			fijkBC.Coord = coordijk.UpAp7r(fijkBC.Coord)
			lastCenter = coordijk.DownAp7r(fijkBC.Coord)
		}

		// The digit is the direction from lastCenter to lastIJK
		diff := coordijk.IJKSub(lastIJK, lastCenter)
		h = h3index.SetIndexDigit(h, r, unitIJKToDigit(diff))
	}

	bc := faceIJKToBaseCell(fijkBC)
	h = h3index.SetBaseCell(h, bc)

	// Apply CCW rotations to align the digit sequence with the base cell's
	// native orientation. From C: for (i = 0; i < numRots; i++) h = _h3Rotate60ccw(h).
	numRots := faceIJKToCCWRot60(fijkBC)
	for i := 0; i < numRots; i++ {
		h = h3Rotate60CCW(h, res)
	}

	return h
}

// unitIJKToDigit returns the digit (direction) corresponding to a unit IJK vector.
// From C: _unitIjkToDigit
func unitIJKToDigit(c coordijk.CoordIJK) int {
	coordijk.IJKNormalize(&c)
	for i, uv := range coordijk.UNIT_VECS {
		if c == uv {
			return i
		}
	}
	return constants.INVALID_DIGIT
}

// faceIJKToBaseCell looks up the base cell number from face and IJK coordinates
// at resolution 0 using the faceIjkBaseCells lookup table.
// From C: _faceIjkToBaseCell.
func faceIJKToBaseCell(fijk FaceIJK) int {
	i, j, k := fijk.Coord.I, fijk.Coord.J, fijk.Coord.K
	if fijk.Face < 0 || fijk.Face >= constants.NumIcosaFaces ||
		i < 0 || i > 2 || j < 0 || j > 2 || k < 0 || k > 2 {
		return h3index.INVALID_BASE_CELL
	}
	return faceIjkBaseCells[fijk.Face][i][j][k].baseCell
}

// faceIJKToCCWRot60 returns the CCW rotation count for aligning the face's
// IJK system with the base cell's native orientation.
// From C: _faceIjkToBaseCellCCWrot60.
func faceIJKToCCWRot60(fijk FaceIJK) int {
	i, j, k := fijk.Coord.I, fijk.Coord.J, fijk.Coord.K
	if fijk.Face < 0 || fijk.Face >= constants.NumIcosaFaces ||
		i < 0 || i > 2 || j < 0 || j > 2 || k < 0 || k > 2 {
		return 0
	}
	return faceIjkBaseCells[fijk.Face][i][j][k].ccwRot60
}

// rotate60CCWDigit rotates a single H3 digit 60 degrees counter-clockwise.
// From C: _rotate60ccw.
// Rotation cycle: K(1)→IK(5)→I(4)→IJ(6)→J(2)→JK(3)→K(1)
func rotate60CCWDigit(d int) int {
	// [CENTER, K, J, JK, I, IK, IJ, INVALID]
	rot := [8]int{0, 5, 3, 1, 6, 4, 2, 7}
	if d < 0 || d > 7 {
		return d
	}
	return rot[d]
}

// h3Rotate60CCW rotates all digits of h by 60 degrees counter-clockwise.
// From C: _h3Rotate60ccw.
func h3Rotate60CCW(h h3index.H3Index, res int) h3index.H3Index {
	for r := 0; r < res; r++ {
		oldDigit := h.IndexDigit(r)
		h = h3index.SetIndexDigit(h, r, rotate60CCWDigit(oldDigit))
	}
	return h
}
