//go:build go1.18
// +build go1.18

package gotest

import "testing"

func FuzzParseLine(f *testing.F) {
	for _, test := range parseLineTests {
		f.Add(test.input)
	}
	f.Fuzz(func(t *testing.T, in string) {
		events := NewParser().parseLine(in)
		if len(events) == 0 {
			t.Fatalf("parseLine(%q) did not return any results", in)
		}
	})
}
