package parser

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	regexStatus   = regexp.MustCompile(`--- (PASS|FAIL|SKIP): (.+) \((\d+\.\d+)(?: seconds|s)\)`)
	regexCoverage = regexp.MustCompile(`^coverage:\s+(\d+\.\d+)%\s+of\s+statements(?:\sin\s.+)?$`)
	regexResult   = regexp.MustCompile(`^(ok|FAIL)\s+([^ ]+)\s+(?:(\d+\.\d+)s|(\[\w+ failed]))(?:\s+coverage:\s+(\d+\.\d+)%\sof\sstatements(?:\sin\s.+)?)?$`)
	regexOutput   = regexp.MustCompile(`(    )*\t(.*)`)
	regexSummary  = regexp.MustCompile(`^(PASS|FAIL|SKIP)$`)
)

type textParser struct {
	cur             string
	capturedPackage string
	tests           []*Test
	report          *Report
	testsTime       int
	seenSummary     bool
	coveragePct     string
	packageCaptures map[string][]string
	buffers         map[string][]string
	pkgName         string
}

func newTextParser(pkgName string) parser {
	return &textParser{
		report:          &Report{make([]Package, 0)},
		pkgName:         pkgName,
		packageCaptures: make(map[string][]string),
		buffers:         make(map[string][]string),
	}
}

func (p *textParser) IngestLine(line string) error {
	if strings.HasPrefix(line, "=== RUN ") {
		// new test
		p.cur = strings.TrimSpace(line[8:])
		p.tests = append(p.tests, &Test{
			Name:   p.cur,
			Result: FAIL,
			Output: make([]string, 0),
		})

		// clear the current build package, so output lines won't be added to that build
		p.capturedPackage = ""
		p.seenSummary = false
	} else if strings.HasPrefix(line, "=== PAUSE ") {
		return nil
	} else if strings.HasPrefix(line, "=== CONT ") {
		p.cur = strings.TrimSpace(line[8:])
		return nil
	} else if matches := regexResult.FindStringSubmatch(line); len(matches) == 6 {
		if matches[5] != "" {
			p.coveragePct = matches[5]
		}
		if strings.HasSuffix(matches[4], "failed]") {
			// the build of the package failed, inject a dummy test into the package
			// which indicate about the failure and contain the failure description.
			p.tests = append(p.tests, &Test{
				Name:   matches[4],
				Result: FAIL,
				Output: p.packageCaptures[matches[2]],
			})
		} else if matches[1] == "FAIL" && len(p.tests) == 0 && len(p.buffers[p.cur]) > 0 {
			// This package didn't have any tests, but it failed with some
			// output. Create a dummy test with the output.
			p.tests = append(p.tests, &Test{
				Name:   "Failure",
				Result: FAIL,
				Output: p.buffers[p.cur],
			})
			p.buffers[p.cur] = p.buffers[p.cur][0:0]
		}

		// all tests in this package are finished
		p.report.Packages = append(p.report.Packages, Package{
			Name:        matches[2],
			Time:        parseTime(matches[3]),
			Tests:       p.tests,
			CoveragePct: p.coveragePct,
		})

		p.buffers[p.cur] = p.buffers[p.cur][0:0]
		p.tests = make([]*Test, 0)
		p.coveragePct = ""
		p.cur = ""
		p.testsTime = 0
	} else if matches := regexStatus.FindStringSubmatch(line); len(matches) == 4 {
		p.cur = matches[2]
		test := findTest(p.tests, p.cur)
		if test == nil {
			return nil
		}

		// test status
		if matches[1] == "PASS" {
			test.Result = PASS
		} else if matches[1] == "SKIP" {
			test.Result = SKIP
		} else {
			test.Result = FAIL
		}
		test.Output = p.buffers[p.cur]

		test.Name = matches[2]
		testTime := parseTime(matches[3]) * 10
		test.Time = testTime
		p.testsTime += testTime
	} else if matches := regexCoverage.FindStringSubmatch(line); len(matches) == 2 {
		p.coveragePct = matches[1]
	} else if matches := regexOutput.FindStringSubmatch(line); p.capturedPackage == "" && len(matches) == 3 {
		// Sub-tests start with one or more series of 4-space indents, followed by a hard tab,
		// followed by the test output
		// Top-level tests start with a hard tab.
		test := findTest(p.tests, p.cur)
		if test == nil {
			return nil
		}
		test.Output = append(test.Output, matches[2])
	} else if strings.HasPrefix(line, "# ") {
		// indicates a capture of build output of a package. set the current build package.
		p.capturedPackage = line[2:]
	} else if p.capturedPackage != "" {
		// current line is build failure capture for the current built package
		p.packageCaptures[p.capturedPackage] = append(p.packageCaptures[p.capturedPackage], line)
	} else if regexSummary.MatchString(line) {
		// don't store any output after the summary
		p.seenSummary = true
	} else if !p.seenSummary {
		// buffer anything else that we didn't recognize
		p.buffers[p.cur] = append(p.buffers[p.cur], line)
	}

	return nil
}

func (p *textParser) Report() (*Report, error) {
	if len(p.tests) > 0 {
		// no result line found
		p.report.Packages = append(p.report.Packages, Package{
			Name:        p.pkgName,
			Time:        p.testsTime,
			Tests:       p.tests,
			CoveragePct: p.coveragePct,
		})
	}

	return p.report, nil
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
