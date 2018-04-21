package parser

import (
	"testing"
	"time"
)

func TestParseSeconds(t *testing.T) {
	tests := []struct {
		in string
		d  time.Duration
	}{
		{"", 0},
		{"4", 4 * time.Second},
		{"0.1", 100 * time.Millisecond},
		{"0.050", 50 * time.Millisecond},
		{"2.003", 2*time.Second + 3*time.Millisecond},
	}

	for _, test := range tests {
		d := parseSeconds(test.in)
		if d != test.d {
			t.Errorf("parseSeconds(%q) == %v, want %v\n", test.in, d, test.d)
		}
	}
}
