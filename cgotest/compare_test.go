//go:build cgo

// Package cgotest compares the native Go H3 implementation against the
// CGO-based Uber H3 binding (github.com/uber/h3-go/v4) as a correctness oracle.
//
// Run with:
//
//	cd cgotest && CGO_ENABLED=1 go test -v ./...
//
// Requires the H3 C library to be installed. On Ubuntu/Debian:
//
//	sudo apt-get install -y cmake
//	git clone --depth 1 https://github.com/uber/h3.git /tmp/h3
//	cd /tmp/h3 && cmake -B build && cmake --build build && sudo cmake --install build && sudo ldconfig
package cgotest

import (
	"fmt"
	"math"
	"sort"
	"testing"

	h3native "github.com/h3-native/h3-go"
	h3cgo "github.com/uber/h3-go/v4"
)

// ============================================================================
// Test constants (inputs only — no hardcoded expected values)
// ============================================================================

const (
	sfLat   = 37.7796
	sfLng   = -122.4193
	googLat = 37.3615593
	googLng = -122.0553238
	epsilon = 1e-9 // absolute tolerance for lat/lng degrees
	relTol  = 1e-9 // relative tolerance for distances and areas
)

// ============================================================================
// Type conversion helpers
// ============================================================================

// nativeToCGOCell converts a native Cell (uint64) to a CGO Cell (int64).
// Valid H3 cells have bit 63 = 0 (mode < 8), so the conversion is safe.
func nativeToCGOCell(c h3native.Cell) h3cgo.Cell {
	return h3cgo.Cell(int64(uint64(c)))
}

// cgoToNativeCell converts a CGO Cell (int64) to a native Cell (uint64).
func cgoToNativeCell(c h3cgo.Cell) h3native.Cell {
	return h3native.Cell(uint64(int64(c)))
}

// nativeToCGOEdge converts a native DirectedEdge (uint64) to a CGO DirectedEdge (int64).
func nativeToCGOEdge(e h3native.DirectedEdge) h3cgo.DirectedEdge {
	return h3cgo.DirectedEdge(int64(uint64(e)))
}

// ============================================================================
// Shared comparison helpers
// ============================================================================

// compareCellSets sorts both slices by uint64 value and compares element-by-element.
func compareCellSets(t *testing.T, native []h3native.Cell, cgo []h3cgo.Cell) {
	t.Helper()
	if len(native) != len(cgo) {
		t.Errorf("cell set size mismatch: native=%d, cgo=%d", len(native), len(cgo))
		return
	}
	sortedNative := make([]uint64, len(native))
	sortedCGO := make([]uint64, len(cgo))
	for i, c := range native {
		sortedNative[i] = uint64(c)
	}
	for i, c := range cgo {
		sortedCGO[i] = uint64(int64(c))
	}
	sort.Slice(sortedNative, func(i, j int) bool { return sortedNative[i] < sortedNative[j] })
	sort.Slice(sortedCGO, func(i, j int) bool { return sortedCGO[i] < sortedCGO[j] })
	for i := range sortedNative {
		if sortedNative[i] != sortedCGO[i] {
			t.Errorf("cell[%d]: native=%#x, cgo=%#x", i, sortedNative[i], sortedCGO[i])
		}
	}
}

// compareNestedCellSets compares two slices of cell slices ring-by-ring.
func compareNestedCellSets(t *testing.T, native [][]h3native.Cell, cgo [][]h3cgo.Cell) {
	t.Helper()
	if len(native) != len(cgo) {
		t.Errorf("ring count mismatch: native=%d, cgo=%d", len(native), len(cgo))
		return
	}
	for i := range native {
		t.Run(fmt.Sprintf("ring%d", i), func(t *testing.T) {
			compareCellSets(t, native[i], cgo[i])
		})
	}
}

// convertGeoPolygon converts a native GeoPolygon to a CGO GeoPolygon.
func convertGeoPolygon(p h3native.GeoPolygon) h3cgo.GeoPolygon {
	outer := make(h3cgo.GeoLoop, len(p.GeoLoop))
	for i, ll := range p.GeoLoop {
		outer[i] = h3cgo.LatLng{Lat: ll.Lat, Lng: ll.Lng}
	}
	holes := make([]h3cgo.GeoLoop, len(p.Holes))
	for i, hole := range p.Holes {
		h := make(h3cgo.GeoLoop, len(hole))
		for j, ll := range hole {
			h[j] = h3cgo.LatLng{Lat: ll.Lat, Lng: ll.Lng}
		}
		holes[i] = h
	}
	return h3cgo.GeoPolygon{GeoLoop: outer, Holes: holes}
}

// approxEq returns true if |a-b| <= eps.
func approxEq(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}

// relApproxEq returns true if |a-b|/|b| <= tol (or |a| <= tol when b==0).
func relApproxEq(a, b, tol float64) bool {
	if b == 0 {
		return math.Abs(a) <= tol
	}
	return math.Abs(a-b)/math.Abs(b) <= tol
}

// ============================================================================
// Core Indexing Tests
// ============================================================================

func TestLatLngToCell_Oracle(t *testing.T) {
	cases := []struct {
		lat, lng float64
		res      int
	}{
		{sfLat, sfLng, 9},
		{googLat, googLng, 5},
		{0, 0, 0},
		{89.9, 0, 4},
		{35.0, 179.9, 7},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%.4f_%.4f_r%d", tc.lat, tc.lng, tc.res), func(t *testing.T) {
			want, err := h3cgo.LatLngToCell(h3cgo.NewLatLng(tc.lat, tc.lng), tc.res)
			if err != nil {
				t.Fatalf("cgo LatLngToCell error: %v", err)
			}
			got := h3native.LatLngToCell(tc.lat, tc.lng, tc.res)
			if uint64(got) != uint64(int64(want)) {
				t.Errorf("got %#x, want %#x", uint64(got), uint64(int64(want)))
			}
		})
	}
}

func TestCellToLatLng_Oracle(t *testing.T) {
	cells := []h3native.Cell{
		h3native.Cell(0x89283082877ffff),
		h3native.Cell(0x85283473fffffff),
	}
	for _, c := range cells {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want, err := h3cgo.CellToLatLng(nativeToCGOCell(c))
			if err != nil {
				t.Fatalf("cgo CellToLatLng error: %v", err)
			}
			got := h3native.CellToLatLng(c)
			if !approxEq(got.Lat, want.Lat, epsilon) {
				t.Errorf("Lat: got %.12f, want %.12f (diff %e)", got.Lat, want.Lat, math.Abs(got.Lat-want.Lat))
			}
			if !approxEq(got.Lng, want.Lng, epsilon) {
				t.Errorf("Lng: got %.12f, want %.12f (diff %e)", got.Lng, want.Lng, math.Abs(got.Lng-want.Lng))
			}
		})
	}
}

func TestCellToBoundary_Oracle(t *testing.T) {
	cells := []h3native.Cell{
		h3native.Cell(0x89283082877ffff),
		h3native.Cell(0x85283473fffffff),
	}
	for _, c := range cells {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want, err := h3cgo.CellToBoundary(nativeToCGOCell(c))
			if err != nil {
				t.Fatalf("cgo CellToBoundary error: %v", err)
			}
			got := h3native.CellToBoundary(c)
			if len(got) != len(want) {
				t.Fatalf("vertex count: got %d, want %d", len(got), len(want))
			}
			for i := range got {
				if !approxEq(got[i].Lat, want[i].Lat, epsilon) {
					t.Errorf("vertex[%d].Lat: got %.12f, want %.12f", i, got[i].Lat, want[i].Lat)
				}
				if !approxEq(got[i].Lng, want[i].Lng, epsilon) {
					t.Errorf("vertex[%d].Lng: got %.12f, want %.12f", i, got[i].Lng, want[i].Lng)
				}
			}
		})
	}
}

// ============================================================================
// Hierarchy Tests
// ============================================================================

func TestCellToParent_Oracle(t *testing.T) {
	cases := []struct {
		cell      h3native.Cell
		parentRes int
	}{
		{h3native.Cell(0x89283082877ffff), 5},
		{h3native.Cell(0x89283082877ffff), 0},
		{h3native.Cell(0x85283473fffffff), 3},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%#x_r%d", uint64(tc.cell), tc.parentRes), func(t *testing.T) {
			want, err := nativeToCGOCell(tc.cell).Parent(tc.parentRes)
			if err != nil {
				t.Fatalf("cgo Parent error: %v", err)
			}
			got := h3native.CellToParent(tc.cell, tc.parentRes)
			if uint64(got) != uint64(int64(want)) {
				t.Errorf("got %#x, want %#x", uint64(got), uint64(int64(want)))
			}
		})
	}
}

func TestCellToChildren_Oracle(t *testing.T) {
	cases := []struct {
		cell     h3native.Cell
		childRes int
	}{
		{h3native.Cell(0x85283473fffffff), 6},
		{h3native.Cell(0x89283082877ffff), 10},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%#x_r%d", uint64(tc.cell), tc.childRes), func(t *testing.T) {
			wantCGO, err := nativeToCGOCell(tc.cell).Children(tc.childRes)
			if err != nil {
				t.Fatalf("cgo Children error: %v", err)
			}
			got := h3native.CellToChildren(tc.cell, tc.childRes)
			compareCellSets(t, got, wantCGO)
		})
	}
}

func TestCellToChildrenSize_Oracle(t *testing.T) {
	cases := []struct {
		cell     h3native.Cell
		childRes int
	}{
		{h3native.Cell(0x85283473fffffff), 6},
		{h3native.Cell(0x85283473fffffff), 7},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%#x_r%d", uint64(tc.cell), tc.childRes), func(t *testing.T) {
			// Derive from CGO: len(Children)
			wantCGO, err := nativeToCGOCell(tc.cell).Children(tc.childRes)
			if err != nil {
				t.Fatalf("cgo Children error: %v", err)
			}
			got := h3native.CellToChildrenSize(tc.cell, tc.childRes)
			if got != int64(len(wantCGO)) {
				t.Errorf("CellToChildrenSize: got %d, want %d", got, len(wantCGO))
			}
		})
	}
}

func TestCellToCenterChild_Oracle(t *testing.T) {
	cases := []struct {
		cell     h3native.Cell
		childRes int
	}{
		{h3native.Cell(0x85283473fffffff), 7},
		{h3native.Cell(0x89283082877ffff), 11},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("%#x_r%d", uint64(tc.cell), tc.childRes), func(t *testing.T) {
			want, err := nativeToCGOCell(tc.cell).CenterChild(tc.childRes)
			if err != nil {
				t.Fatalf("cgo CenterChild error: %v", err)
			}
			got := h3native.CellToCenterChild(tc.cell, tc.childRes)
			if uint64(got) != uint64(int64(want)) {
				t.Errorf("got %#x, want %#x", uint64(got), uint64(int64(want)))
			}
		})
	}
}

func TestCellToChildPos_Oracle(t *testing.T) {
	child := h3native.Cell(0x89283082877ffff)
	parentRes := 7
	want, err := nativeToCGOCell(child).ChildPos(parentRes)
	if err != nil {
		t.Fatalf("cgo ChildPos error: %v", err)
	}
	got := h3native.CellToChildPos(child, parentRes)
	if got != int64(want) {
		t.Errorf("CellToChildPos: got %d, want %d", got, want)
	}
}

// ============================================================================
// Cell Inspection Tests
// ============================================================================

func TestGetResolution_Oracle(t *testing.T) {
	cells := []h3native.Cell{
		h3native.Cell(0x89283082877ffff),
		h3native.Cell(0x85283473fffffff),
	}
	for _, c := range cells {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want := nativeToCGOCell(c).Resolution()
			got := h3native.GetResolution(c)
			if got != want {
				t.Errorf("GetResolution: got %d, want %d", got, want)
			}
		})
	}
}

func TestIsValidCell_Oracle(t *testing.T) {
	cases := []h3native.Cell{
		h3native.Cell(0x89283082877ffff),
		h3native.Cell(0x85283473fffffff),
		h3native.Cell(0),
		h3native.Cell(0xFFFFFFFFFFFFFFFF),
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want := nativeToCGOCell(c).IsValid()
			got := h3native.IsValidCell(c)
			if got != want {
				t.Errorf("IsValidCell: got %v, want %v", got, want)
			}
		})
	}
}

func TestIsPentagon_Oracle(t *testing.T) {
	// Mix of pentagons and hexagons
	cells := h3native.GetPentagonCells(9)
	cells = append(cells, h3native.Cell(0x89283082877ffff))
	for _, c := range cells {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want := nativeToCGOCell(c).IsPentagon()
			got := h3native.IsPentagon(c)
			if got != want {
				t.Errorf("IsPentagon: got %v, want %v", got, want)
			}
		})
	}
}

func TestIsResClassIII_Oracle(t *testing.T) {
	cases := []h3native.Cell{
		h3native.Cell(0x89283082877ffff), // res 9, Class III
		h3native.Cell(0x85283473fffffff), // res 5, Class II
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want := nativeToCGOCell(c).IsResClassIII()
			got := h3native.IsResClassIII(c)
			if got != want {
				t.Errorf("IsResClassIII: got %v, want %v", got, want)
			}
		})
	}
}

func TestGetBaseCellNumber_Oracle(t *testing.T) {
	cells := []h3native.Cell{
		h3native.Cell(0x89283082877ffff),
		h3native.Cell(0x85283473fffffff),
	}
	for _, c := range cells {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want := nativeToCGOCell(c).BaseCellNumber()
			got := h3native.GetBaseCellNumber(c)
			if got != want {
				t.Errorf("GetBaseCellNumber: got %d, want %d", got, want)
			}
		})
	}
}

// ============================================================================
// String Conversion Tests
// ============================================================================

func TestCellToString_Oracle(t *testing.T) {
	cells := []h3native.Cell{
		h3native.Cell(0x89283082877ffff),
		h3native.Cell(0x85283473fffffff),
	}
	for _, c := range cells {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want := h3cgo.CellToString(nativeToCGOCell(c))
			got := h3native.CellToString(c)
			if got != want {
				t.Errorf("CellToString: got %q, want %q", got, want)
			}
		})
	}
}

func TestStringToCell_Oracle(t *testing.T) {
	strs := []string{
		"89283082877ffff",
		"85283473fffffff",
	}
	for _, s := range strs {
		s := s
		t.Run(s, func(t *testing.T) {
			want := h3cgo.CellFromString(s)
			got, err := h3native.StringToCell(s)
			if err != nil {
				t.Fatalf("native StringToCell error: %v", err)
			}
			if uint64(got) != uint64(int64(want)) {
				t.Errorf("StringToCell(%q): got %#x, want %#x", s, uint64(got), uint64(int64(want)))
			}
		})
	}
}

// ============================================================================
// Enumeration Tests
// ============================================================================

func TestGetNumCells_Oracle(t *testing.T) {
	for _, res := range []int{0, 1, 2, 5} {
		res := res
		t.Run(fmt.Sprintf("r%d", res), func(t *testing.T) {
			want := h3cgo.NumCells(res)
			got := h3native.GetNumCells(res)
			if got != int64(want) {
				t.Errorf("GetNumCells(%d): got %d, want %d", res, got, want)
			}
		})
	}
}

func TestGetRes0Cells_Oracle(t *testing.T) {
	wantCGO, err := h3cgo.Res0Cells()
	if err != nil {
		t.Fatalf("cgo Res0Cells error: %v", err)
	}
	got := h3native.GetRes0Cells()
	compareCellSets(t, got, wantCGO)
}

func TestGetPentagonCells_Oracle(t *testing.T) {
	for _, res := range []int{0, 5, 9} {
		res := res
		t.Run(fmt.Sprintf("r%d", res), func(t *testing.T) {
			wantCGO, err := h3cgo.Pentagons(res)
			if err != nil {
				t.Fatalf("cgo Pentagons error: %v", err)
			}
			got := h3native.GetPentagonCells(res)
			compareCellSets(t, got, wantCGO)
		})
	}
}

// ============================================================================
// Grid Traversal Tests
// ============================================================================

func TestGridDisk_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	for _, k := range []int{0, 1, 2} {
		k := k
		t.Run(fmt.Sprintf("k%d", k), func(t *testing.T) {
			wantCGO, err := h3cgo.GridDisk(nativeToCGOCell(origin), k)
			if err != nil {
				t.Fatalf("cgo GridDisk error: %v", err)
			}
			got, err := h3native.GridDisk(origin, k)
			if err != nil {
				t.Fatalf("native GridDisk error: %v", err)
			}
			compareCellSets(t, got, wantCGO)
		})
	}
}

func TestGridDiskDistances_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	for _, k := range []int{0, 1, 2} {
		k := k
		t.Run(fmt.Sprintf("k%d", k), func(t *testing.T) {
			wantCGO, err := h3cgo.GridDiskDistances(nativeToCGOCell(origin), k)
			if err != nil {
				t.Fatalf("cgo GridDiskDistances error: %v", err)
			}
			got, err := h3native.GridDiskDistances(origin, k)
			if err != nil {
				t.Fatalf("native GridDiskDistances error: %v", err)
			}
			compareNestedCellSets(t, got, wantCGO)
		})
	}
}

func TestGridRingUnsafe_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	for _, k := range []int{0, 1, 2} {
		k := k
		t.Run(fmt.Sprintf("k%d", k), func(t *testing.T) {
			wantCGO, err := h3cgo.GridRingUnsafe(nativeToCGOCell(origin), k)
			if err != nil {
				t.Fatalf("cgo GridRingUnsafe error: %v", err)
			}
			got, err := h3native.GridRingUnsafe(origin, k)
			if err != nil {
				t.Fatalf("native GridRingUnsafe error: %v", err)
			}
			compareCellSets(t, got, wantCGO)
		})
	}
}

func TestGridDistance_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	cases := []struct {
		src, dst h3native.Cell
	}{
		{origin, origin},
	}
	if len(disk) > 1 {
		cases = append(cases, struct{ src, dst h3native.Cell }{origin, disk[1]})
	}
	for i, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			want, err := h3cgo.GridDistance(nativeToCGOCell(tc.src), nativeToCGOCell(tc.dst))
			if err != nil {
				t.Fatalf("cgo GridDistance error: %v", err)
			}
			got, err := h3native.GridDistance(tc.src, tc.dst)
			if err != nil {
				t.Fatalf("native GridDistance error: %v", err)
			}
			if got != int64(want) {
				t.Errorf("GridDistance: got %d, want %d", got, want)
			}
		})
	}
}

func TestGridPathCells_Oracle(t *testing.T) {
	src := h3native.Cell(0x89283082877ffff)
	disk, err := h3native.GridDisk(src, 2)
	if err != nil || len(disk) < 3 {
		t.Fatal("failed to get disk cells")
	}
	dst := disk[len(disk)/2]

	wantCGO, err := h3cgo.GridPath(nativeToCGOCell(src), nativeToCGOCell(dst))
	if err != nil {
		t.Fatalf("cgo GridPath error: %v", err)
	}
	got, err := h3native.GridPathCells(src, dst)
	if err != nil {
		t.Fatalf("native GridPathCells error: %v", err)
	}
	if len(got) != len(wantCGO) {
		t.Fatalf("GridPathCells length: got %d, want %d", len(got), len(wantCGO))
	}
	for i := range got {
		if uint64(got[i]) != uint64(int64(wantCGO[i])) {
			t.Errorf("GridPathCells[%d]: got %#x, want %#x", i, uint64(got[i]), uint64(int64(wantCGO[i])))
		}
	}
}

// ============================================================================
// Compaction Tests
// ============================================================================

func TestCompactCells_Oracle(t *testing.T) {
	parent := h3native.Cell(0x85283473fffffff)
	children := h3native.CellToChildren(parent, 6)
	if len(children) == 0 {
		t.Fatal("no children")
	}
	cgoCells := make([]h3cgo.Cell, len(children))
	for i, c := range children {
		cgoCells[i] = nativeToCGOCell(c)
	}
	wantCGO, err := h3cgo.CompactCells(cgoCells)
	if err != nil {
		t.Fatalf("cgo CompactCells error: %v", err)
	}
	got, err := h3native.CompactCells(children)
	if err != nil {
		t.Fatalf("native CompactCells error: %v", err)
	}
	compareCellSets(t, got, wantCGO)
}

func TestUncompactCells_Oracle(t *testing.T) {
	parent := h3native.Cell(0x85283473fffffff)
	res := 6
	wantCGO, err := h3cgo.UncompactCells([]h3cgo.Cell{nativeToCGOCell(parent)}, res)
	if err != nil {
		t.Fatalf("cgo UncompactCells error: %v", err)
	}
	got, err := h3native.UncompactCells([]h3native.Cell{parent}, res)
	if err != nil {
		t.Fatalf("native UncompactCells error: %v", err)
	}
	compareCellSets(t, got, wantCGO)
}

// ============================================================================
// Region Tests
// ============================================================================

func TestPolygonToCells_Oracle(t *testing.T) {
	polygon := h3native.GeoPolygon{
		GeoLoop: []h3native.LatLng{
			{Lat: 37.78, Lng: -122.42},
			{Lat: 37.78, Lng: -122.41},
			{Lat: 37.77, Lng: -122.42},
		},
	}
	res := 9
	cgoPoly := convertGeoPolygon(polygon)
	wantCGO, err := h3cgo.PolygonToCells(cgoPoly, res)
	if err != nil {
		t.Fatalf("cgo PolygonToCells error: %v", err)
	}
	got, err := h3native.PolygonToCells(polygon, res)
	if err != nil {
		t.Fatalf("native PolygonToCells error: %v", err)
	}
	compareCellSets(t, got, wantCGO)
}

func TestCellsToLinkedMultiPolygon_Oracle(t *testing.T) {
	cells := []h3native.Cell{h3native.Cell(0x89283082877ffff)}
	cgoCells := []h3cgo.Cell{nativeToCGOCell(cells[0])}

	wantCGO, err := h3cgo.CellsToMultiPolygon(cgoCells)
	if err != nil {
		t.Fatalf("cgo CellsToMultiPolygon error: %v", err)
	}
	got := h3native.CellsToLinkedMultiPolygon(cells)

	if len(got) != len(wantCGO) {
		t.Errorf("polygon count: got %d, want %d", len(got), len(wantCGO))
		return
	}
	for i := range got {
		if len(got[i].GeoLoop) != len(wantCGO[i].GeoLoop) {
			t.Errorf("polygon[%d] outer loop vertex count: got %d, want %d",
				i, len(got[i].GeoLoop), len(wantCGO[i].GeoLoop))
		}
	}
}

// ============================================================================
// Metrics Tests
// ============================================================================

func TestGreatCircleDistanceKm_Oracle(t *testing.T) {
	cases := []struct{ a, b h3native.LatLng }{
		{
			h3native.LatLng{Lat: sfLat, Lng: sfLng},
			h3native.LatLng{Lat: googLat, Lng: googLng},
		},
		{
			h3native.LatLng{Lat: 0, Lng: 0},
			h3native.LatLng{Lat: 0, Lng: 90},
		},
	}
	for i, tc := range cases {
		tc := tc
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			want := h3cgo.GreatCircleDistanceKm(
				h3cgo.NewLatLng(tc.a.Lat, tc.a.Lng),
				h3cgo.NewLatLng(tc.b.Lat, tc.b.Lng),
			)
			got := h3native.GreatCircleDistanceKm(tc.a, tc.b)
			if !relApproxEq(got, want, relTol) {
				t.Errorf("GreatCircleDistanceKm: got %v, want %v", got, want)
			}
		})
	}
}

func TestGreatCircleDistanceRads_Oracle(t *testing.T) {
	a := h3native.LatLng{Lat: sfLat, Lng: sfLng}
	b := h3native.LatLng{Lat: googLat, Lng: googLng}
	want := h3cgo.GreatCircleDistanceRads(
		h3cgo.NewLatLng(a.Lat, a.Lng),
		h3cgo.NewLatLng(b.Lat, b.Lng),
	)
	got := h3native.GreatCircleDistanceRads(a, b)
	if !relApproxEq(got, want, relTol) {
		t.Errorf("GreatCircleDistanceRads: got %v, want %v", got, want)
	}
}

func TestGreatCircleDistanceM_Oracle(t *testing.T) {
	a := h3native.LatLng{Lat: sfLat, Lng: sfLng}
	b := h3native.LatLng{Lat: googLat, Lng: googLng}
	want := h3cgo.GreatCircleDistanceM(
		h3cgo.NewLatLng(a.Lat, a.Lng),
		h3cgo.NewLatLng(b.Lat, b.Lng),
	)
	got := h3native.GreatCircleDistanceM(a, b)
	if !relApproxEq(got, want, relTol) {
		t.Errorf("GreatCircleDistanceM: got %v, want %v", got, want)
	}
}

func TestCellAreaKm2_Oracle(t *testing.T) {
	cells := []h3native.Cell{
		h3native.Cell(0x89283082877ffff),
		h3native.Cell(0x85283473fffffff),
	}
	for _, c := range cells {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			want, err := h3cgo.CellAreaKm2(nativeToCGOCell(c))
			if err != nil {
				t.Fatalf("cgo CellAreaKm2 error: %v", err)
			}
			got := h3native.CellAreaKm2(c)
			if !relApproxEq(got, want, relTol) {
				t.Errorf("CellAreaKm2: got %v, want %v", got, want)
			}
		})
	}
}

func TestCellAreaRads2_Oracle(t *testing.T) {
	c := h3native.Cell(0x85283473fffffff)
	want, err := h3cgo.CellAreaRads2(nativeToCGOCell(c))
	if err != nil {
		t.Fatalf("cgo CellAreaRads2 error: %v", err)
	}
	got := h3native.CellAreaRads2(c)
	if !relApproxEq(got, want, relTol) {
		t.Errorf("CellAreaRads2: got %v, want %v", got, want)
	}
}

func TestCellAreaM2_Oracle(t *testing.T) {
	c := h3native.Cell(0x85283473fffffff)
	want, err := h3cgo.CellAreaM2(nativeToCGOCell(c))
	if err != nil {
		t.Fatalf("cgo CellAreaM2 error: %v", err)
	}
	got := h3native.CellAreaM2(c)
	if !relApproxEq(got, want, relTol) {
		t.Errorf("CellAreaM2: got %v, want %v", got, want)
	}
}

// TestEdgeLengthKm_Oracle compares average edge length by resolution.
// CGO equivalent: HexagonEdgeLengthAvgKm.
func TestEdgeLengthKm_Oracle(t *testing.T) {
	for _, res := range []int{0, 5, 9} {
		res := res
		t.Run(fmt.Sprintf("r%d", res), func(t *testing.T) {
			want, err := h3cgo.HexagonEdgeLengthAvgKm(res)
			if err != nil {
				t.Fatalf("cgo HexagonEdgeLengthAvgKm error: %v", err)
			}
			got := h3native.EdgeLengthKm(res)
			if !relApproxEq(got, want, relTol) {
				t.Errorf("EdgeLengthKm(%d): got %v, want %v", res, got, want)
			}
		})
	}
}

// TestEdgeLengthM_Oracle compares average edge length in meters by resolution.
// CGO equivalent: HexagonEdgeLengthAvgM.
func TestEdgeLengthM_Oracle(t *testing.T) {
	for _, res := range []int{0, 5, 9} {
		res := res
		t.Run(fmt.Sprintf("r%d", res), func(t *testing.T) {
			want, err := h3cgo.HexagonEdgeLengthAvgM(res)
			if err != nil {
				t.Fatalf("cgo HexagonEdgeLengthAvgM error: %v", err)
			}
			got := h3native.EdgeLengthM(res)
			if !relApproxEq(got, want, relTol) {
				t.Errorf("EdgeLengthM(%d): got %v, want %v", res, got, want)
			}
		})
	}
}

// TestExactEdgeLengthKm_Oracle compares exact edge lengths.
// CGO equivalent: EdgeLengthKm(DirectedEdge).
func TestExactEdgeLengthKm_Oracle(t *testing.T) {
	edges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	if len(edges) == 0 {
		t.Fatal("no directed edges")
	}
	limit := len(edges)
	if limit > 3 {
		limit = 3
	}
	for i, e := range edges[:limit] {
		e := e
		t.Run(fmt.Sprintf("edge%d", i), func(t *testing.T) {
			want, err := h3cgo.EdgeLengthKm(nativeToCGOEdge(e))
			if err != nil {
				t.Fatalf("cgo EdgeLengthKm error: %v", err)
			}
			got := h3native.ExactEdgeLengthKm(e)
			if !relApproxEq(got, want, relTol) {
				t.Errorf("ExactEdgeLengthKm: got %v, want %v", got, want)
			}
		})
	}
}

// TestExactEdgeLengthRads_Oracle compares exact edge lengths in radians.
// CGO equivalent: EdgeLengthRads(DirectedEdge).
func TestExactEdgeLengthRads_Oracle(t *testing.T) {
	edges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	if len(edges) == 0 {
		t.Fatal("no directed edges")
	}
	e := edges[0]
	want, err := h3cgo.EdgeLengthRads(nativeToCGOEdge(e))
	if err != nil {
		t.Fatalf("cgo EdgeLengthRads error: %v", err)
	}
	got := h3native.ExactEdgeLengthRads(e)
	if !relApproxEq(got, want, relTol) {
		t.Errorf("ExactEdgeLengthRads: got %v, want %v", got, want)
	}
}

// TestExactEdgeLengthM_Oracle compares exact edge lengths in meters.
// CGO equivalent: EdgeLengthM(DirectedEdge).
func TestExactEdgeLengthM_Oracle(t *testing.T) {
	edges := h3native.OriginToDirectedEdges(h3native.Cell(0x89283082877ffff))
	if len(edges) == 0 {
		t.Fatal("no directed edges")
	}
	e := edges[0]
	want, err := h3cgo.EdgeLengthM(nativeToCGOEdge(e))
	if err != nil {
		t.Fatalf("cgo EdgeLengthM error: %v", err)
	}
	got := h3native.ExactEdgeLengthM(e)
	if !relApproxEq(got, want, relTol) {
		t.Errorf("ExactEdgeLengthM: got %v, want %v", got, want)
	}
}

// ============================================================================
// Directed Edge Tests
// ============================================================================

func TestAreNeighborCells_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	if len(disk) < 2 {
		t.Fatal("not enough cells in disk")
	}
	cases := []h3native.Cell{origin, disk[1]}
	for _, dst := range cases {
		dst := dst
		t.Run(fmt.Sprintf("%#x", uint64(dst)), func(t *testing.T) {
			want, err := nativeToCGOCell(origin).IsNeighbor(nativeToCGOCell(dst))
			if err != nil {
				t.Fatalf("cgo IsNeighbor error: %v", err)
			}
			got, err := h3native.AreNeighborCells(origin, dst)
			if err != nil {
				t.Fatalf("native AreNeighborCells error: %v", err)
			}
			if got != want {
				t.Errorf("AreNeighborCells: got %v, want %v", got, want)
			}
		})
	}
}

func TestCellsToDirectedEdge_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	if len(disk) < 2 {
		t.Fatal("not enough cells")
	}
	neighbor := disk[1]

	want, err := nativeToCGOCell(origin).DirectedEdge(nativeToCGOCell(neighbor))
	if err != nil {
		t.Fatalf("cgo DirectedEdge error: %v", err)
	}
	got, err := h3native.CellsToDirectedEdge(origin, neighbor)
	if err != nil {
		t.Fatalf("native CellsToDirectedEdge error: %v", err)
	}
	if uint64(got) != uint64(int64(want)) {
		t.Errorf("CellsToDirectedEdge: got %#x, want %#x", uint64(got), uint64(int64(want)))
	}
}

func TestIsValidDirectedEdge_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	edges := h3native.OriginToDirectedEdges(origin)
	if len(edges) == 0 {
		t.Fatal("no edges")
	}
	for i, e := range edges {
		e := e
		t.Run(fmt.Sprintf("edge%d", i), func(t *testing.T) {
			want := nativeToCGOEdge(e).IsValid()
			got := h3native.IsValidDirectedEdge(e)
			if got != want {
				t.Errorf("IsValidDirectedEdge: got %v, want %v", got, want)
			}
		})
	}
}

func TestGetDirectedEdgeOrigin_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	edges := h3native.OriginToDirectedEdges(origin)
	if len(edges) == 0 {
		t.Fatal("no edges")
	}
	e := edges[0]
	want, err := nativeToCGOEdge(e).Origin()
	if err != nil {
		t.Fatalf("cgo Origin error: %v", err)
	}
	got := h3native.GetDirectedEdgeOrigin(e)
	if uint64(got) != uint64(int64(want)) {
		t.Errorf("GetDirectedEdgeOrigin: got %#x, want %#x", uint64(got), uint64(int64(want)))
	}
}

func TestGetDirectedEdgeDestination_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	if len(disk) < 2 {
		t.Fatal("not enough cells")
	}
	neighbor := disk[1]
	e, err := h3native.CellsToDirectedEdge(origin, neighbor)
	if err != nil {
		t.Fatalf("CellsToDirectedEdge error: %v", err)
	}
	want, err := nativeToCGOEdge(e).Destination()
	if err != nil {
		t.Fatalf("cgo Destination error: %v", err)
	}
	got := h3native.GetDirectedEdgeDestination(e)
	if uint64(got) != uint64(int64(want)) {
		t.Errorf("GetDirectedEdgeDestination: got %#x, want %#x", uint64(got), uint64(int64(want)))
	}
}

func TestOriginToDirectedEdges_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	wantCGO, err := nativeToCGOCell(origin).DirectedEdges()
	if err != nil {
		t.Fatalf("cgo DirectedEdges error: %v", err)
	}
	got := h3native.OriginToDirectedEdges(origin)
	if len(got) != len(wantCGO) {
		t.Fatalf("edge count: got %d, want %d", len(got), len(wantCGO))
	}
	// Sort by uint64 value and compare
	gotSorted := make([]uint64, len(got))
	wantSorted := make([]uint64, len(wantCGO))
	for i, e := range got {
		gotSorted[i] = uint64(e)
	}
	for i, e := range wantCGO {
		wantSorted[i] = uint64(int64(e))
	}
	sort.Slice(gotSorted, func(i, j int) bool { return gotSorted[i] < gotSorted[j] })
	sort.Slice(wantSorted, func(i, j int) bool { return wantSorted[i] < wantSorted[j] })
	for i := range gotSorted {
		if gotSorted[i] != wantSorted[i] {
			t.Errorf("edge[%d]: native=%#x, cgo=%#x", i, gotSorted[i], wantSorted[i])
		}
	}
}

func TestDirectedEdgeToBoundary_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	if len(disk) < 2 {
		t.Fatal("not enough cells")
	}
	neighbor := disk[1]
	e, err := h3native.CellsToDirectedEdge(origin, neighbor)
	if err != nil {
		t.Fatalf("CellsToDirectedEdge error: %v", err)
	}
	want, err := nativeToCGOEdge(e).Boundary()
	if err != nil {
		t.Fatalf("cgo edge Boundary error: %v", err)
	}
	got := h3native.DirectedEdgeToBoundary(e)
	if len(got) != len(want) {
		t.Fatalf("DirectedEdgeToBoundary vertex count: got %d, want %d", len(got), len(want))
	}
	for i := range got {
		if !approxEq(got[i].Lat, want[i].Lat, epsilon) {
			t.Errorf("vertex[%d].Lat: got %v, want %v", i, got[i].Lat, want[i].Lat)
		}
		if !approxEq(got[i].Lng, want[i].Lng, epsilon) {
			t.Errorf("vertex[%d].Lng: got %v, want %v", i, got[i].Lng, want[i].Lng)
		}
	}
}

func TestDirectedEdgeToCells_Oracle(t *testing.T) {
	origin := h3native.Cell(0x89283082877ffff)
	disk, _ := h3native.GridDisk(origin, 1)
	if len(disk) < 2 {
		t.Fatal("not enough cells")
	}
	neighbor := disk[1]
	e, err := h3native.CellsToDirectedEdge(origin, neighbor)
	if err != nil {
		t.Fatalf("CellsToDirectedEdge error: %v", err)
	}
	wantCells, err := nativeToCGOEdge(e).Cells()
	if err != nil {
		t.Fatalf("cgo edge Cells error: %v", err)
	}
	got := h3native.DirectedEdgeToCells(e)
	if len(wantCells) != 2 || len(got) != 2 {
		t.Fatalf("expected 2 cells each: native=%d, cgo=%d", len(got), len(wantCells))
	}
	if uint64(got[0]) != uint64(int64(wantCells[0])) {
		t.Errorf("origin cell: got %#x, want %#x", uint64(got[0]), uint64(int64(wantCells[0])))
	}
	if uint64(got[1]) != uint64(int64(wantCells[1])) {
		t.Errorf("dest cell: got %#x, want %#x", uint64(got[1]), uint64(int64(wantCells[1])))
	}
}

// ============================================================================
// Vertex Tests
//
// Note: Native vertex indices use a simplified canonical form that may differ
// from the CGO canonical form. Tests compare geographic positions only.
// ============================================================================

func TestCellToVertex_Oracle(t *testing.T) {
	c := h3native.Cell(0x89283082877ffff) // hexagon, 6 vertices
	for vn := 0; vn < 6; vn++ {
		vn := vn
		t.Run(fmt.Sprintf("v%d", vn), func(t *testing.T) {
			cgoVertex, err := nativeToCGOCell(c).Vertex(vn)
			if err != nil {
				t.Fatalf("cgo Vertex(%d) error: %v", vn, err)
			}
			wantLatLng, err := cgoVertex.LatLng()
			if err != nil {
				t.Fatalf("cgo vertex LatLng error: %v", err)
			}
			nativeVertex := h3native.CellToVertex(c, vn)
			gotLatLng := h3native.VertexToLatLng(nativeVertex)
			// Use slightly looser epsilon for vertices (1e-6 ≈ 0.1m at equator)
			const vertexEps = 1e-6
			if !approxEq(gotLatLng.Lat, wantLatLng.Lat, vertexEps) {
				t.Errorf("vertex[%d].Lat: got %.10f, want %.10f", vn, gotLatLng.Lat, wantLatLng.Lat)
			}
			if !approxEq(gotLatLng.Lng, wantLatLng.Lng, vertexEps) {
				t.Errorf("vertex[%d].Lng: got %.10f, want %.10f", vn, gotLatLng.Lng, wantLatLng.Lng)
			}
		})
	}
}

func TestCellToVertexes_Oracle(t *testing.T) {
	cases := []h3native.Cell{
		h3native.Cell(0x89283082877ffff), // hexagon
	}
	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%#x", uint64(c)), func(t *testing.T) {
			wantVertexes, err := nativeToCGOCell(c).Vertexes()
			if err != nil {
				t.Fatalf("cgo Vertexes error: %v", err)
			}
			got := h3native.CellToVertexes(c)
			if len(got) != len(wantVertexes) {
				t.Errorf("CellToVertexes count: got %d, want %d", len(got), len(wantVertexes))
			}
		})
	}
}

func TestVertexToLatLng_Oracle(t *testing.T) {
	c := h3native.Cell(0x89283082877ffff)
	for vn := 0; vn < 6; vn++ {
		vn := vn
		t.Run(fmt.Sprintf("v%d", vn), func(t *testing.T) {
			cgoVertex, err := nativeToCGOCell(c).Vertex(vn)
			if err != nil {
				t.Fatalf("cgo Vertex(%d) error: %v", vn, err)
			}
			want, err := cgoVertex.LatLng()
			if err != nil {
				t.Fatalf("cgo VertexToLatLng error: %v", err)
			}
			nativeVertex := h3native.CellToVertex(c, vn)
			got := h3native.VertexToLatLng(nativeVertex)
			const vertexEps = 1e-6
			if !approxEq(got.Lat, want.Lat, vertexEps) {
				t.Errorf("v%d Lat: got %.10f, want %.10f", vn, got.Lat, want.Lat)
			}
			if !approxEq(got.Lng, want.Lng, vertexEps) {
				t.Errorf("v%d Lng: got %.10f, want %.10f", vn, got.Lng, want.Lng)
			}
		})
	}
}

func TestIsValidVertex_Oracle(t *testing.T) {
	c := h3native.Cell(0x89283082877ffff)
	// Native-created vertices must be valid according to native IsValidVertex
	vertexes := h3native.CellToVertexes(c)
	for i, v := range vertexes {
		v := v
		if !h3native.IsValidVertex(v) {
			t.Errorf("CellToVertexes()[%d] IsValidVertex = false, want true", i)
		}
	}
	// CGO-created vertices must be valid according to CGO IsValid
	for vn := 0; vn < 6; vn++ {
		cv, err := nativeToCGOCell(c).Vertex(vn)
		if err != nil {
			t.Fatalf("cgo Vertex(%d) error: %v", vn, err)
		}
		if !h3cgo.IsValidVertex(cv) {
			t.Errorf("CGO Vertex(%d).IsValid = false, want true", vn)
		}
	}
	// Zero vertex should be invalid in native
	if h3native.IsValidVertex(h3native.Vertex(0)) {
		t.Error("Vertex(0) IsValidVertex = true, want false")
	}
}
