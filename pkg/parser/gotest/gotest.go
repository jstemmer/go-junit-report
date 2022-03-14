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
	// regexBenchmark captures 3-5 groups: benchmark name, number of times ran, ns/op (with or without decimal), MB/sec (optional), B/op (optional), and allocs/op (optional).
	regexBenchmark = regexp.MustCompile(`^(Benchmark[^ -]+)(?:-\d+\s+|\s+)(\d+)\s+(\d+|\d+\.\d+)\sns\/op(?:\s+(\d+|\d+\.\d+)\sMB\/s)?(?:\s+(\d+)\sB\/op)?(?:\s+(\d+)\sallocs/op)?`)
	regexCoverage  = regexp.MustCompile(`^coverage:\s+(\d+|\d+\.\d+)%\s+of\s+statements(?:\sin\s(.+))?$`)
	regexEndTest   = regexp.MustCompile(`((?:    )*)--- (PASS|FAIL|SKIP): ([^ ]+) \((\d+\.\d+)(?: seconds|s)\)`)
	regexStatus    = regexp.MustCompile(`^(PASS|FAIL|SKIP)$`)
	regexSummary   = regexp.MustCompile(`` +
		// 1: result
		`^(\?|ok|FAIL)` +
		// 2: package name
		`\s+([^ \t]+)` +
		// 3: duration (optional)
		`(?:\s+(\d+\.\d+)s)?` +
		// 4: cached indicator (optional)
		`(?:\s+(\(cached\)))?` +
		// 5: [status message] (optional)
		`(?:\s+(\[[^\]]+\]))?` +
		// 6: coverage percentage (optional)
		// 7: coverage package list (optional)
		`(?:\s+coverage:\s+(\d+\.\d+)%\sof\sstatements(?:\sin\s(.+))?)?$`)
)

// Option defines options that can be passed to gotest.New.
type Option func(*Parser)

// PackageName sets the default package name to use when it cannot be
// determined from the test output.
func PackageName(name string) Option {
	return func(p *Parser) {
		p.packageName = name
	}
}

// New returns a new Go test output parser.
func New(options ...Option) *Parser {
	p := &Parser{}
	for _, option := range options {
		option(p)
	}
	return p
}

// Parser is a Go test output Parser.
type Parser struct {
	packageName string

	events []gtr.Event
}

// Parse parses Go test output from the given io.Reader r.
func (p *Parser) Parse(r io.Reader) ([]gtr.Event, error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		p.parseLine(s.Text())
	}
	return p.events, s.Err()
}

func (p *Parser) parseLine(line string) {
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
	} else if matches := regexSummary.FindStringSubmatch(line); len(matches) == 8 {
		p.summary(matches[1], matches[2], matches[3], matches[4], matches[5], matches[6], matches[7])
	} else if matches := regexCoverage.FindStringSubmatch(line); len(matches) == 3 {
		p.coverage(matches[1], matches[2])
	} else if matches := regexBenchmark.FindStringSubmatch(line); len(matches) == 7 {
		p.benchmark(matches[1], matches[2], matches[3], matches[4], matches[5], matches[6])
	} else if strings.HasPrefix(line, "# ") {
		// TODO(jstemmer): this should just be output; we should detect build output when building report
		fields := strings.Fields(strings.TrimPrefix(line, "# "))
		if len(fields) == 1 || len(fields) == 2 {
			p.buildOutput(fields[0])
		} else {
			p.output(line)
		}
	} else {
		p.output(line)
	}
}

func (p *Parser) add(event gtr.Event) {
	p.events = append(p.events, event)
}

func (p *Parser) runTest(name string) {
	p.add(gtr.Event{Type: "run_test", Name: name})
}

func (p *Parser) pauseTest(name string) {
	p.add(gtr.Event{Type: "pause_test", Name: name})
}

func (p *Parser) contTest(name string) {
	p.add(gtr.Event{Type: "cont_test", Name: name})
}

func (p *Parser) endTest(line, indent, result, name, duration string) {
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

func (p *Parser) status(result string) {
	p.add(gtr.Event{Type: "status", Result: result})
}

func (p *Parser) summary(result, name, duration, cached, status, covpct, packages string) {
	p.add(gtr.Event{
		Type:        "summary",
		Result:      result,
		Name:        name,
		Duration:    parseSeconds(duration),
		Data:        strings.TrimSpace(cached + " " + status),
		CovPct:      parseFloat(covpct),
		CovPackages: parsePackages(packages),
	})
}

func (p *Parser) coverage(percent, packages string) {
	p.add(gtr.Event{
		Type:        "coverage",
		CovPct:      parseFloat(percent),
		CovPackages: parsePackages(packages),
	})
}

func (p *Parser) benchmark(name, iterations, nsPerOp, mbPerSec, bytesPerOp, allocsPerOp string) {
	p.add(gtr.Event{
		Type:        "benchmark",
		Name:        name,
		Iterations:  parseInt(iterations),
		NsPerOp:     parseFloat(nsPerOp),
		MBPerSec:    parseFloat(mbPerSec),
		BytesPerOp:  parseInt(bytesPerOp),
		AllocsPerOp: parseInt(allocsPerOp),
	})
}

func (p *Parser) buildOutput(packageName string) {
	p.add(gtr.Event{
		Type: "build_output",
		Name: packageName,
	})
}

func (p *Parser) output(line string) {
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

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	// ignore error
	pct, _ := strconv.ParseFloat(s, 64)
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
