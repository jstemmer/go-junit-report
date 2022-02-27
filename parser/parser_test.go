package parser

import (
	"strings"
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

func TestReportPrefixPackage(t *testing.T) {
	const prefix = "linux/amd64/"
	originalReport := &Report{
		Packages: []Package{
			{Name: "pkg/a"},
			{Name: "pkg/b"},
			{Name: "pkg/c"},
		},
	}

	prefixedReport := originalReport.PrefixPackage(prefix)

	for i, pkg := range originalReport.Packages {
		if strings.HasPrefix(pkg.Name, prefix) {
			t.Errorf("expecting original report package %q not to be modified", pkg.Name)
		}
		if !strings.HasPrefix(prefixedReport.Packages[i].Name, prefix) {
			t.Errorf("expecting prefixed report package %q to be prefixed with %q", pkg.Name, prefix)
		}
	}
}
