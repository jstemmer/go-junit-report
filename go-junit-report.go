package main

import (
	"flag"
	"fmt"
	"os"
)

var noXmlHeader bool
var packageName string

func init() {
	flag.BoolVar(&noXmlHeader, "no-xml-header", false, "do not print xml header")
	flag.StringVar(&packageName, "package-name", "", "specify a package name (compiled test have no package name in output)")
}

func main() {
	flag.Parse()

	// Read input
	report, err := Parse(os.Stdin, packageName)
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		os.Exit(1)
	}

	// Write xml
	err = JUnitReportXML(report, noXmlHeader, os.Stdout)
	if err != nil {
		fmt.Printf("Error writing XML: %s\n", err)
		os.Exit(1)
	}
}
