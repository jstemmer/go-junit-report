package gotest

import (
	"flag"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/gtr"

	"github.com/google/go-cmp/cmp"
)

var matchTest = flag.String("match", "", "only test testdata matching this pattern")

var tests = []struct {
	name   string
	input  string
	events []gtr.Event
}{
	{
		"run",
		inp("=== RUN TestOne",
			"=== RUN   TestTwo/Subtest",
		),
		[]gtr.Event{
			{Type: "run_test", Name: "TestOne"},
			{Type: "run_test", Name: "TestTwo/Subtest"},
		},
	},
	{
		"pause",
		"=== PAUSE TestOne",
		[]gtr.Event{
			{Type: "pause_test", Name: "TestOne"},
		},
	},
	{
		"cont",
		"=== CONT  TestOne",
		[]gtr.Event{
			{Type: "cont_test", Name: "TestOne"},
		},
	},
	{
		"end",
		inp("--- PASS: TestOne (12.34 seconds)",
			"    --- SKIP: TestOne/Subtest (0.00s)",
			"        --- FAIL: TestOne/Subtest/#01 (0.35s)",
			"some text--- PASS: TestTwo (0.06 seconds)",
		),
		[]gtr.Event{
			{Type: "end_test", Name: "TestOne", Result: "PASS", Duration: 12_340 * time.Millisecond},
			{Type: "end_test", Name: "TestOne/Subtest", Result: "SKIP", Indent: 1},
			{Type: "end_test", Name: "TestOne/Subtest/#01", Result: "FAIL", Duration: 350 * time.Millisecond, Indent: 2},
			{Type: "output", Data: "some text"},
			{Type: "end_test", Name: "TestTwo", Result: "PASS", Duration: 60 * time.Millisecond},
		},
	},
	{
		"status",
		inp("PASS",
			"FAIL",
			"SKIP",
		),
		[]gtr.Event{
			{Type: "status", Result: "PASS"},
			{Type: "status", Result: "FAIL"},
			{Type: "status", Result: "SKIP"},
		},
	},
	{
		"summary",
		inp("ok      package/name/ok 0.100s",
			"FAIL    package/name/failing [build failed]",
			"FAIL    package/other/failing [setup failed]",
			"ok package/other     (cached)",
		),
		[]gtr.Event{
			{Type: "summary", Name: "package/name/ok", Result: "ok", Duration: 100 * time.Millisecond},
			{Type: "summary", Name: "package/name/failing", Result: "FAIL", Data: "[build failed]"},
			{Type: "summary", Name: "package/other/failing", Result: "FAIL", Data: "[setup failed]"},
			{Type: "summary", Name: "package/other", Result: "ok", Data: "(cached)"},
		},
	},
	{
		"coverage",
		inp("coverage: 10% of statements",
			"coverage: 10% of statements in fmt, encoding/xml",
			"coverage: 13.37% of statements",
			"coverage: 99.8% of statements in fmt, encoding/xml",
		),
		[]gtr.Event{
			{Type: "coverage", CovPct: 10},
			{Type: "coverage", CovPct: 10, CovPackages: []string{"fmt", "encoding/xml"}},
			{Type: "coverage", CovPct: 13.37},
			{Type: "coverage", CovPct: 99.8, CovPackages: []string{"fmt", "encoding/xml"}},
		},
	},
	{
		"benchmark",
		inp("BenchmarkOne-8                     2000000	       604 ns/op",
			"BenchmarkTwo-16 30000	52568 ns/op	24879 B/op	494 allocs/op",
			"BenchmarkThree      2000000000	         0.26 ns/op",
			"BenchmarkFour-8         	   10000	    104427 ns/op	  95.76 MB/s	   40629 B/op	       5 allocs/op",
		),
		[]gtr.Event{
			{Type: "benchmark", Name: "BenchmarkOne", Iterations: 2_000_000, NsPerOp: 604},
			{Type: "benchmark", Name: "BenchmarkTwo", Iterations: 30_000, NsPerOp: 52_568, BytesPerOp: 24_879, AllocsPerOp: 494},
			{Type: "benchmark", Name: "BenchmarkThree", Iterations: 2_000_000_000, NsPerOp: 0.26},
			{Type: "benchmark", Name: "BenchmarkFour", Iterations: 10_000, NsPerOp: 104_427, MBPerSec: 95.76, BytesPerOp: 40_629, AllocsPerOp: 5},
		},
	},
	{
		"build output",
		inp("# package/name/failing1",
			"# package/name/failing2 [package/name/failing2.test]",
		),
		[]gtr.Event{
			{Type: "build_output", Name: "package/name/failing1"},
			{Type: "build_output", Name: "package/name/failing2"},
		},
	},
	{
		"output",
		inp("single line stdout",
			"# some more output",
			"\tfile_test.go:11: Error message",
			"\tfile_test.go:12: Longer",
			"\t\terror",
			"\t\tmessage.",
		),
		[]gtr.Event{
			{Type: "output", Data: "single line stdout"},
			{Type: "output", Data: "# some more output"},
			{Type: "output", Data: "\tfile_test.go:11: Error message"},
			{Type: "output", Data: "\tfile_test.go:12: Longer"},
			{Type: "output", Data: "\t\terror"},
			{Type: "output", Data: "\t\tmessage."},
		},
	},
}

func TestParse(t *testing.T) {
	matchRegex := compileMatch(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !matchRegex.MatchString(test.name) || test.input == "" {
				t.SkipNow()
			}
			testParse(t, test.name, test.input, test.events)
		})
	}
}

func testParse(t *testing.T, name, input string, want []gtr.Event) {
	got, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Errorf("Parse(%s) error: %v", name, err)
		return
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Parse returned unexpected events, diff (-got, +want):\n%v", diff)
	}
}

func compileMatch(t *testing.T) *regexp.Regexp {
	rx, err := regexp.Compile(*matchTest)
	if err != nil {
		t.Fatalf("Error compiling -match flag %q: %v", *matchTest, err)
	}
	return rx
}

func inp(lines ...string) string {
	return strings.Join(lines, "\n")
}
