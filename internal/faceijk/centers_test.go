package faceijk

import (
	"fmt"
	"math"
	"testing"

	"github.com/lightboxre/h3-go/internal/h3index"
)

func TestComputeCellCenters(t *testing.T) {
	// Our corrected sfCell
	ourSFCell := h3index.H3Index(0x89283082877ffff)
	sfFijk := H3ToFaceIJK(ourSFCell)
	sfLat, sfLng := FaceIJKToGeo(sfFijk, 9)
	fmt.Printf("Our sfCell (0x89283082877ffff) center: lat=%.8f° lng=%.8f°\n",
		RadsToDegs(sfLat), RadsToDegs(sfLng))

	// googCell center
	googCell := h3index.H3Index(0x85283473fffffff)
	gFijk := H3ToFaceIJK(googCell)
	gLat, gLng := FaceIJKToGeo(gFijk, 5)
	fmt.Printf("googCell (0x85283473fffffff) center: lat=%.8f° lng=%.8f°\n",
		RadsToDegs(gLat), RadsToDegs(gLng))

	// Distance from googleplex input to googCell center
	googInputLat := 37.3615593 * math.Pi / 180.0
	googInputLng := -122.0553238 * math.Pi / 180.0
	dist := GreatCircleDistanceKm(googInputLat, googInputLng, gLat, gLng)
	fmt.Printf("Distance from googleplex input to googCell center: %.1f km\n", dist)
}
