package reader

import (
	"bufio"
	"io"
	"strings"
	"testing"
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

			got, err := r.ReadLine()
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

			got, err = r.ReadLine()
			if err != nil {
				t.Fatalf("ReadLine() returned error %v", err)
			}
			want = line2
			if got != want {
				t.Fatalf("ReadLine() returned incorrect line, got len %d want len %d", len(got), len(want))
			}

			got, err = r.ReadLine()
			if err != io.EOF {
				t.Fatalf("ReadLine() returned unexpected error, got %v want %v\n", err, io.EOF)
			}
			if got != "" {
				t.Fatalf("ReadLine() returned unexpected line, got %v want nothing\n", got)
			}
		})
	}
}
