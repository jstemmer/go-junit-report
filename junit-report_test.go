package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ujiro99/doctest-junit-report/parser"
)

type TestCase struct {
	name        string
	reportName  string
	report      *parser.Report
	noXMLHeader bool
	packageName string
}

var testCases = []TestCase{
	{
		name:       "alternative_macros.txt",
		reportName: "alternative_macros.xml",
		report: &parser.Report{
			TestSuites: []*parser.TestSuite{
				{
					Name: "alternative_macros.cpp",
					Time: 414,
					Tests: []*parser.Test{
						{
							Name:     "custom macros",
							Time:     414,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "alternative_macros.cpp",
						},
					},
				},
			},
		},
	}, {
		name:       "assertion_macros.txt",
		reportName: "assertion_macros.xml",
		report: &parser.Report{
			TestSuites: []*parser.TestSuite{
				{
					Name: "assertion_macros.cpp",
					Time: 4122,
					Tests: []*parser.Test{
						{
							Name:     "normal macros",
							Time:     993,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "expressions should be evaluated only once",
							Time:     61,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "exceptions-related macros",
							Time:     224,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "exceptions-related macros for std::exception",
							Time:     226,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "WARN level of asserts don't fail the test case",
							Time:     481,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "CHECK level of asserts fail the test case but don't abort it",
							Time:     248,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 1",
							Time:     115,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 2",
							Time:     53,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 3",
							Time:     48,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 4",
							Time:     50,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 5",
							Time:     86,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 6",
							Time:     61,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 7",
							Time:     50,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 8",
							Time:     55,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 9",
							Time:     50,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 10",
							Time:     50,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 11",
							Time:     47,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "REQUIRE level of asserts fail and abort the test case - 12",
							Time:     48,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "all binary assertions",
							Time:     904,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
						{
							Name:     "some asserts used in a function called by a test case",
							Time:     272,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "assertion_macros.cpp",
						},
					},
				},
			},
		},
	}, {
		name:       "test_cases_and_suites.txt",
		reportName: "test_cases_and_suites.xml",
		report: &parser.Report{
			TestSuites: []*parser.TestSuite{
				{
					Name: "test_cases_and_suites.cpp",
					Time: 1246,
					Tests: []*parser.Test{
						{
							Name:     "an empty test that will succeed - not part of a test suite",
							Time:     163,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
						{
							Name:     "should fail because of an exception",
							Time:     1008,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
						{
							Name:     "fixtured test - not part of a test suite",
							Time:     75,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
					},
				},
				{
					Name: "scoped test suite",
					Time: 202,
					Tests: []*parser.Test{
						{
							Name:     "part of scoped",
							Time:     107,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
						{
							Name:     "part of scoped 2",
							Time:     95,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
					},
				},
				{
					Name: "some TS",
					Time: 71,
					Tests: []*parser.Test{
						{
							Name:     "part of some TS",
							Time:     71,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
					},
				},
				{
					Name: "ts1",
					Time: 85,
					Tests: []*parser.Test{
						{
							Name:     "normal test in a test suite from a decorator",
							Time:     85,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
					},
				},
				{
					Name: "skipped test cases",
					Time: 86,
					Tests: []*parser.Test{
						{
							Name:     "unskipped",
							Time:     86,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
					},
				},
				{
					Name: "test suite with a description",
					Time: 756,
					Tests: []*parser.Test{
						{
							Name:     "fails - and its allowed",
							Time:     95,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
						{
							Name:     "doesn't fail which is fine",
							Time:     45,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
						{
							Name:     "fails as it should",
							Time:     93,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
						{
							Name:     "doesn't fail but it should have",
							Time:     58,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
						{
							Name:     "fails 1 time as it should",
							Time:     78,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
						{
							Name:     "fails more times as it should",
							Time:     387,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "test_cases_and_suites.cpp",
						},
					},
				},
			},
		},
	}, {
		name:       "subcases.txt",
		reportName: "subcases.xml",
		report: &parser.Report{
			TestSuites: []*parser.TestSuite{
				{
					Name: "subcases.cpp",
					Time: 11024,
					Tests: []*parser.Test{
						{
							Name:     "lots of nested subcases",
							Time:     2737,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "subcases.cpp",
						},
						{
							Name:     "subcases can be used in a separate function as well",
							Time:     437,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "subcases.cpp",
						},
						{
							Name:     "Scenario: vectors can be sized and resized",
							Time:     7850,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "subcases.cpp",
						},
					},
				},
			},
		},
	}, {
		name:       "coverage_maxout.txt",
		reportName: "coverage_maxout.xml",
		report: &parser.Report{
			TestSuites: []*parser.TestSuite{
				{
					Name: "coverage_maxout.cpp",
					Time: 641,
					Tests: []*parser.Test{
						{
							Name:     "doctest internals",
							Time:     641,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "coverage_maxout.cpp",
						},
					},
				},
				{
					Name: "exception related",
					Time: 865,
					Tests: []*parser.Test{
						{
							Name:     "will end from a std::string exception",
							Time:     718,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "coverage_maxout.cpp",
						},
						{
							Name:     "will end from a const char* exception",
							Time:     66,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "coverage_maxout.cpp",
						},
						{
							Name:     "will end from an unknown exception",
							Time:     81,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "coverage_maxout.cpp",
						},
					},
				},
			},
		},
	}, {
		name:       "logging.txt",
		reportName: "logging.xml",
		report: &parser.Report{
			TestSuites: []*parser.TestSuite{
				{
					Name: "logging.cpp",
					Time: 1993,
					Tests: []*parser.Test{
						{
							Name:     "logging the counter of a loop",
							Time:     738,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "logging.cpp",
						},
						{
							Name:     "a test case that will end from an exception",
							Time:     752,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "logging.cpp",
						},
						{
							Name:     "a test case that will end from an exception and should print the unprinted context",
							Time:     109,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "logging.cpp",
						},
						{
							Name:     "third party asserts can report failures to doctest",
							Time:     231,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "logging.cpp",
						},
						{
							Name:     "explicit failures 1",
							Time:     95,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "logging.cpp",
						},
						{
							Name:     "explicit failures 2",
							Time:     68,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "logging.cpp",
						},
					},
				},
			},
		},
	}, {
		name:       "stringification.txt",
		reportName: "stringification.xml",
		report: &parser.Report{
			TestSuites: []*parser.TestSuite{
				{
					Name: "stringification.cpp",
					Time: 2241,
					Tests: []*parser.Test{
						{
							Name:     "all asserts should fail and show how the objects get stringified",
							Time:     2175,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "stringification.cpp",
						},
						{
							Name:     "a test case that registers an exception translator for int and then throws one",
							Time:     66,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "stringification.cpp",
						},
					},
				},
			},
		},
	}, {
		name:       "templated_test_cases.txt",
		reportName: "templated_test_cases.xml",
		report: &parser.Report{
			TestSuites: []*parser.TestSuite{
				{
					Name: "templated_test_cases.cpp",
					Time: 1665,
					Tests: []*parser.Test{
						{
							Name:     "signed integers stuff<int>",
							Time:     409,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "signed integers stuff<short int>",
							Time:     57,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "signed integers stuff<char>",
							Time:     53,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "vector stuff<std::vector<int>>",
							Time:     52,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "default construction<int>",
							Time:     90,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "default construction<short int>",
							Time:     65,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "default construction<char>",
							Time:     63,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "default construction<double>",
							Time:     66,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "default construction<double>",
							Time:     76,
							Result:   parser.PASS,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "multiple types<>",
							Time:     79,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "multiple types<>",
							Time:     78,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "multiple types<>",
							Time:     79,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
						{
							Name:     "bad stringification of type pair<int_pair>",
							Time:     498,
							Result:   parser.FAIL,
							Output:   []string{},
							Filename: "templated_test_cases.cpp",
						},
					},
				},
			},
		},
	},
}

func TestParser(t *testing.T) {
	for _, testCase := range testCases {
		t.Logf("Running: %s", testCase.name)

		file, err := os.Open("tests/" + testCase.name)
		if err != nil {
			t.Fatal(err)
		}

		report, err := parser.Parse(file)
		if err != nil {
			t.Fatalf("error parsing: %s", err)
		}

		if report == nil {
			t.Fatalf("Report == nil")
		}

		expected := testCase.report
		if len(report.TestSuites) != len(expected.TestSuites) {
			t.Fatalf("Report packages == %d, want %d", len(report.TestSuites), len(expected.TestSuites))
		}

		for i, pkg := range report.TestSuites {
			expPkg := expected.TestSuites[i]

			if pkg.Name != expPkg.Name {
				t.Logf("TestSuite.Name %s", pkg.Name)
				t.Errorf("TestSuite.Name == %s, want %s", pkg.Name, expPkg.Name)
			}

			if pkg.Time != expPkg.Time {
				t.Logf("TestSuite.Name %s", pkg.Name)
				t.Errorf("TestSuite.Time == %d, want %d", pkg.Time, expPkg.Time)
			}

			if len(pkg.Tests) != len(expPkg.Tests) {
				t.Logf("TestSuite.Name %s", pkg.Name)
				t.Fatalf("TestSuite Tests == %d, want %d", len(pkg.Tests), len(expPkg.Tests))
			}

			for j, test := range pkg.Tests {
				expTest := expPkg.Tests[j]

				if test.Name != expTest.Name {
					t.Logf("Test.Name %s", test.Name)
					t.Errorf("Test.Name == %s, want %s", test.Name, expTest.Name)
				}

				if test.Time != expTest.Time {
					t.Logf("Test.Name %s", test.Name)
					t.Errorf("Test.Time == %d, want %d", test.Time, expTest.Time)
				}

				if test.Result != expTest.Result {
					t.Logf("Test.Name %s", test.Name)
					t.Errorf("Test.Result == %d, want %d", test.Result, expTest.Result)
				}

				testOutput := strings.Join(test.Output, "\n")
				expTestOutput := strings.Join(expTest.Output, "\n")
				if testOutput != expTestOutput {
					t.Logf("Test.Name %s", test.Name)
					t.Errorf("Test.Output (%s) ==\n%s\n, want\n%s", test.Name, testOutput, expTestOutput)
				}
			}
		}
	}
}

func TestJUnitFormatter(t *testing.T) {
	testJUnitFormatter(t, "", "")
}

func TestVersionFlag(t *testing.T) {
	testJUnitFormatter(t, "custom-version", "")
}

func TestPackageNameFlag(t *testing.T) {
	testJUnitFormatter(t, "", "custom-package")
}

func testJUnitFormatter(t *testing.T, version string, packageName string) {
	for _, testCase := range testCases {
		report, err := loadTestReport(testCase.reportName, version, packageName)
		if err != nil {
			t.Fatal(err)
		}

		var junitReport bytes.Buffer

		if err = JUnitReportXML(testCase.report, testCase.noXMLHeader, version, packageName, &junitReport); err != nil {
			t.Fatal(err)
		}

		if string(junitReport.Bytes()) != report {
			t.Fatalf("Fail: %s Report xml ==\n%s, want\n%s", testCase.name, string(junitReport.Bytes()), report)
		}
	}
}

func loadTestReport(name, version string, packageName string) (string, error) {
	contents, err := ioutil.ReadFile("tests/" + name)
	if err != nil {
		return "", err
	}

	report := string(contents)

	// replace value="1.0" With specified version
	if version != "" {
		report = strings.Replace(report, `<properties></properties>`, fmt.Sprintf(`<properties>
		<property name="version" value="%v"></property>
	</properties>`, version), -1)
	}

	// replace value="package" With specified packageName
	if packageName != "" {
		report = strings.Replace(report, `">`, fmt.Sprintf(`" name="%s">`, packageName), 1)
	}

	return report, nil
}
