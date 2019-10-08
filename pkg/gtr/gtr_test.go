package gtr

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestFromEvents(t *testing.T) {
	events := []Event{
		{Type: "run_test", Name: "TestOne"},
		{Type: "output", Data: "\tHello"},
		{Type: "end_test", Name: "TestOne", Result: "PASS", Duration: 1 * time.Millisecond},
		{Type: "status", Result: "PASS"},
		{Type: "run_test", Name: "TestSkip"},
		{Type: "end_test", Name: "TestSkip", Result: "SKIP", Duration: 1 * time.Millisecond},
		{Type: "summary", Result: "ok", Name: "package/name", Duration: 1 * time.Millisecond},
		{Type: "run_test", Name: "TestOne"},
		{Type: "output", Data: "\tfile_test.go:10: error"},
		{Type: "end_test", Name: "TestOne", Result: "FAIL", Duration: 1 * time.Millisecond},
		{Type: "status", Result: "FAIL"},
		{Type: "summary", Result: "FAIL", Name: "package/name2", Duration: 1 * time.Millisecond},
		{Type: "output", Data: "goarch: amd64"},
		{Type: "benchmark", Name: "BenchmarkOne", NsPerOp: 100},
		{Type: "benchmark", Name: "BenchmarkOne", NsPerOp: 300},
		{Type: "status", Result: "PASS"},
		{Type: "summary", Result: "ok", Name: "package/name3", Duration: 1234 * time.Millisecond},
		{Type: "build_output", Name: "package/failing1"},
		{Type: "output", Data: "error message"},
		{Type: "summary", Result: "FAIL", Name: "package/failing1", Data: "[build failed]"},
	}
	expected := Report{
		Packages: []Package{
			{
				Name:     "package/name",
				Duration: 1 * time.Millisecond,
				Tests: []Test{
					{
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   Pass,
						Output: []string{
							"\tHello", // TODO: strip tabs?
						},
					},
					{
						Name:     "TestSkip",
						Duration: 1 * time.Millisecond,
						Result:   Skip,
					},
				},
			},
			{
				Name:     "package/name2",
				Duration: 1 * time.Millisecond,
				Tests: []Test{
					{
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   Fail,
						Output: []string{
							"\tfile_test.go:10: error",
						},
					},
				},
			},
			{
				Name:     "package/name3",
				Duration: 1234 * time.Millisecond,
				Benchmarks: []Benchmark{
					{
						Name:    "BenchmarkOne",
						Result:  Pass,
						NsPerOp: 100,
					},
					{
						Name:    "BenchmarkOne",
						Result:  Pass,
						NsPerOp: 300,
					},
				},
				Output: []string{"goarch: amd64"},
			},
			{
				Name: "package/failing1",
				BuildError: Error{
					Name:   "package/failing1",
					Cause:  "[build failed]",
					Output: []string{"error message"},
				},
			},
		},
	}

	actual := FromEvents(events, "")
	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("FromEvents report incorrect, diff (-got, +want):\n%v", diff)
	}
}
