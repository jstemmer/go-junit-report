package formatter

import (
	"bytes"
	"encoding/xml"
	"testing"
)

func TestSuites_Unmarshal(t *testing.T) {
	tests := []struct {
		desc        string
		suites      JUnitTestSuites
		noXMLHeader bool
		goVersion   string
	}{
		{
			desc: "Suites should marshal back and forth",
			suites: JUnitTestSuites{
				Suites: []JUnitTestSuite{
					{
						Name: "suite1",
						TestCases: []JUnitTestCase{
							{Name: "test1-1"},
							{Name: "test1-2"},
						},
					},
					{
						Name: "suite2",
						TestCases: []JUnitTestCase{
							{Name: "test2-1"},
							{Name: "test2-2"},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Logf("Test case: %v", test.desc)
		initialBytes, err := xml.Marshal(test.suites)
		if err != nil {
			t.Fatalf("Expected no failure when generating xml; got %v", err)
		}

		var suites JUnitTestSuites
		err = xml.Unmarshal(initialBytes, &suites)
		if err != nil {
			t.Fatalf("Expected no failure when unmarshaling; got %v", err)
		}

		newBytes, err := xml.Marshal(suites)
		if err != nil {
			t.Fatalf("Expected no failure when generating xml again; got %v", err)
		}

		if !bytes.Equal(newBytes, initialBytes) {
			t.Errorf("Expected the same result when marshal/unmarshal/marshal. Expected\n%v\n\t but got\n%v",
				string(initialBytes),
				string(newBytes),
			)
		}
	}
}
