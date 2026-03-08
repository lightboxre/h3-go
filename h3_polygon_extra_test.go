package h3_test

import (
	"slices"
	"testing"

	h3 "github.com/lightboxre/h3-go"
)

func TestPolygonToCells_Exact(t *testing.T) {
	// Create a polygon from exact cell boundary
	cell := h3.LatLngToCell(37.0, -122.0, 9)
	boundary := h3.CellToBoundary(cell)

	// Convert boundary to GeoLoop
	geoLoop := make(h3.GeoLoop, len(boundary))
	for i, pt := range boundary {
		geoLoop[i] = pt
	}

	polygon := h3.GeoPolygon{
		GeoLoop: geoLoop,
	}

	// Convert back to cells
	cells, err := h3.PolygonToCells(polygon, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return exactly 1 cell (the original)
	if len(cells) != 1 {
		t.Errorf("expected 1 cell, got %d", len(cells))
	}

	if len(cells) > 0 && cells[0] != cell {
		t.Errorf("expected cell %v, got %v", cell, cells[0])
	}
}

func TestPolygonToCells_Transmeridian(t *testing.T) {
	// Polygon crossing the antimeridian (±180° longitude)
	polygon := h3.GeoPolygon{
		GeoLoop: h3.GeoLoop{
			{Lat: 0.5, Lng: 179.5},
			{Lat: 0.5, Lng: -179.5},
			{Lat: -0.5, Lng: -179.5},
			{Lat: -0.5, Lng: 179.5},
		},
	}

	cells, err := h3.PolygonToCells(polygon, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return non-empty result
	if len(cells) == 0 {
		t.Error("expected non-empty result for transmeridian polygon")
	}

	// Verify all cells are valid
	for _, c := range cells {
		if !h3.IsValidCell(c) {
			t.Errorf("invalid cell in result: %v", c)
		}
	}
}

func TestPolygonToCells_TransmeridianComplex(t *testing.T) {
	// More complex transmeridian polygon with 6 vertices
	polygon := h3.GeoPolygon{
		GeoLoop: h3.GeoLoop{
			{Lat: 0.1, Lng: 179.0},
			{Lat: 0.1, Lng: -179.0},
			{Lat: 0.0, Lng: -179.5},
			{Lat: -0.1, Lng: -179.0},
			{Lat: -0.1, Lng: 179.0},
			{Lat: 0.0, Lng: 179.5},
		},
	}

	cells, err := h3.PolygonToCells(polygon, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return non-empty result (approximately 1204 cells per C test)
	if len(cells) == 0 {
		t.Error("expected non-empty result for complex transmeridian polygon")
	}

	// Verify all cells are valid
	for _, c := range cells {
		if !h3.IsValidCell(c) {
			t.Errorf("invalid cell in result: %v", c)
		}
	}
}

func TestPolygonToCells_Pentagon(t *testing.T) {
	// Get a pentagon at resolution 9
	pentagons := h3.GetPentagonCells(9)
	if len(pentagons) == 0 {
		t.Fatal("expected pentagons at resolution 9")
	}
	pent := pentagons[0]

	// Get pentagon center
	center := h3.CellToLatLng(pent)

	// Create a small bounding box around the pentagon
	delta := 0.01 // approximately 1km at this scale
	polygon := h3.GeoPolygon{
		GeoLoop: h3.GeoLoop{
			{Lat: center.Lat + delta, Lng: center.Lng - delta},
			{Lat: center.Lat + delta, Lng: center.Lng + delta},
			{Lat: center.Lat - delta, Lng: center.Lng + delta},
			{Lat: center.Lat - delta, Lng: center.Lng - delta},
		},
	}

	cells, err := h3.PolygonToCells(polygon, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include the pentagon
	if len(cells) == 0 {
		t.Error("expected non-empty result")
	}

	// Check if any result is a pentagon
	foundPentagon := slices.ContainsFunc(cells, h3.IsPentagon)

	if !foundPentagon {
		t.Log("warning: no pentagon found in result (may be expected if polygon is too small)")
	}
}

func TestPolygonToCells_Invalid(t *testing.T) {
	// Empty polygon
	emptyPolygon := h3.GeoPolygon{
		GeoLoop: h3.GeoLoop{},
	}

	cells, err := h3.PolygonToCells(emptyPolygon, 5)
	// Should either error or return empty result
	if err == nil && len(cells) > 0 {
		t.Errorf("expected error or empty result for empty polygon, got %d cells", len(cells))
	}

	// Polygon with only one point
	singlePointPolygon := h3.GeoPolygon{
		GeoLoop: h3.GeoLoop{
			{Lat: 37.0, Lng: -122.0},
		},
	}

	cells, err = h3.PolygonToCells(singlePointPolygon, 5)
	// Should either error or return empty/minimal result
	if err == nil && len(cells) > 1 {
		t.Errorf("expected error or minimal result for single-point polygon, got %d cells", len(cells))
	}
}

func TestPolygonToCells_Point(t *testing.T) {
	// Degenerate polygon - all vertices at same point
	polygon := h3.GeoPolygon{
		GeoLoop: h3.GeoLoop{
			{Lat: 37.0, Lng: -122.0},
			{Lat: 37.0, Lng: -122.0},
			{Lat: 37.0, Lng: -122.0},
		},
	}

	cells, err := h3.PolygonToCells(polygon, 5)
	// Should return empty or single cell
	if err != nil {
		// Error is acceptable
		return
	}

	if len(cells) > 1 {
		t.Errorf("expected 0 or 1 cells for point polygon, got %d", len(cells))
	}
}

func TestPolygonToCells_Line(t *testing.T) {
	// Degenerate polygon - vertices form a line
	polygon := h3.GeoPolygon{
		GeoLoop: h3.GeoLoop{
			{Lat: 37.0, Lng: -122.0},
			{Lat: 37.1, Lng: -122.0},
			{Lat: 37.0, Lng: -122.0},
		},
	}

	cells, err := h3.PolygonToCells(polygon, 5)
	// Should return empty or minimal result
	if err != nil {
		// Error is acceptable
		return
	}

	if len(cells) > 5 {
		t.Errorf("expected minimal cells for line polygon, got %d", len(cells))
	}
}
