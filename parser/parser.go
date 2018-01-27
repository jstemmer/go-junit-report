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
	TestSuites []*TestSuite
}

// TestSuite contains the test results of a single package.
type TestSuite struct {
	Name  string
	Time  int
	Tests []*Test
}

// Test contains the results of a single test.
type Test struct {
	Name     string
	Time     int
	Result   Result
	Filename string
	Output   []string
}

var (
	resultMap = map[string]Result{
		// "PASSED!"                     PASS,  // Assertion result of "PASS" is weaker than "FAIL".
		"marking it as not failed":      PASS,
		"FATAL ERROR!":                  FAIL,
		"ERROR!":                        FAIL,
		"TEST CASE FAILED!":             FAIL,
		"Marking it as failed!":         FAIL,
		"marking it as failed!":         FAIL,
		"Test case exceeded time limit": FAIL,
	}
	regexResult      = regexp.MustCompile(`^.*(` + strings.Join(keys(resultMap), "|") + `).*$`)
	regexTestSuite   = regexp.MustCompile(`^TEST SUITE: (.*)$`)
	regexFileName    = regexp.MustCompile(`^(.*)\(.+\)$`)
	regexTimeAndName = regexp.MustCompile(`^(\d+.\d{6}) s:\s+(.*)$`)
)

// Parse parses go test output from reader r and returns a report with the
// results.
func Parse(r io.Reader) (*Report, error) {
	reader := bufio.NewReader(r)

	report := &Report{make([]*TestSuite, 0)}

	// current test result
	current := &Test{
		Name:     "",
		Result:   PASS,
		Filename: "",
		Output:   make([]string, 0),
	}

	// current TestSuite name
	var testSuite string

	// test result separator
	nextToSeparator := false

	// found TestSuite name
	foundTestSuite := false

	// found end of TestCase
	completedTestCase := false

	// parse lines
	for {
		l, _, err := reader.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		line := string(l)

		if strings.HasPrefix(line, "==========") {
			// The file name is written next to the separator.
			nextToSeparator = true

			if completedTestCase {
				// If not found any TestSuite.Name by this line, use filename as TestSuite.Name.
				// This TestSuite is for TestCases right under the file.
				if !foundTestSuite {
					testSuite = current.Filename
				}
				// consumed TestSuite name.
				foundTestSuite = false

				s := findTestSuite(report.TestSuites, testSuite)
				if s == nil {
					// If not created TestSuite of file, create new TestSuite.
					s = &TestSuite{
						Name:  testSuite,
						Tests: make([]*Test, 0),
					}
					report.TestSuites = append(report.TestSuites, s)
				}

				s.Tests = append(s.Tests, current)
				s.Time += current.Time

				// prepares for next TestCase
				current = &Test{
					Name:     "",
					Result:   PASS,
					Filename: "",
					Output:   make([]string, 0),
				}
				completedTestCase = false
			}

		} else if nextToSeparator {
			nextToSeparator = false
			if matches := regexFileName.FindStringSubmatch(line); len(matches) == 2 {
				current.Filename = matches[1]
			}

		} else if matches := regexTestSuite.FindStringSubmatch(line); len(matches) == 2 {
			foundTestSuite = true
			testSuite = matches[1]

		} else if matches := regexResult.FindStringSubmatch(line); len(matches) == 2 {
			// Update test result.
			current.Result = resultMap[matches[1]]

		} else if matches := regexTimeAndName.FindStringSubmatch(line); len(matches) == 3 {
			// Only the time stamp corresponds to the TestCase 1 to 1.
			// Therefore, if find time stamp, judge that current testCase is complete.
			completedTestCase = true

			// Update test time and name.
			current.Time = parseTime(matches[1])
			current.Name = matches[2]
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

func findTestSuite(suites []*TestSuite, name string) *TestSuite {
	for _, s := range suites {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func keys(m map[string]Result) []string {
	i := 0
	ks := make([]string, len(m))
	for k := range m {
		ks[i] = k
		i++
	}
	return ks
}

// Failures counts the number of failed tests in this report
func (r *Report) Failures() int {
	count := 0

	for _, p := range r.TestSuites {
		for _, t := range p.Tests {
			if t.Result == FAIL {
				count++
			}
		}
	}

	return count
}

// Tests counts the number of tests in this report
func (r *Report) Tests() int {
	count := 0

	for _, p := range r.TestSuites {
		count += len(p.Tests)
	}

	return count
}

// Time counts the total times in this report
func (r *Report) Time() int {
	time := 0

	for _, p := range r.TestSuites {
		for _, t := range p.Tests {
			time += t.Time
		}
	}

	return time
}
