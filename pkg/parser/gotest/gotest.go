// Package gotest is a standard Go test output parser.
package gotest

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Event struct {
	Type string

	Id       int
	Name     string
	Result   string
	Duration time.Duration
	Data     string
	Indent   int
	CovPct   float64
	Hints    map[string]string
}

var (
	regexEndTest = regexp.MustCompile(`--- (PASS|FAIL|SKIP): ([^ ]+) \((\d+\.\d+)(?: seconds|s)\)`)
	regexStatus  = regexp.MustCompile(`^(PASS|FAIL|SKIP)$`)
	regexSummary = regexp.MustCompile(`^(ok|FAIL)\s+([^ ]+)\s+` +
		`(?:(\d+\.\d+)s|\(cached\)|(\[\w+ failed]))` +
		`(?:\s+coverage:\s+(\d+\.\d+)%\sof\sstatements(?:\sin\s.+)?)?$`)
	regexCoverage = regexp.MustCompile(`^coverage:\s+(\d+\.\d+)%\s+of\s+statements(?:\sin\s.+)?$`)
)

// Parse parses Go test output from the given io.Reader r.
func Parse(r io.Reader) ([]Event, error) {
	p := &parser{}

	s := bufio.NewScanner(r)
	for s.Scan() {
		p.parseLine(s.Text())
	}
	if s.Err() != nil {
		return nil, s.Err()
	}

	return p.events, nil
}

type parser struct {
	id     int
	events []Event
}

func (p *parser) parseLine(line string) {
	if strings.HasPrefix(line, "=== RUN ") {
		p.runTest(line[8:])
	} else if strings.HasPrefix(line, "=== PAUSE ") {
	} else if strings.HasPrefix(line, "=== CONT ") {
	} else if matches := regexEndTest.FindStringSubmatch(line); len(matches) == 4 {
		p.endTest(matches[1], matches[2], matches[3])
	} else if matches := regexStatus.FindStringSubmatch(line); len(matches) == 2 {
		p.status(matches[1])
	} else if matches := regexSummary.FindStringSubmatch(line); len(matches) == 6 {
		p.summary(matches[1], matches[2], matches[3])
	} else if matches := regexCoverage.FindStringSubmatch(line); len(matches) == 2 {
		p.coverage(matches[1])
	} else {
		p.output(line)
	}
}

func (p *parser) add(event Event) {
	p.events = append(p.events, event)
}

func (p *parser) findTest(name string) int {
	for i := len(p.events) - 1; i >= 0; i-- {
		// TODO: should we only consider tests that haven't ended yet?
		if p.events[i].Type == "run_test" && p.events[i].Name == name {
			return p.events[i].Id
		}
	}
	return -1
}

func (p *parser) runTest(name string) {
	p.id += 1
	p.add(Event{
		Type: "run_test",
		Id:   p.id,
		Name: strings.TrimSpace(name),
	})
}

func (p *parser) endTest(result, name, duration string) {
	p.add(Event{
		Type:     "end_test",
		Id:       p.findTest(name),
		Name:     name,
		Result:   result,
		Duration: parseSeconds(duration),
	})
}

func (p *parser) status(result string) {
	p.add(Event{
		Type:   "status",
		Result: result,
	})
}

func (p *parser) summary(result, name, duration string) {
	p.add(Event{
		Type:     "summary",
		Result:   result,
		Name:     name,
		Duration: parseSeconds(duration),
	})
}

func (p *parser) coverage(percent string) {
	// ignore error
	pct, _ := strconv.ParseFloat(percent, 64)
	p.add(Event{
		Type:   "coverage",
		CovPct: pct,
	})
}

func (p *parser) output(line string) {
	// TODO: Count indentations, however don't assume every tab is an indentation
	var indent int
	for indent = 0; strings.HasPrefix(line, "\t"); indent++ {
		line = line[1:]
	}
	p.add(Event{
		Type:   "output",
		Data:   line,
		Indent: indent,
	})
}

func parseSeconds(s string) time.Duration {
	if s == "" {
		return time.Duration(0)
	}
	// ignore error
	d, _ := time.ParseDuration(s + "s")
	return d
}
