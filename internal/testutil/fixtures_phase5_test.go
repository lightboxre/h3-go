package testutil

import (
	"testing"
)

func TestParseCellParentFile(t *testing.T) {
	cases, err := ParseCellParentFile("../../testdata/cell_to_parent.txt")
	if err != nil {
		t.Fatalf("ParseCellParentFile failed: %v", err)
	}

	if len(cases) == 0 {
		t.Fatal("Expected at least one case")
	}

	// Check first case
	if cases[0].ChildCell == 0 {
		t.Error("Expected non-zero child cell")
	}
	if cases[0].ParentRes < 0 {
		t.Error("Expected non-negative parent resolution")
	}
	if cases[0].ParentCell == 0 {
		t.Error("Expected non-zero parent cell")
	}

	t.Logf("Parsed %d cell parent cases", len(cases))
}

func TestParseGridDistanceFile(t *testing.T) {
	cases, err := ParseCellPairFile("../../testdata/grid_distance.txt")
	if err != nil {
		t.Fatalf("ParseCellPairFile failed: %v", err)
	}

	if len(cases) == 0 {
		t.Fatal("Expected at least one case")
	}

	// Check first case
	if cases[0].Cell1 == 0 {
		t.Error("Expected non-zero cell1")
	}
	if cases[0].Distance < 0 {
		t.Error("Expected non-negative distance")
	}

	t.Logf("Parsed %d grid distance cases", len(cases))
}

func TestParseLatLngCellRes5File(t *testing.T) {
	cases, err := ParseLatLngCellFile("../../testdata/latlng_to_cell_res5.txt")
	if err != nil {
		t.Fatalf("ParseLatLngCellFile failed: %v", err)
	}

	if len(cases) == 0 {
		t.Fatal("Expected at least one case")
	}

	// Check that all cases are res 5
	for i, c := range cases {
		if c.Res != 5 {
			t.Errorf("Case %d: expected res 5, got %d", i, c.Res)
		}
		if c.Cell == 0 {
			t.Errorf("Case %d: expected non-zero cell", i)
		}
	}

	t.Logf("Parsed %d lat/lng to cell res 5 cases", len(cases))
}
