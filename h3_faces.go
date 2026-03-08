package h3

import (
	"github.com/lightboxre/h3-go/internal/faceijk"
	"github.com/lightboxre/h3-go/internal/h3index"
)

// GetIcosahedronFaces returns the icosahedron face IDs (0-19) that cell c overlaps.
// Returns sorted unique face IDs. Returns error for invalid cells.
//
// Pentagon cells always intersect 5 faces. Hexagon cells intersect 1 or 2 faces.
// For Class II pentagons (even resolutions), this function uses the first child
// pentagon instead, as Class II pentagons have all vertices on face edges.
func GetIcosahedronFaces(c Cell) ([]int, error) {
	return faceijk.GetCellFaces(h3index.H3Index(c))
}
