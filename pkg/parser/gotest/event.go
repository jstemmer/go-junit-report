package gotest

import "time"

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
	MBPerSec    float64
	BytesPerOp  int64
	AllocsPerOp int64
}
