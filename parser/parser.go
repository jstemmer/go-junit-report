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
	regexResult    = regexp.MustCompile(`^.*(` + strings.Join(keys(resultMap), "|") + `).*$`)
	regexTestSuite = regexp.MustCompile(`^TEST SUITE: (.*)$`)
	regexTestCase  = regexp.MustCompile(`^(TEST CASE:  |  Scenario: )(.*)$`)
	regexFileName  = regexp.MustCompile(`^(.*)\(.+\)$`)
	regexTime      = regexp.MustCompile(`^(\d+.\d{6}) s:.*`)
)

// Parse parses go test output from reader r and returns a report with the
// results.
func Parse(r io.Reader) (*Report, error) {
	reader := bufio.NewReader(r)

	report := &Report{make([]*TestSuite, 0)}

	// current test
	var cur string

	// current test file name
	var filename string

	// current test suite name
	var testSuite string

	// test result separator
	nextToSeparator := false

	// found test suite name
	foundTestSuite := false

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

		} else if nextToSeparator {
			if matches := regexFileName.FindStringSubmatch(line); len(matches) == 2 {
				filename = matches[1]
			}
			nextToSeparator = false

		} else if matches := regexTestSuite.FindStringSubmatch(line); len(matches) == 2 {
			testSuite = matches[1]

			// If not created, create new TestSuite.
			s := findTestSuite(report.TestSuites, testSuite)
			if s == nil {
				report.TestSuites = append(report.TestSuites, &TestSuite{
					Name:  testSuite,
					Tests: make([]*Test, 0),
				})
			}

			foundTestSuite = true

		} else if matches := regexTestCase.FindStringSubmatch(line); len(matches) == 3 {
			// If not found any TestSuite by this line, use filename as TestSuite.
			if !foundTestSuite {
				testSuite = filename
			}

			// If not created TestSuite of file, create new TestSuite.
			// This TestSuite is for TestCases right under the file.
			s := findTestSuite(report.TestSuites, testSuite)
			if s == nil {
				s = &TestSuite{
					Name:  filename,
					Tests: make([]*Test, 0),
				}
				report.TestSuites = append(report.TestSuites, s)
			}

			// If not created TestCase by this line, create new TestCase.
			cur = matches[2]
			t := findTestCase(s.Tests, cur)
			if t == nil {
				s.Tests = append(s.Tests, &Test{
					Name:     cur,
					Result:   PASS,
					Filename: filename,
					Output:   make([]string, 0),
				})
			}

			// consumed
			foundTestSuite = false

		} else if matches := regexTime.FindStringSubmatch(line); len(matches) == 2 {
			// Update test time.
			s := findTestSuite(report.TestSuites, testSuite)
			t := findTestCase(s.Tests, cur)
			t.Time = parseTime(matches[1])
			s.Time += t.Time

		} else if matches := regexResult.FindStringSubmatch(line); len(matches) == 2 {
			// Update test result.
			s := findTestSuite(report.TestSuites, testSuite)
			t := findTestCase(s.Tests, cur)
			t.Result = resultMap[matches[1]]
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

func findTestCase(tests []*Test, name string) *Test {
	for _, t := range tests {
		if t.Name == name {
			return t
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
