package formatter

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/parser"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName xml.Name `xml:"testsuites"`
	Suites  []JUnitTestSuite
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
	// for benchmarks
	Bytes  string `xml:"bytes,attr,omitempty"`
	Allocs string `xml:"allocs,attr,omitempty"`
}

// JUnitSkipMessage contains the reason why a testcase was skipped.
type JUnitSkipMessage struct {
	Message string `xml:"message,attr"`
}

// JUnitProperty represents a key/value pair used to define properties.
type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// JUnitFailure contains data related to a failed test.
type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

// JUnitReportXML writes a JUnit xml representation of the given report to w
// in the format described at http://windyroad.org/dl/Open%20Source/JUnit.xsd
func JUnitReportXML(report *parser.Report, noXMLHeader bool, goVersion string, w io.Writer) error {
	suites := JUnitTestSuites{}

	// convert Report to JUnit test suites
	for _, pkg := range report.Packages {
		var tests int
		if len(pkg.Tests) >= 1 && len(pkg.Benchmarks) >= 1 {
			tests = len(pkg.Tests) + len(pkg.Benchmarks)
		} else if len(pkg.Benchmarks) >= 1 {
			tests = len(pkg.Benchmarks)
		} else {
			tests = len(pkg.Tests)
		}

		ts := JUnitTestSuite{
			Tests:      tests,
			Failures:   0,
			Time:       formatTime(pkg.Duration),
			Name:       pkg.Name,
			Properties: []JUnitProperty{},
			TestCases:  []JUnitTestCase{},
		}

		classname := pkg.Name
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
			classname = pkg.Name[idx+1:]
		}

		// properties
		if goVersion == "" {
			// if goVersion was not specified as a flag, fall back to version reported by runtime
			goVersion = runtime.Version()
		}
		ts.Properties = append(ts.Properties, JUnitProperty{"go.version", goVersion})
		if pkg.CoveragePct != "" {
			ts.Properties = append(ts.Properties, JUnitProperty{"coverage.statements.pct", pkg.CoveragePct})
		}

		// individual test cases
		for _, test := range pkg.Tests {
			testCase := JUnitTestCase{
				Classname: classname,
				Name:      test.Name,
				Time:      formatTime(test.Duration),
				Failure:   nil,
			}

			if test.Result == parser.FAIL {
				ts.Failures++
				testCase.Failure = &JUnitFailure{
					Message:  "Failed",
					Type:     "",
					Contents: strings.Join(test.Output, "\n"),
				}
			}

			if test.Result == parser.SKIP {
				testCase.SkipMessage = &JUnitSkipMessage{strings.Join(test.Output, "\n")}
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		// individual benchmarks
		for _, benchmark := range pkg.Benchmarks {
			benchmarkCase := JUnitTestCase{
				Classname: classname,
				Name:      benchmark.Name,
				Time:      formatBenchmarkTime(benchmark.Duration),
				Failure:   nil,
			}

			if benchmark.Bytes != 0 {
				benchmarkCase.Bytes = strconv.Itoa(benchmark.Bytes)
			}
			if benchmark.Allocs != 0 {
				benchmarkCase.Allocs = strconv.Itoa(benchmark.Allocs)
			}

			ts.TestCases = append(ts.TestCases, benchmarkCase)
		}

		suites.Suites = append(suites.Suites, ts)
	}

	// to xml
	bytes, err := xml.MarshalIndent(suites, "", "\t")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(w)

	if !noXMLHeader {
		writer.WriteString(xml.Header)
	}

	writer.Write(bytes)
	writer.WriteByte('\n')
	writer.Flush()

	return nil
}

func formatTime(d time.Duration) string {
	return fmt.Sprintf("%.3f", d.Seconds())
}

func formatBenchmarkTime(d time.Duration) string {
	return fmt.Sprintf("%.9f", d.Seconds())
}
