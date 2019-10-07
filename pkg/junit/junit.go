// Package junit defines a JUnit XML report and includes convenience methods
// for working with these reports.
package junit

import (
	"encoding/xml"
	"fmt"
	"time"
)

// Testsuites is a collection of JUnit testsuites.
type Testsuites struct {
	XMLName xml.Name `xml:"testsuites"`

	Name     string `xml:"name,attr,omitempty"`
	Time     string `xml:"time,attr,omitempty"` // total duration in seconds
	Tests    int    `xml:"tests,attr,omitempty"`
	Errors   int    `xml:"errors,attr,omitempty"`
	Failures int    `xml:"failures,attr,omitempty"`
	Disabled int    `xml:"disabled,attr,omitempty"`

	Suites []Testsuite `xml:"testsuite,omitempty"`
}

// AddSuite adds a Testsuite and updates this testssuites' totals.
func (t *Testsuites) AddSuite(ts Testsuite) {
	t.Suites = append(t.Suites, ts)
	t.Tests += ts.Tests
	t.Errors += ts.Errors
	t.Failures += ts.Failures
	t.Disabled += ts.Disabled
}

// Testsuite is a single JUnit testsuite containing testcases.
type Testsuite struct {
	// required attributes
	Name  string `xml:"name,attr"`
	Tests int    `xml:"tests,attr"`

	// optional attributes
	Disabled  int    `xml:"disabled,attr,omitempty"`
	Errors    int    `xml:"errors,attr"`
	Failures  int    `xml:"failures,attr"`
	Hostname  string `xml:"hostname,attr,omitempty"`
	ID        int    `xml:"id,attr,omitempty"`
	Package   string `xml:"package,attr,omitempty"`
	Skipped   int    `xml:"skipped,attr,omitempty"`
	Time      string `xml:"time,attr"`                // duration in seconds
	Timestamp string `xml:"timestamp,attr,omitempty"` // date and time in ISO8601

	Properties []Property `xml:"properties>property,omitempty"`
	Testcases  []Testcase `xml:"testcase,omitempty"`
	SystemOut  string     `xml:"system-out,omitempty"`
	SystemErr  string     `xml:"system-err,omitempty"`
}

func (t *Testsuite) AddProperty(name, value string) {
	t.Properties = append(t.Properties, Property{Name: name, Value: value})
}

func (t *Testsuite) AddTestcase(tc Testcase) {
	t.Testcases = append(t.Testcases, tc)
	t.Tests += 1

	if tc.Error != nil {
		t.Errors += 1
	}

	if tc.Failure != nil {
		t.Failures += 1
	}
}

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
	SystemOut string  `xml:"system-out,omitempty"`
	SystemErr string  `xml:"system-err,omitempty"`
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
	Data    string `xml:",chardata"`
}

// FormatDuration returns the JUnit string representation of the given
// duration.
func FormatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}

// FormatBenchmarkTime returns the JUnit string representation of the given
// benchmark time.
func FormatBenchmarkTime(d time.Duration) string {
	return fmt.Sprintf("%.9f", d.Seconds())
}
