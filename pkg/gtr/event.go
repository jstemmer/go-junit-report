package gtr

import "time"

type Result int

const (
	Unknown Result = iota
	Pass
	Fail
	Skip
)

func (r Result) String() string {
	switch r {
	case Unknown:
		return "UNKNOWN"
	case Pass:
		return "PASS"
	case Fail:
		return "FAIL"
	case Skip:
		return "SKIP"
	default:
		panic("invalid Result")
	}
}

// TODO: provide some common types, or have a custom type (e.g.
// identifier:type, where identifier is a unique identifier for a particular
// parser.)

// Event is a single event in a test or benchmark.
type Event struct {
	Type string

	Name     string
	Result   string
	Duration time.Duration
	Data     string
	Indent   int

	// Code coverage
	CovPct      float64
	CovPackages []string

	// Benchmarks
	Iterations  int64
	NsPerOp     float64
	BytesPerOp  int64
	AllocsPerOp int64
}
