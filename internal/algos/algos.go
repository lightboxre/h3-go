// Package algos implements H3 grid algorithms for disk, ring, path, and compact/uncompact operations.
// Ported from the C H3 reference implementation (algos.c).
package algos

import (
	"errors"
	"fmt"

	"github.com/h3-native/h3-go/internal/constants"
	"github.com/h3-native/h3-go/internal/coordijk"
	"github.com/h3-native/h3-go/internal/faceijk"
	"github.com/h3-native/h3-go/internal/h3index"
)

// Sentinel errors
var (
	// ErrPentagon is returned when a pentagon is encountered in grid disk traversal.
	ErrPentagon = errors.New("pentagon encountered in grid disk traversal")
	// ErrDomain is returned when an invalid parameter is provided.
	ErrDomain = errors.New("invalid domain parameter")
	// ErrCellInvalid is returned when an invalid H3 cell is encountered.
	ErrCellInvalid = errors.New("invalid H3 cell")
	// ErrFailed is returned for general failures.
	ErrFailed = errors.New("operation failed")
	// ErrNotConnected is returned when cells are not grid-connected.
	ErrNotConnected = errors.New("cells not connected")
)

// DIRECTIONS lists the 6 directions for CCW traversal (excluding CENTER).
// These are digit values used for traversing a hexagonal ring counterclockwise.
// Order: J, JK, K, IK, I, IJ
var DIRECTIONS = [6]int{
	constants.J_AXES_DIGIT,  // 2
	constants.JK_AXES_DIGIT, // 3
	constants.K_AXES_DIGIT,  // 1
	constants.IK_AXES_DIGIT, // 5
	constants.I_AXES_DIGIT,  // 4
	constants.IJ_AXES_DIGIT, // 6
}

// NEXT_RING_DIRECTION is the direction used to move to the next outward ring.
const NEXT_RING_DIRECTION = constants.I_AXES_DIGIT

// NEW_DIGIT_II[from][to] gives the new digit when moving from digit 'from'
// in direction 'to' in a Class II cell.
// Transcribed from C's algos.c: newDigitII
var NEW_DIGIT_II = [7][7]int{
	{constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT},
	{constants.K_AXES_DIGIT, constants.I_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.J_AXES_DIGIT, constants.CENTER_DIGIT},
	{constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.K_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.IK_AXES_DIGIT},
	{constants.JK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.J_AXES_DIGIT},
	{constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.K_AXES_DIGIT},
	{constants.IK_AXES_DIGIT, constants.J_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.I_AXES_DIGIT},
	{constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.IK_AXES_DIGIT, constants.J_AXES_DIGIT, constants.K_AXES_DIGIT, constants.I_AXES_DIGIT, constants.JK_AXES_DIGIT},
}

// NEW_ADJUSTMENT_II[from][to] gives the adjustment digit (move at coarser level)
// when moving from digit 'from' in direction 'to' in a Class II cell.
var NEW_ADJUSTMENT_II = [7][7]int{
	{constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT},
	{constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.CENTER_DIGIT, constants.IK_AXES_DIGIT, constants.CENTER_DIGIT},
	{constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.J_AXES_DIGIT},
	{constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT},
	{constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.I_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IJ_AXES_DIGIT},
	{constants.CENTER_DIGIT, constants.IK_AXES_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.CENTER_DIGIT},
	{constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.J_AXES_DIGIT, constants.CENTER_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.IJ_AXES_DIGIT},
}

// NEW_DIGIT_III[from][to] gives the new digit when moving from digit 'from'
// in direction 'to' in a Class III cell.
var NEW_DIGIT_III = [7][7]int{
	{constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT},
	{constants.K_AXES_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT},
	{constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT},
	{constants.JK_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.J_AXES_DIGIT},
	{constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT},
	{constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.I_AXES_DIGIT},
	{constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT},
}

// NEW_ADJUSTMENT_III[from][to] gives the adjustment digit (move at coarser level)
// when moving from digit 'from' in direction 'to' in a Class III cell.
var NEW_ADJUSTMENT_III = [7][7]int{
	{constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT},
	{constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.CENTER_DIGIT, constants.JK_AXES_DIGIT, constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.CENTER_DIGIT},
	{constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.J_AXES_DIGIT, constants.J_AXES_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.IJ_AXES_DIGIT},
	{constants.CENTER_DIGIT, constants.JK_AXES_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT},
	{constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.I_AXES_DIGIT},
	{constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.IK_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.CENTER_DIGIT},
	{constants.CENTER_DIGIT, constants.CENTER_DIGIT, constants.IJ_AXES_DIGIT, constants.CENTER_DIGIT, constants.I_AXES_DIGIT, constants.CENTER_DIGIT, constants.IJ_AXES_DIGIT},
}

// MaxGridDiskSize returns the maximum number of cells in a grid disk of radius k.
// Formula: 3*k*(k+1) + 1
func MaxGridDiskSize(k int) (int64, error) {
	if k < 0 {
		return 0, ErrDomain
	}
	return 3*int64(k)*(int64(k)+1) + 1, nil
}

// GridDisk produces all H3 cells within k steps of the origin cell.
// Includes the origin cell.
func GridDisk(origin h3index.H3Index, k int) ([]h3index.H3Index, error) {
	cells, err := GridDiskUnsafe(origin, k)
	if err == nil {
		return cells, nil
	}
	// Fall back to safe BFS-based approach
	return gridDiskSafe(origin, k)
}

// GridDiskDistances produces all H3 cells within k steps, with distances.
// output[dist] contains cells at exactly distance dist.
func GridDiskDistances(origin h3index.H3Index, k int) ([][]h3index.H3Index, error) {
	cells, err := GridDiskDistancesUnsafe(origin, k)
	if err == nil {
		return cells, nil
	}
	// Fall back to safe approach
	return gridDiskDistancesSafe(origin, k)
}

// GridDiskUnsafe is a fast version that fails if a pentagon is encountered.
// Returns (cells, error) where error is non-nil if a pentagon is encountered.
func GridDiskUnsafe(origin h3index.H3Index, k int) ([]h3index.H3Index, error) {
	if k < 0 {
		return nil, ErrDomain
	}

	maxSize, err := MaxGridDiskSize(k)
	if err != nil {
		return nil, err
	}

	out := make([]h3index.H3Index, maxSize)
	idx := 0

	// Add origin
	out[idx] = origin
	idx++

	if h3index.IsPentagon(origin) {
		return nil, ErrPentagon
	}

	if k == 0 {
		return out[:idx], nil
	}

	// Traverse rings from 1 to k
	ring := 1
	direction := 0
	i := 0
	rotations := 0
	current := origin

	for ring <= k {
		if direction == 0 && i == 0 {
			// Move to the start of the next ring
			neighborResult := h3NeighborRotations(current, NEXT_RING_DIRECTION, &rotations, &current)
			if neighborResult != nil {
				return nil, neighborResult
			}

			if h3index.IsPentagon(current) {
				return nil, ErrPentagon
			}
		}

		neighborResult := h3NeighborRotations(current, DIRECTIONS[direction], &rotations, &current)
		if neighborResult != nil {
			return nil, neighborResult
		}

		out[idx] = current
		idx++

		i++
		// Check if end of this side of the k-ring
		if i == ring {
			i = 0
			direction++
			// Check if end of this ring
			if direction == 6 {
				direction = 0
				ring++
			}
		}

		if h3index.IsPentagon(current) {
			return nil, ErrPentagon
		}
	}

	return out[:idx], nil
}

// GridDiskDistancesUnsafe is the fast version with distances.
func GridDiskDistancesUnsafe(origin h3index.H3Index, k int) ([][]h3index.H3Index, error) {
	if k < 0 {
		return nil, ErrDomain
	}

	maxSize, err := MaxGridDiskSize(k)
	if err != nil {
		return nil, err
	}

	out := make([]h3index.H3Index, maxSize)
	distances := make([]int, maxSize)
	idx := 0

	// Add origin
	out[idx] = origin
	distances[idx] = 0
	idx++

	if h3index.IsPentagon(origin) {
		return nil, ErrPentagon
	}

	if k == 0 {
		result := make([][]h3index.H3Index, 1)
		result[0] = []h3index.H3Index{origin}
		return result, nil
	}

	// Traverse rings from 1 to k
	ring := 1
	direction := 0
	i := 0
	rotations := 0
	current := origin

	for ring <= k {
		if direction == 0 && i == 0 {
			// Move to the start of the next ring
			neighborResult := h3NeighborRotations(current, NEXT_RING_DIRECTION, &rotations, &current)
			if neighborResult != nil {
				return nil, neighborResult
			}

			if h3index.IsPentagon(current) {
				return nil, ErrPentagon
			}
		}

		neighborResult := h3NeighborRotations(current, DIRECTIONS[direction], &rotations, &current)
		if neighborResult != nil {
			return nil, neighborResult
		}

		out[idx] = current
		distances[idx] = ring
		idx++

		i++
		// Check if end of this side of the k-ring
		if i == ring {
			i = 0
			direction++
			// Check if end of this ring
			if direction == 6 {
				direction = 0
				ring++
			}
		}

		if h3index.IsPentagon(current) {
			return nil, ErrPentagon
		}
	}

	// Group by distance
	result := make([][]h3index.H3Index, k+1)
	for j := 0; j < idx; j++ {
		dist := distances[j]
		result[dist] = append(result[dist], out[j])
	}

	return result, nil
}

// GridRingUnsafe produces the cells at exactly k distance from origin.
// Fails if pentagon encountered. Returns error if pentagon.
func GridRingUnsafe(origin h3index.H3Index, k int) ([]h3index.H3Index, error) {
	if k < 0 {
		return nil, ErrDomain
	}

	if k == 0 {
		return []h3index.H3Index{origin}, nil
	}

	out := make([]h3index.H3Index, 6*k)
	idx := 0
	rotations := 0

	if h3index.IsPentagon(origin) {
		return nil, ErrPentagon
	}

	// Move to the start of ring k
	current := origin
	for ring := 0; ring < k; ring++ {
		neighborResult := h3NeighborRotations(current, NEXT_RING_DIRECTION, &rotations, &current)
		if neighborResult != nil {
			return nil, neighborResult
		}

		if h3index.IsPentagon(current) {
			return nil, ErrPentagon
		}
	}

	lastIndex := current
	out[idx] = current
	idx++

	// Traverse the ring
	for direction := 0; direction < 6; direction++ {
		for pos := 0; pos < k; pos++ {
			neighborResult := h3NeighborRotations(current, DIRECTIONS[direction], &rotations, &current)
			if neighborResult != nil {
				return nil, neighborResult
			}

			// Skip the very last index, it was already added
			if pos != k-1 || direction != 5 {
				out[idx] = current
				idx++

				if h3index.IsPentagon(current) {
					return nil, ErrPentagon
				}
			}
		}
	}

	// Check for pentagonal distortion
	if lastIndex != current {
		return nil, ErrPentagon
	}

	return out[:idx], nil
}

// GridPathCells returns all cells on the path from src to dst (inclusive).
// Uses H3's canonical algorithm: at each step pick the neighbor that minimises
// grid distance to dst; on ties, prefer the neighbor whose IJK displacement
// from the current cell matches the normalised IJK direction from current→dst.
func GridPathCells(src, dst h3index.H3Index) ([]h3index.H3Index, error) {
	if src == dst {
		return []h3index.H3Index{src}, nil
	}

	dist, err := GridDistance(src, dst)
	if err != nil {
		return nil, err
	}

	path := make([]h3index.H3Index, dist+1)
	path[0] = src
	path[dist] = dst

	// Hoist dst face/coord outside the step loop — O(1) per step instead of O(BFS).
	dstFijk := faceijk.H3ToFaceIJK(dst)

	current := src
	for step := int64(1); step < dist; step++ {
		neighbors := getNeighbors(current)
		best := h3index.H3_NULL
		bestGridDist := int64(1 << 62)
		bestIJKMatch := false

		// Compute the canonical step direction from current toward dst in IJK space.
		currentFijk := faceijk.H3ToFaceIJK(current)
		sameFace := currentFijk.Face == dstFijk.Face
		var wantI, wantJ, wantK int
		if sameFace {
			delta := coordijk.CoordIJK{
				I: dstFijk.Coord.I - currentFijk.Coord.I,
				J: dstFijk.Coord.J - currentFijk.Coord.J,
				K: dstFijk.Coord.K - currentFijk.Coord.K,
			}
			coordijk.IJKNormalize(&delta)
			if delta.I > 0 {
				wantI = 1
			}
			if delta.J > 0 {
				wantJ = 1
			}
			if delta.K > 0 {
				wantK = 1
			}
		}

		for _, n := range neighbors {
			if n == h3index.H3_NULL {
				continue
			}
			// O(1) distance using IJK math where possible; BFS only for cross-face.
			var d int64
			nFijk := faceijk.H3ToFaceIJK(n)
			if nFijk.Face == dstFijk.Face {
				d = int64(coordijk.IJKDistance(nFijk.Coord, dstFijk.Coord))
			} else {
				d, err = GridDistance(n, dst)
				if err != nil {
					continue
				}
			}
			nMatches := false
			if sameFace && nFijk.Face == currentFijk.Face {
				dir := coordijk.CoordIJK{
					I: nFijk.Coord.I - currentFijk.Coord.I,
					J: nFijk.Coord.J - currentFijk.Coord.J,
					K: nFijk.Coord.K - currentFijk.Coord.K,
				}
				coordijk.IJKNormalize(&dir)
				di, dj, dk := 0, 0, 0
				if dir.I > 0 {
					di = 1
				}
				if dir.J > 0 {
					dj = 1
				}
				if dir.K > 0 {
					dk = 1
				}
				nMatches = (di == wantI && dj == wantJ && dk == wantK)
			}
			if d < bestGridDist || (d == bestGridDist && nMatches && !bestIJKMatch) {
				bestGridDist = d
				bestIJKMatch = nMatches
				best = n
			}
		}

		if best == h3index.H3_NULL {
			return nil, fmt.Errorf("path failed at step %d", step)
		}

		path[step] = best
		current = best
	}

	return path, nil
}

// GridPathCellsSize returns the number of cells in the path from src to dst.
func GridPathCellsSize(src, dst h3index.H3Index) (int64, error) {
	dist, err := GridDistance(src, dst)
	if err != nil {
		return 0, err
	}
	return dist + 1, nil
}

// GridDistance returns the grid distance between two cells.
func GridDistance(src, dst h3index.H3Index) (int64, error) {
	if src == dst {
		return 0, nil
	}

	// Fast O(1) path: same icosahedral face → use IJK distance.
	srcFijk := faceijk.H3ToFaceIJK(src)
	dstFijk := faceijk.H3ToFaceIJK(dst)
	if srcFijk.Face == dstFijk.Face {
		d := coordijk.IJKDistance(srcFijk.Coord, dstFijk.Coord)
		return int64(d), nil
	}

	// Cross-face fallback: BFS.
	seen := map[h3index.H3Index]int64{src: 0}
	queue := []h3index.H3Index{src}
	const maxDepth = int64(1000)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		dist := seen[current]

		if dist >= maxDepth {
			return -1, ErrNotConnected
		}

		neighbors := getNeighbors(current)
		for _, n := range neighbors {
			if n == h3index.H3_NULL {
				continue
			}
			if _, exists := seen[n]; !exists {
				seen[n] = dist + 1
				if n == dst {
					return dist + 1, nil
				}
				queue = append(queue, n)
			}
		}
	}

	return -1, ErrNotConnected
}

// CompactCells compacts a set of cells (removes cells whose parent is also present).
func CompactCells(cells []h3index.H3Index) ([]h3index.H3Index, error) {
	if len(cells) == 0 {
		return nil, nil
	}

	cellSet := make(map[h3index.H3Index]struct{}, len(cells))
	for _, c := range cells {
		if !h3index.IsValid(c) {
			return nil, ErrCellInvalid
		}
		cellSet[c] = struct{}{}
	}

	// Iteratively compact from finest to coarsest resolution
	changed := true
	for changed {
		changed = false
		parentCounts := make(map[h3index.H3Index]int)
		parentRes := make(map[h3index.H3Index]int)

		// Count children for each parent
		for h := range cellSet {
			res := h3index.H3Index(h).Resolution()
			if res > 0 {
				parent := getParent(h, res-1)
				parentCounts[parent]++
				parentRes[parent] = res - 1
			}
		}

		// Replace full sets of children with their parent
		for parent, count := range parentCounts {
			expectedChildren := getChildrenCount(parent)
			if count == expectedChildren {
				// Remove all children
				res := parentRes[parent]
				children := getChildren(parent, res+1)
				for _, child := range children {
					delete(cellSet, child)
				}
				// Add parent
				cellSet[parent] = struct{}{}
				changed = true
			}
		}
	}

	result := make([]h3index.H3Index, 0, len(cellSet))
	for h := range cellSet {
		result = append(result, h)
	}
	return result, nil
}

// UncompactCells expands a compacted set of cells to the given resolution.
func UncompactCells(cells []h3index.H3Index, res int) ([]h3index.H3Index, error) {
	if res < 0 || res > constants.MaxH3Res {
		return nil, ErrDomain
	}

	var result []h3index.H3Index
	for _, h := range cells {
		if !h3index.IsValid(h) {
			return nil, ErrCellInvalid
		}

		cellRes := h3index.H3Index(h).Resolution()
		if cellRes > res {
			return nil, fmt.Errorf("cell resolution %d > target %d", cellRes, res)
		}
		if cellRes == res {
			result = append(result, h)
			continue
		}

		// Get all children at target resolution
		children := getChildren(h, res)
		result = append(result, children...)
	}
	return result, nil
}

// UncompactCellsSize returns the number of cells after uncompacting.
func UncompactCellsSize(cells []h3index.H3Index, res int) (int64, error) {
	if res < 0 || res > constants.MaxH3Res {
		return 0, ErrDomain
	}

	var count int64
	for _, h := range cells {
		if !h3index.IsValid(h) {
			return 0, ErrCellInvalid
		}

		cellRes := h3index.H3Index(h).Resolution()
		if cellRes > res {
			return 0, fmt.Errorf("cell resolution %d > target %d", cellRes, res)
		}
		if cellRes == res {
			count++
			continue
		}

		// Count children at target resolution
		resDiff := res - cellRes
		childCount := int64(1)
		for i := 0; i < resDiff; i++ {
			if h3index.IsPentagon(h) && i == 0 {
				childCount *= 6 // Pentagons have 6 children (one center + 5 neighbors)
			} else {
				childCount *= 7 // Regular hexagons have 7 children
			}
		}
		count += childCount
	}
	return count, nil
}

// Helper functions

// H3NeighborRotations returns the hexagon index neighboring the origin, in the direction dir.
// It applies rotation tracking as cells cross face boundaries.
// This is the critical function for grid traversal.
// This is exported for use by the directed edge functions.
func H3NeighborRotations(origin h3index.H3Index, dir int, rotations *int, out *h3index.H3Index) error {
	return h3NeighborRotations(origin, dir, rotations, out)
}

// h3NeighborRotations is the internal implementation.
func h3NeighborRotations(origin h3index.H3Index, dir int, rotations *int, out *h3index.H3Index) error {
	current := origin

	if dir < constants.CENTER_DIGIT || dir >= constants.NUM_DIGITS {
		return ErrFailed
	}

	// Ensure rotations is modulo 6
	*rotations = *rotations % 6
	for i := 0; i < *rotations; i++ {
		dir = rotate60ccw(dir)
	}

	newRotations := 0
	oldBaseCell := current.BaseCell()
	if oldBaseCell < 0 || oldBaseCell >= constants.NumBaseCells {
		return ErrCellInvalid
	}

	oldLeadingDigit := leadingNonZeroDigit(current)

	// Adjust the indexing digits and, if needed, the base cell
	r := current.Resolution() - 1

	for {
		if r == -1 {
			// Base cell transition
			newBaseCell := h3index.BaseCellNeighbor(oldBaseCell, dir)
			newRotations = h3index.BaseCellNeighborRotations(oldBaseCell, dir)

			if newBaseCell == h3index.INVALID_BASE_CELL {
				// Adjust for the deleted k vertex at the base cell level
				newBaseCell = h3index.BaseCellNeighbor(oldBaseCell, constants.IK_AXES_DIGIT)
				newRotations = h3index.BaseCellNeighborRotations(oldBaseCell, constants.IK_AXES_DIGIT)

				// Perform the adjustment for the k-subsequence we're skipping over
				current = rotate60ccwIndex(current)
				*rotations = *rotations + 1
			}

			current = h3index.SetBaseCell(current, newBaseCell)
			break
		} else {
			// IndexDigit/SetIndexDigit use 0-indexed resolution (0=coarsest digit).
			// C uses 1-indexed: C position r+1 = Go index r.
			oldDigit := current.IndexDigit(r)
			var nextDir int

			if oldDigit == constants.INVALID_DIGIT {
				return ErrCellInvalid
			}

			// Note: In C code, isResolutionClassIII checks resolution level r+1.
			// Class II: even resolution, Class III: odd resolution.
			// Class III (odd) uses NEW_DIGIT_II tables (per C source).
			if isResClassIII(r + 1) {
				// For Class III resolution, use Class II tables (per C source algos.c)
				current = h3index.SetIndexDigit(current, r, NEW_DIGIT_II[oldDigit][dir])
				nextDir = NEW_ADJUSTMENT_II[oldDigit][dir]
			} else {
				// For Class II resolution, use Class III tables (per C source algos.c)
				current = h3index.SetIndexDigit(current, r, NEW_DIGIT_III[oldDigit][dir])
				nextDir = NEW_ADJUSTMENT_III[oldDigit][dir]
			}

			if nextDir != constants.CENTER_DIGIT {
				dir = nextDir
				r--
			} else {
				// No more adjustment to perform
				break
			}
		}
	}

	newBaseCell := current.BaseCell()
	if h3index.IsBaseCellPentagon(newBaseCell) {
		alreadyAdjustedKSubsequence := false

		// Force rotation out of missing k-axes sub-sequence
		if leadingNonZeroDigit(current) == constants.K_AXES_DIGIT {
			if oldBaseCell != newBaseCell {
				// Traversed into the deleted k subsequence of a pentagon base cell
				// Rotate out depending on face orientation
				if baseCellIsCwOffset(newBaseCell, getBaseCellHomeFace(oldBaseCell)) {
					current = rotate60cwIndex(current)
				} else {
					current = rotate60ccwIndex(current)
				}
				alreadyAdjustedKSubsequence = true
			} else {
				// Traversed into deleted k subsequence from within same pentagon
				if oldLeadingDigit == constants.CENTER_DIGIT {
					// Undefined: the k direction is deleted from here
					return ErrPentagon
				} else if oldLeadingDigit == constants.JK_AXES_DIGIT {
					current = rotate60ccwIndex(current)
					*rotations = *rotations + 1
				} else if oldLeadingDigit == constants.IK_AXES_DIGIT {
					current = rotate60cwIndex(current)
					*rotations = *rotations + 5
				} else {
					return ErrFailed
				}
			}
		}

		// Apply pentagon rotations
		for i := 0; i < newRotations; i++ {
			current = rotatePent60ccw(current)
		}

		// Account for differing orientation of base cells
		if oldBaseCell != newBaseCell {
			if isBaseCellPolarPentagon(newBaseCell) {
				// Polar base cells behave differently
				if oldBaseCell != 118 && oldBaseCell != 8 &&
					leadingNonZeroDigit(current) != constants.JK_AXES_DIGIT {
					*rotations = *rotations + 1
				}
			} else if leadingNonZeroDigit(current) == constants.IK_AXES_DIGIT &&
				!alreadyAdjustedKSubsequence {
				*rotations = *rotations + 1
			}
		}
	} else {
		// Regular hexagon: apply standard rotations
		for i := 0; i < newRotations; i++ {
			current = rotate60ccwIndex(current)
		}
	}

	*rotations = (*rotations + newRotations) % 6
	*out = current
	return nil
}

// getNeighbors returns all immediate neighbors of the given cell.
func getNeighbors(h h3index.H3Index) []h3index.H3Index {
	isPent := h3index.IsPentagon(h)
	var neighbors []h3index.H3Index

	// Start from K_AXES_DIGIT for hexagons, J_AXES_DIGIT for pentagons
	startDir := constants.K_AXES_DIGIT
	if isPent {
		startDir = constants.J_AXES_DIGIT
	}

	for dir := startDir; dir < constants.NUM_DIGITS; dir++ {
		rotations := 0
		var neighbor h3index.H3Index
		err := h3NeighborRotations(h, dir, &rotations, &neighbor)
		if err == nil {
			neighbors = append(neighbors, neighbor)
		}
	}

	return neighbors
}

// gridDiskSafe is a BFS fallback for pentagons.
func gridDiskSafe(origin h3index.H3Index, k int) ([]h3index.H3Index, error) {
	if k < 0 {
		return nil, ErrDomain
	}

	seen := make(map[h3index.H3Index]struct{})
	queue := []h3index.H3Index{origin}
	seen[origin] = struct{}{}

	for step := 0; step < k; step++ {
		var next []h3index.H3Index
		for _, h := range queue {
			neighbors := getNeighbors(h)
			for _, n := range neighbors {
				if n != h3index.H3_NULL {
					if _, exists := seen[n]; !exists {
						seen[n] = struct{}{}
						next = append(next, n)
					}
				}
			}
		}
		queue = next
	}

	result := make([]h3index.H3Index, 0, len(seen))
	for h := range seen {
		result = append(result, h)
	}
	return result, nil
}

// gridDiskDistancesSafe is a BFS fallback with distances.
func gridDiskDistancesSafe(origin h3index.H3Index, k int) ([][]h3index.H3Index, error) {
	if k < 0 {
		return nil, ErrDomain
	}

	result := make([][]h3index.H3Index, k+1)
	seen := make(map[h3index.H3Index]struct{})
	queue := []h3index.H3Index{origin}
	seen[origin] = struct{}{}
	result[0] = []h3index.H3Index{origin}

	for step := 0; step < k; step++ {
		var next []h3index.H3Index
		for _, h := range queue {
			neighbors := getNeighbors(h)
			for _, n := range neighbors {
				if n != h3index.H3_NULL {
					if _, exists := seen[n]; !exists {
						seen[n] = struct{}{}
						next = append(next, n)
						result[step+1] = append(result[step+1], n)
					}
				}
			}
		}
		queue = next
	}

	return result, nil
}

// rotate60ccw rotates a direction 60 degrees counter-clockwise.
func rotate60ccw(dir int) int {
	switch dir {
	case constants.K_AXES_DIGIT:
		return constants.IK_AXES_DIGIT
	case constants.IK_AXES_DIGIT:
		return constants.I_AXES_DIGIT
	case constants.I_AXES_DIGIT:
		return constants.IJ_AXES_DIGIT
	case constants.IJ_AXES_DIGIT:
		return constants.J_AXES_DIGIT
	case constants.J_AXES_DIGIT:
		return constants.JK_AXES_DIGIT
	case constants.JK_AXES_DIGIT:
		return constants.K_AXES_DIGIT
	default:
		return dir
	}
}

// rotate60ccwIndex rotates all digits of an H3Index 60 degrees counter-clockwise.
func rotate60ccwIndex(h h3index.H3Index) h3index.H3Index {
	res := h.Resolution()
	for r := 0; r < res; r++ {
		digit := h.IndexDigit(r)
		h = h3index.SetIndexDigit(h, r, rotate60ccw(digit))
	}
	return h
}

// rotate60cwIndex rotates all digits of an H3Index 60 degrees clockwise.
func rotate60cwIndex(h h3index.H3Index) h3index.H3Index {
	// Rotate clockwise 5 times is the same as counter-clockwise once
	for i := 0; i < 5; i++ {
		h = rotate60ccwIndex(h)
	}
	return h
}

// rotatePent60ccw rotates a pentagon H3Index 60 degrees counter-clockwise.
// Pentagon rotations skip the deleted K subsequence.
func rotatePent60ccw(h h3index.H3Index) h3index.H3Index {
	res := h.Resolution()
	for r := 0; r < res; r++ {
		digit := h.IndexDigit(r)
		// Skip K_AXES_DIGIT for pentagons
		newDigit := rotate60ccw(digit)
		if newDigit == constants.K_AXES_DIGIT {
			newDigit = rotate60ccw(newDigit)
		}
		h = h3index.SetIndexDigit(h, r, newDigit)
	}
	return h
}

// leadingNonZeroDigit returns the first non-zero digit of an H3Index.
func leadingNonZeroDigit(h h3index.H3Index) int {
	res := h.Resolution()
	for r := 0; r < res; r++ {
		digit := h.IndexDigit(r)
		if digit != constants.CENTER_DIGIT {
			return digit
		}
	}
	return constants.CENTER_DIGIT
}

// isResClassIII returns true if the resolution is Class III (odd).
func isResClassIII(res int) bool {
	return (res % 2) == 1
}

// getParent returns the parent cell at the given resolution.
func getParent(h h3index.H3Index, res int) h3index.H3Index {
	if res < 0 || res >= h.Resolution() {
		return h
	}

	parent := h
	parent = h3index.SetResolution(parent, res)

	// Set all digits beyond res to INVALID_DIGIT
	for r := res; r < constants.MaxH3Res; r++ {
		parent = h3index.SetIndexDigit(parent, r, constants.INVALID_DIGIT)
	}

	return parent
}

// getChildren returns all children of a cell at the target resolution.
func getChildren(h h3index.H3Index, targetRes int) []h3index.H3Index {
	currentRes := h.Resolution()
	if targetRes <= currentRes {
		return []h3index.H3Index{h}
	}

	// Start with the parent
	children := []h3index.H3Index{h}

	// Iteratively generate children for each resolution level
	for res := currentRes; res < targetRes; res++ {
		var nextLevel []h3index.H3Index
		for _, parent := range children {
			// Generate children for this parent
			childDigits := []int{constants.CENTER_DIGIT, constants.K_AXES_DIGIT, constants.J_AXES_DIGIT,
				constants.JK_AXES_DIGIT, constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT}

			// Pentagons don't have K_AXES_DIGIT children
			if h3index.IsPentagon(parent) && res == currentRes {
				childDigits = []int{constants.CENTER_DIGIT, constants.J_AXES_DIGIT, constants.JK_AXES_DIGIT,
					constants.I_AXES_DIGIT, constants.IK_AXES_DIGIT, constants.IJ_AXES_DIGIT}
			}

			for _, digit := range childDigits {
				child := parent
				child = h3index.SetResolution(child, res+1)
				child = h3index.SetIndexDigit(child, res, digit)
				nextLevel = append(nextLevel, child)
			}
		}
		children = nextLevel
	}

	return children
}

// getChildrenCount returns the number of children a cell has at the next resolution.
func getChildrenCount(h h3index.H3Index) int {
	if h3index.IsPentagon(h) {
		return 6 // Pentagons have 6 children (no K_AXES_DIGIT)
	}
	return 7 // Hexagons have 7 children
}

// baseCellIsCwOffset returns true if the base cell has a clockwise offset on the given face.
// This is a simplified implementation; the full version requires baseCellData.
func baseCellIsCwOffset(baseCell int, face int) bool {
	// Simplified: assume no CW offset for now
	// Full implementation would check baseCellData[baseCell].cwOffsetPent
	return false
}

// getBaseCellHomeFace returns the home face of a base cell.
// This is a simplified implementation.
func getBaseCellHomeFace(baseCell int) int {
	// Simplified: return base cell mod 20 (number of icosahedron faces)
	return baseCell % 20
}

// isBaseCellPolarPentagon returns true if the base cell is a polar pentagon.
func isBaseCellPolarPentagon(baseCell int) bool {
	// Polar pentagons are base cells 4 and 117
	return baseCell == 4 || baseCell == 117
}
