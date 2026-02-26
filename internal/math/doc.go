// Package h3math provides mathematical functions for H3 geospatial operations.
//
// This package includes:
//
//   - Great circle distance calculations (in radians, kilometers, and meters)
//   - Cell area computations using spherical polygon geometry and L'Huilier's theorem
//   - Average and exact edge length functions for H3 cells
//
// All calculations use a spherical Earth model with the WGS84 authalic radius:
//
//   - Earth radius: 6371.007180918475 km
//
// Area calculations use L'Huilier's theorem for spherical triangle areas, which
// provides better numerical stability than Girard's theorem for small triangles.
//
// Edge length statistics are based on the H3 specification, with exact calculations
// for resolutions 0-6 and extrapolated values for resolutions 7-15.
package h3math
