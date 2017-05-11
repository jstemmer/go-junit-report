package parser

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Result represents a test result.
type Result int

// Test result constants
const (
	PASS Result = iota
	FAIL
	SKIP
)

// Report is a collection of package tests.
type Report struct {
	Packages []Package
}

// Package contains the test results of a single package.
type Package struct {
	Name        string
	Time        int
	Tests       []*Test
	CoveragePct string
}

// Test contains the results of a single test.
type Test struct {
	Name   string
	Time   int
	Result Result
	Output []string
}

var (
	regexStatus   = regexp.MustCompile(`^\s*--- (PASS|FAIL|SKIP): (.+) \((\d+\.\d+)(?: seconds|s)\)$`)
	regexCoverage = regexp.MustCompile(`^coverage:\s+(\d+\.\d+)%\s+of\s+statements$`)
	regexResult   = regexp.MustCompile(`^(ok|FAIL)\s+([^ ]+)\s+(?:(\d+\.\d+)s|(\[\w+ failed]))(?:\s+coverage:\s+(\d+\.\d+)%\sof\sstatements)?$`)
	regexOutput   = regexp.MustCompile(`(    )*\t(.*)`)
)

// Parse parses go test output from reader r and returns a report with the
// results. An optional pkgName can be given, which is used in case a package
// result line is missing.
func Parse(r io.Reader, pkgName string) (*Report, error) {
	reader := bufio.NewReader(r)

	report := &Report{make([]Package, 0)}

	// keep track of tests we find
	var tests []*Test

	// sum of tests' time, use this if current test has no result line (when it is compiled test)
	testsTime := 0

	// current test
	var cur string

	// coverage percentage report for current package
	var coveragePct string

	// stores mapping between package name and output of build failures
	var packageCaptures = map[string][]string{}

	// the name of the package which it's build failure output is being captured
	var capturedPackage string

	// capture any non-test output
	var buffer []string

	// parse lines
	for {
		l, _, err := reader.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		line := string(l)

		if strings.HasPrefix(line, "=== RUN ") {
			// new test
			cur = strings.TrimSpace(line[8:])
			tests = append(tests, &Test{
				Name:   cur,
				Result: FAIL,
				Output: make([]string, 0),
			})

			// clear the current build package, so output lines won't be added to that build
			capturedPackage = ""
		} else if matches := regexResult.FindStringSubmatch(line); len(matches) == 6 {
			if matches[5] != "" {
				coveragePct = matches[5]
			}
			if strings.HasSuffix(matches[4], "failed]") {
				// the build of the package failed, inject a dummy test into the package
				// which indicate about the failure and contain the failure description.
				tests = append(tests, &Test{
					Name:   matches[4],
					Result: FAIL,
					Output: packageCaptures[matches[2]],
				})
			} else if matches[1] == "FAIL" && len(tests) == 0 && len(buffer) > 0 {
				// This package didn't have any tests, but it failed with some
				// output. Create a dummy test with the output.
				tests = append(tests, &Test{
					Name:   "Failure",
					Result: FAIL,
					Output: buffer,
				})
				buffer = buffer[0:0]
			}

			// all tests in this package are finished
			report.Packages = append(report.Packages, Package{
				Name:        matches[2],
				Time:        parseTime(matches[3]),
				Tests:       tests,
				CoveragePct: coveragePct,
			})

			buffer = buffer[0:0]
			tests = make([]*Test, 0)
			coveragePct = ""
			cur = ""
			testsTime = 0
		} else if matches := regexStatus.FindStringSubmatch(line); len(matches) == 4 {
			cur = matches[2]
			test := findTest(tests, cur)
			if test == nil {
				continue
			}

			// test status
			if matches[1] == "PASS" {
				test.Result = PASS
			} else if matches[1] == "SKIP" {
				test.Result = SKIP
			} else {
				test.Result = FAIL
			}

			test.Name = matches[2]
			testTime := parseTime(matches[3]) * 10
			test.Time = testTime
			testsTime += testTime
		} else if matches := regexCoverage.FindStringSubmatch(line); len(matches) == 2 {
			coveragePct = matches[1]
		} else if matches := regexOutput.FindStringSubmatch(line); capturedPackage == "" && len(matches) == 3 {
			// Sub-tests start with one or more series of 4-space indents, followed by a hard tab,
			// followed by the test output
			// Top-level tests start with a hard tab.
			test := findTest(tests, cur)
			if test == nil {
				continue
			}
			test.Output = append(test.Output, matches[2])
		} else if strings.HasPrefix(line, "# ") {
			// indicates a capture of build output of a package. set the current build package.
			capturedPackage = line[2:]
		} else if capturedPackage != "" {
			// current line is build failure capture for the current built package
			packageCaptures[capturedPackage] = append(packageCaptures[capturedPackage], line)
		} else {
			// buffer anything else that we didn't recognize
			buffer = append(buffer, line)
		}
	}

	if len(tests) > 0 {
		// no result line found
		report.Packages = append(report.Packages, Package{
			Name:        pkgName,
			Time:        testsTime,
			Tests:       tests,
			CoveragePct: coveragePct,
		})
	}

	return report, nil
}

func parseTime(time string) int {
	t, err := strconv.Atoi(strings.Replace(time, ".", "", -1))
	if err != nil {
		return 0
	}
	return t
}

func findTest(tests []*Test, name string) *Test {
	for i := len(tests) - 1; i >= 0; i-- {
		if tests[i].Name == name {
			return tests[i]
		}
	}
	return nil
}

// Failures counts the number of failed tests in this report
func (r *Report) Failures() int {
	count := 0

	for _, p := range r.Packages {
		for _, t := range p.Tests {
			if t.Result == FAIL {
				count++
			}
		}
	}

	return count
}
