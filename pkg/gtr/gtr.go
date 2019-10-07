// Package gtr defines a standard test report format and provides convenience
// methods to create and convert reports.
package gtr

import (
	"fmt"
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/parser/gotest"
	"github.com/jstemmer/go-junit-report/v2/pkg/junit"
)

type Result int

const (
	// TODO: move these to event and don't make the all-caps
	UNKNOWN Result = iota
	PASS
	FAIL
	SKIP
)

func (r Result) String() string {
	switch r {
	case UNKNOWN:
		return "UNKNOWN"
	case PASS:
		return "PASS"
	case FAIL:
		return "FAIL"
	case SKIP:
		return "SKIP"
	default:
		panic("invalid result")
	}
}

type Report struct {
	Packages []Package
}

func (r *Report) HasFailures() bool {
	for _, pkg := range r.Packages {
		for _, t := range pkg.Tests {
			if t.Result == FAIL {
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

	Tests []Test
}

type Test struct {
	Name     string
	Duration time.Duration
	Result   Result
	Output   []string
}

// FromEvents creates a Report from the given list of events.
func FromEvents(events []gotest.Event) Report {
	report := NewReportBuilder()
	for _, ev := range events {
		switch ev.Type {
		case "run_test":
			report.CreateTest(ev.Name)
		case "end_test":
			report.EndTest(ev.Name, ev.Result, ev.Duration)
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

		for _, test := range pkg.Tests {
			tc := junit.Testcase{
				Classname: pkg.Name,
				Name:      test.Name,
				Time:      junit.FormatDuration(test.Duration),
			}
			if test.Result == FAIL {
				tc.Failure = &junit.Result{
					Message: "Failed",
					Data:    strings.Join(test.Output, "\n"),
				}
			} else if test.Result == SKIP {
				tc.Skipped = &junit.Result{
					Data: strings.Join(test.Output, "\n"),
				}
			}
			suite.AddTestcase(tc)
		}
		suites.AddSuite(suite)
	}
	return suites
}
