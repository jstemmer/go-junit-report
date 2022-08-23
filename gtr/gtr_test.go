package gtr

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTrimPrefixSpaces(t *testing.T) {
	tests := []struct {
		input  string
		indent int
		want   string
	}{
		{"", 0, ""},
		{"\ttest", 0, "test"}, // prefixed tabs are always trimmed
		{"    test", 0, "test"},
		{"        test", 0, "    test"},
		{"            test", 2, "test"},
		{"      test", 1, "      test"}, // prefix is not a multiple of 4
		{"    \t    test", 3, "    test"},
	}

	for _, test := range tests {
		got := TrimPrefixSpaces(test.input, test.indent)
		if got != test.want {
			t.Errorf("TrimPrefixSpaces(%q, %d) incorrect, got %q, want %q", test.input, test.indent, got, test.want)
		}
	}
}

func TestSetProperty(t *testing.T) {
	pkg := Package{}
	pkg.SetProperty("a", "b")
	pkg.SetProperty("c", "d")
	pkg.SetProperty("a", "e")

	want := []Property{{Name: "c", Value: "d"}, {Name: "a", Value: "e"}}
	if diff := cmp.Diff(want, pkg.Properties); diff != "" {
		t.Errorf("SetProperty got unexpected diff: %s", diff)
	}
}
