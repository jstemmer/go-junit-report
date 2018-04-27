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
	{"03-skip.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestOne"},
			{Type: "end_test", Id: 1, Name: "TestOne", Result: "SKIP", Duration: 20 * time.Millisecond},
			{Type: "output", Data: "file_test.go:11: Skip message", Indent: 1},
			{Type: "run_test", Id: 2, Name: "TestTwo"},
			{Type: "end_test", Id: 2, Name: "TestTwo", Result: "PASS", Duration: 130 * time.Millisecond},
			{Type: "status", Result: "PASS"},
			{Type: "summary", Result: "ok", Name: "package/name", Duration: 150 * time.Millisecond},
		}},
	{"04-go_1_4.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestOne"},
			{Type: "end_test", Id: 1, Name: "TestOne", Result: "PASS", Duration: 60 * time.Millisecond},
			{Type: "run_test", Id: 2, Name: "TestTwo"},
			{Type: "end_test", Id: 2, Name: "TestTwo", Result: "PASS", Duration: 100 * time.Millisecond},
			{Type: "status", Result: "PASS"},
			{Type: "summary", Result: "ok", Name: "package/name", Duration: 160 * time.Millisecond},
		}},
	// Test 05 is skipped, because it was actually testing XML output
	{"06-mixed.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestOne"},
			{Type: "end_test", Id: 1, Name: "TestOne", Result: "PASS", Duration: 60 * time.Millisecond},
			{Type: "run_test", Id: 2, Name: "TestTwo"},
			{Type: "end_test", Id: 2, Name: "TestTwo", Result: "PASS", Duration: 100 * time.Millisecond},
			{Type: "status", Result: "PASS"},
			{Type: "summary", Result: "ok", Name: "package/name1", Duration: 160 * time.Millisecond},
			{Type: "run_test", Id: 3, Name: "TestOne"},
			{Type: "end_test", Id: 3, Name: "TestOne", Result: "FAIL", Duration: 20 * time.Millisecond},
			{Type: "output", Data: "file_test.go:11: Error message", Indent: 1},
			{Type: "output", Data: "file_test.go:11: Longer", Indent: 1},
			{Type: "output", Data: "error", Indent: 2},
			{Type: "output", Data: "message.", Indent: 2},
			{Type: "run_test", Id: 4, Name: "TestTwo"},
			{Type: "end_test", Id: 4, Name: "TestTwo", Result: "PASS", Duration: 130 * time.Millisecond},
			{Type: "status", Result: "FAIL"},
			{Type: "output", Data: "exit status 1", Indent: 0},
			{Type: "summary", Result: "FAIL", Name: "package/name2", Duration: 151 * time.Millisecond},
		}},
	{"07-compiled_test.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestOne"},
			{Type: "end_test", Id: 1, Name: "TestOne", Result: "PASS", Duration: 60 * time.Millisecond},
			{Type: "run_test", Id: 2, Name: "TestTwo"},
			{Type: "end_test", Id: 2, Name: "TestTwo", Result: "PASS", Duration: 100 * time.Millisecond},
			{Type: "status", Result: "PASS"},
		}},
	{"08-parallel.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestDoFoo"},
			{Type: "run_test", Id: 2, Name: "TestDoFoo2"},
			{Type: "end_test", Id: 1, Name: "TestDoFoo", Result: "PASS", Duration: 270 * time.Millisecond},
			{Type: "output", Data: "cov_test.go:10: DoFoo log 1", Indent: 1},
			{Type: "output", Data: "cov_test.go:10: DoFoo log 2", Indent: 1},
			{Type: "end_test", Id: 2, Name: "TestDoFoo2", Result: "PASS", Duration: 160 * time.Millisecond},
			{Type: "output", Data: "cov_test.go:21: DoFoo2 log 1", Indent: 1},
			{Type: "output", Data: "cov_test.go:21: DoFoo2 log 2", Indent: 1},
			{Type: "status", Result: "PASS"},
			{Type: "summary", Result: "ok", Name: "package/name", Duration: 440 * time.Millisecond},
		}},
	{"09-coverage.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestZ"},
			{Type: "end_test", Id: 1, Name: "TestZ", Result: "PASS", Duration: 60 * time.Millisecond},
			{Type: "run_test", Id: 2, Name: "TestA"},
			{Type: "end_test", Id: 2, Name: "TestA", Result: "PASS", Duration: 100 * time.Millisecond},
			{Type: "status", Result: "PASS"},
			{Type: "coverage", CovPct: 13.37},
			{Type: "summary", Result: "ok", Name: "package/name", Duration: 160 * time.Millisecond},
		}},
	{"10-multipkg-coverage.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestA"},
			{Type: "end_test", Id: 1, Name: "TestA", Result: "PASS", Duration: 100 * time.Millisecond},
			{Type: "run_test", Id: 2, Name: "TestB"},
			{Type: "end_test", Id: 2, Name: "TestB", Result: "PASS", Duration: 300 * time.Millisecond},
			{Type: "status", Result: "PASS"},
			{Type: "coverage", CovPct: 10},
			{Type: "summary", Result: "ok", Name: "package1/foo", Duration: 400 * time.Millisecond, CovPct: 10},
			{Type: "run_test", Id: 3, Name: "TestC"},
			{Type: "end_test", Id: 3, Name: "TestC", Result: "PASS", Duration: 4200 * time.Millisecond},
			{Type: "status", Result: "PASS"},
			{Type: "coverage", CovPct: 99.8},
			{Type: "summary", Result: "ok", Name: "package2/bar", Duration: 4200 * time.Millisecond, CovPct: 99.8},
		}},
	{"11-go_1_5.txt",
		[]Event{
			{Type: "run_test", Id: 1, Name: "TestOne"},
			{Type: "end_test", Id: 1, Name: "TestOne", Result: "PASS", Duration: 20 * time.Millisecond},
			{Type: "run_test", Id: 2, Name: "TestTwo"},
			{Type: "end_test", Id: 2, Name: "TestTwo", Result: "PASS", Duration: 30 * time.Millisecond},
			{Type: "status", Result: "PASS"},
			{Type: "summary", Result: "ok", Name: "package/name", Duration: 50 * time.Millisecond},
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
