// Package testutil provides utilities for loading and parsing test fixtures.
package testutil

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LatLngCell represents a lat/lng → cell test case.
type LatLngCell struct {
	Lat  float64
	Lng  float64
	Res  int
	Cell uint64 // expected H3 cell index
}

// ParseLatLngCellFile parses a fixture file with format:
// lat lng res cell
// Example: 37.3615593 -122.0553238 5 85283473fffffff
// Lines starting with # are comments.
// Empty lines are ignored.
func ParseLatLngCellFile(path string) ([]LatLngCell, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var cases []LatLngCell
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			return nil, fmt.Errorf("line %d: expected 4 fields (lat lng res cell), got %d", lineNum, len(fields))
		}

		lat, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid lat: %w", lineNum, err)
		}

		lng, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid lng: %w", lineNum, err)
		}

		res, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid resolution: %w", lineNum, err)
		}

		cell, err := strconv.ParseUint(fields[3], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid cell (expected hex): %w", lineNum, err)
		}

		cases = append(cases, LatLngCell{
			Lat:  lat,
			Lng:  lng,
			Res:  res,
			Cell: cell,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return cases, nil
}

// CellPair represents a pair of cells for testing neighbor relationships, distances, etc.
type CellPair struct {
	Cell1    uint64
	Cell2    uint64
	Distance int64 // expected grid distance, -1 if not neighbors
}

// ParseCellPairFile parses a fixture file with format:
// cell1 cell2 distance
// Example: 85283473fffffff 85283477fffffff 1
func ParseCellPairFile(path string) ([]CellPair, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var pairs []CellPair
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("line %d: expected 3 fields (cell1 cell2 distance), got %d", lineNum, len(fields))
		}

		cell1, err := strconv.ParseUint(fields[0], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid cell1: %w", lineNum, err)
		}

		cell2, err := strconv.ParseUint(fields[1], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid cell2: %w", lineNum, err)
		}

		distance, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid distance: %w", lineNum, err)
		}

		pairs = append(pairs, CellPair{
			Cell1:    cell1,
			Cell2:    cell2,
			Distance: distance,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return pairs, nil
}

// CellParent represents a child cell → parent cell relationship.
type CellParent struct {
	ChildCell  uint64
	ParentRes  int
	ParentCell uint64
}

// ParseCellParentFile parses a fixture file with format:
// child_cell parent_res parent_cell
// Example: 89283082877ffff 8 8828308287fffff
func ParseCellParentFile(path string) ([]CellParent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	var cases []CellParent
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("line %d: expected 3 fields (child_cell parent_res parent_cell), got %d", lineNum, len(fields))
		}

		childCell, err := strconv.ParseUint(fields[0], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid child_cell: %w", lineNum, err)
		}

		parentRes, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid parent_res: %w", lineNum, err)
		}

		parentCell, err := strconv.ParseUint(fields[2], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid parent_cell: %w", lineNum, err)
		}

		cases = append(cases, CellParent{
			ChildCell:  childCell,
			ParentRes:  parentRes,
			ParentCell: parentCell,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	return cases, nil
}
