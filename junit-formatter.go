package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type JUnitTestSuite struct {
	XMLName   xml.Name `xml:"testsuite"`
	Tests     int      `xml:"tests,attr"`
	Failures  int      `xml:"failures,attr"`
	Time      string   `xml:"time,attr"`
	Name      string   `xml:"name,attr"`
	TestCases []JUnitTestCase
}

type JUnitTestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	Classname string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Time      string   `xml:"time,attr"`
	Failure   *string  `xml:"failure"`
}

func JUnitReportXML(report *Report, w io.Writer) error {
	suites := []JUnitTestSuite{}

	// convert Report to JUnit test suites
	for _, pkg := range report.Packages {
		ts := JUnitTestSuite{
			Tests:     len(pkg.Tests),
			Failures:  0,
			Time:      formatTime(pkg.Time),
			Name:      pkg.Name,
			TestCases: []JUnitTestCase{},
		}

		classname := pkg.Name
		if idx := strings.LastIndex(classname, "/"); idx > -1 && idx < len(pkg.Name) {
			classname = pkg.Name[idx+1:]
		}

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

				// TODO: set error message
				msg := ""
				testCase.Failure = &msg
			}

			ts.TestCases = append(ts.TestCases, testCase)
		}

		suites = append(suites, ts)
	}

	// to xml
	bytes, err := xml.MarshalIndent(suites, "", "\t")
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(w)

	// remove newline from xml.Header, because xml.MarshalIndent starts with a newline
	writer.WriteString(xml.Header[:len(xml.Header)-1])
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
