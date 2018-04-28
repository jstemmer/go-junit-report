// Package gtr generates Go Test Reports from a collection of Events.
package gtr

import (
	"fmt"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/parser/gotest"
)

type Result int

const (
	// TODO: move these to event and don't make the all-caps
	UNKNOWN Result = iota
	PASS
	FAIL
	SKIP
)

func (r Result) String() string {
	switch r {
	case UNKNOWN:
		return "UNKNOWN"
	case PASS:
		return "PASS"
	case FAIL:
		return "FAIL"
	case SKIP:
		return "SKIP"
	default:
		panic("invalid result")
	}
}

type Report struct {
	Packages []Package
}

type Package struct {
	Name     string
	Duration time.Duration
	Coverage float64

	Tests []Test
}

type Test struct {
	Name     string
	Duration time.Duration
	Result   Result
	Output   []string
}

// FromEvents creates a Report from the given list of test events.
func FromEvents(events []gotest.Event) Report {
	report := NewReportBuilder()
	for _, ev := range events {
		switch ev.Type {
		case "run_test":
			report.CreateTest(ev.Id, ev.Name)
		case "end_test":
			report.UpdateTest(ev.Id, ev.Result, ev.Duration)
		case "status": // ignore for now
		case "summary":
			report.CreatePackage(ev.Name, ev.Duration)
		default:
			fmt.Printf("unhandled event type: %v\n", ev.Type)
		}
	}
	return report.Build()
}

// ReportBuilder builds Reports.
type ReportBuilder struct {
	packages []Package
	ids      []int
	tests    map[int]Test
}

func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{
		tests: make(map[int]Test),
	}
}

func (b *ReportBuilder) flush() {
	if len(b.tests) > 0 {
		b.CreatePackage("unknown", 0)
	}
}

func (b *ReportBuilder) Build() Report {
	b.flush()
	return Report{Packages: b.packages}
}

func (b *ReportBuilder) CreateTest(id int, name string) {
	b.ids = append(b.ids, id)
	b.tests[id] = Test{Name: name}
}

func (b *ReportBuilder) UpdateTest(id int, result string, duration time.Duration) {
	t := b.tests[id]
	t.Result = parseResult(result)
	t.Duration = duration
	b.tests[id] = t
}

func (b *ReportBuilder) CreatePackage(name string, duration time.Duration) {
	var tests []Test
	for _, id := range b.ids {
		tests = append(tests, b.tests[id])
	}
	b.packages = append(b.packages, Package{
		Name:     name,
		Duration: duration,
		Tests:    tests,
	})
	b.ids = nil
	b.tests = make(map[int]Test)
}

func parseResult(r string) Result {
	switch r {
	case "PASS":
		return PASS
	case "FAIL":
		return FAIL
	case "SKIP":
		return SKIP
	default:
		fmt.Printf("unknown result: %q\n", r)
		return UNKNOWN
	}
}
