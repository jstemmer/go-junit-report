package gtr

import "time"

// TODO: provide some common types, or have a custom type (e.g.
// identifier:type, where identifier is a unique identifier for a particular
// parser.)

// Event is a single event in a test or benchmark.
type Event struct {
	Type string

	Name        string
	Result      string
	Duration    time.Duration
	Data        string
	Indent      int
	CovPct      float64
	CovPackages []string
}
