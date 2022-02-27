package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/jstemmer/go-junit-report/parser"
)

var (
	Version   = "v1.0.0-dev"
	Revision  = "HEAD"
	BuildTime string
)

var (
	noXMLHeader   = flag.Bool("no-xml-header", false, "do not print xml header")
	packageName   = flag.String("package-name", "", "specify a package name (compiled test have no package name in output)")
	packagePrefix = flag.String("package-prefix", "", "specify a package prefix that can be used to namespace a test suite")
	goVersionFlag = flag.String("go-version", "", "specify the value to use for the go.version property in the generated XML")
	setExitCode   = flag.Bool("set-exit-code", false, "set exit code to 1 if tests failed")
	version       = flag.Bool("version", false, "print version")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("go-junit-report %s %s (%s)\n", Version, BuildTime, Revision)
		return
	}

	if flag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "%s does not accept positional arguments\n", os.Args[0])
		flag.Usage()
		os.Exit(1)
	}

	// Read input
	report, err := parser.Parse(os.Stdin, *packageName)
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		os.Exit(1)
	}

	// Replace report with a version that has prefixed package names before formatting
	if *packagePrefix != "" {
		report = report.PrefixPackage(*packagePrefix)
	}

	// Write xml
	err = formatter.JUnitReportXML(report, *noXMLHeader, *goVersionFlag, os.Stdout)
	if err != nil {
		fmt.Printf("Error writing XML: %s\n", err)
		os.Exit(1)
	}

	if *setExitCode && report.Failures() > 0 {
		os.Exit(1)
	}
}
