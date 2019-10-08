package gotest

import (
	"io"
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
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	want := `some other output
=== RUN   TestOK
`

	if diff := cmp.Diff(string(got), want); diff != "" {
		t.Errorf("unexpected result from jsonReader, diff (-got, +want):\n%s\n", diff)
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

		if diff := cmp.Diff(string(buf[:n]), string(want)); diff != "" {
			t.Fatalf("unexpected result from jsonReader, diff (-got, +want):\n%s\n", diff)
		}
	}

	_, err := r.Read(buf)
	if err != io.EOF {
		t.Fatalf("unexpected result from jsonReader: got %v, want %v", err, io.EOF)
	}
}
