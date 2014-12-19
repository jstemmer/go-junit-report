package main

import (
	"flag"
	"fmt"
	"os"
)

var noXmlHeader bool
var compiledTest bool
var packageName string

func init() {
	flag.BoolVar(&noXmlHeader, "noXmlHeader", false, "do not print xml header")
	flag.BoolVar(&compiledTest, "compiledTest", false, "if the test is compiled with 'go test -c'")
	flag.StringVar(&packageName, "packageName", "", "package name. if the test is compiled, there is no package name in the output, will use the one specified by this parameter.")
}

func main() {
	flag.Parse()

	// Read input
	report, err := Parse(os.Stdin, compiledTest, packageName)
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
