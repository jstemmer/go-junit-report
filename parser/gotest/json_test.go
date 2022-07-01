package gotest

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var input = `some other output
{"Time":"2019-10-09T00:00:00.708139047+00:00","Action":"output","Package":"package/name/ok","Test":"TestOK"}
{"Time":"2019-10-09T00:00:00.708139047+00:00","Action":"output","Package":"package/name/ok","Test":"TestOK","Output":"=== RUN   TestOK\n"}
`

func TestJSONReaderReadAll(t *testing.T) {
	r := newJSONReader(strings.NewReader(input))
	got, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	want := `some other output
=== RUN   TestOK
`

	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("unexpected result from jsonReader, diff (-want, +got):\n%s\n", diff)
	}
}

func TestJSONReaderReadSmallBuffer(t *testing.T) {
	expected := [][]byte{
		[]byte("some"),
		[]byte(" oth"),
		[]byte("er o"),
		[]byte("utpu"),
		[]byte("t\n"),
		[]byte("=== "),
		[]byte("RUN "),
		[]byte("  Te"),
		[]byte("stOK"),
		[]byte("\n"),
	}

	r := newJSONReader(strings.NewReader(input))
	buf := make([]byte, 4)
	for _, want := range expected {
		n, err := r.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}

		got := buf[:n]
		if diff := cmp.Diff(string(want), string(got)); diff != "" {
			t.Fatalf("unexpected result from jsonReader, diff (-want, +got):\n%s\n", diff)
		}
	}

	_, err := r.Read(buf)
	if err != io.EOF {
		t.Fatalf("unexpected error from jsonReader: got %v, want %v", err, io.EOF)
	}
}
