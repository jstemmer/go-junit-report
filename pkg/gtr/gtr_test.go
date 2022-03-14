package gtr

import "testing"

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
