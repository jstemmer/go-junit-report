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
	Output string
}

var (
	regexStatus = regexp.MustCompile(`^--- (PASS|FAIL): (.+) \((\d+\.\d+) seconds\)$`)
	regexResult = regexp.MustCompile(`^(ok|FAIL)\s+(.+)\s(\d+\.\d+)s$`)
)

func Parse(r io.Reader) (*Report, error) {
	reader := bufio.NewReader(r)

	report := &Report{make([]Package, 0)}

	// keep track of tests we find
	tests := make([]Test, 0)

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
		if test == nil {
			// expecting new test or package result
			if strings.HasPrefix(line, "=== RUN ") {
				test = &Test{
					Name: line[8:],
				}
			} else if matches := regexResult.FindStringSubmatch(line); len(matches) == 4 {
				report.Packages = append(report.Packages, Package{
					Name:  matches[2],
					Time:  parseTime(matches[3]),
					Tests: tests,
				})

				tests = make([]Test, 0)
			}
		} else {
			// expecting test status
			if matches := regexStatus.FindStringSubmatch(line); len(matches) == 4 {
				if matches[1] == "PASS" {
					test.Result = PASS
				} else {
					test.Result = FAIL
				}

				test.Name = matches[2]
				test.Time = parseTime(matches[3]) * 10

				tests = append(tests, *test)
				test = nil
			}
		}
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
