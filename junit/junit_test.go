package junit

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/jstemmer/go-junit-report/v2/gtr"

	"github.com/google/go-cmp/cmp"
)

func TestCreateFromReport(t *testing.T) {
	report := gtr.Report{
		Packages: []gtr.Package{
			{
				Name:       "package/name",
				Timestamp:  time.Date(2022, 6, 26, 0, 0, 0, 0, time.UTC),
				Duration:   1 * time.Second,
				Coverage:   0.9,
				Output:     []string{"output"},
				Properties: map[string]string{"go.version": "go1.18"},
				Tests: []gtr.Test{
					{
						Name:   "TestPass",
						Result: gtr.Pass,
						Output: []string{"ok"},
					},
					{
						Name:   "TestFail",
						Result: gtr.Fail,
						Output: []string{"fail"},
					},
					{
						Name:   "TestSkip",
						Result: gtr.Skip,
					},
					{
						Name:   "TestIncomplete",
						Result: gtr.Unknown,
					},
				},
				BuildError: gtr.Error{Name: "Build error"},
				RunError:   gtr.Error{Name: "Run error"},
			},
		},
	}

	want := Testsuites{
		Tests:    6,
		Errors:   3,
		Failures: 1,
		Skipped:  1,
		Suites: []Testsuite{
			{
				Name:      "package/name",
				Tests:     6,
				Errors:    3,
				ID:        0,
				Failures:  1,
				Skipped:   1,
				Time:      "1.000",
				Timestamp: "2022-06-26T00:00:00Z",
				Properties: &[]Property{
					{Name: "go.version", Value: "go1.18"},
					{Name: "coverage.statements.pct", Value: "0.90"},
				},
				Testcases: []Testcase{
					{
						Name:      "TestPass",
						Classname: "package/name",
						Time:      "0.000",
						SystemOut: &Output{Data: "ok"},
					},
					{
						Name:      "TestFail",
						Classname: "package/name",
						Time:      "0.000",
						Failure:   &Result{Message: "Failed", Data: "fail"},
					},
					{
						Name:      "TestSkip",
						Classname: "package/name",
						Time:      "0.000",
						Skipped:   &Result{Message: "Skipped"},
					},
					{
						Name:      "TestIncomplete",
						Classname: "package/name",
						Time:      "0.000",
						Error:     &Result{Message: "No test result found"},
					},
					{
						Classname: "Build error",
						Time:      "0.000",
						Error:     &Result{Message: "Build error"},
					},
					{
						Name:      "Failure",
						Classname: "Run error",
						Time:      "0.000",
						Error:     &Result{Message: "Runtime error"},
					},
				},
				SystemOut: &Output{Data: "output"},
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
				ID:         0,
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
