package main

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Result int

const (
	PASS Result = iota
	FAIL
	SKIP
)

type Report struct {
	Packages []Package
}

type Package struct {
	Name  string
	Time  int
	Tests []Test
}

type Test struct {
	Name   string
	Time   int
	Result Result
	Output []string
}

var (
	regexStatus = regexp.MustCompile(`^--- (PASS|FAIL|SKIP): (.+) \((\d+\.\d+)(?: seconds|s)\)$`)
	regexResult = regexp.MustCompile(`^(ok|FAIL)\s+(.+)\s(\d+\.\d+)s$`)
)

func Parse(r io.Reader, specifiedPackageName string) (*Report, error) {
	reader := bufio.NewReader(r)

	report := &Report{make([]Package, 0)}

	// keep track of tests we find
	tests := make([]Test, 0)

	// sum of tests' time, use this if current test has no result line (when it is compiled test)
	testsTime := 0

	// current test
	var test *Test

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
			// start of a new test
			if test != nil {
				tests = append(tests, *test)
			}

			test = &Test{
				Name:   line[8:],
				Result: FAIL,
				Output: make([]string, 0),
			}
		} else if matches := regexResult.FindStringSubmatch(line); len(matches) == 4 {
			// all tests in this package are finished
			if test != nil {
				tests = append(tests, *test)
				test = nil
			}

			report.Packages = append(report.Packages, Package{
				Name:  matches[2],
				Time:  parseTime(matches[3]),
				Tests: tests,
			})

			tests = make([]Test, 0)
			testsTime = 0
		} else if test != nil {
			if matches := regexStatus.FindStringSubmatch(line); len(matches) == 4 {
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
			} else if strings.HasPrefix(line, "\t") {
				// test output
				test.Output = append(test.Output, line[1:])
			}
		}
	}

	if test != nil {
		tests = append(tests, *test)
	}

	if len(tests) > 0 { //no result line found
		report.Packages = append(report.Packages, Package{
			Name:  specifiedPackageName,
			Time:  testsTime,
			Tests: tests,
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
