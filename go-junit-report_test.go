package main

import (
	"os"
	"testing"
)

type TestCase struct {
	name           string
	expectedReport Report
}

var testCases []TestCase = []TestCase{
	{
		name: "01-pass.txt",
		expectedReport: Report{
			Packages: []Package{
				{
					Name: "package/name",
					Time: 160,
					Tests: []Test{
						{
							Name:   "TestOne",
							Time:   60,
							Result: PASS,
							Output: "",
						},
						{
							Name:   "TestTwo",
							Time:   100,
							Result: PASS,
							Output: "",
						},
					},
				},
			},
		},
	},
	{
		name: "02-fail.txt",
		expectedReport: Report{
			Packages: []Package{
				{
					Name: "package/name",
					Time: 151,
					Tests: []Test{
						{
							Name:   "TestOne",
							Time:   20,
							Result: FAIL,
							Output: "",
						},
						{
							Name:   "TestTwo",
							Time:   130,
							Result: PASS,
							Output: "",
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

		report, err := Parse(file)
		if err != nil {
			t.Fatalf("error parsing: %s", err)
		}

		if report == nil {
			t.Fatalf("Report == nil")
		}

		expected := testCase.expectedReport
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

				if test.Output != expTest.Output {
					t.Errorf("Test.Output == %s, want %s", test.Output, expTest.Output)
				}
			}
		}
	}
}
