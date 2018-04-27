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
