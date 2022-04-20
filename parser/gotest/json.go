package gotest

import (
	"bufio"
	"encoding/json"
	"io"
	"time"

	"github.com/jstemmer/go-junit-report/v2/gtr"
)

// NewJSONParser returns a new Go test json output parser.
func NewJSONParser(options ...Option) *JSONParser {
	return &JSONParser{gp: NewParser(options...)}
}

// Parser is a Go test json output Parser.
type JSONParser struct {
	gp *Parser
}

// Parse parses Go test json output from the given io.Reader r and returns
// gtr.Report.
func (p *JSONParser) Parse(r io.Reader) (gtr.Report, error) {
	return p.gp.Parse(newJSONReader(r))
}

// Events returns the events created by the parser.
func (p *JSONParser) Events() []Event {
	return p.gp.Events()
}

type jsonEvent struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

type jsonReader struct {
	r   *bufio.Reader
	buf []byte
}

func newJSONReader(reader io.Reader) *jsonReader {
	return &jsonReader{r: bufio.NewReader(reader)}
}

func (j *jsonReader) Read(p []byte) (int, error) {
	var err error
	for len(j.buf) == 0 {
		j.buf, err = j.readNextLine()
		if err != nil {
			return 0, err
		}
	}
	n := copy(p, j.buf)
	j.buf = j.buf[n:]
	return n, nil
}

func (j jsonReader) readNextLine() ([]byte, error) {
	line, err := j.r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) == 0 || line[0] != '{' {
		return line, nil
	}
	var event jsonEvent
	if err := json.Unmarshal(line, &event); err != nil {
		return nil, err
	}
	return []byte(event.Output), nil
}
