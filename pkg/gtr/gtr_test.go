package gtr

import (
	"testing"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/parser/gotest"

	"github.com/google/go-cmp/cmp"
)

func TestFromEvents(t *testing.T) {
	events := []gtop.Event{
		{Type: "run_test", Id: 1, Name: "TestOne"},
		{Type: "output", Data: "\tHello"},
		{Type: "end_test", Id: 1, Name: "TestOne", Result: "PASS", Duration: 1 * time.Millisecond},
		{Type: "status", Result: "PASS"},
		{Type: "run_test", Id: 2, Name: "TestSkip"},
		{Type: "end_test", Id: 2, Name: "TestSkip", Result: "SKIP", Duration: 1 * time.Millisecond},
		{Type: "summary", Result: "ok", Name: "package/name", Duration: 1 * time.Millisecond},
		{Type: "run_test", Id: 3, Name: "TestOne"},
		{Type: "output", Data: "\tfile_test.go:10: error"},
		{Type: "end_test", Id: 3, Name: "TestOne", Result: "FAIL", Duration: 1 * time.Millisecond},
		{Type: "status", Result: "FAIL"},
		{Type: "summary", Result: "FAIL", Name: "package/name2", Duration: 1 * time.Millisecond},
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
						Result:   PASS,
						Output: []string{
							"\tHello", // TODO: strip tabs?
						},
					},
					{
						Name:     "TestSkip",
						Duration: 1 * time.Millisecond,
						Result:   SKIP,
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
						Result:   FAIL,
						Output: []string{
							"\tfile_test.go:10: error",
						},
					},
				},
			},
		},
	}

	actual := FromEvents(events)
	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("FromEvents report incorrect, diff (-got, +want):\n%v", diff)
	}
}
