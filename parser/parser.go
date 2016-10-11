package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
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
	regexResult   = regexp.MustCompile(`^(ok|FAIL)\s+(.+)\s(\d+\.\d+)s(?:\s+coverage:\s+(\d+\.\d+)%\s+of\s+statements)?$`)
	regexColors   = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// Parse parses go test output from reader r and returns a report with the
// results. An optional pkgName can be given, which is used in case a package
// result line is missing.
func Parse(r io.Reader, pkgName string) (*Report, error) {
	reader := bufio.NewReader(r)

	report := &Report{make([]Package, 0)}

	// keep track of tests we find
	var tests []*Test
	running := map[string][]*Test{}

	// test for last seen result line
	var lastResult *Test

	// sum of tests' time, use this if current test has no result line (when it is compiled test)
	testsTime := 0

	// coverage percentage report for current package
	var coveragePct string

	// parse lines
	for {
		l, _, err := reader.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// Remove ANSI colors
		line := string(regexColors.ReplaceAll(l, []byte{}))

		if strings.HasPrefix(line, "=== RUN ") {
			// new test
			name := strings.TrimSpace(line[8:])
			test := &Test{
				Name:   name,
				Result: FAIL,
				Output: make([]string, 0),
			}
			if len(running[test.Name]) > 0 {
				fmt.Fprintf(os.Stderr, "multiple tests with name %v running concurrently\n", test.Name)
			}
			running[test.Name] = append(running[test.Name], test)
			tests = append(tests, test)
		} else if matches := regexResult.FindStringSubmatch(line); len(matches) == 5 {
			if matches[4] != "" {
				coveragePct = matches[4]
			}

			if len(running) > 0 {
				// `go test` should ensure that individual packets' tests' output are serialized,
				// so there should be no running tests at package boundaries.
				fmt.Fprintf(os.Stderr, "%v tests still running after package\n", len(running))
			}

			// all tests in this package are finished
			report.Packages = append(report.Packages, Package{
				Name:        matches[2],
				Time:        parseTime(matches[3]),
				Tests:       tests,
				CoveragePct: coveragePct,
			})

			tests = nil
			running = map[string][]*Test{}
			coveragePct = ""
			testsTime = 0
			lastResult = nil
		} else if matches := regexStatus.FindStringSubmatch(line); len(matches) == 4 {
			name := matches[2]
			if len(running[name]) == 0 {
				fmt.Fprintf(os.Stderr, "Got test result for %v but no tests are running\n", name)
				lastResult = &Test{
					Name:   name,
					Result: FAIL,
					Output: make([]string, 0),
				}
			} else {
				lastResult = running[name][0]
				if len(running[name]) > 1 {
					running[name] = running[name][1:]
				} else {
					delete(running, name)
				}
			}

			// test status
			if matches[1] == "PASS" {
				lastResult.Result = PASS
			} else if matches[1] == "SKIP" {
				lastResult.Result = SKIP
			} else {
				lastResult.Result = FAIL
			}

			lastResult.Name = matches[2]
			testTime := parseTime(matches[3]) * 10
			lastResult.Time = testTime
			testsTime += testTime
			lastResult = lastResult
		} else if matches := regexCoverage.FindStringSubmatch(line); len(matches) == 2 {
			coveragePct = matches[1]
		} else if strings.HasPrefix(line, "\t") {
			// test output
			if lastResult == nil {
				continue
			}
			lastResult.Output = append(lastResult.Output, line[1:])
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
