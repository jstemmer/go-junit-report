package main

import (
	"flag"
	"fmt"
	"os"
)

var noXmlHeader bool

func init() {
	flag.BoolVar(&noXmlHeader, "no-xml-header", false, "do not print xml header")
}

func main() {
	flag.Parse()

	// Read input
	report, err := Parse(os.Stdin)
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
