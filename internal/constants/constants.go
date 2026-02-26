// Package constants defines compile-time constants for the H3 library.
// Values are transcribed from the C H3 reference implementation.
package constants

// Resolution limits.
const (
	MaxH3Res      = 15
	NumBaseCells  = 122
	NumIcosaFaces = 20
	NumPentagons  = 12
)

// Floating-point constants.
const (
	// M_AP7_ROT_RADS is the rotation angle between Class II and Class III
	// resolution axes: asin(sqrt(3.0/28.0)).
	// Exact value from H3 C source: constants.h
	M_AP7_ROT_RADS = 0.333473172251832115336090755351601070065900389

	// M_SIN60 is sin(60°) = sqrt(3)/2.
	M_SIN60 = 0.8660254037844386

	// M_SQRT3_2 is sqrt(3)/2 (alias of M_SIN60).
	M_SQRT3_2 = 0.8660254037844386

	// RES0_U_GNOMONIC is the gnomonic scaling factor for resolution 0.
	RES0_U_GNOMONIC = 0.38196601125010500003

	// EPSILON is the floating-point comparison threshold.
	EPSILON = 1e-16

	// DBL_EPSILON is similar to C DBL_EPSILON for internal checks.
	DBL_EPSILON = 2.2204460492503131e-16
)

// Earth radius constants (WGS84 authalic sphere).
const (
	EarthRadiusKm = 6371.007180918475
	EarthRadiusM  = EarthRadiusKm * 1000.0
)

// Pentagon base cell indices (12 total).
var PentagonBaseCells = [NumPentagons]int{
	4, 14, 24, 38, 49, 58, 63, 72, 83, 97, 107, 117,
}

// maxDimByCIIres[r] is the max IJK dimension for a Class II cell at resolution r.
// From the C source: H3_EXPORT(maxGridDiskSize).
var MaxDimByCIIres = [MaxH3Res + 1]int{
	2,        // res 0
	-1,       // res 1 (class III — unused here)
	14,       // res 2
	-1,       // res 3
	98,       // res 4
	-1,       // res 5
	686,      // res 6
	-1,       // res 7
	4802,     // res 8
	-1,       // res 9
	33614,    // res 10
	-1,       // res 11
	235298,   // res 12
	-1,       // res 13
	1647086,  // res 14
	-1,       // res 15
}

// unitScaleByCIIres[r] is the unit scale for Class II grids at resolution r.
var UnitScaleByCIIres = [MaxH3Res + 1]float64{
	1.0,
	-1,
	7.0,
	-1,
	49.0,
	-1,
	343.0,
	-1,
	2401.0,
	-1,
	16807.0,
	-1,
	117649.0,
	-1,
	823543.0,
	-1,
}

// H3 index mode values.
const (
	H3_CELL_MODE         = 1
	H3_DIRECTEDEDGE_MODE = 2
	H3_VERTEX_MODE       = 3
)

// Bit-field positions for the H3Index packed integer.
const (
	H3_MODE_OFFSET       = 59
	H3_RESERVED_OFFSET   = 56
	H3_RES_OFFSET        = 52
	H3_BC_OFFSET         = 45
	H3_PER_DIGIT_OFFSET  = 3

	H3_MODE_MASK       uint64 = 0xF
	H3_RESERVED_MASK   uint64 = 0x7
	H3_RES_MASK        uint64 = 0xF
	H3_BC_MASK         uint64 = 0x7F
	H3_DIGIT_MASK      uint64 = 0x7
)

// Direction digits used in H3 index encoding.
const (
	CENTER_DIGIT    = 0
	K_AXES_DIGIT    = 1
	J_AXES_DIGIT    = 2
	JK_AXES_DIGIT   = 3
	I_AXES_DIGIT    = 4
	IK_AXES_DIGIT   = 5
	IJ_AXES_DIGIT   = 6
	INVALID_DIGIT   = 7
	NUM_DIGITS      = 7
)

// UNUSED_DIGIT is the filler for cells coarser than the resolution.
const UNUSED_DIGIT = 7
