package gotest

import (
	"testing"

	"github.com/jstemmer/go-junit-report/v2/gtr"

	"github.com/google/go-cmp/cmp"
)

func TestGroupBenchmarksByName(t *testing.T) {
	tests := []struct {
		in   []gtr.Benchmark
		want []gtr.Benchmark
	}{
		{nil, nil},
		{
			[]gtr.Benchmark{{Name: "BenchmarkFailed", Result: gtr.Fail}},
			[]gtr.Benchmark{{Name: "BenchmarkFailed", Result: gtr.Fail}},
		},
		{
			[]gtr.Benchmark{
				{Name: "BenchmarkOne", Result: gtr.Pass, NsPerOp: 10, MBPerSec: 400, BytesPerOp: 1, AllocsPerOp: 2},
				{Name: "BenchmarkOne", Result: gtr.Pass, NsPerOp: 20, MBPerSec: 300, BytesPerOp: 1, AllocsPerOp: 4},
				{Name: "BenchmarkOne", Result: gtr.Pass, NsPerOp: 30, MBPerSec: 200, BytesPerOp: 1, AllocsPerOp: 8},
				{Name: "BenchmarkOne", Result: gtr.Pass, NsPerOp: 40, MBPerSec: 100, BytesPerOp: 5, AllocsPerOp: 2},
			},
			[]gtr.Benchmark{
				{Name: "BenchmarkOne", Result: gtr.Pass, NsPerOp: 25, MBPerSec: 250, BytesPerOp: 2, AllocsPerOp: 4},
			},
		},
		{
			[]gtr.Benchmark{
				{Name: "BenchmarkMixed", Result: gtr.Unknown},
				{Name: "BenchmarkMixed", Result: gtr.Pass, NsPerOp: 10, MBPerSec: 400, BytesPerOp: 1, AllocsPerOp: 2},
				{Name: "BenchmarkMixed", Result: gtr.Pass, NsPerOp: 40, MBPerSec: 100, BytesPerOp: 3, AllocsPerOp: 4},
				{Name: "BenchmarkMixed", Result: gtr.Fail},
			},
			[]gtr.Benchmark{
				{Name: "BenchmarkMixed", Result: gtr.Fail, NsPerOp: 25, MBPerSec: 250, BytesPerOp: 2, AllocsPerOp: 3},
			},
		},
	}

	for _, test := range tests {
		got := groupBenchmarksByName(test.in)
		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Errorf("groupBenchmarksByName result incorrect, diff (-want, +got):\n%s\n", diff)
		}
	}
}
