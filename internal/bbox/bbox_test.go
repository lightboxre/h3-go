package bbox

import (
	"math"
	"testing"
)

const epsilon = 1e-10

func TestBBoxIsTransmeridian(t *testing.T) {
	tests := []struct {
		name string
		bbox BBox
		want bool
	}{
		{
			name: "standard bbox",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			want: false,
		},
		{
			name: "transmeridian bbox",
			bbox: BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
			want: true,
		},
		{
			name: "zero bbox",
			bbox: BBox{North: 0, South: 0, East: 0, West: 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BBoxIsTransmeridian(tt.bbox); got != tt.want {
				t.Errorf("BBoxIsTransmeridian() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBBoxWidth(t *testing.T) {
	tests := []struct {
		name string
		bbox BBox
		want float64
	}{
		{
			name: "standard bbox",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			want: 2.0,
		},
		{
			name: "transmeridian bbox",
			bbox: BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
			want: 2*math.Pi - 6.0, // 2π + (East - West) = 2π + (-3 - 3) = 2π - 6
		},
		{
			name: "zero width bbox",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: 1.0},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BBoxWidth(tt.bbox); math.Abs(got-tt.want) > epsilon {
				t.Errorf("BBoxWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBBoxHeight(t *testing.T) {
	tests := []struct {
		name string
		bbox BBox
		want float64
	}{
		{
			name: "standard bbox",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			want: 1.0,
		},
		{
			name: "zero height bbox",
			bbox: BBox{North: 0.5, South: 0.5, East: 1.0, West: -1.0},
			want: 0.0,
		},
		{
			name: "large height bbox",
			bbox: BBox{North: math.Pi / 2, South: -math.Pi / 2, East: 1.0, West: -1.0},
			want: math.Pi,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BBoxHeight(tt.bbox); math.Abs(got-tt.want) > epsilon {
				t.Errorf("BBoxHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBBoxCenter(t *testing.T) {
	tests := []struct {
		name    string
		bbox    BBox
		wantLat float64
		wantLng float64
	}{
		{
			name:    "standard bbox",
			bbox:    BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			wantLat: 0.0,
			wantLng: 0.0,
		},
		{
			name:    "transmeridian bbox",
			bbox:    BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
			wantLat: 0.0,
			wantLng: (3.0 + (-3.0 + 2*math.Pi)) / 2.0, // ((West + (East + 2π)) / 2
		},
		{
			name:    "zero bbox",
			bbox:    BBox{North: 0, South: 0, East: 0, West: 0},
			wantLat: 0.0,
			wantLng: 0.0,
		},
		{
			name:    "offset bbox",
			bbox:    BBox{North: 1.0, South: 0.0, East: 2.0, West: 1.0},
			wantLat: 0.5,
			wantLng: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLat, gotLng := BBoxCenter(tt.bbox)
			if math.Abs(gotLat-tt.wantLat) > epsilon {
				t.Errorf("BBoxCenter() lat = %v, want %v", gotLat, tt.wantLat)
			}
			if math.Abs(gotLng-tt.wantLng) > epsilon {
				t.Errorf("BBoxCenter() lng = %v, want %v", gotLng, tt.wantLng)
			}
		})
	}
}

func TestBBoxContains(t *testing.T) {
	tests := []struct {
		name string
		bbox BBox
		lat  float64
		lng  float64
		want bool
	}{
		{
			name: "standard bbox - center point",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			lat:  0.0,
			lng:  0.0,
			want: true,
		},
		{
			name: "standard bbox - outside north",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			lat:  1.0,
			lng:  0.0,
			want: false,
		},
		{
			name: "standard bbox - outside east",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			lat:  0.0,
			lng:  2.0,
			want: false,
		},
		{
			name: "standard bbox - on boundary north",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			lat:  0.5,
			lng:  0.0,
			want: true,
		},
		{
			name: "standard bbox - on boundary east",
			bbox: BBox{North: 0.5, South: -0.5, East: 1.0, West: -1.0},
			lat:  0.0,
			lng:  1.0,
			want: true,
		},
		{
			name: "transmeridian bbox - west side",
			bbox: BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
			lat:  0.0,
			lng:  3.1,
			want: true,
		},
		{
			name: "transmeridian bbox - east side",
			bbox: BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
			lat:  0.0,
			lng:  -3.1,
			want: true,
		},
		{
			name: "transmeridian bbox - middle (outside)",
			bbox: BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
			lat:  0.0,
			lng:  0.0,
			want: false,
		},
		{
			name: "transmeridian bbox - on west boundary",
			bbox: BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
			lat:  0.0,
			lng:  3.0,
			want: true,
		},
		{
			name: "transmeridian bbox - on east boundary",
			bbox: BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
			lat:  0.0,
			lng:  -3.0,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BBoxContains(tt.bbox, tt.lat, tt.lng); got != tt.want {
				t.Errorf("BBoxContains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBBoxOverlap(t *testing.T) {
	tests := []struct {
		name string
		a    BBox
		b    BBox
		want bool
	}{
		{
			name: "two overlapping standard bboxes",
			a:    BBox{North: 1.0, South: 0.0, East: 1.0, West: 0.0},
			b:    BBox{North: 2.0, South: 0.5, East: 2.0, West: 0.5},
			want: true,
		},
		{
			name: "two non-overlapping standard bboxes (latitude)",
			a:    BBox{North: 1.0, South: 0.0, East: 1.0, West: 0.0},
			b:    BBox{North: -0.5, South: -1.0, East: 1.0, West: 0.0},
			want: false,
		},
		{
			name: "two non-overlapping standard bboxes (longitude)",
			a:    BBox{North: 1.0, South: 0.0, East: 1.0, West: 0.0},
			b:    BBox{North: 1.0, South: 0.0, East: 3.0, West: 2.0},
			want: false,
		},
		{
			name: "identical bboxes",
			a:    BBox{North: 1.0, South: 0.0, East: 1.0, West: 0.0},
			b:    BBox{North: 1.0, South: 0.0, East: 1.0, West: 0.0},
			want: true,
		},
		{
			name: "transmeridian and standard overlap",
			a:    BBox{North: 1.0, South: 0.0, East: -3.0, West: 3.0},
			b:    BBox{North: 1.0, South: 0.0, East: -2.5, West: -3.5},
			want: true,
		},
		{
			name: "two transmeridian bboxes",
			a:    BBox{North: 1.0, South: 0.0, East: -3.0, West: 3.0},
			b:    BBox{North: 1.0, South: 0.0, East: -2.0, West: 2.5},
			want: true,
		},
		{
			name: "transmeridian and standard no overlap",
			a:    BBox{North: 1.0, South: 0.0, East: -3.0, West: 3.0},
			b:    BBox{North: 1.0, South: 0.0, East: 1.0, West: 0.0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BBoxOverlap(tt.a, tt.b); got != tt.want {
				t.Errorf("BBoxOverlap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBBoxFromGeoLoop(t *testing.T) {
	tests := []struct {
		name string
		lats []float64
		lngs []float64
		want BBox
	}{
		{
			name: "empty geoloop",
			lats: []float64{},
			lngs: []float64{},
			want: BBox{},
		},
		{
			name: "single point",
			lats: []float64{0.5},
			lngs: []float64{1.0},
			want: BBox{North: 0.5, South: 0.5, East: 1.0, West: 1.0},
		},
		{
			name: "simple rectangle",
			lats: []float64{0.0, 1.0, 1.0, 0.0},
			lngs: []float64{0.0, 0.0, 1.0, 1.0},
			want: BBox{North: 1.0, South: 0.0, East: 1.0, West: 0.0},
		},
		{
			name: "transmeridian polygon",
			lats: []float64{0.0, 0.5, 0.0, -0.5},
			lngs: []float64{3.0, math.Pi, -3.0, -math.Pi + 0.5},
			want: BBox{North: 0.5, South: -0.5, East: -3.0, West: 3.0},
		},
		{
			name: "polygon crossing antimeridian",
			lats: []float64{0.0, 0.0, 0.5, 0.5},
			lngs: []float64{math.Pi - 0.1, -math.Pi + 0.1, -math.Pi + 0.1, math.Pi - 0.1},
			want: BBox{North: 0.5, South: 0.0, East: -math.Pi + 0.1, West: math.Pi - 0.1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BBoxFromGeoLoop(tt.lats, tt.lngs)
			if math.Abs(got.North-tt.want.North) > epsilon {
				t.Errorf("BBoxFromGeoLoop() North = %v, want %v", got.North, tt.want.North)
			}
			if math.Abs(got.South-tt.want.South) > epsilon {
				t.Errorf("BBoxFromGeoLoop() South = %v, want %v", got.South, tt.want.South)
			}
			if math.Abs(got.East-tt.want.East) > epsilon {
				t.Errorf("BBoxFromGeoLoop() East = %v, want %v", got.East, tt.want.East)
			}
			if math.Abs(got.West-tt.want.West) > epsilon {
				t.Errorf("BBoxFromGeoLoop() West = %v, want %v", got.West, tt.want.West)
			}
		})
	}
}

func TestBBoxFromGeoLoopTransmeridian(t *testing.T) {
	// Test case specifically for detecting transmeridian crossing
	lats := []float64{0.1, 0.1, -0.1, -0.1}
	lngs := []float64{3.0, -3.0, -3.0, 3.0}

	bbox := BBoxFromGeoLoop(lats, lngs)

	if !BBoxIsTransmeridian(bbox) {
		t.Error("Expected transmeridian bbox")
	}

	// Verify it contains points on both sides
	if !BBoxContains(bbox, 0.0, 3.1) {
		t.Error("Expected bbox to contain point on west side")
	}
	if !BBoxContains(bbox, 0.0, -3.1) {
		t.Error("Expected bbox to contain point on east side")
	}
	if BBoxContains(bbox, 0.0, 0.0) {
		t.Error("Expected bbox to NOT contain point in middle")
	}
}
