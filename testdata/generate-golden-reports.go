//go:generate go run generate-golden-reports.go -w

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jstemmer/go-junit-report/v2/internal/gojunitreport"
)

var verbose bool

var configs = map[string]gojunitreport.Config{
	"005-no-xml-header.txt": {SkipXMLHeader: true},
	"006-mixed.txt":         {SkipXMLHeader: true},
	"007-compiled_test.txt": {PackageName: "test/package"},
}

func main() {
	var writeFiles bool
	var id int
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.BoolVar(&writeFiles, "w", false, "write output xml files")
	flag.IntVar(&id, "id", 0, "generate report for given id only")
	flag.Parse()

	files, err := filepath.Glob("*.txt")
	if err != nil {
		exitf("error listing files: %v\n", err)
	}

	var idPrefix string
	if id > 0 {
		idPrefix = fmt.Sprintf("%03d-", id)
	}
	for _, file := range files {
		if idPrefix != "" && !strings.HasPrefix(file, idPrefix) {
			continue
		}

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

	config := configs[inputFile]
	config.Parser = "gotest"
	if strings.HasSuffix(inputFile, ".gojson.txt") {
		config.Parser = "gojson"
	}

	config.Hostname = "hostname"
	config.TimestampFunc = func() time.Time {
		return time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	config.Properties = map[string]string{"go.version": "1.0"}

	_, err = config.Run(in, out)
	return err
}
