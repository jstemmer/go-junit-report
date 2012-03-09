package main

import (
	"strings"
	"testing"
)

func TestOutputPass(t *testing.T) {
	testOutputPass := `=== RUN TestOne
--- PASS: TestOne (0.06 seconds)
=== RUN TestTwo
--- PASS: TestTwo (0.10 seconds)
PASS
ok  	package/name 0.160s`

	expected := Report{
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
	}

	report, err := Parse(strings.NewReader(testOutputPass))
	if err != nil {
		t.Fatalf("error parsing: %s", err)
	}

	if report == nil {
		t.Fatalf("Report == nil")
	}

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

const testOutputFail = `=== RUN TestOne
--- FAIL: TestOne (0.02 seconds)
	file_test.go:11: Error message
	file_test.go:11: Longer
		error
		message.
=== RUN TestTwo
--- PASS: TestTwo (0.13 seconds)
FAIL
exit status 1
FAIL	package/name 0.151s`

func TestOutputFail(t *testing.T) {
	_, err := Parse(strings.NewReader(testOutputFail))
	if err != nil {
		t.Fatalf("error parsing: %s", err)
	}
}
