package junit

import (
	"bytes"
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
				Properties: []gtr.Property{{Name: "go.version", Value: "go1.18"}},
				Tests: []gtr.Test{
					{
						Name:   "TestPass",
						Result: gtr.Pass,
						Output: []string{"ok"},
					},
					{
						Name:   "TestEscapeOutput",
						Result: gtr.Pass,
						Output: []string{"\x00\v\f \t\\"},
					},
					{
						Name:   "TestRemoveOutputANSI",
						Result: gtr.Pass,
						Output: []string{"This contains some", "\x1b[38;5;140mANSI\x1b[0m", "sequence"},
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
		Tests:    8,
		Errors:   3,
		Failures: 1,
		Skipped:  1,
		Suites: []Testsuite{
			{
				Name:      "package/name",
				Tests:     8,
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
						Name:      "TestEscapeOutput",
						Classname: "package/name",
						Time:      "0.000",
						SystemOut: &Output{Data: `��� 	\`},
					},
					{
						Name:      "TestRemoveOutputANSI",
						Classname: "package/name",
						Time:      "0.000",
						SystemOut: &Output{Data: "This contains some\nANSI\nsequence"},
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
				Name:      "suite1",
				Tests:     1,
				Errors:    1,
				Failures:  1,
				Hostname:  "localhost",
				ID:        0,
				Package:   "package",
				Skipped:   1,
				Time:      "12.345",
				Timestamp: "2012-03-09T14:38:06+01:00",
				Properties: &[]Property{
					{Name: "key", Value: "value"},
				},
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

func TestWriteXML(t *testing.T) {
	want := `<testsuites tests="1">
	<testsuite name="Example" tests="1" failures="0" errors="0" id="0" time="">
		<testcase name="Test" classname=""></testcase>
	</testsuite>
</testsuites>
`

	var suites Testsuites

	ts := Testsuite{Name: "Example"}
	ts.AddTestcase(Testcase{Name: "Test"})
	suites.AddSuite(ts)

	var buf bytes.Buffer
	if err := suites.WriteXML(&buf); err != nil {
		t.Fatalf("WriteXML failed: %v\n", err)
	}

	got := buf.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("WriteXML mismatch, diff (-want +got):\n%s\n", diff)
	}
}
