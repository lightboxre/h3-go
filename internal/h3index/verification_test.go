package h3index

import "testing"

// TestRealWorldH3Indices tests with actual H3 indices that can be verified
// against the official H3 library or online tools.
func TestRealWorldH3Indices(t *testing.T) {
	tests := []struct {
		name     string
		hexStr   string
		res      int
		bc       int
		valid    bool
		pentagon bool
	}{
		{
			name:     "San Francisco res 9",
			hexStr:   "8928308280fffff",
			res:      9,
			bc:       20,
			valid:    true,
			pentagon: false,
		},
		{
			name:     "Seattle res 7",
			hexStr:   "872830828ffffff",
			res:      7,
			bc:       20,
			valid:    true,
			pentagon: false,
		},
		{
			name:     "Pentagon res 0 base cell 4",
			hexStr:   "8009fffffffffff",
			res:      0,
			bc:       4,
			valid:    true,
			pentagon: true,
		},
		{
			name:     "Pentagon res 0 base cell 14",
			hexStr:   "801dfffffffffff",
			res:      0,
			bc:       14,
			valid:    true,
			pentagon: true,
		},
		{
			name:     "Invalid - bad mode",
			hexStr:   "0000000000000001",
			res:      0,
			bc:       0,
			valid:    false,
			pentagon: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, err := FromString(tt.hexStr)
			if err != nil {
				t.Fatalf("FromString error: %v", err)
			}

			if tt.valid {
				if !IsValid(h) {
					t.Errorf("%s: expected valid, got invalid", tt.name)
				}

				if h.Resolution() != tt.res {
					t.Errorf("%s: resolution = %d, want %d", tt.name, h.Resolution(), tt.res)
				}

				if h.BaseCell() != tt.bc {
					t.Errorf("%s: base cell = %d, want %d", tt.name, h.BaseCell(), tt.bc)
				}

				if IsPentagon(h) != tt.pentagon {
					t.Errorf("%s: pentagon = %v, want %v", tt.name, IsPentagon(h), tt.pentagon)
				}
			} else {
				if IsValid(h) {
					t.Errorf("%s: expected invalid, got valid", tt.name)
				}
			}
		})
	}
}
