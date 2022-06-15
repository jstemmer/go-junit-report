package gotest

import (
	"testing"

	"github.com/jstemmer/go-junit-report/v2/gtr"

	"github.com/google/go-cmp/cmp"
)

func TestGroupBenchmarksByName(t *testing.T) {
	tests := []struct {
		name string
		in   []gtr.Test
		want []gtr.Test
	}{
		{"nil", nil, nil},
		{
			"one failing benchmark",
			[]gtr.Test{{ID: 1, Name: "BenchmarkFailed", Result: gtr.Fail, Data: map[string]interface{}{}}},
			[]gtr.Test{{ID: 1, Name: "BenchmarkFailed", Result: gtr.Fail, Data: map[string]interface{}{}}},
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
				{ID: 1, Name: "BenchmarkOne", Result: gtr.Pass, Data: map[string]interface{}{key: Benchmark{NsPerOp: 25, MBPerSec: 250, BytesPerOp: 2, AllocsPerOp: 4}}},
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
				{ID: 1, Name: "BenchmarkMixed", Result: gtr.Fail, Data: map[string]interface{}{key: Benchmark{NsPerOp: 25, MBPerSec: 250, BytesPerOp: 2, AllocsPerOp: 3}}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := newReportBuilder()
			got := b.groupBenchmarksByName(test.in)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("groupBenchmarksByName result incorrect, diff (-want, +got):\n%s\n", diff)
			}
		})
	}
}
