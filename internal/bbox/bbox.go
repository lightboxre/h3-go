// Package bbox provides geographic bounding box operations for H3.
// All coordinates are in radians.
package bbox

import (
	"math"
)

// BBox is a geographic bounding box defined by lat/lng in radians.
type BBox struct {
	North float64 // maximum latitude (radians)
	South float64 // minimum latitude (radians)
	East  float64 // maximum longitude (radians)
	West  float64 // minimum longitude (radians)
}

// BBoxIsTransmeridian returns true if the BBox crosses the antimeridian
// (i.e., East < West in the normalized representation).
func BBoxIsTransmeridian(bbox BBox) bool {
	return bbox.East < bbox.West
}

// BBoxWidth returns the "width" of the bounding box in longitude.
// For transmeridian boxes: 2π + (East - West).
func BBoxWidth(bbox BBox) float64 {
	if BBoxIsTransmeridian(bbox) {
		return bbox.East - bbox.West + 2*math.Pi
	}
	return bbox.East - bbox.West
}

// BBoxHeight returns the height in latitude.
func BBoxHeight(bbox BBox) float64 {
	return bbox.North - bbox.South
}

// BBoxCenter returns the center point of a BBox.
// For antimeridian-crossing boxes, the center longitude is the midpoint
// of the wrapped range.
func BBoxCenter(bbox BBox) (lat, lng float64) {
	lat = (bbox.North + bbox.South) / 2.0
	east := bbox.East

	// For transmeridian boxes, shift east by 2π before averaging
	if BBoxIsTransmeridian(bbox) {
		east += 2 * math.Pi
	}

	lng = (east + bbox.West) / 2.0
	return lat, lng
}

// BBoxContains returns true if the BBox contains the given point.
// Must handle antimeridian crossing (East < West).
func BBoxContains(bbox BBox, lat, lng float64) bool {
	// Check latitude bounds
	if lat < bbox.South || lat > bbox.North {
		return false
	}

	// Check longitude with antimeridian handling
	if BBoxIsTransmeridian(bbox) {
		// Transmeridian: point longitude must be >= West OR <= East
		return lng >= bbox.West || lng <= bbox.East
	}

	// Standard case: point longitude must be between West and East
	return lng >= bbox.West && lng <= bbox.East
}

// bboxOverlapsLng returns true if two longitude ranges overlap.
// Handles antimeridian crossing.
func bboxOverlapsLng(a, b BBox) bool {
	aIsTransmeridian := BBoxIsTransmeridian(a)
	bIsTransmeridian := BBoxIsTransmeridian(b)

	// If neither crosses antimeridian, simple range overlap
	if !aIsTransmeridian && !bIsTransmeridian {
		return a.West <= b.East && b.West <= a.East
	}

	// If both cross antimeridian, they must overlap
	if aIsTransmeridian && bIsTransmeridian {
		return true
	}

	// One crosses, one doesn't
	if aIsTransmeridian {
		// a crosses: check if b overlaps either side
		return b.West <= a.East || b.East >= a.West
	}

	// b crosses: check if a overlaps either side
	return a.West <= b.East || a.East >= b.West
}

// BBoxOverlap returns true if two BBoxes overlap.
// Must handle antimeridian crossing.
func BBoxOverlap(a, b BBox) bool {
	// Check latitude overlap first
	if a.South > b.North || b.South > a.North {
		return false
	}

	// Check longitude overlap with antimeridian handling
	return bboxOverlapsLng(a, b)
}

// BBoxFromGeoLoop computes the bounding box of a GeoLoop (slice of lat/lng pairs).
// lats and lngs are in radians.
// Must detect antimeridian crossing: if consecutive longitudes differ by > π,
// it's an antimeridian crossing.
func BBoxFromGeoLoop(lats, lngs []float64) BBox {
	if len(lats) == 0 || len(lngs) == 0 {
		return BBox{}
	}

	// Initialize with first vertex
	bbox := BBox{
		North: lats[0],
		South: lats[0],
		East:  lngs[0],
		West:  lngs[0],
	}

	// Track min/max for detecting transmeridian crossing
	minPosLng := math.MaxFloat64
	maxNegLng := -math.MaxFloat64
	isTransmeridian := false

	prevLng := lngs[len(lngs)-1] // Start with last vertex for wrapping

	for i := range lats {
		lat := lats[i]
		lng := lngs[i]

		// Update latitude bounds
		if lat < bbox.South {
			bbox.South = lat
		}
		if lat > bbox.North {
			bbox.North = lat
		}

		// Update longitude bounds (standard)
		if lng < bbox.West {
			bbox.West = lng
		}
		if lng > bbox.East {
			bbox.East = lng
		}

		// Detect antimeridian crossing between consecutive vertices
		if math.Abs(prevLng-lng) > math.Pi {
			isTransmeridian = true

			// Track min positive and max negative longitudes
			if lng > 0 {
				if lng < minPosLng {
					minPosLng = lng
				}
			} else {
				if lng > maxNegLng {
					maxNegLng = lng
				}
			}
		}

		prevLng = lng
	}

	// If transmeridian, adjust East/West
	if isTransmeridian {
		bbox.East = maxNegLng
		bbox.West = minPosLng
	}

	return bbox
}
