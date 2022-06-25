package gotest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jstemmer/go-junit-report/v2/gtr"

	"github.com/google/go-cmp/cmp"
)

var (
	testTimestamp     = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	testTimestampFunc = func() time.Time { return testTimestamp }
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
			parser := NewParser()
			parser.parseLine(test.input)
			got := parser.events
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("parseLine(%q) returned unexpected events, diff (-want, +got):\n%v", test.input, diff)
			}
		})
	}
}

func TestParseLargeLine(t *testing.T) {
	tests := []struct {
		desc      string
		inputSize int
	}{
		{"small size", 128},
		{"under buf size", 4095},
		{"buf size", 4096},
		{"multiple of buf size ", 4096 * 2},
		{"not multiple of buf size", 10 * 1024},
		{"bufio.MaxScanTokenSize", bufio.MaxScanTokenSize},
		{"over bufio.MaxScanTokenSize", bufio.MaxScanTokenSize + 1},
		{"under limit", maxLineSize - 1},
		{"at limit", maxLineSize},
		{"just over limit", maxLineSize + 1},
		{"over limit", maxLineSize + 128},
	}

	createInput := func(lines ...string) *bytes.Buffer {
		buf := &bytes.Buffer{}
		buf.WriteString("=== RUN TestOne\n--- PASS: TestOne (0.00s)\n")
		buf.WriteString(strings.Join(lines, "\n"))
		return buf
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			line1 := string(make([]byte, test.inputSize))
			line2 := "other line"
			report, err := NewParser().Parse(createInput(line1, line2))
			if err != nil {
				t.Fatalf("Parse() returned error %v", err)
			} else if len(report.Packages) != 1 {
				t.Fatalf("Parse() returned unexpected number of packages, got %d want 1.", len(report.Packages))
			} else if len(report.Packages[0].Output) != 2 {
				t.Fatalf("Parse() returned unexpected number of output lines, got %d want 1.", len(report.Packages[0].Output))
			}

			want := line1
			if len(want) > maxLineSize {
				want = want[:maxLineSize]
			}
			if got := report.Packages[0].Output[0]; got != want {
				t.Fatalf("Parse() output line1 mismatch, got len %d want len %d", len(got), len(want))
			}
			if report.Packages[0].Output[1] != line2 {
				t.Fatalf("Parse() output line2 mismatch, got %v want %v", report.Packages[0].Output[1], line2)
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
	want := gtr.Report{
		Packages: []gtr.Package{
			{
				Name:      "package/name",
				Duration:  1 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						ID:       1,
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Pass,
						Output: []string{
							"\tHello", // TODO: strip tabs?
						},
						Data: map[string]interface{}{},
					},
					{
						ID:       2,
						Name:     "TestSkip",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Skip,
						Data:     map[string]interface{}{},
					},
				},
			},
			{
				Name:      "package/name2",
				Duration:  1 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						ID:       3,
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Fail,
						Output: []string{
							"\tfile_test.go:10: error",
						},
						Data: map[string]interface{}{},
					},
				},
			},
			{
				Name:      "package/name3",
				Duration:  1234 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						ID:     4,
						Name:   "BenchmarkOne",
						Result: gtr.Pass,
						Data:   map[string]interface{}{key: Benchmark{NsPerOp: 100}},
					},
					{
						ID:     5,
						Name:   "BenchmarkTwo",
						Result: gtr.Fail,
						Data:   map[string]interface{}{},
					},
				},
				Output: []string{"goarch: amd64"},
			},
			{
				Name:      "package/failing1",
				Timestamp: testTimestamp,
				BuildError: gtr.Error{
					ID:     6,
					Name:   "package/failing1",
					Cause:  "[build failed]",
					Output: []string{"error message"},
				},
			},
		},
	}

	parser := NewParser(TimestampFunc(testTimestampFunc))
	got := parser.report(events)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("FromEvents report incorrect, diff (-want, +got):\n%v", diff)
	}
}

func TestSubtestModes(t *testing.T) {
	events := []Event{
		{Type: "run_test", Name: "TestParent"},
		{Type: "output", Data: "TestParent before"},
		{Type: "run_test", Name: "TestParent/Subtest#1"},
		{Type: "output", Data: "Subtest#1 output"},
		{Type: "run_test", Name: "TestParent/Subtest#2"},
		{Type: "output", Data: "Subtest#2 output"},
		{Type: "cont_test", Name: "TestParent"},
		{Type: "output", Data: "TestParent after"},
		{Type: "end_test", Name: "TestParent", Result: "PASS", Duration: 1 * time.Millisecond},
		{Type: "end_test", Name: "TestParent/Subtest#1", Result: "FAIL", Duration: 2 * time.Millisecond},
		{Type: "end_test", Name: "TestParent/Subtest#2", Result: "PASS", Duration: 3 * time.Millisecond},
		{Type: "output", Data: "output"},
		{Type: "summary", Result: "FAIL", Name: "package/name", Duration: 1 * time.Millisecond},
	}

	tests := []struct {
		name string
		mode SubtestMode
		want gtr.Report
	}{
		{
			name: "ignore subtest parent results",
			mode: IgnoreParentResults,
			want: gtr.Report{
				Packages: []gtr.Package{
					{
						Name:      "package/name",
						Duration:  1 * time.Millisecond,
						Timestamp: testTimestamp,
						Tests: []gtr.Test{
							{
								ID:       1,
								Name:     "TestParent",
								Duration: 1 * time.Millisecond,
								Result:   gtr.Pass,
								Output:   []string{"TestParent before", "TestParent after"},
								Data:     map[string]interface{}{},
							},
							{
								ID:       2,
								Name:     "TestParent/Subtest#1",
								Duration: 2 * time.Millisecond,
								Result:   gtr.Fail,
								Output:   []string{"Subtest#1 output"},
								Data:     map[string]interface{}{},
							},
							{
								ID:       3,
								Name:     "TestParent/Subtest#2",
								Duration: 3 * time.Millisecond,
								Result:   gtr.Pass,
								Output:   []string{"Subtest#2 output"},
								Data:     map[string]interface{}{},
							},
						},
						Output: []string{"output"},
					},
				},
			},
		},
		{
			name: "exclude subtest parents",
			mode: ExcludeParents,
			want: gtr.Report{
				Packages: []gtr.Package{
					{
						Name:      "package/name",
						Duration:  1 * time.Millisecond,
						Timestamp: testTimestamp,
						Tests: []gtr.Test{
							{
								ID:       2,
								Name:     "TestParent/Subtest#1",
								Duration: 2 * time.Millisecond,
								Result:   gtr.Fail,
								Output:   []string{"Subtest#1 output"},
								Data:     map[string]interface{}{},
							},
							{
								ID:       3,
								Name:     "TestParent/Subtest#2",
								Duration: 3 * time.Millisecond,
								Result:   gtr.Pass,
								Output:   []string{"Subtest#2 output"},
								Data:     map[string]interface{}{},
							},
						},
						Output: []string{"TestParent before", "TestParent after", "output"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := NewParser(TimestampFunc(testTimestampFunc), SetSubtestMode(test.mode))
			got := parser.report(events)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Invalid report created from events, diff (-want, +got):\n%v", diff)
			}
		})
	}
}
