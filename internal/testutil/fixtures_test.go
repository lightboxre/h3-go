package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseLatLngCellFile(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := `# Test file
# lat lng res cell
37.3615593 -122.0553238 5 85283473fffffff
0.0 0.0 0 8075fffffffffff

# Another line
51.5074 -0.1278 5 8531e7bffffffff
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	cases, err := ParseLatLngCellFile(testFile)
	if err != nil {
		t.Fatalf("ParseLatLngCellFile() error = %v", err)
	}

	// Should have 3 cases (2 data lines + 1 blank line ignored)
	if len(cases) != 3 {
		t.Errorf("ParseLatLngCellFile() returned %d cases, want 3", len(cases))
	}

	// Check first case
	if cases[0].Lat != 37.3615593 {
		t.Errorf("cases[0].Lat = %f, want 37.3615593", cases[0].Lat)
	}
	if cases[0].Lng != -122.0553238 {
		t.Errorf("cases[0].Lng = %f, want -122.0553238", cases[0].Lng)
	}
	if cases[0].Res != 5 {
		t.Errorf("cases[0].Res = %d, want 5", cases[0].Res)
	}
	if cases[0].Cell != 0x85283473fffffff {
		t.Errorf("cases[0].Cell = %#x, want 0x85283473fffffff", cases[0].Cell)
	}
}

func TestParseCellPairFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := `# Test file
85283473fffffff 85283477fffffff 1
8928308280fffff 8928308280fffff 0
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	pairs, err := ParseCellPairFile(testFile)
	if err != nil {
		t.Fatalf("ParseCellPairFile() error = %v", err)
	}

	if len(pairs) != 2 {
		t.Errorf("ParseCellPairFile() returned %d pairs, want 2", len(pairs))
	}

	if pairs[0].Cell1 != 0x85283473fffffff {
		t.Errorf("pairs[0].Cell1 = %#x, want 0x85283473fffffff", pairs[0].Cell1)
	}
	if pairs[0].Distance != 1 {
		t.Errorf("pairs[0].Distance = %d, want 1", pairs[0].Distance)
	}
}

func TestParseLatLngCellFileNotFound(t *testing.T) {
	_, err := ParseLatLngCellFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ParseLatLngCellFile() should return error for nonexistent file")
	}
}
