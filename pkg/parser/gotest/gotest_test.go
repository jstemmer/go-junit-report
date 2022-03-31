package gotest

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jstemmer/go-junit-report/v2/pkg/gtr"
)

var (
	testTimestamp = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
)

type parseLineTest struct {
	input  string
	events interface{}
}

var parseLineTests = []parseLineTest{
	{
		"=== RUN TestOne",
		Event{Type: "run_test", Name: "TestOne"},
	},
	{
		"=== RUN   TestTwo/Subtest",
		Event{Type: "run_test", Name: "TestTwo/Subtest"},
	},
	{
		"=== PAUSE TestOne",
		Event{Type: "pause_test", Name: "TestOne"},
	},
	{
		"=== CONT  TestOne",
		Event{Type: "cont_test", Name: "TestOne"},
	},
	{
		"--- PASS: TestOne (12.34 seconds)",
		Event{Type: "end_test", Name: "TestOne", Result: "PASS", Duration: 12_340 * time.Millisecond},
	},
	{
		"    --- SKIP: TestOne/Subtest (0.00s)",
		Event{Type: "end_test", Name: "TestOne/Subtest", Result: "SKIP", Indent: 1},
	},
	{
		"        --- FAIL: TestOne/Subtest/#01 (0.35s)",
		Event{Type: "end_test", Name: "TestOne/Subtest/#01", Result: "FAIL", Duration: 350 * time.Millisecond, Indent: 2},
	},
	{
		"some text--- PASS: TestTwo (0.06 seconds)",
		[]Event{
			{Type: "output", Data: "some text"},
			{Type: "end_test", Name: "TestTwo", Result: "PASS", Duration: 60 * time.Millisecond},
		},
	},
	{
		"PASS",
		Event{Type: "status", Result: "PASS"},
	},
	{
		"FAIL",
		Event{Type: "status", Result: "FAIL"},
	},
	{
		"SKIP",
		Event{Type: "status", Result: "SKIP"},
	},
	{
		"ok      package/name/ok 0.100s",
		Event{Type: "summary", Name: "package/name/ok", Result: "ok", Duration: 100 * time.Millisecond},
	},
	{
		"FAIL    package/name/failing [build failed]",
		Event{Type: "summary", Name: "package/name/failing", Result: "FAIL", Data: "[build failed]"},
	},
	{
		"FAIL    package/other/failing [setup failed]",
		Event{Type: "summary", Name: "package/other/failing", Result: "FAIL", Data: "[setup failed]"},
	},
	{
		"ok package/other     (cached)",
		Event{Type: "summary", Name: "package/other", Result: "ok", Data: "(cached)"},
	},
	{
		"ok  	package/name 0.400s  coverage: 10.0% of statements",
		Event{Type: "summary", Name: "package/name", Result: "ok", Duration: 400 * time.Millisecond, CovPct: 10},
	},
	{
		"ok  	package/name 4.200s  coverage: 99.8% of statements in fmt, encoding/xml",
		Event{Type: "summary", Name: "package/name", Result: "ok", Duration: 4200 * time.Millisecond, CovPct: 99.8, CovPackages: []string{"fmt", "encoding/xml"}},
	},
	{
		"?   	package/name	[no test files]",
		Event{Type: "summary", Name: "package/name", Result: "?", Data: "[no test files]"},
	},
	{
		"ok  	package/name	0.001s [no tests to run]",
		Event{Type: "summary", Name: "package/name", Result: "ok", Duration: 1 * time.Millisecond, Data: "[no tests to run]"},
	},
	{
		"ok  	package/name	(cached) [no tests to run]",
		Event{Type: "summary", Name: "package/name", Result: "ok", Data: "(cached) [no tests to run]"},
	},
	{
		"coverage: 10% of statements",
		Event{Type: "coverage", CovPct: 10},
	},
	{
		"coverage: 10% of statements in fmt, encoding/xml",
		Event{Type: "coverage", CovPct: 10, CovPackages: []string{"fmt", "encoding/xml"}},
	},
	{
		"coverage: 13.37% of statements",
		Event{Type: "coverage", CovPct: 13.37},
	},
	{
		"coverage: 99.8% of statements in fmt, encoding/xml",
		Event{Type: "coverage", CovPct: 99.8, CovPackages: []string{"fmt", "encoding/xml"}},
	},
	{
		"BenchmarkOK",
		Event{Type: "run_benchmark", Name: "BenchmarkOK"},
	},
	{
		"BenchmarkOne-8                     2000000	       604 ns/op",
		Event{Type: "benchmark", Name: "BenchmarkOne", Iterations: 2_000_000, NsPerOp: 604},
	},
	{
		"BenchmarkTwo-16 30000	52568 ns/op	24879 B/op	494 allocs/op",
		Event{Type: "benchmark", Name: "BenchmarkTwo", Iterations: 30_000, NsPerOp: 52_568, BytesPerOp: 24_879, AllocsPerOp: 494},
	},
	{
		"BenchmarkThree      2000000000	         0.26 ns/op",
		Event{Type: "benchmark", Name: "BenchmarkThree", Iterations: 2_000_000_000, NsPerOp: 0.26},
	},
	{
		"BenchmarkFour-8         	   10000	    104427 ns/op	  95.76 MB/s	   40629 B/op	       5 allocs/op",
		Event{Type: "benchmark", Name: "BenchmarkFour", Iterations: 10_000, NsPerOp: 104_427, MBPerSec: 95.76, BytesPerOp: 40_629, AllocsPerOp: 5},
	},
	{
		"--- BENCH: BenchmarkOK-8",
		Event{Type: "end_benchmark", Name: "BenchmarkOK", Result: "BENCH"},
	},
	{
		"--- FAIL: BenchmarkError",
		Event{Type: "end_benchmark", Name: "BenchmarkError", Result: "FAIL"},
	},
	{
		"--- SKIP: BenchmarkSkip",
		Event{Type: "end_benchmark", Name: "BenchmarkSkip", Result: "SKIP"},
	},
	{
		"# package/name/failing1",
		Event{Type: "build_output", Name: "package/name/failing1"},
	},
	{
		"# package/name/failing2 [package/name/failing2.test]",
		Event{Type: "build_output", Name: "package/name/failing2"},
	},
	{
		"single line stdout",
		Event{Type: "output", Data: "single line stdout"},
	},
	{
		"# some more output",
		Event{Type: "output", Data: "# some more output"},
	},
	{
		"\tfile_test.go:11: Error message",
		Event{Type: "output", Data: "\tfile_test.go:11: Error message"},
	},
	{
		"\tfile_test.go:12: Longer",
		Event{Type: "output", Data: "\tfile_test.go:12: Longer"},
	},
	{
		"\t\terror",
		Event{Type: "output", Data: "\t\terror"},
	},
	{
		"\t\tmessage.",
		Event{Type: "output", Data: "\t\tmessage."},
	},
}

func TestParseLine(t *testing.T) {
	for i, test := range parseLineTests {
		var want []Event
		switch e := test.events.(type) {
		case Event:
			want = []Event{e}
		case []Event:
			want = e
		default:
			panic("invalid events type")
		}

		var types []string
		for _, e := range want {
			types = append(types, e.Type)
		}

		name := fmt.Sprintf("%d %s", i+1, strings.Join(types, ","))
		t.Run(name, func(t *testing.T) {
			parser := New()
			parser.parseLine(test.input)
			got := parser.events
			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("parseLine(%q) returned unexpected events, diff (-want, +got):\n%v", test.input, diff)
			}
		})
	}
}

func TestReport(t *testing.T) {
	events := []Event{
		{Type: "run_test", Name: "TestOne"},
		{Type: "output", Data: "\tHello"},
		{Type: "end_test", Name: "TestOne", Result: "PASS", Duration: 1 * time.Millisecond},
		{Type: "status", Result: "PASS"},
		{Type: "run_test", Name: "TestSkip"},
		{Type: "end_test", Name: "TestSkip", Result: "SKIP", Duration: 1 * time.Millisecond},
		{Type: "summary", Result: "ok", Name: "package/name", Duration: 1 * time.Millisecond},
		{Type: "run_test", Name: "TestOne"},
		{Type: "output", Data: "\tfile_test.go:10: error"},
		{Type: "end_test", Name: "TestOne", Result: "FAIL", Duration: 1 * time.Millisecond},
		{Type: "status", Result: "FAIL"},
		{Type: "summary", Result: "FAIL", Name: "package/name2", Duration: 1 * time.Millisecond},
		{Type: "output", Data: "goarch: amd64"},
		{Type: "run_benchmark", Name: "BenchmarkOne"},
		{Type: "benchmark", Name: "BenchmarkOne", NsPerOp: 100},
		{Type: "end_benchmark", Name: "BenchmarkOne", Result: "BENCH"},
		{Type: "run_benchmark", Name: "BenchmarkTwo"},
		{Type: "benchmark", Name: "BenchmarkTwo"},
		{Type: "end_benchmark", Name: "BenchmarkTwo", Result: "FAIL"},
		{Type: "status", Result: "PASS"},
		{Type: "summary", Result: "ok", Name: "package/name3", Duration: 1234 * time.Millisecond},
		{Type: "build_output", Name: "package/failing1"},
		{Type: "output", Data: "error message"},
		{Type: "summary", Result: "FAIL", Name: "package/failing1", Data: "[build failed]"},
	}
	expected := gtr.Report{
		Packages: []gtr.Package{
			{
				Name:      "package/name",
				Duration:  1 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Pass,
						Output: []string{
							"\tHello", // TODO: strip tabs?
						},
					},
					{
						Name:     "TestSkip",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Skip,
					},
				},
			},
			{
				Name:      "package/name2",
				Duration:  1 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Fail,
						Output: []string{
							"\tfile_test.go:10: error",
						},
					},
				},
			},
			{
				Name:      "package/name3",
				Duration:  1234 * time.Millisecond,
				Timestamp: testTimestamp,
				Benchmarks: []gtr.Benchmark{
					{
						Name:    "BenchmarkOne",
						Result:  gtr.Pass,
						NsPerOp: 100,
					},
					{
						Name:    "BenchmarkTwo",
						Result:  gtr.Fail,
					},
				},
				Output: []string{"goarch: amd64"},
			},
			{
				Name:      "package/failing1",
				Timestamp: testTimestamp,
				BuildError: gtr.Error{
					Name:   "package/failing1",
					Cause:  "[build failed]",
					Output: []string{"error message"},
				},
			},
		},
	}

	parser := New(TimestampFunc(func() time.Time { return testTimestamp }))
	actual := parser.report(events)
	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("FromEvents report incorrect, diff (-want, +got):\n%v", diff)
	}
}
