//go:generate go run generate-golden.go -w

package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/junit"
	"github.com/jstemmer/go-junit-report/v2/pkg/parser/gotest"
)

var verbose bool

type Settings struct {
	skipXMLHeader bool
	packageName   string
}

var fileSettings = map[string]Settings{
	"005-no_xml_header.txt": {skipXMLHeader: true},
	"006-mixed.txt":         {skipXMLHeader: true},
	"007-compiled_test.txt": {packageName: "test/package"},
}

func main() {
	var writeFiles bool
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.BoolVar(&writeFiles, "w", false, "write output xml files")
	flag.Parse()

	files, err := filepath.Glob("*.txt")
	if err != nil {
		exitf("error listing files: %v\n", err)
	}

	for _, file := range files {
		outName := outputName(file)
		if err := createReportFromInput(file, outName, writeFiles); err != nil {
			logf("error creating report: %v\n", err)
			continue
		}
		if writeFiles {
			logf("report written to %s\n", outName)
		}
	}
}

func logf(msg string, args ...interface{}) {
	if verbose {
		fmt.Printf(msg, args...)
	}
}

func exitf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(1)
}

func outputName(input string) string {
	dir, name := filepath.Split(input)
	var out string
	if idx := strings.IndexByte(name, '-'); idx > -1 {
		out = input[:idx+1] + "report.xml"
	} else {
		out = strings.TrimSuffix(name, filepath.Ext(name)) + "report.xml"
	}
	return filepath.Join(dir, out)
}

func createReportFromInput(inputFile, outputFile string, write bool) error {
	in, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer in.Close()

	out := io.Discard
	if write {
		f, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}
	return writeReport(in, out, fileSettings[inputFile])
}

func writeReport(in io.Reader, out io.Writer, settings Settings) error {
	parser := gotest.New(
		gotest.PackageName(settings.packageName),
		gotest.TimestampFunc(func() time.Time {
			return time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		}),
	)

	report, err := parser.Parse(in)
	if err != nil {
		return err
	}
	for i := range report.Packages {
		report.Packages[i].SetProperty("go.version", "1.0")
	}
	testsuites := junit.CreateFromReport(report, "hostname")

	if !settings.skipXMLHeader {
		if _, err := fmt.Fprintf(out, xml.Header); err != nil {
			return err
		}
	}

	enc := xml.NewEncoder(out)
	enc.Indent("", "\t")
	if err := enc.Encode(testsuites); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "\n")
	return err
}
