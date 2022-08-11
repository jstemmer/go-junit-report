package reader

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const testingLimit = 4 * 1024 * 1024

func TestLimitedLineReader(t *testing.T) {
	tests := []struct {
		desc      string
		inputSize int
	}{
		{"small size", 128},
		{"under buf size", 4095},
		{"buf size", 4096},
		{"multiple of buf size ", 4096 * 2},
		{"not multiple of buf size", 10 * 1024},
		{"bufio.MaxScanTokenSize", bufio.MaxScanTokenSize},
		{"over bufio.MaxScanTokenSize", bufio.MaxScanTokenSize + 1},
		{"under limit", testingLimit - 1},
		{"at limit", testingLimit},
		{"just over limit", testingLimit + 1},
		{"over limit", testingLimit + 128},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			line1 := string(make([]byte, test.inputSize))
			line2 := "other line"
			input := strings.NewReader(strings.Join([]string{line1, line2}, "\n"))
			r := NewLimitedLineReader(input, testingLimit)

			got, _, err := r.ReadLine()
			if err != nil {
				t.Fatalf("ReadLine() returned error %v", err)
			}

			want := line1
			if len(line1) > testingLimit {
				want = want[:testingLimit]
			}
			if got != want {
				t.Fatalf("ReadLine() returned incorrect line, got len %d want len %d", len(got), len(want))
			}

			got, _, err = r.ReadLine()
			if err != nil {
				t.Fatalf("ReadLine() returned error %v", err)
			}
			want = line2
			if got != want {
				t.Fatalf("ReadLine() returned incorrect line, got len %d want len %d", len(got), len(want))
			}

			got, _, err = r.ReadLine()
			if err != io.EOF {
				t.Fatalf("ReadLine() returned unexpected error, got %v want %v\n", err, io.EOF)
			}
			if got != "" {
				t.Fatalf("ReadLine() returned unexpected line, got %v want nothing\n", got)
			}
		})
	}
}

func TestJSONEventReader(t *testing.T) {
	input := `some other output
{"Time":"2019-10-09T00:00:00.708139047+00:00","Action":"output","Package":"package/name/ok","Test":"TestOK"}
{"Time":"2019-10-09T00:00:00.708139047+00:00","Action":"output","Package":"package/name/ok","Test":"TestOK","Output":"=== RUN   TestOK\n"}
`
	want := []struct {
		line     string
		metadata *Metadata
	}{
		{"some other output", nil},
		{"=== RUN   TestOK", &Metadata{Package: "package/name/ok"}},
	}

	r := NewJSONEventReader(strings.NewReader(input))
	for i := 0; i < len(want); i++ {
		line, metadata, err := r.ReadLine()
		if err == io.EOF {
			return
		} else if err != nil {
			t.Fatalf("ReadLine() returned error %v", err)
		}

		if diff := cmp.Diff(want[i].line, line); diff != "" {
			t.Errorf("ReadLine() returned incorrect line, diff (-want, +got):\n%s\n", diff)
		}
		if diff := cmp.Diff(want[i].metadata, metadata); diff != "" {
			t.Errorf("ReadLine() Returned incorrect metadata, diff (-want, +got):\n%s\n", diff)
		}
	}
}
