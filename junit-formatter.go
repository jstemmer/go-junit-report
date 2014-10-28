package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"runtime"
	"strings"
)

type JUnitTestSuites struct {
	XMLName	   xml.Name         `xml:"testsuites"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Time       string           `xml:"time,attr"`
	Name       string           `xml:"name,attr"`
        TestSuites []JUnitTestSuite
}

type JUnitTestSuite struct {
	XMLName    xml.Name        `xml:"testsuite"`
	Tests      int             `xml:"tests,attr"`
	Failures   int             `xml:"failures,attr"`
	Time       string          `xml:"time,attr"`
	Name       string          `xml:"name,attr"`
	Properties []JUnitProperty `xml:"properties>property,omitempty"`
	TestCases  []JUnitTestCase
}

type JUnitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Classname string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
}

type JUnitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type JUnitFailure struct {
	Message  string `xml:"message,attr"`
	Contents string `xml:",chardata"`
}

func NewJUnitProperty(name, value string) JUnitProperty {
	return JUnitProperty{
		Name:  name,
		Value: value,
	}
}

// JUnitReportXML writes a junit xml representation of the given report to w
// in the format described at http://windyroad.org/dl/Open%20Source/JUnit.xsd
func JUnitReportXML(report *Report, w io.Writer) error {
	if len(report.Packages) == 0 {
		return nil
	}

	suites := JUnitTestSuites{
		Name:       report.Packages[0].Name,  // There a better value to use?
		Tests:      0,
		Failures:   0,
		TestSuites: []JUnitTestSuite{},
	}

	// convert Report to JUnit test suites
	totalTime := 0
	for _, pkg := range report.Packages {
		ts := JUnitTestSuite{
			Tests:      len(pkg.Tests),
			Failures:   0,
			Time:       formatTime(pkg.Time),
			Name:       pkg.Name,
			Properties: []JUnitProperty{},
			TestCases:  []JUnitTestCase{},
		}

		classname := pkg.Name
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
			classname = pkg.Name[idx+1:]
		}

		// properties
		ts.Properties = append(ts.Properties, NewJUnitProperty("go.version", runtime.Version()))

		// individual test cases
		for _, test := range pkg.Tests {
			testCase := JUnitTestCase{
				Classname: classname,
				Name:      test.Name,
				Time:      formatTime(test.Time),
				Failure:   nil,
			}

			if test.Result == FAIL {
				ts.Failures += 1

				testCase.Failure = &JUnitFailure{
					Message:  "Failed",
					Contents: strings.Join(test.Output, "\n"),
				}
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		totalTime += pkg.Time
		suites.Tests += ts.Tests
		suites.Failures += ts.Failures
		suites.TestSuites = append(suites.TestSuites, ts)
	}
	suites.Time = formatTime(totalTime)

	// to xml
	bytes, err := xml.MarshalIndent(suites, "", "\t")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(w)

	// remove newline from xml.Header, because xml.MarshalIndent starts with a newline
	writer.WriteString(xml.Header[:len(xml.Header)-1])
	writer.WriteByte('\n')
	writer.Write(bytes)
	writer.WriteByte('\n')
	writer.Flush()

	return nil
}

func countFailures(tests []Test) (result int) {
	for _, test := range tests {
		if test.Result == FAIL {
			result += 1
		}
	}
	return
}

func formatTime(time int) string {
	return fmt.Sprintf("%.3f", float64(time)/1000.0)
}
