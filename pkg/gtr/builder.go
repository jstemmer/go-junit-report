package gtr

import (
	"time"
)

// ReportBuilder builds Reports.
type ReportBuilder struct {
	packages    []Package
	tests       map[int]Test
	benchmarks  map[int]Benchmark
	buildErrors map[int]Error
	runErrors   map[int]Error

	// state
	nextId   int // next free id
	lastId   int // last test id // TODO: stack?
	output   []string
	coverage float64

	// defaults
	packageName string
}

// TODO: move gtr.FromEvents to ReportBuilder so we could have report builders for different parsers.
func NewReportBuilder(packageName string) *ReportBuilder {
	return &ReportBuilder{
		tests:       make(map[int]Test),
		benchmarks:  make(map[int]Benchmark),
		buildErrors: make(map[int]Error),
		runErrors:   make(map[int]Error),
		nextId:      1,
		packageName: packageName,
	}
}

func (b *ReportBuilder) newId() int {
	id := b.nextId
	b.lastId = id
	b.nextId += 1
	return id
}

func (b *ReportBuilder) flush() {
	// Create package in case we have tests or benchmarks that didn't get a
	// summary.
	if len(b.tests) > 0 || len(b.benchmarks) > 0 {
		b.CreatePackage(b.packageName, "", 0, "")
	}
}

func (b *ReportBuilder) Build() Report {
	b.flush()
	return Report{Packages: b.packages}
}

func (b *ReportBuilder) CreateTest(name string) {
	b.tests[b.newId()] = Test{Name: name}
}

func (b *ReportBuilder) PauseTest(name string) {
	b.lastId = 0
}

func (b *ReportBuilder) ContinueTest(name string) {
	b.lastId = b.findTest(name)
}

func (b *ReportBuilder) EndTest(name, result string, duration time.Duration, level int) {
	b.lastId = b.findTest(name)
	if b.lastId < 0 {
		// test did not exist, create one
		// TODO: Likely reason is that the user ran go test without the -v
		// flag, should we report this somewhere?
		b.CreateTest(name)
	}

	t := b.tests[b.lastId]
	t.Result = parseResult(result)
	t.Duration = duration
	t.Level = level
	b.tests[b.lastId] = t
}

func (b *ReportBuilder) End() {
	b.lastId = 0
}

func (b *ReportBuilder) Benchmark(name string, iterations int64, nsPerOp, mbPerSec float64, bytesPerOp, allocsPerOp int64) {
	b.benchmarks[b.newId()] = Benchmark{
		Name:        name,
		Result:      Pass,
		Iterations:  iterations,
		NsPerOp:     nsPerOp,
		MBPerSec:    mbPerSec,
		BytesPerOp:  bytesPerOp,
		AllocsPerOp: allocsPerOp,
	}
}

func (b *ReportBuilder) CreateBuildError(packageName string) {
	b.buildErrors[b.newId()] = Error{Name: packageName}
}

func (b *ReportBuilder) CreatePackage(name, result string, duration time.Duration, data string) {
	pkg := Package{
		Name:     name,
		Duration: duration,
	}

	// Build errors are treated somewhat differently. Rather than having a
	// single package with all build errors collected so far, we only care
	// about the build errors for this particular package.
	for id, buildErr := range b.buildErrors {
		if buildErr.Name == name {
			if len(b.tests) > 0 || len(b.benchmarks) > 0 {
				panic("unexpected tests and/or benchmarks found in build error package")
			}
			buildErr.Duration = duration
			buildErr.Cause = data
			pkg.BuildError = buildErr
			b.packages = append(b.packages, pkg)

			delete(b.buildErrors, id)
			// TODO: reset state
			// TODO: buildErrors shouldn't reset/use nextId/lastId, they're more like a global cache
			return
		}
	}

	// If we've collected output, but there were no tests or benchmarks then
	// either there were no tests, or there was some other non-build error.
	if len(b.output) > 0 && len(b.tests) == 0 && len(b.benchmarks) == 0 {
		if parseResult(result) == Fail {
			pkg.RunError = Error{
				Name:   name,
				Output: b.output,
			}
		}
		b.packages = append(b.packages, pkg)

		// TODO: reset state
		b.output = nil
		return
	}

	// If the summary result says we failed, but there were no failing tests
	// then something else must have failed.
	if parseResult(result) == Fail && (len(b.tests) > 0 || len(b.benchmarks) > 0) && !b.containsFailingTest() {
		pkg.RunError = Error{
			Name:   name,
			Output: b.output,
		}
		b.output = nil
	}

	// Collect tests and benchmarks for this package, maintaining insertion order.
	var tests []Test
	var benchmarks []Benchmark
	for id := 1; id < b.nextId; id++ {
		if t, ok := b.tests[id]; ok {
			tests = append(tests, t)
		}
		if bm, ok := b.benchmarks[id]; ok {
			benchmarks = append(benchmarks, bm)
		}
	}

	pkg.Coverage = b.coverage
	pkg.Output = b.output
	pkg.Tests = tests
	pkg.Benchmarks = benchmarks
	b.packages = append(b.packages, pkg)

	b.nextId = 1
	b.lastId = 0
	b.output = nil
	b.coverage = 0
	b.tests = make(map[int]Test)
	b.benchmarks = make(map[int]Benchmark)
}

func (b *ReportBuilder) Coverage(pct float64, packages []string) {
	b.coverage = pct
}

func (b *ReportBuilder) AppendOutput(line string) {
	if b.lastId <= 0 {
		b.output = append(b.output, line)
		return
	}

	if t, ok := b.tests[b.lastId]; ok {
		t.Output = append(t.Output, line)
		b.tests[b.lastId] = t
	} else if bm, ok := b.benchmarks[b.lastId]; ok {
		bm.Output = append(bm.Output, line)
		b.benchmarks[b.lastId] = bm
	} else if be, ok := b.buildErrors[b.lastId]; ok {
		be.Output = append(be.Output, line)
		b.buildErrors[b.lastId] = be
	} else {
		b.output = append(b.output, line)
	}
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

func (b *ReportBuilder) findBenchmark(name string) int {
	// check if this benchmark was lastId
	if bm, ok := b.benchmarks[b.lastId]; ok && bm.Name == name {
		return b.lastId
	}
	for id := len(b.benchmarks); id >= 0; id-- {
		if b.benchmarks[id].Name == name {
			return id
		}
	}
	return -1
}

func (b *ReportBuilder) containsFailingTest() bool {
	for _, test := range b.tests {
		if test.Result == Fail || test.Result == Unknown {
			return true
		}
	}
	return false
}

func parseResult(r string) Result {
	switch r {
	case "PASS":
		return Pass
	case "FAIL":
		return Fail
	case "SKIP":
		return Skip
	default:
		return Unknown
	}
}
