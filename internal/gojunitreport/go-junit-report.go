package gojunitreport

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jstemmer/go-junit-report/v2/gtr"
	"github.com/jstemmer/go-junit-report/v2/junit"
	"github.com/jstemmer/go-junit-report/v2/parser/gotest"
)

type parser interface {
	Parse(r io.Reader) (gtr.Report, error)
	Events() []gotest.Event
}

// Config contains the go-junit-report command configuration.
type Config struct {
	Parser        string
	Hostname      string
	PackageName   string
	SkipXMLHeader bool
	SubtestMode   gotest.SubtestMode
	Properties    map[string]string
	TimestampFunc func() time.Time

	// For debugging
	PrintEvents bool
}

// Run runs the go-junit-report command and returns the generated report.
func (c Config) Run(input io.Reader, output io.Writer) (*gtr.Report, error) {
	var p parser
	switch c.Parser {
	case "gotest":
		p = gotest.NewParser(c.gotestOptions()...)
	case "gojson":
		p = gotest.NewJSONParser(c.gotestOptions()...)
	default:
		return nil, fmt.Errorf("invalid parser: %s", c.Parser)
	}

	report, err := p.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("error parsing input: %w", err)
	}

	if c.PrintEvents {
		enc := json.NewEncoder(os.Stderr)
		for _, event := range p.Events() {
			if err := enc.Encode(event); err != nil {
				return nil, err
			}
		}
	}

	for i := range report.Packages {
		for k, v := range c.Properties {
			report.Packages[i].SetProperty(k, v)
		}
	}

	if err = c.writeXML(output, report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (c Config) writeXML(w io.Writer, report gtr.Report) error {
	testsuites := junit.CreateFromReport(report, c.Hostname)

	if !c.SkipXMLHeader {
		_, err := fmt.Fprintf(w, xml.Header)
		if err != nil {
			return err
		}
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "\t")
	if err := enc.Encode(testsuites); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "\n")
	return err
}

func (c Config) gotestOptions() []gotest.Option {
	return []gotest.Option{
		gotest.PackageName(c.PackageName),
		gotest.SetSubtestMode(c.SubtestMode),
		gotest.TimestampFunc(c.TimestampFunc),
	}
}
