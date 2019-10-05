// Package gotest is a standard Go test output parser.
package gotest

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/gtr"
)

var (
	// regexBenchmark captures 3-5 groups: benchmark name, number of times ran, ns/op (with or without decimal), B/op (optional), and allocs/op (optional).
	regexBenchmark = regexp.MustCompile(`^(Benchmark[^ -]+)(?:-\d+\s+|\s+)(\d+)\s+(\d+|\d+\.\d+)\sns\/op(?:\s+(\d+)\sB\/op)?(?:\s+(\d+)\sallocs/op)?`)
	regexCoverage  = regexp.MustCompile(`^coverage:\s+(\d+|\d+\.\d+)%\s+of\s+statements(?:\sin\s(.+))?$`)
	regexEndTest   = regexp.MustCompile(`((?:    )*)--- (PASS|FAIL|SKIP): ([^ ]+) \((\d+\.\d+)(?: seconds|s)\)`)
	regexStatus    = regexp.MustCompile(`^(PASS|FAIL|SKIP)$`)
	regexSummary   = regexp.MustCompile(`^(ok|FAIL)\s+([^ ]+)\s+` + `(?:(\d+\.\d+)s|(\(cached\)|\[\w+ failed]))` + `(?:\s+coverage:\s+(\d+\.\d+)%\sof\sstatements(?:\sin\s(.+))?)?$`)
)

// Parse parses Go test output from the given io.Reader r.
func Parse(r io.Reader) ([]gtr.Event, error) {
	p := &parser{}

	s := bufio.NewScanner(r)
	for s.Scan() {
		p.parseLine(s.Text())
	}
	return p.events, s.Err()
}

type parser struct {
	events []gtr.Event
}

func (p *parser) parseLine(line string) {
	if strings.HasPrefix(line, "=== RUN ") {
		p.runTest(strings.TrimSpace(line[8:]))
	} else if strings.HasPrefix(line, "=== PAUSE ") {
		p.pauseTest(strings.TrimSpace(line[10:]))
	} else if strings.HasPrefix(line, "=== CONT ") {
		p.contTest(strings.TrimSpace(line[9:]))
	} else if matches := regexEndTest.FindStringSubmatch(line); len(matches) == 5 {
		p.endTest(line, matches[1], matches[2], matches[3], matches[4])
	} else if matches := regexStatus.FindStringSubmatch(line); len(matches) == 2 {
		p.status(matches[1])
	} else if matches := regexSummary.FindStringSubmatch(line); len(matches) == 7 {
		p.summary(matches[1], matches[2], matches[3], matches[4], matches[5], matches[6])
	} else if matches := regexCoverage.FindStringSubmatch(line); len(matches) == 3 {
		p.coverage(matches[1], matches[2])
	} else if matches := regexBenchmark.FindStringSubmatch(line); len(matches) == 6 {
		p.benchmark(matches[1], matches[2], matches[3], matches[4], matches[5])
	} else {
		p.output(line)
	}
}

func (p *parser) add(event gtr.Event) {
	p.events = append(p.events, event)
}

func (p *parser) runTest(name string) {
	p.add(gtr.Event{Type: "run_test", Name: name})
}

func (p *parser) pauseTest(name string) {
	p.add(gtr.Event{Type: "pause_test", Name: name})
}

func (p *parser) contTest(name string) {
	p.add(gtr.Event{Type: "cont_test", Name: name})
}

func (p *parser) endTest(line, indent, result, name, duration string) {
	if idx := strings.Index(line, fmt.Sprintf("%s--- %s:", indent, result)); idx > 0 {
		p.output(line[:idx])
	}
	_, n := stripIndent(indent)
	p.add(gtr.Event{
		Type:     "end_test",
		Name:     name,
		Result:   result,
		Indent:   n,
		Duration: parseSeconds(duration),
	})
}

func (p *parser) status(result string) {
	p.add(gtr.Event{Type: "status", Result: result})
}

func (p *parser) summary(result, name, duration, data, covpct, packages string) {
	p.add(gtr.Event{
		Type:        "summary",
		Result:      result,
		Name:        name,
		Duration:    parseSeconds(duration),
		Data:        data,
		CovPct:      parseCoverage(covpct),
		CovPackages: parsePackages(packages),
	})
}

func (p *parser) coverage(percent, packages string) {
	p.add(gtr.Event{
		Type:        "coverage",
		CovPct:      parseCoverage(percent),
		CovPackages: parsePackages(packages),
	})
}

func (p *parser) benchmark(name, iterations, timePerOp, bytesPerOp, allocsPerOp string) {
	p.add(gtr.Event{
		Type:        "benchmark",
		Name:        name,
		Iterations:  parseInt(iterations),
		NsPerOp:     parseInt(timePerOp),
		BytesPerOp:  parseInt(bytesPerOp),
		AllocsPerOp: parseInt(allocsPerOp),
	})
}

func (p *parser) output(line string) {
	p.add(gtr.Event{Type: "output", Data: line})
}

func parseSeconds(s string) time.Duration {
	if s == "" {
		return time.Duration(0)
	}
	// ignore error
	d, _ := time.ParseDuration(s + "s")
	return d
}

func parseCoverage(percent string) float64 {
	if percent == "" {
		return 0
	}
	// ignore error
	pct, _ := strconv.ParseFloat(percent, 64)
	return pct
}

func parsePackages(pkgList string) []string {
	if len(pkgList) == 0 {
		return nil
	}
	return strings.Split(pkgList, ", ")
}

func parseInt(s string) int64 {
	// ignore error
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func stripIndent(line string) (string, int) {
	var indent int
	for indent = 0; strings.HasPrefix(line, "    "); indent++ {
		line = line[4:]
	}
	return line, indent
}
