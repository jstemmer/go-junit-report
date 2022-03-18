// Package junit defines a JUnit XML report and includes convenience methods
// for working with these reports.
package junit

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/gtr"
)

// Testsuites is a collection of JUnit testsuites.
type Testsuites struct {
	XMLName xml.Name `xml:"testsuites"`

	Name     string `xml:"name,attr,omitempty"`
	Time     string `xml:"time,attr,omitempty"` // total duration in seconds
	Tests    int    `xml:"tests,attr,omitempty"`
	Errors   int    `xml:"errors,attr,omitempty"`
	Failures int    `xml:"failures,attr,omitempty"`
	Skipped  int    `xml:"skipped,attr,omitempty"`
	Disabled int    `xml:"disabled,attr,omitempty"`

	Suites []Testsuite `xml:"testsuite,omitempty"`
}

// AddSuite adds a Testsuite and updates this testssuites' totals.
func (t *Testsuites) AddSuite(ts Testsuite) {
	t.Suites = append(t.Suites, ts)
	t.Tests += ts.Tests
	t.Errors += ts.Errors
	t.Failures += ts.Failures
	t.Skipped += ts.Skipped
	t.Disabled += ts.Disabled
}

// Testsuite is a single JUnit testsuite containing testcases.
type Testsuite struct {
	// required attributes
	Name     string `xml:"name,attr"`
	Tests    int    `xml:"tests,attr"`
	Failures int    `xml:"failures,attr"`
	Errors   int    `xml:"errors,attr"`

	// optional attributes
	Disabled  int    `xml:"disabled,attr,omitempty"`
	Hostname  string `xml:"hostname,attr,omitempty"`
	ID        int    `xml:"id,attr,omitempty"`
	Package   string `xml:"package,attr,omitempty"`
	Skipped   int    `xml:"skipped,attr,omitempty"`
	Time      string `xml:"time,attr"`                // duration in seconds
	Timestamp string `xml:"timestamp,attr,omitempty"` // date and time in ISO8601

	Properties *[]Property `xml:"properties>property,omitempty"`
	Testcases  []Testcase  `xml:"testcase,omitempty"`
	SystemOut  *Output     `xml:"system-out,omitempty"`
	SystemErr  *Output     `xml:"system-err,omitempty"`
}

// AddProperty adds a property with the given name and value to this Testsuite.
func (t *Testsuite) AddProperty(name, value string) {
	prop := Property{Name: name, Value: value}
	if t.Properties == nil {
		t.Properties = &[]Property{prop}
		return
	}
	props := append(*t.Properties, prop)
	t.Properties = &props
}

// AddTestcase adds Testcase tc to this Testsuite.
func (t *Testsuite) AddTestcase(tc Testcase) {
	t.Testcases = append(t.Testcases, tc)
	t.Tests += 1

	if tc.Error != nil {
		t.Errors += 1
	}

	if tc.Failure != nil {
		t.Failures += 1
	}

	if tc.Skipped != nil {
		t.Skipped += 1
	}
}

// SetTimestamp sets the timestamp in this Testsuite.
func (ts *Testsuite) SetTimestamp(t time.Time) {
	ts.Timestamp = t.Format(time.RFC3339)
}

// Testcase represents a single test with its results.
type Testcase struct {
	// required attributes
	Name      string `xml:"name,attr"`
	Classname string `xml:"classname,attr"`

	// optional attributes
	Time   string `xml:"time,attr,omitempty"` // duration in seconds
	Status string `xml:"status,attr,omitempty"`

	Skipped   *Result `xml:"skipped,omitempty"`
	Error     *Result `xml:"error,omitempty"`
	Failure   *Result `xml:"failure,omitempty"`
	SystemOut *Output `xml:"system-out,omitempty"`
	SystemErr *Output `xml:"system-err,omitempty"`
}

// Property represents a key/value pair.
type Property struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Result represents the result of a single test.
type Result struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr,omitempty"`
	Data    string `xml:",cdata"`
}

// Output represents output written to stdout or sderr.
type Output struct {
	Data string `xml:",cdata"`
}

// CreateFromReport creates a JUnit representation of the given gtr.Report.
func CreateFromReport(report gtr.Report, hostname string) Testsuites {
	var suites Testsuites
	for _, pkg := range report.Packages {
		var duration time.Duration
		suite := Testsuite{
			Name:     pkg.Name,
			Hostname: hostname,
		}

		if !pkg.Timestamp.IsZero() {
			suite.SetTimestamp(pkg.Timestamp)
		}

		for k, v := range pkg.Properties {
			suite.AddProperty(k, v)
		}

		if len(pkg.Output) > 0 {
			suite.SystemOut = &Output{Data: formatOutput(pkg.Output, 0)}
		}

		if pkg.Coverage > 0 {
			suite.AddProperty("coverage.statements.pct", fmt.Sprintf("%.2f", pkg.Coverage))
		}

		for _, test := range pkg.Tests {
			duration += test.Duration

			tc := Testcase{
				Classname: pkg.Name,
				Name:      test.Name,
				Time:      formatDuration(test.Duration),
			}

			if test.Result == gtr.Fail {
				tc.Failure = &Result{
					Message: "Failed",
					Data:    formatOutput(test.Output, test.Level),
				}
			} else if test.Result == gtr.Skip {
				tc.Skipped = &Result{
					Message: formatOutput(test.Output, test.Level),
				}
			} else if test.Result == gtr.Unknown {
				tc.Error = &Result{
					Message: "No test result found",
					Data:    formatOutput(test.Output, test.Level),
				}
			}

			suite.AddTestcase(tc)
		}

		for _, bm := range groupBenchmarksByName(pkg.Benchmarks) {
			tc := Testcase{
				Classname: pkg.Name,
				Name:      bm.Name,
				Time:      formatBenchmarkTime(time.Duration(bm.NsPerOp)),
			}

			if bm.Result == gtr.Fail {
				tc.Failure = &Result{
					Message: "Failed",
				}
			}

			suite.AddTestcase(tc)
		}

		// JUnit doesn't have a good way of dealing with build or runtime
		// errors that happen before a test has started, so we create a single
		// failing test that contains the build error details.
		if pkg.BuildError.Name != "" {
			tc := Testcase{
				Classname: pkg.BuildError.Name,
				Name:      pkg.BuildError.Cause,
				Time:      formatDuration(0),
				Error: &Result{
					Message: "Build error",
					Data:    strings.Join(pkg.BuildError.Output, "\n"),
				},
			}
			suite.AddTestcase(tc)
		}

		if pkg.RunError.Name != "" {
			tc := Testcase{
				Classname: pkg.RunError.Name,
				Name:      "Failure",
				Time:      formatDuration(0),
				Error: &Result{
					Message: "Run error",
					Data:    strings.Join(pkg.RunError.Output, "\n"),
				},
			}
			suite.AddTestcase(tc)
		}

		if (pkg.Duration) == 0 {
			suite.Time = formatDuration(duration)
		} else {
			suite.Time = formatDuration(pkg.Duration)
		}
		suites.AddSuite(suite)
	}
	return suites
}

// formatDuration returns the JUnit string representation of the given
// duration.
func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}

// formatBenchmarkTime returns the JUnit string representation of the given
// benchmark time.
func formatBenchmarkTime(d time.Duration) string {
	return fmt.Sprintf("%.9f", d.Seconds())
}

// formatOutput trims the test whitespace prefix from each line and joins all
// the lines.
func formatOutput(output []string, indent int) string {
	var lines []string
	for _, line := range output {
		lines = append(lines, gtr.TrimPrefixSpaces(line, indent))
	}
	return strings.Join(lines, "\n")
}

func groupBenchmarksByName(benchmarks []gtr.Benchmark) []gtr.Benchmark {
	var grouped []gtr.Benchmark

	benchmap := make(map[string][]gtr.Benchmark)
	for _, bm := range benchmarks {
		if _, ok := benchmap[bm.Name]; !ok {
			grouped = append(grouped, gtr.Benchmark{Name: bm.Name})
		}
		benchmap[bm.Name] = append(benchmap[bm.Name], bm)
	}

	for i, bm := range grouped {
		for _, b := range benchmap[bm.Name] {
			bm.NsPerOp += b.NsPerOp
			bm.MBPerSec += b.MBPerSec
			bm.BytesPerOp += b.BytesPerOp
			bm.AllocsPerOp += b.AllocsPerOp
		}
		n := len(benchmap[bm.Name])
		grouped[i].NsPerOp = bm.NsPerOp / float64(n)
		grouped[i].MBPerSec = bm.MBPerSec / float64(n)
		grouped[i].BytesPerOp = bm.BytesPerOp / int64(n)
		grouped[i].AllocsPerOp = bm.AllocsPerOp / int64(n)
	}

	return grouped
}
