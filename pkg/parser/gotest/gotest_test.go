package gotest

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/gtr"

	"github.com/google/go-cmp/cmp"
)

type parseLineTest struct {
	input  string
	events interface{}
}

var parseLineTests = []parseLineTest{
	{
		"=== RUN TestOne",
		gtr.Event{Type: "run_test", Name: "TestOne"},
	},
	{
		"=== RUN   TestTwo/Subtest",
		gtr.Event{Type: "run_test", Name: "TestTwo/Subtest"},
	},
	{
		"=== PAUSE TestOne",
		gtr.Event{Type: "pause_test", Name: "TestOne"},
	},
	{
		"=== CONT  TestOne",
		gtr.Event{Type: "cont_test", Name: "TestOne"},
	},
	{
		"--- PASS: TestOne (12.34 seconds)",
		gtr.Event{Type: "end_test", Name: "TestOne", Result: "PASS", Duration: 12_340 * time.Millisecond},
	},
	{
		"    --- SKIP: TestOne/Subtest (0.00s)",
		gtr.Event{Type: "end_test", Name: "TestOne/Subtest", Result: "SKIP", Indent: 1},
	},
	{
		"        --- FAIL: TestOne/Subtest/#01 (0.35s)",
		gtr.Event{Type: "end_test", Name: "TestOne/Subtest/#01", Result: "FAIL", Duration: 350 * time.Millisecond, Indent: 2},
	},
	{
		"some text--- PASS: TestTwo (0.06 seconds)",
		[]gtr.Event{
			{Type: "output", Data: "some text"},
			{Type: "end_test", Name: "TestTwo", Result: "PASS", Duration: 60 * time.Millisecond},
		},
	},
	{
		"PASS",
		gtr.Event{Type: "status", Result: "PASS"},
	},
	{
		"FAIL",
		gtr.Event{Type: "status", Result: "FAIL"},
	},
	{
		"SKIP",
		gtr.Event{Type: "status", Result: "SKIP"},
	},
	{
		"ok      package/name/ok 0.100s",
		gtr.Event{Type: "summary", Name: "package/name/ok", Result: "ok", Duration: 100 * time.Millisecond},
	},
	{
		"FAIL    package/name/failing [build failed]",
		gtr.Event{Type: "summary", Name: "package/name/failing", Result: "FAIL", Data: "[build failed]"},
	},
	{
		"FAIL    package/other/failing [setup failed]",
		gtr.Event{Type: "summary", Name: "package/other/failing", Result: "FAIL", Data: "[setup failed]"},
	},
	{
		"ok package/other     (cached)",
		gtr.Event{Type: "summary", Name: "package/other", Result: "ok", Data: "(cached)"},
	},
	{
		"ok  	package/name 0.400s  coverage: 10.0% of statements",
		gtr.Event{Type: "summary", Name: "package/name", Result: "ok", Duration: 400 * time.Millisecond, CovPct: 10},
	},
	{
		"ok  	package/name 4.200s  coverage: 99.8% of statements in fmt, encoding/xml",
		gtr.Event{Type: "summary", Name: "package/name", Result: "ok", Duration: 4200 * time.Millisecond, CovPct: 99.8, CovPackages: []string{"fmt", "encoding/xml"}},
	},
	{
		"?   	package/name	[no test files]",
		gtr.Event{Type: "summary", Name: "package/name", Result: "?", Data: "[no test files]"},
	},
	{
		"ok  	package/name	0.001s [no tests to run]",
		gtr.Event{Type: "summary", Name: "package/name", Result: "ok", Duration: 1 * time.Millisecond, Data: "[no tests to run]"},
	},
	{
		"ok  	package/name	(cached) [no tests to run]",
		gtr.Event{Type: "summary", Name: "package/name", Result: "ok", Data: "(cached) [no tests to run]"},
	},
	{
		"coverage: 10% of statements",
		gtr.Event{Type: "coverage", CovPct: 10},
	},
	{
		"coverage: 10% of statements in fmt, encoding/xml",
		gtr.Event{Type: "coverage", CovPct: 10, CovPackages: []string{"fmt", "encoding/xml"}},
	},
	{
		"coverage: 13.37% of statements",
		gtr.Event{Type: "coverage", CovPct: 13.37},
	},
	{
		"coverage: 99.8% of statements in fmt, encoding/xml",
		gtr.Event{Type: "coverage", CovPct: 99.8, CovPackages: []string{"fmt", "encoding/xml"}},
	},
	{
		"BenchmarkOne-8                     2000000	       604 ns/op",
		gtr.Event{Type: "benchmark", Name: "BenchmarkOne", Iterations: 2_000_000, NsPerOp: 604},
	},
	{
		"BenchmarkTwo-16 30000	52568 ns/op	24879 B/op	494 allocs/op",
		gtr.Event{Type: "benchmark", Name: "BenchmarkTwo", Iterations: 30_000, NsPerOp: 52_568, BytesPerOp: 24_879, AllocsPerOp: 494},
	},
	{
		"BenchmarkThree      2000000000	         0.26 ns/op",
		gtr.Event{Type: "benchmark", Name: "BenchmarkThree", Iterations: 2_000_000_000, NsPerOp: 0.26},
	},
	{
		"BenchmarkFour-8         	   10000	    104427 ns/op	  95.76 MB/s	   40629 B/op	       5 allocs/op",
		gtr.Event{Type: "benchmark", Name: "BenchmarkFour", Iterations: 10_000, NsPerOp: 104_427, MBPerSec: 95.76, BytesPerOp: 40_629, AllocsPerOp: 5},
	},
	{
		"# package/name/failing1",
		gtr.Event{Type: "build_output", Name: "package/name/failing1"},
	},
	{
		"# package/name/failing2 [package/name/failing2.test]",
		gtr.Event{Type: "build_output", Name: "package/name/failing2"},
	},
	{
		"single line stdout",
		gtr.Event{Type: "output", Data: "single line stdout"},
	},
	{
		"# some more output",
		gtr.Event{Type: "output", Data: "# some more output"},
	},
	{
		"\tfile_test.go:11: Error message",
		gtr.Event{Type: "output", Data: "\tfile_test.go:11: Error message"},
	},
	{
		"\tfile_test.go:12: Longer",
		gtr.Event{Type: "output", Data: "\tfile_test.go:12: Longer"},
	},
	{
		"\t\terror",
		gtr.Event{Type: "output", Data: "\t\terror"},
	},
	{
		"\t\tmessage.",
		gtr.Event{Type: "output", Data: "\t\tmessage."},
	},
}

func TestParseLine(t *testing.T) {
	for i, test := range parseLineTests {
		var want []gtr.Event
		switch e := test.events.(type) {
		case gtr.Event:
			want = []gtr.Event{e}
		case []gtr.Event:
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
				t.Errorf("parseLine(%q) returned unexpected events, diff (-got, +want):\n%v", test.input, diff)
			}
		})
	}
}
