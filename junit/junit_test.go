package junit

import (
	"encoding/xml"
	"testing"

	"github.com/jstemmer/go-junit-report/v2/gtr"

	"github.com/google/go-cmp/cmp"
)

func TestCreateFromReport(t *testing.T) {
	// TODO: complete this report
	report := gtr.Report{
		Packages: []gtr.Package{
			{
				Benchmarks: []gtr.Benchmark{
					{
						Name:   "BenchmarkFail",
						Result: gtr.Fail,
					},
				},
			},
		},
	}

	want := Testsuites{
		Tests:    1,
		Failures: 1,
		Suites: []Testsuite{
			{
				Tests:    1,
				Failures: 1,
				Time:     "0.000",
				Testcases: []Testcase{
					{
						Name:    "BenchmarkFail",
						Time:    "0.000000000",
						Failure: &Result{Message: "Failed"},
					},
				},
			},
		},
	}

	got := CreateFromReport(report, "")
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("CreateFromReport incorrect, diff (-want, +got):\n%s\n", diff)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	want := Testsuites{
		Name:     "name",
		Time:     "12.345",
		Tests:    1,
		Errors:   1,
		Failures: 1,
		Disabled: 1,
		Suites: []Testsuite{
			{
				Name:       "suite1",
				Tests:      1,
				Errors:     1,
				Failures:   1,
				Hostname:   "localhost",
				ID:         1,
				Package:    "package",
				Skipped:    1,
				Time:       "12.345",
				Timestamp:  "2012-03-09T14:38:06+01:00",
				Properties: properties("key", "value"),
				Testcases: []Testcase{
					{
						Name:      "test1",
						Classname: "class",
						Time:      "12.345",
						Status:    "status",
						Skipped:   &Result{Message: "skipped", Type: "type", Data: "data"},
						Error:     &Result{Message: "error", Type: "type", Data: "data"},
						Failure:   &Result{Message: "failure", Type: "type", Data: "data"},
						SystemOut: &Output{"system-out"},
						SystemErr: &Output{"system-err"},
					},
				},
				SystemOut: &Output{"system-out"},
				SystemErr: &Output{"system-err"},
			},
		},
	}

	data, err := xml.MarshalIndent(want, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	var got Testsuites
	if err := xml.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	want.XMLName.Local = "testsuites"
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unmarshal result incorrect, diff (-want +got):\n%s\n", diff)
	}
}

func properties(keyvals ...string) *[]Property {
	if len(keyvals)%2 != 0 {
		panic("invalid keyvals specified")
	}
	var props []Property
	for i := 0; i < len(keyvals); i += 2 {
		props = append(props, Property{keyvals[i], keyvals[i+1]})
	}
	return &props
}

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
