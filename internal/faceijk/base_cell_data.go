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

// baseCellCWOffsetPent maps base cell number to the two clockwise-offset faces
// used by pentagon rotation handling. Non-pentagons use {0, 0}.
// Data from C: baseCellData[NUM_BASE_CELLS].cwOffsetPent
var baseCellCWOffsetPent = [122][2]int{
	{0, 0},   // base cell 0
	{0, 0},   // base cell 1
	{0, 0},   // base cell 2
	{0, 0},   // base cell 3
	{-1, -1}, // base cell 4 (pentagon)
	{0, 0},   // base cell 5
	{0, 0},   // base cell 6
	{0, 0},   // base cell 7
	{0, 0},   // base cell 8
	{0, 0},   // base cell 9
	{0, 0},   // base cell 10
	{0, 0},   // base cell 11
	{0, 0},   // base cell 12
	{0, 0},   // base cell 13
	{2, 6},   // base cell 14 (pentagon)
	{0, 0},   // base cell 15
	{0, 0},   // base cell 16
	{0, 0},   // base cell 17
	{0, 0},   // base cell 18
	{0, 0},   // base cell 19
	{0, 0},   // base cell 20
	{0, 0},   // base cell 21
	{0, 0},   // base cell 22
	{0, 0},   // base cell 23
	{1, 5},   // base cell 24 (pentagon)
	{0, 0},   // base cell 25
	{0, 0},   // base cell 26
	{0, 0},   // base cell 27
	{0, 0},   // base cell 28
	{0, 0},   // base cell 29
	{0, 0},   // base cell 30
	{0, 0},   // base cell 31
	{0, 0},   // base cell 32
	{0, 0},   // base cell 33
	{0, 0},   // base cell 34
	{0, 0},   // base cell 35
	{0, 0},   // base cell 36
	{0, 0},   // base cell 37
	{3, 7},   // base cell 38 (pentagon)
	{0, 0},   // base cell 39
	{0, 0},   // base cell 40
	{0, 0},   // base cell 41
	{0, 0},   // base cell 42
	{0, 0},   // base cell 43
	{0, 0},   // base cell 44
	{0, 0},   // base cell 45
	{0, 0},   // base cell 46
	{0, 0},   // base cell 47
	{0, 0},   // base cell 48
	{0, 9},   // base cell 49 (pentagon)
	{0, 0},   // base cell 50
	{0, 0},   // base cell 51
	{0, 0},   // base cell 52
	{0, 0},   // base cell 53
	{0, 0},   // base cell 54
	{0, 0},   // base cell 55
	{0, 0},   // base cell 56
	{0, 0},   // base cell 57
	{4, 8},   // base cell 58 (pentagon)
	{0, 0},   // base cell 59
	{0, 0},   // base cell 60
	{0, 0},   // base cell 61
	{0, 0},   // base cell 62
	{11, 15}, // base cell 63 (pentagon)
	{0, 0},   // base cell 64
	{0, 0},   // base cell 65
	{0, 0},   // base cell 66
	{0, 0},   // base cell 67
	{0, 0},   // base cell 68
	{0, 0},   // base cell 69
	{0, 0},   // base cell 70
	{0, 0},   // base cell 71
	{12, 16}, // base cell 72 (pentagon)
	{0, 0},   // base cell 73
	{0, 0},   // base cell 74
	{0, 0},   // base cell 75
	{0, 0},   // base cell 76
	{0, 0},   // base cell 77
	{0, 0},   // base cell 78
	{0, 0},   // base cell 79
	{0, 0},   // base cell 80
	{0, 0},   // base cell 81
	{0, 0},   // base cell 82
	{10, 19}, // base cell 83 (pentagon)
	{0, 0},   // base cell 84
	{0, 0},   // base cell 85
	{0, 0},   // base cell 86
	{0, 0},   // base cell 87
	{0, 0},   // base cell 88
	{0, 0},   // base cell 89
	{0, 0},   // base cell 90
	{0, 0},   // base cell 91
	{0, 0},   // base cell 92
	{0, 0},   // base cell 93
	{0, 0},   // base cell 94
	{0, 0},   // base cell 95
	{0, 0},   // base cell 96
	{13, 17}, // base cell 97 (pentagon)
	{0, 0},   // base cell 98
	{0, 0},   // base cell 99
	{0, 0},   // base cell 100
	{0, 0},   // base cell 101
	{0, 0},   // base cell 102
	{0, 0},   // base cell 103
	{0, 0},   // base cell 104
	{0, 0},   // base cell 105
	{0, 0},   // base cell 106
	{14, 18}, // base cell 107 (pentagon)
	{0, 0},   // base cell 108
	{0, 0},   // base cell 109
	{0, 0},   // base cell 110
	{0, 0},   // base cell 111
	{0, 0},   // base cell 112
	{0, 0},   // base cell 113
	{0, 0},   // base cell 114
	{0, 0},   // base cell 115
	{0, 0},   // base cell 116
	{-1, -1}, // base cell 117 (pentagon)
	{0, 0},   // base cell 118
	{0, 0},   // base cell 119
	{0, 0},   // base cell 120
	{0, 0},   // base cell 121
}

func baseCellIsCwOffset(baseCell, testFace int) bool {
	offsets := baseCellCWOffsetPent[baseCell]
	return offsets[0] == testFace || offsets[1] == testFace
}
