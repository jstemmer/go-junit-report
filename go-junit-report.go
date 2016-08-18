package main

import (
	"flag"
	"fmt"
	"os"
	"io"

	"github.com/jstemmer/go-junit-report/parser"
)

var (
	noXMLHeader bool
	packageName string
	setExitCode bool
	outFile     string
)

func init() {
	flag.BoolVar(&noXMLHeader, "no-xml-header", false, "do not print xml header")
	flag.StringVar(&packageName, "package-name", "", "specify a package name (compiled test have no package name in output)")
	flag.BoolVar(&setExitCode, "set-exit-code", false, "set exit code to 1 if tests failed")
	flag.StringVar(&outFile, "out", "", "write output to file instead of stdout")
}

func main() {
	flag.Parse()

	var output io.Writer = os.Stdout
	var input io.Reader = os.Stdin
	if outFile != "" {
		var err error
		if output, err = os.Create(outFile); err != nil {
			fmt.Printf("Error opening file: %s\n", err)
			os.Exit(1)
		}
		input = io.TeeReader(os.Stdin, os.Stdout)
	}

	// Read input
	report, err := parser.Parse(input, packageName)
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		os.Exit(1)
	}

	// Write xml
	err = JUnitReportXML(report, noXMLHeader, output)
	if err != nil {
		fmt.Printf("Error writing XML: %s\n", err)
		os.Exit(1)
	}

	if setExitCode && report.Failures() > 0 {
		os.Exit(1)
	}
}
