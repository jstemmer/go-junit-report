package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/ujiro99/doctest-junit-report/parser"
)

// JUnitTestSuites is a collection of JUnit test suites.
type JUnitTestSuites struct {
	XMLName    xml.Name        `xml:"testsuites"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr,omitempty"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	Suites     []JUnitTestSuite
}

// JUnitTestSuite is a single JUnit test suite which may contain many
// testcases.
type JUnitTestSuite struct {
	XMLName   xml.Name `xml:"testsuite"`
	Tests     int      `xml:"tests,attr"`
	Failures  int      `xml:"failures,attr"`
	Time      string   `xml:"time,attr"`
	Name      string   `xml:"name,attr"`
	TestCases []JUnitTestCase
}

// JUnitTestCase is a single test case with its result.
type JUnitTestCase struct {
	XMLName     xml.Name          `xml:"testcase"`
	Classname   string            `xml:"classname,attr"`
	Name        string            `xml:"name,attr"`
	Time        string            `xml:"time,attr"`
	SkipMessage *JUnitSkipMessage `xml:"skipped,omitempty"`
	Failure     *JUnitFailure     `xml:"failure,omitempty"`
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
func JUnitReportXML(report *parser.Report, noXMLHeader bool, version string, packageName string, w io.Writer) error {
	suites := JUnitTestSuites{
		Name:       packageName,
		Tests:      report.Tests(),
		Failures:   report.Failures(),
		Properties: []JUnitProperty{},
		Time:       formatTime(report.Time()),
	}

	if version != "" {
		suites.Properties = append(suites.Properties, JUnitProperty{"version", version})
	}

	// convert Report to JUnit test suites
	for _, pkg := range report.TestSuites {
		ts := JUnitTestSuite{
			Tests:     len(pkg.Tests),
			Failures:  0,
			Time:      formatTime(pkg.Time),
			Name:      pkg.Name,
			TestCases: []JUnitTestCase{},
		}

		// individual test cases
		for _, test := range pkg.Tests {
			testCase := JUnitTestCase{
				Classname: test.Filename,
				Name:      test.Name,
				Time:      formatTime(test.Time),
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

func formatTime(time int) string {
	return fmt.Sprintf("%.6f", float64(time)/1000000.0)
}
