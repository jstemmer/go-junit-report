package gotest

import (
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/v2/gtr"
	"github.com/jstemmer/go-junit-report/v2/parser/gotest/internal/collector"
)

const (
	globalID = 0
)

// reportBuilder helps build a test Report from a collection of events.
//
// The reportBuilder keeps track of the active context whenever a test or build
// error is created. This is necessary because the test parser do not contain
// any state themselves and simply just emit an event for every line that is
// read. By tracking the active context, any output that is appended to the
// reportBuilder gets attributed to the correct test or build error.
type reportBuilder struct {
	packages    []gtr.Package
	tests       map[int]gtr.Test
	buildErrors map[int]gtr.Error
	runErrors   map[int]gtr.Error

	// state
	nextID    int               // next free unused id
	lastID    int               // most recently created id
	output    *collector.Output // output collected for each id
	coverage  float64           // coverage percentage
	parentIDs map[int]struct{}  // set of test id's that contain subtests

	// options
	packageName   string
	subtestMode   SubtestMode
	timestampFunc func() time.Time
}

// newReportBuilder creates a new reportBuilder.
func newReportBuilder() *reportBuilder {
	return &reportBuilder{
		tests:         make(map[int]gtr.Test),
		buildErrors:   make(map[int]gtr.Error),
		runErrors:     make(map[int]gtr.Error),
		nextID:        1,
		output:        collector.New(),
		parentIDs:     make(map[int]struct{}),
		timestampFunc: time.Now,
	}
}

// newID returns a new unique id and sets the active context this id.
func (b *reportBuilder) newID() int {
	id := b.nextID
	b.lastID = id
	b.nextID += 1
	return id
}

// flush creates a new package in this report containing any tests we've
// collected so far. This is necessary when a test did not end with a summary.
func (b *reportBuilder) flush() {
	if len(b.tests) > 0 {
		b.CreatePackage(b.packageName, "", 0, "")
	}
}

// Build returns the new Report containing all the tests and output created so
// far.
func (b *reportBuilder) Build() gtr.Report {
	b.flush()
	return gtr.Report{Packages: b.packages}
}

// CreateTest adds a test with the given name to the report, and marks it as
// active.
func (b *reportBuilder) CreateTest(name string) {
	if parentID, ok := b.findTestParentID(name); ok {
		b.parentIDs[parentID] = struct{}{}
	}
	id := b.newID()
	b.tests[id] = gtr.NewTest(id, name)
}

// PauseTest marks the active context as no longer active. Any results or
// output added to the report after calling PauseTest will no longer be assumed
// to belong to this test.
func (b *reportBuilder) PauseTest(name string) {
	b.lastID = 0
}

// ContinueTest finds the test with the given name and marks it as active. If
// more than one test exist with this name, the most recently created test will
// be used.
func (b *reportBuilder) ContinueTest(name string) {
	b.lastID, _ = b.findTest(name)
}

// EndTest finds the test with the given name, sets the result, duration and
// level. If more than one test exists with this name, the most recently
// created test will be used. If no test exists with this name, a new test is
// created.
func (b *reportBuilder) EndTest(name, result string, duration time.Duration, level int) {
	id, ok := b.findTest(name)
	if !ok {
		// test did not exist, create one
		// TODO: Likely reason is that the user ran go test without the -v
		// flag, should we report this somewhere?
		b.CreateTest(name)
		id = b.lastID
	}

	t := b.tests[id]
	t.Result = parseResult(result)
	t.Duration = duration
	t.Level = level
	b.tests[id] = t
	b.lastID = 0
}

// End marks the active context as no longer active.
func (b *reportBuilder) End() {
	b.lastID = 0
}

// CreateBenchmark adds a benchmark with the given name to the report, and
// marks it as active. If more than one benchmark exists with this name, the
// most recently created benchmark will be updated. If no benchmark exists with
// this name, a new benchmark is created.
func (b *reportBuilder) CreateBenchmark(name string) {
	b.CreateTest(name)
}

// BenchmarkResult updates an existing or adds a new test with the given
// results and marks it as active. If an existing test with this name exists
// but without result, then that one is updated. Otherwise a new one is added
// to the report.
func (b *reportBuilder) BenchmarkResult(name string, iterations int64, nsPerOp, mbPerSec float64, bytesPerOp, allocsPerOp int64) {
	id, ok := b.findTest(name)
	if !ok || b.tests[id].Result != gtr.Unknown {
		b.CreateTest(name)
		id = b.lastID
	}

	benchmark := Benchmark{iterations, nsPerOp, mbPerSec, bytesPerOp, allocsPerOp}
	test := gtr.NewTest(id, name)
	test.Result = gtr.Pass
	test.Duration = benchmark.ApproximateDuration()
	SetBenchmarkData(&test, benchmark)
	b.tests[id] = test
}

// EndBenchmark finds the benchmark with the given name and sets the result. If
// more than one benchmark exists with this name, the most recently created
// benchmark will be used. If no benchmark exists with this name, a new
// benchmark is created.
func (b *reportBuilder) EndBenchmark(name, result string) {
	b.EndTest(name, result, 0, 0)
}

// CreateBuildError creates a new build error and marks it as active.
func (b *reportBuilder) CreateBuildError(packageName string) {
	id := b.newID()
	b.buildErrors[id] = gtr.Error{ID: id, Name: packageName}
}

// CreatePackage adds a new package with the given name to the Report. This
// package contains all the build errors, output, tests and benchmarks created
// so far. Afterwards all state is reset.
func (b *reportBuilder) CreatePackage(name, result string, duration time.Duration, data string) {
	pkg := gtr.Package{
		Name:     name,
		Duration: duration,
	}

	if b.timestampFunc != nil {
		pkg.Timestamp = b.timestampFunc()
	}

	// Build errors are treated somewhat differently. Rather than having a
	// single package with all build errors collected so far, we only care
	// about the build errors for this particular package.
	for id, buildErr := range b.buildErrors {
		if buildErr.Name == name {
			if len(b.tests) > 0 {
				panic("unexpected tests found in build error package")
			}
			buildErr.ID = id
			buildErr.Duration = duration
			buildErr.Cause = data
			buildErr.Output = b.output.Get(id)

			pkg.BuildError = buildErr
			b.packages = append(b.packages, pkg)

			delete(b.buildErrors, id)
			// TODO: reset state
			// TODO: buildErrors shouldn't reset/use nextID/lastID, they're more like a global cache
			return
		}
	}

	// If we've collected output, but there were no tests then either there
	// actually were no tests, or there was some other non-build error.
	if b.output.Contains(globalID) && len(b.tests) == 0 {
		if parseResult(result) == gtr.Fail {
			pkg.RunError = gtr.Error{
				Name:   name,
				Output: b.output.Get(globalID),
			}
		} else if b.output.Contains(globalID) {
			pkg.Output = b.output.Get(globalID)
		}
		b.packages = append(b.packages, pkg)
		b.output.Clear(globalID)
		return
	}

	// If the summary result says we failed, but there were no failing tests
	// then something else must have failed.
	if parseResult(result) == gtr.Fail && len(b.tests) > 0 && !b.containsFailures() {
		pkg.RunError = gtr.Error{
			Name:   name,
			Output: b.output.Get(globalID),
		}
		b.output.Clear(globalID)
	}

	// Collect tests for this package, maintaining insertion order.
	var tests []gtr.Test
	for id := 1; id < b.nextID; id++ {
		if t, ok := b.tests[id]; ok {
			if b.isParent(id) {
				if b.subtestMode == IgnoreParentResults {
					t.Result = gtr.Pass
				} else if b.subtestMode == ExcludeParents {
					b.output.Merge(id, globalID)
					continue
				}
			}
			t.Output = b.output.Get(id)
			tests = append(tests, t)
			continue
		}
	}
	tests = b.groupBenchmarksByName(tests)

	pkg.Coverage = b.coverage
	pkg.Output = b.output.Get(globalID)
	pkg.Tests = tests
	b.packages = append(b.packages, pkg)

	// reset state, except for nextID to ensure all id's are unique.
	b.lastID = 0
	b.output.Clear(globalID)
	b.coverage = 0
	b.tests = make(map[int]gtr.Test)
	b.parentIDs = make(map[int]struct{})
}

// Coverage sets the code coverage percentage.
func (b *reportBuilder) Coverage(pct float64, packages []string) {
	b.coverage = pct
}

// AppendOutput appends the given text to the currently active context. If no
// active context exists, the output is assumed to belong to the package.
func (b *reportBuilder) AppendOutput(text string) {
	b.output.Append(b.lastID, text)
}

// findTest returns the id of the most recently created test with the given
// name if it exists.
func (b *reportBuilder) findTest(name string) (int, bool) {
	// check if this test was lastID
	if t, ok := b.tests[b.lastID]; ok && t.Name == name {
		return b.lastID, true
	}
	for i := b.nextID; i > 0; i-- {
		if test, ok := b.tests[i]; ok && test.Name == name {
			return i, true
		}
	}
	return 0, false
}

func (b *reportBuilder) findTestParentID(name string) (int, bool) {
	parent := dropLastSegment(name)
	for parent != "" {
		if id, ok := b.findTest(parent); ok {
			return id, true
		}
		parent = dropLastSegment(parent)
	}
	return 0, false
}

func (b *reportBuilder) isParent(id int) bool {
	_, ok := b.parentIDs[id]
	return ok
}

func dropLastSegment(name string) string {
	if idx := strings.LastIndexByte(name, '/'); idx >= 0 {
		return name[:idx]
	}
	return ""
}

// containsFailures return true if the current list of tests contains at least
// one failing test or an unknown result.
func (b *reportBuilder) containsFailures() bool {
	for _, test := range b.tests {
		if test.Result == gtr.Fail || test.Result == gtr.Unknown {
			return true
		}
	}
	return false
}

// parseResult returns a Result for the given string r.
func parseResult(r string) gtr.Result {
	switch r {
	case "PASS":
		return gtr.Pass
	case "FAIL":
		return gtr.Fail
	case "SKIP":
		return gtr.Skip
	case "BENCH":
		return gtr.Pass
	default:
		return gtr.Unknown
	}
}

func (b *reportBuilder) groupBenchmarksByName(tests []gtr.Test) []gtr.Test {
	if len(tests) == 0 {
		return nil
	}

	var grouped []gtr.Test
	byName := make(map[string][]gtr.Test)
	for _, test := range tests {
		if !strings.HasPrefix(test.Name, "Benchmark") {
			// If this test is not a benchmark, we won't group it by name but
			// just add it to the final result.
			grouped = append(grouped, test)
			continue
		}
		if _, ok := byName[test.Name]; !ok {
			grouped = append(grouped, gtr.NewTest(test.ID, test.Name))
		}
		byName[test.Name] = append(byName[test.Name], test)
	}

	for i, group := range grouped {
		if !strings.HasPrefix(group.Name, "Benchmark") {
			continue
		}
		var (
			ids   []int
			total Benchmark
			count int
		)
		for _, test := range byName[group.Name] {
			ids = append(ids, test.ID)
			if test.Result != gtr.Pass {
				continue
			}

			if bench, ok := GetBenchmarkData(test); ok {
				total.Iterations += bench.Iterations
				total.NsPerOp += bench.NsPerOp
				total.MBPerSec += bench.MBPerSec
				total.BytesPerOp += bench.BytesPerOp
				total.AllocsPerOp += bench.AllocsPerOp
				count++
			}
		}

		group.Duration = combinedDuration(byName[group.Name])
		group.Result = groupResults(byName[group.Name])
		group.Output = b.output.GetAll(ids...)
		if count > 0 {
			total.Iterations /= int64(count)
			total.NsPerOp /= float64(count)
			total.MBPerSec /= float64(count)
			total.BytesPerOp /= int64(count)
			total.AllocsPerOp /= int64(count)
			SetBenchmarkData(&group, total)
		}
		grouped[i] = group
	}
	return grouped
}

func combinedDuration(tests []gtr.Test) time.Duration {
	var total time.Duration
	for _, test := range tests {
		total += test.Duration
	}
	return total
}

func groupResults(tests []gtr.Test) gtr.Result {
	var result gtr.Result
	for _, test := range tests {
		if test.Result == gtr.Fail {
			return gtr.Fail
		}
		if result != gtr.Pass {
			result = test.Result
		}
	}
	return result
}
