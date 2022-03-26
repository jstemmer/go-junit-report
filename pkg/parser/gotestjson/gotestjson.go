package gotestjson

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
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

type TestEvent struct {
	Time time.Time // encodes as an RFC3339-format string
	//    run    - the test has started running
	//    pause  - the test has been paused
	//    cont   - the test has continued running
	//    pass   - the test passed
	//    bench  - the benchmark printed log output but did not fail
	//    fail   - the test or benchmark failed
	//    output - the test printed output
	//    skip   - the test was skipped or the package contained no tests
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

// Option defines options that can be passed to gotestjson.New.
type Option func(*Parser)

// PackageName is an Option that sets the default package name to use when it
// cannot be determined from the test output.
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

// Parser is a Go test json output Parser.
type Parser struct {
	packageName string
}

// Parse parses Go test json output from the given io.Reader r and returns
// gtr.Report.
func (p *Parser) Parse(r io.Reader) (gtr.Report, error) {
	rb := gtr.NewReportBuilder()
	rb.PackageName = p.packageName
	de := json.NewDecoder(r)
	for {
		event := TestEvent{}
		err := de.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return gtr.Report{}, err
		}
		p.addEvent(rb, &event)
	}

	return rb.Build(), nil
}

func (p *Parser) addEvent(rb *gtr.ReportBuilder, ev *TestEvent) {
	if ev.Test == "" {
		switch ev.Action {
		case "run", "pause", "cont", "bench":
			return
		case "pass":
			rb.TimestampFunc = func() time.Time {
				return ev.Time
			}
			rb.CreatePackage(ev.Package, "ok", time.Duration(ev.Elapsed*1e9), "")
			return
		case "fail":
			rb.TimestampFunc = func() time.Time {
				return ev.Time
			}
			rb.CreatePackage(ev.Package, "FAIL", time.Duration(ev.Elapsed*1e9), "")
			return
		case "skip":
			rb.TimestampFunc = func() time.Time {
				return ev.Time
			}
			rb.CreatePackage(ev.Package, "?", time.Duration(ev.Elapsed*1e9), "")
			return
		case "output":
			line := strings.TrimSuffix(ev.Output, "\n")

			if regexStatus.MatchString(line) {
				rb.End()
				return
			}

			if regexSummary.MatchString(line) {
				return
			}

			rb.AppendOutput(line)
			return
		}
	}

	switch ev.Action {
	case "run":
		rb.CreateTest(ev.Test)
		return
	case "pause":
		rb.PauseTest(ev.Test)
		return
	case "cont":
		rb.ContinueTest(ev.Test)
		return
	case "pass", "fail", "skip":
		rb.EndTest(ev.Test, strings.ToUpper(ev.Action), time.Duration(ev.Elapsed*1e9), 0)
		return
	case "bench":
		// todo
		return
	case "output":
		if strings.HasPrefix(ev.Output, "=== ") || strings.HasPrefix(ev.Output, "--- ") {
			return
		}

		line := strings.TrimSuffix(ev.Output, "\n")
		rb.AppendOutput(line)
		return
	}
}

func stripIndent(line string) (string, int) {
	var indent int
	for indent = 0; strings.HasPrefix(line, "    "); indent++ {
		line = line[4:]
	}
	return line, indent
}
