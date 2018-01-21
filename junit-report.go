package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ujiro99/doctest-junit-report/parser"
)

var (
	noXMLHeader bool
	packageName string
	versionFlag string
	setExitCode bool
)

func init() {
	flag.BoolVar(&noXMLHeader, "no-xml-header", false, "do not print xml header")
	flag.StringVar(&packageName, "package-name", "", "specify a package name (compiled test have no package name in output)")
	flag.StringVar(&versionFlag, "version", "", "specify the value to use for the version property in the generated XML")
	flag.BoolVar(&setExitCode, "set-exit-code", false, "set exit code to 1 if tests failed")
}

func main() {
	flag.Parse()

	// Read input
	report, err := parser.Parse(os.Stdin)
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		os.Exit(1)
	}

	// Write xml
	err = JUnitReportXML(report, noXMLHeader, versionFlag, packageName, os.Stdout)
	if err != nil {
		fmt.Printf("Error writing XML: %s\n", err)
		os.Exit(1)
	}

	if setExitCode && report.Failures() > 0 {
		os.Exit(1)
	}
}
