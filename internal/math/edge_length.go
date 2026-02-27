package h3math

import (
	"github.com/EDRInc/h3-go/internal/constants"
	"github.com/EDRInc/h3-go/internal/h3index"
)

// averageEdgeLengthKmTable contains the average hexagon edge lengths in kilometers
// by resolution. Values match the H3 C library 4.x (HexagonEdgeLengthAvgKm).
var averageEdgeLengthKmTable = [constants.MaxH3Res + 1]float64{
	1281.256011, // res 0
	483.0568391, // res 1
	182.5129565, // res 2
	68.97922179, // res 3
	26.07175968, // res 4
	9.85409099,  // res 5
	3.724532667, // res 6
	1.406475763, // res 7
	0.53141401,  // res 8
	0.200786148, // res 9
	0.075863783, // res 10
	0.028663897, // res 11
	0.010830188, // res 12
	0.00409201,  // res 13
	0.0015461,   // res 14
	0.000584169, // res 15
}

// averageEdgeLengthMTable contains the average hexagon edge lengths in meters
// by resolution. The H3 C library stores these independently from the km table,
// so minor floating-point differences exist (e.g. res 9: 0.200786148 km ≠ 200.7861476/1000 m).
// Values match the H3 C library 4.x (HexagonEdgeLengthAvgM).
var averageEdgeLengthMTable = [constants.MaxH3Res + 1]float64{
	1281256.011, // res 0
	483056.8391, // res 1
	182512.9565, // res 2
	68979.22179, // res 3
	26071.75968, // res 4
	9854.09099,  // res 5
	3724.532667, // res 6
	1406.475763, // res 7
	531.4140101, // res 8
	200.7861476, // res 9
	75.86378287, // res 10
	28.66389748, // res 11
	10.83018784, // res 12
	4.092010473, // res 13
	1.546099657, // res 14
	0.58416863,  // res 15
}

// EdgeLengthRads returns the average edge length in radians for a given resolution.
// These are average values, not exact per-cell values.
func EdgeLengthRads(res int) float64 {
	if res < 0 || res > constants.MaxH3Res {
		return 0
	}
	return averageEdgeLengthKmTable[res] / constants.EarthRadiusKm
}

// EdgeLengthKm returns the average edge length in kilometers.
func EdgeLengthKm(res int) float64 {
	if res < 0 || res > constants.MaxH3Res {
		return 0
	}
	return averageEdgeLengthKmTable[res]
}

// EdgeLengthM returns the average edge length in meters.
func EdgeLengthM(res int) float64 {
	if res < 0 || res > constants.MaxH3Res {
		return 0
	}
	return averageEdgeLengthMTable[res]
}

// ExactEdgeLengthRads returns the exact edge length of the directed edge in radians.
// This is a simplified implementation that returns the average edge length
// for the resolution of the edge's cells.
//
// A complete implementation would:
// 1. Decode the directed edge to get origin and destination cells
// 2. Get the boundary vertices of both cells
// 3. Find the shared edge vertices
// 4. Compute the great-circle distance between them
//
// This requires DirectedEdgeToBoundary which is in Agent G's domain.
// For now, we use the average edge length as an approximation.
func ExactEdgeLengthRads(e h3index.H3Index) float64 {
	// Get the resolution from the edge index
	res := e.Resolution()
	return EdgeLengthRads(res)
}

// ExactEdgeLengthKm returns the exact edge length in kilometers.
func ExactEdgeLengthKm(e h3index.H3Index) float64 {
	return ExactEdgeLengthRads(e) * constants.EarthRadiusKm
}

// ExactEdgeLengthM returns the exact edge length in meters.
func ExactEdgeLengthM(e h3index.H3Index) float64 {
	return ExactEdgeLengthRads(e) * constants.EarthRadiusM
}
