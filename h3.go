// Package h3 provides a native Go implementation of the H3 geospatial indexing system.
// Public API matches github.com/uber/h3-go/v4 exactly.
//
// All lat/lng inputs and outputs are in DEGREES (the public API uses degrees).
// Internally, the faceijk package uses RADIANS.
package h3

import (
	"fmt"
	"math"

	"github.com/lightboxre/h3-go/internal/algos"
	"github.com/lightboxre/h3-go/internal/constants"
	"github.com/lightboxre/h3-go/internal/faceijk"
	"github.com/lightboxre/h3-go/internal/h3index"

	h3math "github.com/lightboxre/h3-go/internal/math"
	polygonpkg "github.com/lightboxre/h3-go/internal/polygon"
)

// Cell is an H3 index representing a hexagonal or pentagonal cell.
type Cell uint64

// DirectedEdge is an H3 index representing a directed edge between two cells.
type DirectedEdge uint64

// Vertex is an H3 index representing a vertex of an H3 cell.
type Vertex uint64

// LatLng is a geographic coordinate in degrees.
// IMPORTANT: Public API uses DEGREES. Internally, faceijk uses RADIANS.
type LatLng struct {
	Lat float64
	Lng float64
}

// CellBoundary is a slice of LatLng defining a cell's boundary polygon.
// Maximum 10 points (6 for hexagons, potential boundary crossings).
type CellBoundary []LatLng

// GeoLoop is a slice of LatLng defining a polygon loop (no holes).
type GeoLoop []LatLng

// GeoPolygon is a polygon with an outer boundary and optional holes.
type GeoPolygon struct {
	GeoLoop GeoLoop
	Holes   []GeoLoop
}

// GeoMultiPolygon is a collection of polygons.
type GeoMultiPolygon []GeoPolygon

// CoordIJ represents a discrete IJ coordinate pair.
type CoordIJ struct {
	I int
	J int
}

// ============================================================================
// Core Indexing Functions
// ============================================================================

// LatLngToCell returns the H3 cell containing the given lat/lng at resolution res.
// lat and lng are in DEGREES.
func LatLngToCell(lat, lng float64, res int) Cell {
	if res < 0 || res > constants.MaxH3Res {
		return Cell(h3index.H3_NULL)
	}
	latRad := faceijk.DegsToRads(lat)
	lngRad := faceijk.DegsToRads(lng)
	fijk := faceijk.GeoToFaceIJK(latRad, lngRad, res)
	h := faceIJKToH3(fijk, res)
	return Cell(h)
}

// CellToLatLng returns the center lat/lng of a cell in DEGREES.
func CellToLatLng(c Cell) LatLng {
	h := h3index.H3Index(c)
	if !h3index.IsValid(h) {
		return LatLng{}
	}
	fijk := faceijk.H3ToFaceIJK(h)
	res := h.Resolution()
	latRad, lngRad := faceijk.FaceIJKToGeo(fijk, res)
	return LatLng{
		Lat: faceijk.RadsToDegs(latRad),
		Lng: faceijk.RadsToDegs(lngRad),
	}
}

// CellToBoundary returns the boundary polygon of a cell as lat/lng in DEGREES.
func CellToBoundary(c Cell) CellBoundary {
	h := h3index.H3Index(c)
	if !h3index.IsValid(h) {
		return CellBoundary{}
	}
	fijk := faceijk.H3ToFaceIJK(h)
	res := h.Resolution()
	isPent := h3index.IsPentagon(h)
	geoPoints := faceijk.FaceIJKToGeoBoundary(fijk, res, isPent)

	boundary := make(CellBoundary, len(geoPoints))
	for i, p := range geoPoints {
		boundary[i] = LatLng{
			Lat: faceijk.RadsToDegs(p.Lat),
			Lng: faceijk.RadsToDegs(p.Lng),
		}
	}
	return boundary
}

// ============================================================================
// Hierarchy Functions
// ============================================================================

// CellToParent returns the parent cell of c at resolution parentRes.
func CellToParent(c Cell, parentRes int) Cell {
	h := h3index.H3Index(c)
	res := h.Resolution()
	if parentRes < 0 || parentRes > res {
		return Cell(h3index.H3_NULL)
	}
	if !h3index.IsValid(h) {
		return Cell(h3index.H3_NULL)
	}

	// Set resolution to parentRes
	h = h3index.SetResolution(h, parentRes)

	// Set all digits from parentRes to MaxH3Res-1 to INVALID_DIGIT (7)
	// Note: digit indices are 0-14 (for res 1-15), stored at positions 14..0
	for r := parentRes; r < constants.MaxH3Res; r++ {
		h = h3index.SetIndexDigit(h, r, constants.INVALID_DIGIT)
	}

	return Cell(h)
}

// CellToChildren returns all children of c at resolution childRes.
func CellToChildren(c Cell, childRes int) []Cell {
	h := h3index.H3Index(c)
	res := h.Resolution()
	if childRes < res || childRes > constants.MaxH3Res {
		return nil
	}
	if !h3index.IsValid(h) {
		return nil
	}

	size := CellToChildrenSize(c, childRes)
	children := make([]Cell, 0, size)
	enumerateChildren(h, res, childRes, &children)
	return children
}

// enumerateChildren recursively generates all children at the target resolution.
func enumerateChildren(h h3index.H3Index, currentRes, targetRes int, out *[]Cell) {
	if currentRes == targetRes {
		*out = append(*out, Cell(h))
		return
	}

	// Get the number of valid directions based on whether current cell is pentagon
	isPent := h3index.IsPentagon(h)

	// Enumerate all 7 directions (0-6), skip K_AXES_DIGIT (1) for pentagons
	for d := range constants.NUM_DIGITS {
		if isPent && d == constants.K_AXES_DIGIT {
			continue // Skip K-axis direction for pentagons
		}

		// Build child at next resolution level
		child := h3index.SetResolution(h, currentRes+1)
		child = h3index.SetIndexDigit(child, currentRes, d)

		// Set remaining digits to INVALID_DIGIT
		for r := currentRes + 1; r < constants.MaxH3Res; r++ {
			child = h3index.SetIndexDigit(child, r, constants.INVALID_DIGIT)
		}

		enumerateChildren(child, currentRes+1, targetRes, out)
	}
}

// CellToChildrenSize returns the number of children of c at resolution childRes.
func CellToChildrenSize(c Cell, childRes int) int64 {
	h := h3index.H3Index(c)
	res := h.Resolution()
	diff := childRes - res
	if diff < 0 {
		return 0
	}
	if diff == 0 {
		return 1
	}

	// Pentagon children: 1 + 5*(7^diff - 1)/6
	if h3index.IsPentagon(h) {
		pow7 := int64(1)
		for range diff {
			pow7 *= 7
		}
		return 1 + 5*(pow7-1)/6
	}

	// Hexagon children: 7^diff
	result := int64(1)
	for range diff {
		result *= 7
	}
	return result
}

// CellToCenterChild returns the center child of c at resolution childRes.
func CellToCenterChild(c Cell, childRes int) Cell {
	h := h3index.H3Index(c)
	res := h.Resolution()
	if childRes < res || childRes > constants.MaxH3Res {
		return Cell(h3index.H3_NULL)
	}
	if !h3index.IsValid(h) {
		return Cell(h3index.H3_NULL)
	}

	// Set resolution to childRes
	h = h3index.SetResolution(h, childRes)

	// Set all new digits (from res to childRes-1) to CENTER_DIGIT (0)
	for r := res; r < childRes; r++ {
		h = h3index.SetIndexDigit(h, r, constants.CENTER_DIGIT)
	}

	// Set remaining digits to INVALID_DIGIT
	for r := childRes; r < constants.MaxH3Res; r++ {
		h = h3index.SetIndexDigit(h, r, constants.INVALID_DIGIT)
	}

	return Cell(h)
}

// CellToChildPos returns the position of c among its parent's children.
// Result is 0-indexed.
func CellToChildPos(c Cell, parentRes int) int64 {
	h := h3index.H3Index(c)
	cellRes := h.Resolution()
	if parentRes > cellRes || parentRes < 0 {
		return -1
	}
	if !h3index.IsValid(h) {
		return -1
	}
	if parentRes == cellRes {
		return 0
	}

	parent := CellToParent(c, parentRes)

	// Fast path: hexagon parent — pure base-7 digit encode (no allocations).
	if !IsPentagon(parent) {
		pos := int64(0)
		for r := parentRes; r < cellRes; r++ {
			pos = pos*7 + int64(h.IndexDigit(r))
		}
		return pos
	}

	// Pentagon parent: fall back to linear search (rare).
	children := CellToChildren(parent, cellRes)
	for i, child := range children {
		if child == c {
			return int64(i)
		}
	}
	return -1
}

// ChildPosToCell returns the cell at position childPos among parent's children
// at resolution childRes.
func ChildPosToCell(childPos int64, parent Cell, childRes int) Cell {
	if childPos < 0 {
		return Cell(h3index.H3_NULL)
	}
	parentRes := GetResolution(parent)
	if childRes < parentRes || childRes > constants.MaxH3Res {
		return Cell(h3index.H3_NULL)
	}
	if !IsValidCell(parent) {
		return Cell(h3index.H3_NULL)
	}
	if childRes == parentRes {
		if childPos == 0 {
			return parent
		}
		return Cell(h3index.H3_NULL)
	}

	// Fast path: hexagon parent — pure base-7 mixed-radix decode (stack only).
	if !IsPentagon(parent) {
		depth := childRes - parentRes
		var digits [15]int
		pos := childPos
		for i := depth - 1; i >= 0; i-- {
			digits[i] = int(pos % 7)
			pos /= 7
		}
		if pos != 0 {
			return Cell(h3index.H3_NULL) // childPos out of range
		}
		h := h3index.H3Index(parent)
		h = h3index.SetResolution(h, childRes)
		for i := range depth {
			h = h3index.SetIndexDigit(h, parentRes+i, digits[i])
		}
		return Cell(h)
	}

	// Pentagon parent: fall back to full enumeration (rare).
	children := CellToChildren(parent, childRes)
	if childPos >= int64(len(children)) {
		return Cell(h3index.H3_NULL)
	}
	return children[childPos]
}

// ============================================================================
// Cell Information Functions
// ============================================================================

// GetResolution returns the resolution (0-15) of the given cell.
func GetResolution(c Cell) int {
	return h3index.H3Index(c).Resolution()
}

// GetBaseCellNumber returns the base cell number (0-121) of the cell.
func GetBaseCellNumber(c Cell) int {
	return h3index.H3Index(c).BaseCell()
}

// IsValidCell returns true if c is a valid H3 cell index.
func IsValidCell(c Cell) bool {
	return h3index.IsValid(h3index.H3Index(c))
}

// IsPentagon returns true if c is a pentagonal cell.
func IsPentagon(c Cell) bool {
	return h3index.IsPentagon(h3index.H3Index(c))
}

// IsResClassIII returns true if c is at a Class III (odd) resolution.
func IsResClassIII(c Cell) bool {
	return h3index.IsResClassIII(h3index.H3Index(c))
}

// ============================================================================
// String Conversion Functions
// ============================================================================

// CellToString returns the hexadecimal string representation of a cell.
func CellToString(c Cell) string {
	return h3index.ToString(h3index.H3Index(c))
}

// StringToCell parses a hexadecimal string and returns the cell.
func StringToCell(s string) (Cell, error) {
	h, err := h3index.FromString(s)
	if err != nil {
		return Cell(h3index.H3_NULL), err
	}
	return Cell(h), nil
}

// H3ToString is an alias for CellToString (for compatibility).
func H3ToString(h uint64) string {
	return h3index.ToString(h3index.H3Index(h))
}

// StringToH3 parses a hexadecimal string as a raw uint64.
func StringToH3(s string) (uint64, error) {
	h, err := h3index.FromString(s)
	return uint64(h), err
}

// ============================================================================
// Enumeration Functions
// ============================================================================

// GetNumCells returns the total number of cells at resolution res.
func GetNumCells(res int) int64 {
	return h3index.GetNumCells(res)
}

// GetRes0Cells returns all 122 resolution-0 base cells.
func GetRes0Cells() []Cell {
	raw := h3index.GetRes0Cells()
	cells := make([]Cell, len(raw))
	for i, h := range raw {
		cells[i] = Cell(h)
	}
	return cells
}

// GetPentagonCells returns all 12 pentagon cells at the given resolution.
func GetPentagonCells(res int) []Cell {
	raw := h3index.GetPentagonCells(res)
	cells := make([]Cell, len(raw))
	for i, h := range raw {
		cells[i] = Cell(h)
	}
	return cells
}

// ============================================================================
// Grid Traversal Functions
// ============================================================================

// MaxGridDiskSize returns the maximum number of cells in a grid disk of radius k.
func MaxGridDiskSize(k int) (int64, error) {
	return algos.MaxGridDiskSize(k)
}

// GridDisk produces all H3 cells within k steps of origin (including origin).
// Falls back to a safe (slower) BFS if a pentagon is encountered.
func GridDisk(origin Cell, k int) ([]Cell, error) {
	cells, err := algos.GridDisk(h3index.H3Index(origin), k)
	if err != nil {
		return nil, err
	}
	return h3SliceToCells(cells), nil
}

// GridDiskDistances produces all cells within k steps grouped by distance.
// out[d] contains cells at exactly distance d from origin.
func GridDiskDistances(origin Cell, k int) ([][]Cell, error) {
	rings, err := algos.GridDiskDistances(h3index.H3Index(origin), k)
	if err != nil {
		return nil, err
	}
	out := make([][]Cell, len(rings))
	for i, ring := range rings {
		out[i] = h3SliceToCells(ring)
	}
	return out, nil
}

// GridDiskUnsafe is a fast version of GridDisk that fails on pentagons.
func GridDiskUnsafe(origin Cell, k int) ([]Cell, error) {
	cells, err := algos.GridDiskUnsafe(h3index.H3Index(origin), k)
	if err != nil {
		return nil, err
	}
	return h3SliceToCells(cells), nil
}

// GridDiskDistancesUnsafe is a fast version of GridDiskDistances that fails on pentagons.
func GridDiskDistancesUnsafe(origin Cell, k int) ([][]Cell, error) {
	rings, err := algos.GridDiskDistancesUnsafe(h3index.H3Index(origin), k)
	if err != nil {
		return nil, err
	}
	out := make([][]Cell, len(rings))
	for i, ring := range rings {
		out[i] = h3SliceToCells(ring)
	}
	return out, nil
}

// GridRingUnsafe produces the hollow ring of cells at exactly distance k.
// Returns an error if a pentagon is encountered.
func GridRingUnsafe(origin Cell, k int) ([]Cell, error) {
	cells, err := algos.GridRingUnsafe(h3index.H3Index(origin), k)
	if err != nil {
		return nil, err
	}
	return h3SliceToCells(cells), nil
}

// GridPathCells returns the cells forming the shortest path from src to dst.
func GridPathCells(src, dst Cell) ([]Cell, error) {
	cells, err := algos.GridPathCells(h3index.H3Index(src), h3index.H3Index(dst))
	if err != nil {
		return nil, err
	}
	return h3SliceToCells(cells), nil
}

// GridPathCellsSize returns the number of cells in the path from src to dst.
func GridPathCellsSize(src, dst Cell) (int64, error) {
	return algos.GridPathCellsSize(h3index.H3Index(src), h3index.H3Index(dst))
}

// GridDistance returns the grid distance (number of hops) between src and dst.
func GridDistance(src, dst Cell) (int64, error) {
	return algos.GridDistance(h3index.H3Index(src), h3index.H3Index(dst))
}

// CompactCells compresses a set of cells to a minimal mixed-resolution set.
func CompactCells(cells []Cell) ([]Cell, error) {
	in := cellsToH3Slice(cells)
	out, err := algos.CompactCells(in)
	if err != nil {
		return nil, err
	}
	return h3SliceToCells(out), nil
}

// UncompactCells expands a compacted set of cells to the given resolution.
func UncompactCells(cells []Cell, res int) ([]Cell, error) {
	in := cellsToH3Slice(cells)
	out, err := algos.UncompactCells(in, res)
	if err != nil {
		return nil, err
	}
	return h3SliceToCells(out), nil
}

// UncompactCellsSize returns the number of cells when uncompacted to res.
func UncompactCellsSize(cells []Cell, res int) (int64, error) {
	return algos.UncompactCellsSize(cellsToH3Slice(cells), res)
}

// ============================================================================
// Region Functions
// ============================================================================

// PolygonToCells fills a GeoPolygon with H3 cells at the given resolution.
// polygon.GeoLoop and Holes contain lat/lng in DEGREES.
func PolygonToCells(polygon GeoPolygon, res int) ([]Cell, error) {
	outerLats, outerLngs := geoLoopToRadSlices(polygon.GeoLoop)

	holesLats := make([][]float64, len(polygon.Holes))
	holesLngs := make([][]float64, len(polygon.Holes))
	for i, hole := range polygon.Holes {
		holesLats[i], holesLngs[i] = geoLoopToRadSlices(hole)
	}

	cells, err := polygonpkg.PolygonToCells(outerLats, outerLngs, holesLats, holesLngs, res, 0)
	if err != nil {
		return nil, err
	}
	return h3SliceToCells(cells), nil
}

// CellsToLinkedMultiPolygon converts a set of cells to the merged polygon outline.
// Returns vertices in DEGREES.
func CellsToLinkedMultiPolygon(cells []Cell) GeoMultiPolygon {
	result := polygonpkg.CellsToLinkedMultiPolygon(cellsToH3Slice(cells))
	if result == nil {
		return nil
	}
	out := make(GeoMultiPolygon, len(result))
	for i, poly := range result {
		out[i] = GeoPolygon{
			GeoLoop: geoPointsToLoop(poly.Outer.Points),
			Holes:   make([]GeoLoop, len(poly.Holes)),
		}
		for j, hole := range poly.Holes {
			out[i].Holes[j] = geoPointsToLoop(hole.Points)
		}
	}
	return out
}

// ============================================================================
// Metrics Functions
// ============================================================================

// GreatCircleDistanceRads returns the great-circle distance in radians.
// Input coordinates are in DEGREES.
func GreatCircleDistanceRads(a, b LatLng) float64 {
	aLat := a.Lat * degsToRads
	aLng := a.Lng * degsToRads
	bLat := b.Lat * degsToRads
	bLng := b.Lng * degsToRads
	return h3math.GreatCircleDistanceRads(aLat, aLng, bLat, bLng)
}

// GreatCircleDistanceKm returns the great-circle distance in kilometers.
// Input coordinates are in DEGREES.
func GreatCircleDistanceKm(a, b LatLng) float64 {
	return h3math.GreatCircleDistanceKm(
		a.Lat*degsToRads, a.Lng*degsToRads,
		b.Lat*degsToRads, b.Lng*degsToRads,
	)
}

// GreatCircleDistanceM returns the great-circle distance in meters.
// Input coordinates are in DEGREES.
func GreatCircleDistanceM(a, b LatLng) float64 {
	return h3math.GreatCircleDistanceM(
		a.Lat*degsToRads, a.Lng*degsToRads,
		b.Lat*degsToRads, b.Lng*degsToRads,
	)
}

// CellAreaRads2 returns the area of the cell in steradians.
func CellAreaRads2(c Cell) float64 {
	return h3math.CellAreaRads2(h3index.H3Index(c))
}

// CellAreaKm2 returns the area of the cell in square kilometers.
func CellAreaKm2(c Cell) float64 {
	return h3math.CellAreaKm2(h3index.H3Index(c))
}

// CellAreaM2 returns the area of the cell in square meters.
func CellAreaM2(c Cell) float64 {
	return h3math.CellAreaM2(h3index.H3Index(c))
}

// EdgeLengthRads returns the average edge length in radians for resolution res.
func EdgeLengthRads(res int) float64 {
	return h3math.EdgeLengthRads(res)
}

// EdgeLengthKm returns the average edge length in kilometers for resolution res.
func EdgeLengthKm(res int) float64 {
	return h3math.EdgeLengthKm(res)
}

// EdgeLengthM returns the average edge length in meters for resolution res.
func EdgeLengthM(res int) float64 {
	return h3math.EdgeLengthM(res)
}

// ExactEdgeLengthRads returns the exact length of a directed edge in radians.
// Computes the great-circle distance between the two shared boundary vertices.
func ExactEdgeLengthRads(e DirectedEdge) float64 {
	boundary := DirectedEdgeToBoundary(e)
	if len(boundary) < 2 {
		return 0
	}
	return GreatCircleDistanceRads(boundary[0], boundary[len(boundary)-1])
}

// ExactEdgeLengthKm returns the exact length of a directed edge in kilometers.
func ExactEdgeLengthKm(e DirectedEdge) float64 {
	return ExactEdgeLengthRads(e) * constants.EarthRadiusKm
}

// ExactEdgeLengthM returns the exact length of a directed edge in meters.
func ExactEdgeLengthM(e DirectedEdge) float64 {
	return ExactEdgeLengthRads(e) * constants.EarthRadiusM
}

// ============================================================================
// Internal Helper Functions
// ============================================================================

// degsToRads converts degrees to radians.
const degsToRads = math.Pi / 180.0

// h3SliceToCells converts []H3Index to []Cell.
func h3SliceToCells(hs []h3index.H3Index) []Cell {
	cells := make([]Cell, len(hs))
	for i, h := range hs {
		cells[i] = Cell(h)
	}
	return cells
}

// cellsToH3Slice converts []Cell to []H3Index.
func cellsToH3Slice(cells []Cell) []h3index.H3Index {
	hs := make([]h3index.H3Index, len(cells))
	for i, c := range cells {
		hs[i] = h3index.H3Index(c)
	}
	return hs
}

// geoLoopToRadSlices converts a GeoLoop (degrees) to lat/lng slices in radians.
func geoLoopToRadSlices(loop GeoLoop) (lats, lngs []float64) {
	lats = make([]float64, len(loop))
	lngs = make([]float64, len(loop))
	for i, pt := range loop {
		lats[i] = pt.Lat * degsToRads
		lngs[i] = pt.Lng * degsToRads
	}
	return lats, lngs
}

// geoPointsToLoop converts GeoPoints (radians) to a GeoLoop (degrees).
func geoPointsToLoop(pts []faceijk.GeoPoint) GeoLoop {
	loop := make(GeoLoop, len(pts))
	for i, pt := range pts {
		loop[i] = LatLng{
			Lat: pt.Lat / degsToRads,
			Lng: pt.Lng / degsToRads,
		}
	}
	return loop
}

// faceIJKToH3 converts a FaceIJK coordinate to an H3Index at the given resolution.
func faceIJKToH3(fijk faceijk.FaceIJK, res int) h3index.H3Index {
	return faceijk.FaceIJKToH3(fijk, res)
}

// Error returns a formatted error message.
func Error(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
