package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/redorb/go-junit-report/parser"
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
		name:       "01-pass.txt",
		reportName: "01-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 160,
					Tests: []*parser.Test{
						{
							Name:   "TestZ",
							Time:   60,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestA",
							Time:   100,
							Result: parser.PASS,
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
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 151,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   20,
							Result: parser.FAIL,
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
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "03-skip.txt",
		reportName: "03-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 150,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   20,
							Result: parser.SKIP,
							Output: []string{
								"file_test.go:11: Skip message",
							},
						},
						{
							Name:   "TestTwo",
							Time:   130,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "04-go_1_4.txt",
		reportName: "04-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 160,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   60,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "05-no_xml_header.txt",
		reportName: "05-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 160,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   60,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
			},
		},
		noXMLHeader: true,
	},
	{
		name:       "06-mixed.txt",
		reportName: "06-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name1",
					Time: 160,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   60,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
				{
					Name: "package/name2",
					Time: 151,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   20,
							Result: parser.FAIL,
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
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
			},
		},
		noXMLHeader: true,
	},
	{
		name:       "07-compiled_test.txt",
		reportName: "07-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "test/package",
					Time: 160,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   60,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
			},
		},
		packageName: "test/package",
	},
	{
		name:       "08-parallel.txt",
		reportName: "08-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "github.com/dmitris/test-go-junit-report",
					Time: 440,
					Tests: []*parser.Test{
						{
							Name:   "TestDoFoo",
							Time:   270,
							Result: parser.PASS,
							Output: []string{"cov_test.go:10: DoFoo log 1", "cov_test.go:10: DoFoo log 2"},
						},
						{
							Name:   "TestDoFoo2",
							Time:   160,
							Result: parser.PASS,
							Output: []string{"cov_test.go:21: DoFoo2 log 1", "cov_test.go:21: DoFoo2 log 2"},
						},
					},
				},
			},
		},
	},
	{
		name:       "09-coverage.txt",
		reportName: "09-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 160,
					Tests: []*parser.Test{
						{
							Name:   "TestZ",
							Time:   60,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestA",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
					},
					CoveragePct: "13.37",
				},
			},
		},
	},
	{
		name:       "10-multipkg-coverage.txt",
		reportName: "10-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package1/foo",
					Time: 400,
					Tests: []*parser.Test{
						{
							Name:   "TestA",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestB",
							Time:   300,
							Result: parser.PASS,
							Output: []string{},
						},
					},
					CoveragePct: "10.0",
				},
				{
					Name: "package2/bar",
					Time: 4200,
					Tests: []*parser.Test{
						{
							Name:   "TestC",
							Time:   4200,
							Result: parser.PASS,
							Output: []string{},
						},
					},
					CoveragePct: "99.8",
				},
			},
		},
	},
	{
		name:       "11-go_1_5.txt",
		reportName: "11-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 50,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   20,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo",
							Time:   30,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "12-go_1_7.txt",
		reportName: "12-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 50,
					Tests: []*parser.Test{
						{
							Name:   "TestOne",
							Time:   10,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestOne/Child",
							Time:   20,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestOne/Child#01",
							Time:   30,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestOne/Child=02",
							Time:   40,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo",
							Time:   10,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo/Child",
							Time:   20,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo/Child#01",
							Time:   30,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestTwo/Child=02",
							Time:   40,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestThree",
							Time:   10,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestThree/a#1",
							Time:   20,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestThree/a#1/b#1",
							Time:   30,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestThree/a#1/b#1/c#1",
							Time:   40,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestFour",
							Time:   20,
							Result: parser.FAIL,
							Output: []string{},
						},
						{
							Name:   "TestFour/#00",
							Time:   0,
							Result: parser.FAIL,
							Output: []string{
								"example.go:12: Expected abc  OBTAINED:",
								"	xyz",
								"example.go:123: Expected and obtained are different.",
							},
						},
						{
							Name:   "TestFour/#01",
							Time:   0,
							Result: parser.SKIP,
							Output: []string{
								"example.go:1234: Not supported yet.",
							},
						},
						{
							Name:   "TestFour/#02",
							Time:   0,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestFive",
							Time:   0,
							Result: parser.SKIP,
							Output: []string{
								"example.go:1392: Not supported yet.",
							},
						},
						{
							Name:   "TestSix",
							Time:   0,
							Result: parser.FAIL,
							Output: []string{
								"example.go:371: This should not fail!",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "13-syntax-error.txt",
		reportName: "13-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name/passing1",
					Time: 100,
					Tests: []*parser.Test{
						{
							Name:   "TestA",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
				{
					Name: "package/name/passing2",
					Time: 100,
					Tests: []*parser.Test{
						{
							Name:   "TestB",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
				{
					Name: "package/name/failing1",
					Tests: []*parser.Test{
						{
							Name:   "[build failed]",
							Time:   0,
							Result: parser.FAIL,
							Output: []string{
								"failing1/failing_test.go:15: undefined: x",
							},
						},
					},
				},
				{
					Name: "package/name/failing2",
					Tests: []*parser.Test{
						{
							Name:   "[build failed]",
							Time:   0,
							Result: parser.FAIL,
							Output: []string{
								"failing2/another_failing_test.go:20: undefined: y",
							},
						},
					},
				},
				{
					Name: "package/name/setupfailing1",
					Tests: []*parser.Test{
						{
							Name:   "[setup failed]",
							Time:   0,
							Result: parser.FAIL,
							Output: []string{
								"setupfailing1/failing_test.go:4: cannot find package \"other/package\" in any of:",
								"\t/path/vendor (vendor tree)",
								"\t/path/go/root (from $GOROOT)",
								"\t/path/go/path (from $GOPATH)",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "14-panic.txt",
		reportName: "14-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/panic",
					Time: 3,
					Tests: []*parser.Test{
						{
							Name:   "Failure",
							Result: parser.FAIL,
							Output: []string{
								"panic: init",
								"stacktrace",
							},
						},
					},
				},
				{
					Name: "package/panic2",
					Time: 3,
					Tests: []*parser.Test{
						{
							Name:   "Failure",
							Result: parser.FAIL,
							Output: []string{
								"panic: init",
								"stacktrace",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "15-empty.txt",
		reportName: "15-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name:  "package/empty",
					Time:  1,
					Tests: []*parser.Test{},
				},
			},
		},
	},
	{
		name:       "16-repeated-names.txt",
		reportName: "16-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/repeated-names",
					Time: 1,
					Tests: []*parser.Test{
						{
							Name:   "TestRepeat",
							Time:   0,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestRepeat",
							Time:   0,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestRepeat",
							Time:   0,
							Result: parser.PASS,
							Output: []string{},
						},
					},
				},
			},
		},
	},
	{
		name:       "17-race.txt",
		reportName: "17-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "race_test",
					Time: 15,
					Tests: []*parser.Test{
						{
							Name:   "TestRace",
							Time:   0,
							Result: parser.FAIL,
							Output: []string{
								"test output",
								"2 0xc4200153d0",
								"==================",
								"WARNING: DATA RACE",
								"Write at 0x00c4200153d0 by goroutine 7:",
								"  race_test.TestRace.func1()",
								"      race_test.go:13 +0x3b",
								"",
								"Previous write at 0x00c4200153d0 by goroutine 6:",
								"  race_test.TestRace()",
								"      race_test.go:15 +0x136",
								"  testing.tRunner()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:657 +0x107",
								"",
								"Goroutine 7 (running) created at:",
								"  race_test.TestRace()",
								"      race_test.go:14 +0x125",
								"  testing.tRunner()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:657 +0x107",
								"",
								"Goroutine 6 (running) created at:",
								"  testing.(*T).Run()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:697 +0x543",
								"  testing.runTests.func1()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:882 +0xaa",
								"  testing.tRunner()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:657 +0x107",
								"  testing.runTests()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:888 +0x4e0",
								"  testing.(*M).Run()",
								"      /usr/local/Cellar/go/1.8.3/libexec/src/testing/testing.go:822 +0x1c3",
								"  main.main()",
								"      _test/_testmain.go:52 +0x20f",
								"==================",
								"testing.go:610: race detected during execution of test",
							},
						},
					},
				},
			},
		},
	},
	{
		name:       "18-coverpkg.txt",
		reportName: "18-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package1/foo",
					Time: 400,
					Tests: []*parser.Test{
						{
							Name:   "TestA",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestB",
							Time:   300,
							Result: parser.PASS,
							Output: []string{},
						},
					},
					CoveragePct: "10.0",
				},
				{
					Name: "package2/bar",
					Time: 4200,
					Tests: []*parser.Test{
						{
							Name:   "TestC",
							Time:   4200,
							Result: parser.PASS,
							Output: []string{},
						},
					},
					CoveragePct: "99.8",
				},
			},
		},
	},
	{
		name:       "19-pass.txt",
		reportName: "19-report.xml",
		report: &parser.Report{
			Packages: []parser.Package{
				{
					Name: "package/name",
					Time: 160,
					Tests: []*parser.Test{
						{
							Name:   "TestZ",
							Time:   60,
							Result: parser.PASS,
							Output: []string{},
						},
						{
							Name:   "TestA",
							Time:   100,
							Result: parser.PASS,
							Output: []string{},
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

		report, err := parser.Parse(file, testCase.packageName)
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
					t.Errorf("Test.Result == %d, want %d", test.Result, expTest.Result)
				}

				testOutput := strings.Join(test.Output, "\n")
				expTestOutput := strings.Join(expTest.Output, "\n")
				if testOutput != expTestOutput {
					t.Errorf("Test.Output (%s) ==\n%s\n, want\n%s", test.Name, testOutput, expTestOutput)
				}
			}
			if pkg.CoveragePct != expPkg.CoveragePct {
				t.Errorf("Package.CoveragePct == %s, want %s", pkg.CoveragePct, expPkg.CoveragePct)
			}
		}
	}
}

func TestJUnitFormatter(t *testing.T) {
	testJUnitFormatter(t, "")
}

func TestVersionFlag(t *testing.T) {
	testJUnitFormatter(t, "custom-version")
}

func testJUnitFormatter(t *testing.T, goVersion string) {
	for _, testCase := range testCases {
		report, err := loadTestReport(testCase.reportName, goVersion)
		if err != nil {
			t.Fatal(err)
		}

		var junitReport bytes.Buffer

		if err = JUnitReportXML(testCase.report, testCase.noXMLHeader, goVersion, &junitReport); err != nil {
			t.Fatal(err)
		}

		if string(junitReport.Bytes()) != report {
			t.Fatalf("Fail: %s Report xml ==\n%s, want\n%s", testCase.name, string(junitReport.Bytes()), report)
		}
	}
}

func loadTestReport(name, goVersion string) (string, error) {
	contents, err := ioutil.ReadFile("tests/" + name)
	if err != nil {
		return "", err
	}

	if goVersion == "" {
		// if goVersion is not specified, default to runtime version
		goVersion = runtime.Version()
	}

	// replace value="1.0" With actual version
	report := strings.Replace(string(contents), `value="1.0"`, fmt.Sprintf(`value="%s"`, goVersion), -1)

	return report, nil
}
