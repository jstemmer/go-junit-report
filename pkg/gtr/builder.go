package gtr

import (
	"fmt"
	"time"
)

// ReportBuilder builds Reports.
type ReportBuilder struct {
	packages []Package
	tests    map[int]Test

	// state
	nextId int // next free id
	lastId int // last test id // TODO: stack?
	output []string
}

func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{
		tests:  make(map[int]Test),
		nextId: 1,
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

func (b *ReportBuilder) CreateTest(name string) {
	id := b.nextId
	b.lastId = id

	b.nextId += 1
	b.tests[id] = Test{Name: name}
}

func (b *ReportBuilder) EndTest(name, result string, duration time.Duration) {
	id := b.findTest(name)
	b.lastId = id

	t := b.tests[id]
	t.Result = parseResult(result)
	t.Duration = duration
	b.tests[id] = t
}

func (b *ReportBuilder) CreatePackage(name string, duration time.Duration) {
	var tests []Test
	for id := 1; id < b.nextId; id++ {
		tests = append(tests, b.tests[id])
	}
	b.packages = append(b.packages, Package{
		Name:     name,
		Duration: duration,
		Tests:    tests,
		Output:   b.output,
	})

	b.tests = make(map[int]Test)
	b.output = nil
	b.nextId = 1
	b.lastId = 0
}

func (b *ReportBuilder) AppendOutput(line string) {
	if b.lastId <= 0 {
		b.output = append(b.output, line)
		return
	}
	t := b.tests[b.lastId]
	t.Output = append(t.Output, line)
	b.tests[b.lastId] = t
}

func (b *ReportBuilder) findTest(name string) int {
	// check if this test was lastId
	if t, ok := b.tests[b.lastId]; ok && t.Name == name {
		return b.lastId
	}
	for id := len(b.tests); id >= 0; id-- {
		if b.tests[id].Name == name {
			return id
		}
	}
	return -1
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
