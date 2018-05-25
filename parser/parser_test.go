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

func TestParseNanoseconds(t *testing.T) {
	tests := []struct {
		in string
		d  time.Duration
	}{
		{"", 0},
		{"0.1", 0 * time.Nanosecond},
		{"0.9", 0 * time.Nanosecond},
		{"4", 4 * time.Nanosecond},
		{"5000", 5 * time.Microsecond},
		{"2000003", 2*time.Millisecond + 3*time.Nanosecond},
	}

	for _, test := range tests {
		d := parseNanoseconds(test.in)
		if d != test.d {
			t.Errorf("parseSeconds(%q) == %v, want %v\n", test.in, d, test.d)
		}
	}
}
