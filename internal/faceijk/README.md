# Package faceijk

This package implements face-IJK coordinate system conversions for the H3 geospatial indexing system.

## Overview

The face-IJK system maps geographic coordinates (lat/lng) to positions on the 20 faces of an icosahedron, with IJK coordinates representing positions on each face. This is a fundamental component of the H3 indexing system.

## Key Components

### Data Structures

- **FaceIJK**: Represents a position on a specific icosahedron face with IJK coordinates
- **GeoPoint**: Latitude/longitude pair in radians

### Lookup Tables

All lookup tables are transcribed exactly from the C reference implementation:

- **faceCenterGeo[20]**: Lat/lng coordinates (radians) of icosahedron face centers
- **faceCenterPoint[20]**: 3D unit vectors representing face centers
- **faceAxesAzRadsCII[20][3]**: Azimuth angles defining IJK axes on each face for Class II resolutions

### Core Functions

#### Geographic Conversions

- **GeoToFaceIJK**: Converts lat/lng (radians) to face-IJK coordinates at a given resolution
  - Uses 3D distance to find nearest face
  - Applies gnomonic projection to convert to planar coordinates
  - Quantizes to IJK grid

- **FaceIJKToGeo**: Converts face-IJK coordinates back to lat/lng (radians)
  - Reverses the gnomonic projection
  - Uses spherical trigonometry to compute final coordinates

#### Gnomonic Projection

The gnomonic projection is the key geometric transformation:

- **geoToHex2d**: Projects spherical coordinates onto the tangent plane at a face center
- **hex2dToGeo**: Reverses the projection from planar coordinates back to spherical

#### Coordinate Conversions

- **hex2dToCoordIJK**: Converts 2D hex coordinates to IJK coordinates
- Uses formulas: j = y / sqrt(3)/2, i = x + 0.5*j

### Utility Functions (latlng.go)

- **GreatCircleDistanceRads**: Haversine formula for spherical distance
- **DegsToRads / RadsToDegs**: Angular unit conversions
- **ConstrainLat / ConstrainLng**: Normalize coordinates to valid ranges
- **GeoAzimuthRads**: Calculate bearing between two points

## Implementation Notes

### Class II vs Class III Resolutions

H3 uses two grid classes:
- **Class II** (even resolutions): Aligned hexagonal grids
- **Class III** (odd resolutions): Rotated grids, requiring M_AP7_ROT_RADS rotation

The gnomonic projection functions handle both classes by applying rotation when needed.

### Resolution Scaling

Each resolution level scales by a factor of 7 (aperture-7 hexagonal grid). The unit scale factors are:
- Resolution 0: 1.0
- Resolution 2: 7.0
- Resolution 4: 49.0
- etc.

### Pentagon Handling

The 12 pentagon cells in H3 require special handling in boundary calculations and face transitions. The current implementation includes basic pentagon support in FaceIJKToGeoBoundary.

## Testing

The test suite includes:
- Unit tests for all conversion functions
- Round-trip tests to verify consistency
- Validation that face centers match between representations
- Boundary calculation tests for hexagons and pentagons

## Future Enhancements

The current implementation provides:
- ✅ Core geographic conversions (GeoToFaceIJK, FaceIJKToGeo)
- ✅ Gnomonic projection functions
- ✅ Basic boundary calculation
- ⚠️  H3ToFaceIJK (simplified - needs base cell face lookup table)
- ⚠️  FaceIJKToGeoBoundary (simplified vertex calculation)

For production use, these functions need:
- Complete base cell face data integration
- More accurate vertex positioning for cell boundaries
- Face transition handling for cells near face edges
- Pentagon distortion corrections

## References

This implementation is based on the Uber H3 C reference implementation:
- https://github.com/uber/h3/blob/master/src/h3lib/lib/faceijk.c
- https://github.com/uber/h3/blob/master/src/h3lib/lib/latLng.c
