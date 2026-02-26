// Package faceijk - base cell home face-IJK lookup table.
// Transcribed from C: src/h3lib/lib/baseCells.c (baseCellData[]).
package faceijk

// baseCellFIJK holds the home icosahedron face and IJK coordinates at
// resolution 0 for a single base cell.
type baseCellFIJK struct {
	Face, I, J, K int
}

// baseCellHomeFIJK maps base cell number (0-121) to its home face-IJK at res 0.
// Data from C: baseCellData[NUM_BASE_CELLS].homeFijk
var baseCellHomeFIJK = [122]baseCellFIJK{
	{1, 1, 0, 0},  // base cell 0
	{2, 1, 1, 0},  // base cell 1
	{1, 0, 0, 0},  // base cell 2
	{2, 1, 0, 0},  // base cell 3
	{0, 2, 0, 0},  // base cell 4 (pentagon)
	{1, 1, 1, 0},  // base cell 5
	{1, 0, 0, 1},  // base cell 6
	{2, 0, 0, 0},  // base cell 7
	{0, 1, 0, 0},  // base cell 8
	{2, 0, 1, 0},  // base cell 9
	{1, 0, 1, 0},  // base cell 10
	{1, 0, 1, 1},  // base cell 11
	{3, 1, 0, 0},  // base cell 12
	{3, 1, 1, 0},  // base cell 13
	{11, 2, 0, 0}, // base cell 14 (pentagon)
	{4, 1, 0, 0},  // base cell 15
	{0, 0, 0, 0},  // base cell 16
	{6, 0, 1, 0},  // base cell 17
	{0, 0, 0, 1},  // base cell 18
	{2, 0, 1, 1},  // base cell 19
	{7, 0, 0, 1},  // base cell 20
	{2, 0, 0, 1},  // base cell 21
	{0, 1, 1, 0},  // base cell 22
	{6, 0, 0, 1},  // base cell 23
	{10, 2, 0, 0}, // base cell 24 (pentagon)
	{6, 0, 0, 0},  // base cell 25
	{3, 0, 0, 0},  // base cell 26
	{11, 1, 0, 0}, // base cell 27
	{4, 1, 1, 0},  // base cell 28
	{3, 0, 1, 0},  // base cell 29
	{0, 0, 1, 1},  // base cell 30
	{4, 0, 0, 0},  // base cell 31
	{5, 0, 1, 0},  // base cell 32
	{0, 0, 1, 0},  // base cell 33
	{7, 0, 1, 0},  // base cell 34
	{11, 1, 1, 0}, // base cell 35
	{7, 0, 0, 0},  // base cell 36
	{10, 1, 0, 0}, // base cell 37
	{12, 2, 0, 0}, // base cell 38 (pentagon)
	{6, 1, 0, 1},  // base cell 39
	{7, 1, 0, 1},  // base cell 40
	{4, 0, 0, 1},  // base cell 41
	{3, 0, 0, 1},  // base cell 42
	{3, 0, 1, 1},  // base cell 43
	{4, 0, 1, 0},  // base cell 44
	{6, 1, 0, 0},  // base cell 45
	{11, 0, 0, 0}, // base cell 46
	{8, 0, 0, 1},  // base cell 47
	{5, 0, 0, 1},  // base cell 48
	{14, 2, 0, 0}, // base cell 49 (pentagon)
	{5, 0, 0, 0},  // base cell 50
	{12, 1, 0, 0}, // base cell 51
	{10, 1, 1, 0}, // base cell 52
	{4, 0, 1, 1},  // base cell 53
	{12, 1, 1, 0}, // base cell 54
	{7, 1, 0, 0},  // base cell 55
	{11, 0, 1, 0}, // base cell 56
	{10, 0, 0, 0}, // base cell 57
	{13, 2, 0, 0}, // base cell 58 (pentagon)
	{10, 0, 0, 1}, // base cell 59
	{11, 0, 0, 1}, // base cell 60
	{9, 0, 1, 0},  // base cell 61
	{8, 0, 1, 0},  // base cell 62
	{6, 2, 0, 0},  // base cell 63 (pentagon)
	{8, 0, 0, 0},  // base cell 64
	{9, 0, 0, 1},  // base cell 65
	{14, 1, 0, 0}, // base cell 66
	{5, 1, 0, 1},  // base cell 67
	{16, 0, 1, 1}, // base cell 68
	{8, 1, 0, 1},  // base cell 69
	{5, 1, 0, 0},  // base cell 70
	{12, 0, 0, 0}, // base cell 71
	{7, 2, 0, 0},  // base cell 72 (pentagon)
	{12, 0, 1, 0}, // base cell 73
	{10, 0, 1, 0}, // base cell 74
	{9, 0, 0, 0},  // base cell 75
	{13, 1, 0, 0}, // base cell 76
	{16, 0, 0, 1}, // base cell 77
	{15, 0, 1, 1}, // base cell 78
	{15, 0, 1, 0}, // base cell 79
	{16, 0, 1, 0}, // base cell 80
	{14, 1, 1, 0}, // base cell 81
	{13, 1, 1, 0}, // base cell 82
	{5, 2, 0, 0},  // base cell 83 (pentagon)
	{8, 1, 0, 0},  // base cell 84
	{14, 0, 0, 0}, // base cell 85
	{9, 1, 0, 1},  // base cell 86
	{14, 0, 0, 1}, // base cell 87
	{17, 0, 0, 1}, // base cell 88
	{12, 0, 0, 1}, // base cell 89
	{16, 0, 0, 0}, // base cell 90
	{17, 0, 1, 1}, // base cell 91
	{15, 0, 0, 1}, // base cell 92
	{16, 1, 0, 1}, // base cell 93
	{9, 1, 0, 0},  // base cell 94
	{15, 0, 0, 0}, // base cell 95
	{13, 0, 0, 0}, // base cell 96
	{8, 2, 0, 0},  // base cell 97 (pentagon)
	{13, 0, 1, 0}, // base cell 98
	{17, 1, 0, 1}, // base cell 99
	{19, 0, 1, 0}, // base cell 100
	{14, 0, 1, 0}, // base cell 101
	{19, 0, 1, 1}, // base cell 102
	{17, 0, 1, 0}, // base cell 103
	{13, 0, 0, 1}, // base cell 104
	{17, 0, 0, 0}, // base cell 105
	{16, 1, 0, 0}, // base cell 106
	{9, 2, 0, 0},  // base cell 107 (pentagon)
	{15, 1, 0, 1}, // base cell 108
	{15, 1, 0, 0}, // base cell 109
	{18, 0, 1, 1}, // base cell 110
	{18, 0, 0, 1}, // base cell 111
	{19, 0, 0, 1}, // base cell 112
	{17, 1, 0, 0}, // base cell 113
	{19, 0, 0, 0}, // base cell 114
	{18, 0, 1, 0}, // base cell 115
	{18, 1, 0, 1}, // base cell 116
	{19, 2, 0, 0}, // base cell 117 (pentagon)
	{19, 1, 0, 0}, // base cell 118
	{18, 0, 0, 0}, // base cell 119
	{19, 1, 0, 1}, // base cell 120
	{18, 1, 0, 0}, // base cell 121
}
