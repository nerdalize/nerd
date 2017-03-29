package nerd

import "testing"

func TestSemVer(t *testing.T) {
	testCases := []struct {
		big   string
		small string
	}{
		{
			big:   "2.0.0",
			small: "1.0.0",
		},
		{
			big:   "2.0.0",
			small: "1.3.3",
		},
		{
			big:   "0.2.0",
			small: "0.1.0",
		},
		{
			big:   "0.2.0",
			small: "0.1.2",
		},
		{
			big:   "0.0.2",
			small: "0.0.1",
		},
		{
			big:   "0.1.0",
			small: "0.0.1",
		},
		{
			big:   "2.0.0",
			small: "1.20.30",
		},
		{
			big:   "20.0.0",
			small: "19.2.3",
		},
	}
	for _, tc := range testCases {
		big, err := ParseSemVer(tc.big)
		if err != nil {
			t.Fatalf("Failed to parse semver %v (big)", tc.big)
		}
		small, err := ParseSemVer(tc.small)
		if err != nil {
			t.Fatalf("Failed to parse semver %v (small)", tc.small)
		}
		if !big.GreaterThan(small) {
			t.Errorf("Expected %v to be a larger semver than %v", tc.big, tc.small)
		}
	}
}

func TestSemVerFail(t *testing.T) {
	ver := "ver.ver.ver"
	_, err := ParseSemVer(ver)
	if err == nil {
		t.Errorf("Parsing %v should raise an error.", ver)
	}
}
