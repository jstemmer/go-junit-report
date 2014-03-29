package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"
)

type TestCase struct {
	name       string
	reportName string
	report     *Report
}

var testCases []TestCase = []TestCase{
	{
		name:       "01-pass.txt",
		reportName: "01-report.xml",
		report: &Report{
			Packages: []Package{
				{
					Name: "package/name",
					Time: 160,
					Tests: []Test{
						{
							Name:   "TestOne",
							Time:   60,
							Result: PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo",
							Time:   100,
							Result: PASS,
							Output: []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "02-fail.txt",
		reportName: "02-report.xml",
		report: &Report{
			Packages: []Package{
				{
					Name: "package/name",
					Time: 151,
					Tests: []Test{
						{
							Name:   "TestOne",
							Time:   20,
							Result: FAIL,
							Output: []string{
								"file_test.go:11: Error message",
								"file_test.go:11: Longer",
								"\terror",
								"\tmessage.",
							},
						},
						{
							Name:   "TestTwo",
							Time:   130,
							Result: PASS,
							Output: []string{},
						},
						{
							Name:   "TestThree",
							Time:   0,
							Result: SKIP,
							Output: []string{
								"Skipped because I'm broken",
							},
						},
					},
				},
			},
		},
	},
}

func TestParser(t *testing.T) {
	for _, testCase := range testCases {
		file, err := os.Open("tests/" + testCase.name)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		report, err := Parse(file)
		if err != nil {
			t.Fatalf("error parsing: %s", err)
		}

		if report == nil {
			t.Fatalf("Report == nil")
		}

		expected := testCase.report
		if len(report.Packages) != len(expected.Packages) {
			t.Fatalf("Report packages == %d, want %d", len(report.Packages), len(expected.Packages))
		}

		for i, pkg := range report.Packages {
			expPkg := expected.Packages[i]

			if pkg.Name != expPkg.Name {
				t.Errorf("Package.Name == %s, want %s", pkg.Name, expPkg.Name)
			}

			if pkg.Time != expPkg.Time {
				t.Errorf("Package.Time == %d, want %d", pkg.Time, expPkg.Time)
			}

			if len(pkg.Tests) != len(expPkg.Tests) {
				t.Fatalf("Package Tests == %d, want %d", len(pkg.Tests), len(expPkg.Tests))
			}

			for j, test := range pkg.Tests {
				expTest := expPkg.Tests[j]

				if test.Name != expTest.Name {
					t.Errorf("Test.Name == %s, want %s", test.Name, expTest.Name)
				}

				if test.Time != expTest.Time {
					t.Errorf("Test.Time == %d, want %d", test.Time, expTest.Time)
				}

				if test.Result != expTest.Result {
					t.Errorf("Test.Result == %#v, want %#v", test, expTest)
				}

				testOutput := strings.Join(test.Output, "\n")
				expTestOutput := strings.Join(expTest.Output, "\n")
				if testOutput != expTestOutput {
					t.Errorf("Test.Output ==\n%s\n, want\n%s\n", testOutput, expTestOutput)
				}
			}
		}
	}
}

func TestJUnitFormatter(t *testing.T) {
	for _, testCase := range testCases {
		report, err := loadTestReport(testCase.reportName)
		if err != nil {
			t.Fatal(err)
		}

		var junitReport bytes.Buffer

		if err = JUnitReportXML(testCase.report, &junitReport); err != nil {
			t.Fatal(err)
		}

		if string(junitReport.Bytes()) != report {
			t.Fatalf("Report xml ==\n%s, want\n%s\n", string(junitReport.Bytes()), report)
		}
	}
}

func loadTestReport(name string) (string, error) {
	contents, err := ioutil.ReadFile("tests/" + name)
	if err != nil {
		return "", err
	}

	// replace value="1.0" With actual version
	report := strings.Replace(string(contents), `value="1.0"`, fmt.Sprintf(`value="%s"`, runtime.Version()), 1)

	return report, nil
}
