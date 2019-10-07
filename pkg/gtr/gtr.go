// Package gtr defines a standard test report format and provides convenience
// methods to create and convert reports.
package gtr

import (
	"fmt"
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/junit"
)

var (
	propPrefixes   = map[string]bool{"goos": true, "goarch": true, "pkg": true}
	propFieldsFunc = func(r rune) bool { return r == ':' || r == ' ' }
)

type Report struct {
	Packages []Package
}

func (r *Report) HasFailures() bool {
	for _, pkg := range r.Packages {
		for _, t := range pkg.Tests {
			if t.Result == Fail {
				return true
			}
		}
	}
	return false
}

type Package struct {
	Name     string
	Duration time.Duration
	Coverage float64
	Output   []string

	Tests      []Test
	Benchmarks []Benchmark
}

type Test struct {
	Name     string
	Duration time.Duration
	Result   Result
	Output   []string
}

type Benchmark struct {
	Name        string
	Result      Result
	Output      []string
	Iterations  int64
	NsPerOp     float64
	MBPerSec    float64
	BytesPerOp  int64
	AllocsPerOp int64
}

// FromEvents creates a Report from the given list of events.
func FromEvents(events []Event) Report {
	report := NewReportBuilder()
	for _, ev := range events {
		switch ev.Type {
		case "run_test":
			report.CreateTest(ev.Name)
		case "end_test":
			report.EndTest(ev.Name, ev.Result, ev.Duration)
		case "benchmark":
			report.Benchmark(ev.Name, ev.Iterations, ev.NsPerOp, ev.MBPerSec, ev.BytesPerOp, ev.AllocsPerOp)
		case "status": // ignore for now
		case "summary":
			report.CreatePackage(ev.Name, ev.Duration)
		case "output":
			report.AppendOutput(ev.Data)
		default:
			fmt.Printf("unhandled event type: %v\n", ev.Type)
		}
	}
	return report.Build()
}

// JUnit converts the given report to a collection of JUnit Testsuites.
func JUnit(report Report) junit.Testsuites {
	var suites junit.Testsuites
	for _, pkg := range report.Packages {
		suite := junit.Testsuite{
			Name:  pkg.Name,
			Tests: len(pkg.Tests),
			Time:  junit.FormatDuration(pkg.Duration),
		}

		for _, line := range pkg.Output {
			if fields := strings.FieldsFunc(line, propFieldsFunc); len(fields) == 2 && propPrefixes[fields[0]] {
				suite.AddProperty(fields[0], fields[1])
			}
		}

		for _, test := range pkg.Tests {
			tc := junit.Testcase{
				Classname: pkg.Name,
				Name:      test.Name,
				Time:      junit.FormatDuration(test.Duration),
			}
			if test.Result == Fail {
				tc.Failure = &junit.Result{
					Message: "Failed",
					Data:    strings.Join(test.Output, "\n"),
				}
			} else if test.Result == Skip {
				tc.Skipped = &junit.Result{
					Data: strings.Join(test.Output, "\n"),
				}
			}
			suite.AddTestcase(tc)
		}

		for _, bm := range mergeBenchmarks(pkg.Benchmarks) {
			tc := junit.Testcase{
				Classname: pkg.Name,
				Name:      bm.Name,
				Time:      junit.FormatBenchmarkTime(time.Duration(bm.NsPerOp)),
			}
			if bm.Result == Fail {
				tc.Failure = &junit.Result{
					Message: "Failed",
				}
			}
			suite.AddTestcase(tc)
		}

		suites.AddSuite(suite)
	}
	return suites
}

func mergeBenchmarks(benchmarks []Benchmark) []Benchmark {
	var merged []Benchmark

	benchmap := make(map[string][]Benchmark)
	for _, bm := range benchmarks {
		if _, ok := benchmap[bm.Name]; !ok {
			merged = append(merged, Benchmark{Name: bm.Name})
		}
		benchmap[bm.Name] = append(benchmap[bm.Name], bm)
	}

	for i, bm := range merged {
		for _, b := range benchmap[bm.Name] {
			bm.NsPerOp += b.NsPerOp
			bm.MBPerSec += b.MBPerSec
			bm.BytesPerOp += b.BytesPerOp
			bm.AllocsPerOp += b.AllocsPerOp
		}
		n := len(benchmap[bm.Name])
		merged[i].NsPerOp = bm.NsPerOp / float64(n)
		merged[i].MBPerSec = bm.MBPerSec / float64(n)
		merged[i].BytesPerOp = bm.BytesPerOp / int64(n)
		merged[i].AllocsPerOp = bm.AllocsPerOp / int64(n)
	}

	return merged
}
