// Package gotest is a standard Go test output parser.
package gotest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/v2/gtr"
)

var (
	// regexBenchInfo captures 3-5 groups: benchmark name, number of times ran, ns/op (with or without decimal), MB/sec (optional), B/op (optional), and allocs/op (optional).
	regexBenchmark    = regexp.MustCompile(`^(Benchmark[^ -]+)$`)
	regexBenchSummary = regexp.MustCompile(`^(Benchmark[^ -]+)(?:-\d+\s+|\s+)(\d+)\s+(\d+|\d+\.\d+)\sns\/op(?:\s+(\d+|\d+\.\d+)\sMB\/s)?(?:\s+(\d+)\sB\/op)?(?:\s+(\d+)\sallocs/op)?`)
	regexCoverage     = regexp.MustCompile(`^coverage:\s+(\d+|\d+\.\d+)%\s+of\s+statements(?:\sin\s(.+))?$`)
	regexEndBenchmark = regexp.MustCompile(`^--- (BENCH|FAIL|SKIP): (Benchmark[^ -]+)(?:-\d+)?$`)
	regexEndTest      = regexp.MustCompile(`((?:    )*)--- (PASS|FAIL|SKIP): ([^ ]+) \((\d+\.\d+)(?: seconds|s)\)`)
	regexStatus       = regexp.MustCompile(`^(PASS|FAIL|SKIP)$`)
	regexSummary      = regexp.MustCompile(`` +
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

// PackageName is an Option that sets the default package name to use when it
// cannot be determined from the test output.
func PackageName(name string) Option {
	return func(p *Parser) {
		p.packageName = name
	}
}

// TimestampFunc is an Option that sets the timestamp function that is used to
// determine the current time when creating the Report. This can be used to
// override the default behaviour of using time.Now().
func TimestampFunc(f func() time.Time) Option {
	return func(p *Parser) {
		p.timestampFunc = f
	}
}

// SubtestMode configures how Go subtests should be handled by the parser.
type SubtestMode string

const (
	// SubtestModeDefault is the default subtest mode. It treats tests with
	// subtests as any other tests.
	SubtestModeDefault SubtestMode = ""

	// IgnoreParentResults ignores test results for tests with subtests. Use
	// this mode if you use subtest parents for common setup/teardown, but are
	// not interested in counting them as failed tests. Ignoring their results
	// still preserves these tests and their captured output in the report.
	IgnoreParentResults SubtestMode = "ignore-parent-results"

	// ExcludeParents excludes tests that contain subtests from the report.
	// Note that the subtests themselves are not removed. Use this mode if you
	// use subtest parents for common setup/teardown, but are not actually
	// interested in their presence in the created report. If output was
	// captured for tests that are removed, the output is preserved in the
	// global report output.
	ExcludeParents SubtestMode = "exclude-parents"
)

// ParseSubtestMode returns a SubtestMode for the given string.
func ParseSubtestMode(in string) (SubtestMode, error) {
	switch in {
	case string(IgnoreParentResults):
		return IgnoreParentResults, nil
	case string(ExcludeParents):
		return ExcludeParents, nil
	default:
		return SubtestModeDefault, fmt.Errorf("unknown subtest mode: %v", in)
	}
}

// SetSubtestMode is an Option to change how the parser handles tests with
// subtests. See the documentation for the individual SubtestModes for more
// information.
func SetSubtestMode(mode SubtestMode) Option {
	return func(p *Parser) {
		p.subtestMode = mode
	}
}

const (
	// maxLineSize is the maximum amount of bytes we'll read for a single line.
	// Lines longer than maxLineSize will be truncated.
	maxLineSize = 4 * 1024 * 1024
)

// Parser is a Go test output Parser.
type Parser struct {
	packageName string
	subtestMode SubtestMode

	timestampFunc func() time.Time

	events []Event
}

// NewParser returns a new Go test output parser.
func NewParser(options ...Option) *Parser {
	p := &Parser{}
	for _, option := range options {
		option(p)
	}
	return p
}

// Parse parses Go test output from the given io.Reader r and returns
// gtr.Report.
func (p *Parser) Parse(r io.Reader) (gtr.Report, error) {
	p.events = nil
	s := bufio.NewReader(r)
	for {
		line, isPrefix, err := s.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return gtr.Report{}, err
		}

		if !isPrefix {
			p.parseLine(string(line))
			continue
		}

		// Line is incomplete, keep reading until we reach the end of the line.
		var buf bytes.Buffer
		buf.Write(line) // ignore err, always nil
		for isPrefix {
			line, isPrefix, err = s.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				return gtr.Report{}, err
			}

			if buf.Len() >= maxLineSize {
				// Stop writing to buf if we exceed maxLineSize. We continue
				// reading however to make sure we consume the entire line.
				continue
			}

			buf.Write(line) // ignore err, always nil
		}

		if buf.Len() > maxLineSize {
			buf.Truncate(maxLineSize)
		}

		// Lines that exceed bufio.MaxScanTokenSize are not expected to contain
		// any relevant test infrastructure output, so instead of parsing them
		// we treat them as regular output to increase performance.
		//
		// Parser used a bufio.Scanner in the past, which only supported
		// reading lines up to bufio.MaxScanTokenSize in length. Since this
		// turned out to be fine in almost all cases, it seemed an appropriate
		// value to use to decide whether or not to attempt parsing this line.
		if buf.Len() > bufio.MaxScanTokenSize {
			p.output(buf.String())
		} else {
			p.parseLine(buf.String())
		}
	}
	return p.report(p.events), nil
}

// report generates a gtr.Report from the given list of events.
func (p *Parser) report(events []Event) gtr.Report {
	rb := newReportBuilder()
	rb.packageName = p.packageName
	rb.subtestMode = p.subtestMode
	if p.timestampFunc != nil {
		rb.timestampFunc = p.timestampFunc
	}
	for _, ev := range events {
		switch ev.Type {
		case "run_test":
			rb.CreateTest(ev.Name)
		case "pause_test":
			rb.PauseTest(ev.Name)
		case "cont_test":
			rb.ContinueTest(ev.Name)
		case "end_test":
			rb.EndTest(ev.Name, ev.Result, ev.Duration, ev.Indent)
		case "run_benchmark":
			rb.CreateBenchmark(ev.Name)
		case "benchmark":
			rb.BenchmarkResult(ev.Name, ev.Iterations, ev.NsPerOp, ev.MBPerSec, ev.BytesPerOp, ev.AllocsPerOp)
		case "end_benchmark":
			rb.EndBenchmark(ev.Name, ev.Result)
		case "status":
			rb.End()
		case "summary":
			rb.CreatePackage(ev.Name, ev.Result, ev.Duration, ev.Data)
		case "coverage":
			rb.Coverage(ev.CovPct, ev.CovPackages)
		case "build_output":
			rb.CreateBuildError(ev.Name)
		case "output":
			rb.AppendOutput(ev.Data)
		default:
			fmt.Printf("unhandled event type: %v\n", ev.Type)
		}
	}
	return rb.Build()
}

// Events returns the events created by the parser.
func (p *Parser) Events() []Event {
	events := make([]Event, len(p.events))
	copy(events, p.events)
	return events
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
	} else if matches := regexBenchmark.FindStringSubmatch(line); len(matches) == 2 {
		p.runBench(matches[1])
	} else if matches := regexBenchSummary.FindStringSubmatch(line); len(matches) == 7 {
		p.benchSummary(matches[1], matches[2], matches[3], matches[4], matches[5], matches[6])
	} else if matches := regexEndBenchmark.FindStringSubmatch(line); len(matches) == 3 {
		p.endBench(matches[1], matches[2])
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

func (p *Parser) add(event Event) {
	p.events = append(p.events, event)
}

func (p *Parser) runTest(name string) {
	p.add(Event{Type: "run_test", Name: name})
}

func (p *Parser) pauseTest(name string) {
	p.add(Event{Type: "pause_test", Name: name})
}

func (p *Parser) contTest(name string) {
	p.add(Event{Type: "cont_test", Name: name})
}

func (p *Parser) endTest(line, indent, result, name, duration string) {
	if idx := strings.Index(line, fmt.Sprintf("%s--- %s:", indent, result)); idx > 0 {
		p.output(line[:idx])
	}
	_, n := stripIndent(indent)
	p.add(Event{
		Type:     "end_test",
		Name:     name,
		Result:   result,
		Indent:   n,
		Duration: parseSeconds(duration),
	})
}

func (p *Parser) status(result string) {
	p.add(Event{Type: "status", Result: result})
}

func (p *Parser) summary(result, name, duration, cached, status, covpct, packages string) {
	p.add(Event{
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
	p.add(Event{
		Type:        "coverage",
		CovPct:      parseFloat(percent),
		CovPackages: parsePackages(packages),
	})
}

func (p *Parser) runBench(name string) {
	p.add(Event{
		Type: "run_benchmark",
		Name: name,
	})
}

func (p *Parser) benchSummary(name, iterations, nsPerOp, mbPerSec, bytesPerOp, allocsPerOp string) {
	p.add(Event{
		Type:        "benchmark",
		Name:        name,
		Iterations:  parseInt(iterations),
		NsPerOp:     parseFloat(nsPerOp),
		MBPerSec:    parseFloat(mbPerSec),
		BytesPerOp:  parseInt(bytesPerOp),
		AllocsPerOp: parseInt(allocsPerOp),
	})
}

func (p *Parser) endBench(result, name string) {
	p.add(Event{
		Type:   "end_benchmark",
		Name:   name,
		Result: result,
	})
}

func (p *Parser) buildOutput(packageName string) {
	p.add(Event{
		Type: "build_output",
		Name: packageName,
	})
}

func (p *Parser) output(line string) {
	p.add(Event{Type: "output", Data: line})
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
