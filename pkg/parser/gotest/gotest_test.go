package gotest

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const testdataRoot = "../../../testdata/"

var tests = []struct {
	in       string
	expected []Event
}{
	{"01-pass.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestZ"},
			{Type: "end_test", Id: 1, Name: "TestZ", Result: "PASS", Duration: 60 * time.Millisecond},
			{Type: "run_test", Id: 2, Name: "TestA"},
			{Type: "end_test", Id: 2, Name: "TestA", Result: "PASS", Duration: 100 * time.Millisecond},
			{Type: "status", Result: "PASS"},
			{Type: "summary", Result: "ok", Name: "package/name", Duration: 160 * time.Millisecond},
		}},
	{"02-fail.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestOne"},
			{Type: "end_test", Id: 1, Name: "TestOne", Result: "FAIL", Duration: 20 * time.Millisecond},
			{Type: "output", Data: "file_test.go:11: Error message", Indent: 1},
			{Type: "output", Data: "file_test.go:11: Longer", Indent: 1},
			{Type: "output", Data: "error", Indent: 2},
			{Type: "output", Data: "message.", Indent: 2},
			{Type: "run_test", Id: 2, Name: "TestTwo"},
			{Type: "end_test", Id: 2, Name: "TestTwo", Result: "PASS", Duration: 130 * time.Millisecond},
			{Type: "status", Result: "FAIL"},
			{Type: "output", Data: "exit status 1", Indent: 0},
			{Type: "summary", Result: "FAIL", Name: "package/name", Duration: 151 * time.Millisecond},
		}},
}

func TestParse(t *testing.T) {
	for _, test := range tests {
		f, err := os.Open(testdataRoot + test.in)
		if err != nil {
			t.Errorf("error reading %s: %v", test.in, err)
			continue
		}
		actual, err := Parse(f)
		f.Close()
		if err != nil {
			t.Errorf("Parse(%s) error: %v", test.in, err)
			continue
		}

		if diff := cmp.Diff(actual, test.expected); diff != "" {
			t.Errorf("Parse %s returned unexpected events, diff (-got, +want):\n%v", test.in, diff)
		}
	}
}
