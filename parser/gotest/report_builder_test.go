package gotest

import (
	"fmt"
	"testing"
	"time"

	"github.com/jstemmer/go-junit-report/v2/gtr"
	"github.com/jstemmer/go-junit-report/v2/parser/gotest/internal/collector"

	"github.com/google/go-cmp/cmp"
)

func TestReport(t *testing.T) {
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
		{Type: "run_benchmark", Name: "BenchmarkOne"},
		{Type: "benchmark", Name: "BenchmarkOne", NsPerOp: 100},
		{Type: "end_benchmark", Name: "BenchmarkOne", Result: "BENCH"},
		{Type: "run_benchmark", Name: "BenchmarkTwo"},
		{Type: "benchmark", Name: "BenchmarkTwo"},
		{Type: "end_benchmark", Name: "BenchmarkTwo", Result: "FAIL"},
		{Type: "status", Result: "PASS"},
		{Type: "summary", Result: "ok", Name: "package/name3", Duration: 1234 * time.Millisecond},
		{Type: "build_output", Name: "package/failing1"},
		{Type: "output", Data: "error message"},
		{Type: "summary", Result: "FAIL", Name: "package/failing1", Data: "[build failed]"},
	}
	want := gtr.Report{
		Packages: []gtr.Package{
			{
				Name:      "package/name",
				Duration:  1 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						ID:       1,
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Pass,
						Output: []string{
							"\tHello", // TODO: strip tabs?
						},
						Data: map[string]interface{}{},
					},
					{
						ID:       2,
						Name:     "TestSkip",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Skip,
						Data:     map[string]interface{}{},
					},
				},
			},
			{
				Name:      "package/name2",
				Duration:  1 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						ID:       3,
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Fail,
						Output: []string{
							"\tfile_test.go:10: error",
						},
						Data: map[string]interface{}{},
					},
				},
			},
			{
				Name:      "package/name3",
				Duration:  1234 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						ID:     4,
						Name:   "BenchmarkOne",
						Result: gtr.Pass,
						Data:   map[string]interface{}{key: Benchmark{NsPerOp: 100}},
					},
					{
						ID:     5,
						Name:   "BenchmarkTwo",
						Result: gtr.Fail,
						Data:   map[string]interface{}{},
					},
				},
				Output: []string{"goarch: amd64"},
			},
			{
				Name:      "package/failing1",
				Timestamp: testTimestamp,
				BuildError: gtr.Error{
					ID:     6,
					Name:   "package/failing1",
					Cause:  "[build failed]",
					Output: []string{"error message"},
				},
			},
		},
	}

	rb := newReportBuilder()
	rb.timestampFunc = testTimestampFunc
	for _, ev := range events {
		rb.ProcessEvent(ev)
	}
	got := rb.Build()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Incorrect report created, diff (-want, +got):\n%v", diff)
	}
}

func TestBuildReportMultiplePackages(t *testing.T) {
	events := []Event{
		{Package: "package/name1", Type: "run_test", Name: "TestOne"},
		{Package: "package/name2", Type: "run_test", Name: "TestOne"},
		{Package: "package/name1", Type: "output", Data: "\tHello"},
		{Package: "package/name1", Type: "end_test", Name: "TestOne", Result: "PASS", Duration: 1 * time.Millisecond},
		{Package: "package/name2", Type: "output", Data: "\tfile_test.go:10: error"},
		{Package: "package/name2", Type: "end_test", Name: "TestOne", Result: "FAIL", Duration: 1 * time.Millisecond},
		{Package: "package/name2", Type: "status", Result: "FAIL"},
		{Package: "package/name2", Type: "summary", Result: "FAIL", Name: "package/name2", Duration: 1 * time.Millisecond},
		{Package: "package/name1", Type: "status", Result: "PASS"},
		{Package: "package/name1", Type: "summary", Result: "ok", Name: "package/name1", Duration: 1 * time.Millisecond},
	}

	want := gtr.Report{
		Packages: []gtr.Package{
			{
				Name:      "package/name2",
				Duration:  1 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						ID:       2,
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Fail,
						Output:   []string{"\tfile_test.go:10: error"},
						Data:     make(map[string]interface{}),
					},
				},
			},
			{
				Name:      "package/name1",
				Duration:  1 * time.Millisecond,
				Timestamp: testTimestamp,
				Tests: []gtr.Test{
					{
						ID:       1,
						Name:     "TestOne",
						Duration: 1 * time.Millisecond,
						Result:   gtr.Pass,
						Output:   []string{"\tHello"},
						Data:     make(map[string]interface{}),
					},
				},
			},
		},
	}

	rb := newReportBuilder()
	rb.timestampFunc = testTimestampFunc
	for _, ev := range events {
		rb.ProcessEvent(ev)
	}
	got := rb.Build()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Incorrect report created, diff (-want, +got):\n%v", diff)
	}
}

func TestSubtestModes(t *testing.T) {
	events := []Event{
		{Type: "run_test", Name: "TestParent"},
		{Type: "output", Data: "TestParent before"},
		{Type: "run_test", Name: "TestParent/Subtest#1"},
		{Type: "output", Data: "Subtest#1 output"},
		{Type: "run_test", Name: "TestParent/Subtest#2"},
		{Type: "output", Data: "Subtest#2 output"},
		{Type: "cont_test", Name: "TestParent"},
		{Type: "output", Data: "TestParent after"},
		{Type: "end_test", Name: "TestParent", Result: "PASS", Duration: 1 * time.Millisecond},
		{Type: "end_test", Name: "TestParent/Subtest#1", Result: "FAIL", Duration: 2 * time.Millisecond},
		{Type: "end_test", Name: "TestParent/Subtest#2", Result: "PASS", Duration: 3 * time.Millisecond},
		{Type: "output", Data: "output"},
		{Type: "summary", Result: "FAIL", Name: "package/name", Duration: 1 * time.Millisecond},
	}

	tests := []struct {
		name string
		mode SubtestMode
		want gtr.Report
	}{
		{
			name: "ignore subtest parent results",
			mode: IgnoreParentResults,
			want: gtr.Report{
				Packages: []gtr.Package{
					{
						Name:      "package/name",
						Duration:  1 * time.Millisecond,
						Timestamp: testTimestamp,
						Tests: []gtr.Test{
							{
								ID:       1,
								Name:     "TestParent",
								Duration: 1 * time.Millisecond,
								Result:   gtr.Pass,
								Output:   []string{"TestParent before", "TestParent after"},
								Data:     map[string]interface{}{},
							},
							{
								ID:       2,
								Name:     "TestParent/Subtest#1",
								Duration: 2 * time.Millisecond,
								Result:   gtr.Fail,
								Output:   []string{"Subtest#1 output"},
								Data:     map[string]interface{}{},
							},
							{
								ID:       3,
								Name:     "TestParent/Subtest#2",
								Duration: 3 * time.Millisecond,
								Result:   gtr.Pass,
								Output:   []string{"Subtest#2 output"},
								Data:     map[string]interface{}{},
							},
						},
						Output: []string{"output"},
					},
				},
			},
		},
		{
			name: "exclude subtest parents",
			mode: ExcludeParents,
			want: gtr.Report{
				Packages: []gtr.Package{
					{
						Name:      "package/name",
						Duration:  1 * time.Millisecond,
						Timestamp: testTimestamp,
						Tests: []gtr.Test{
							{
								ID:       2,
								Name:     "TestParent/Subtest#1",
								Duration: 2 * time.Millisecond,
								Result:   gtr.Fail,
								Output:   []string{"Subtest#1 output"},
								Data:     map[string]interface{}{},
							},
							{
								ID:       3,
								Name:     "TestParent/Subtest#2",
								Duration: 3 * time.Millisecond,
								Result:   gtr.Pass,
								Output:   []string{"Subtest#2 output"},
								Data:     map[string]interface{}{},
							},
						},
						Output: []string{"TestParent before", "TestParent after", "output"},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rb := newReportBuilder()
			rb.timestampFunc = testTimestampFunc
			rb.subtestMode = test.mode
			for _, ev := range events {
				rb.ProcessEvent(ev)
			}
			got := rb.Build()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Invalid report created from events, diff (-want, +got):\n%v", diff)
			}
		})
	}
}

func TestGroupBenchmarksByName(t *testing.T) {
	output := collector.New()
	for i := 1; i <= 4; i++ {
		output.AppendToID(i, fmt.Sprintf("output-%d", i))
	}

	tests := []struct {
		name string
		in   []gtr.Test
		want []gtr.Test
	}{
		{"nil", nil, nil},
		{
			"one failing benchmark",
			[]gtr.Test{{ID: 1, Name: "BenchmarkFailed", Result: gtr.Fail, Data: map[string]interface{}{}}},
			[]gtr.Test{{ID: 1, Name: "BenchmarkFailed", Result: gtr.Fail, Output: []string{"output-1"}, Data: map[string]interface{}{}}},
		},
		{
			"four passing benchmarks",
			[]gtr.Test{
				{ID: 1, Name: "BenchmarkOne", Result: gtr.Pass, Data: map[string]interface{}{key: Benchmark{NsPerOp: 10, MBPerSec: 400, BytesPerOp: 1, AllocsPerOp: 2}}},
				{ID: 2, Name: "BenchmarkOne", Result: gtr.Pass, Data: map[string]interface{}{key: Benchmark{NsPerOp: 20, MBPerSec: 300, BytesPerOp: 1, AllocsPerOp: 4}}},
				{ID: 3, Name: "BenchmarkOne", Result: gtr.Pass, Data: map[string]interface{}{key: Benchmark{NsPerOp: 30, MBPerSec: 200, BytesPerOp: 1, AllocsPerOp: 8}}},
				{ID: 4, Name: "BenchmarkOne", Result: gtr.Pass, Data: map[string]interface{}{key: Benchmark{NsPerOp: 40, MBPerSec: 100, BytesPerOp: 5, AllocsPerOp: 2}}},
			},
			[]gtr.Test{
				{ID: 1, Name: "BenchmarkOne", Result: gtr.Pass, Output: []string{"output-1", "output-2", "output-3", "output-4"}, Data: map[string]interface{}{key: Benchmark{NsPerOp: 25, MBPerSec: 250, BytesPerOp: 2, AllocsPerOp: 4}}},
			},
		},
		{
			"four mixed result benchmarks",
			[]gtr.Test{
				{ID: 1, Name: "BenchmarkMixed", Result: gtr.Unknown},
				{ID: 2, Name: "BenchmarkMixed", Result: gtr.Pass, Data: map[string]interface{}{key: Benchmark{NsPerOp: 10, MBPerSec: 400, BytesPerOp: 1, AllocsPerOp: 2}}},
				{ID: 3, Name: "BenchmarkMixed", Result: gtr.Pass, Data: map[string]interface{}{key: Benchmark{NsPerOp: 40, MBPerSec: 100, BytesPerOp: 3, AllocsPerOp: 4}}},
				{ID: 4, Name: "BenchmarkMixed", Result: gtr.Fail},
			},
			[]gtr.Test{
				{ID: 1, Name: "BenchmarkMixed", Result: gtr.Fail, Output: []string{"output-1", "output-2", "output-3", "output-4"}, Data: map[string]interface{}{key: Benchmark{NsPerOp: 25, MBPerSec: 250, BytesPerOp: 2, AllocsPerOp: 3}}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := groupBenchmarksByName(test.in, output)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("groupBenchmarksByName result incorrect, diff (-want, +got):\n%s\n", diff)
			}
		})
	}
}
