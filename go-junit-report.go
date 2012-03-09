package main

import (
	"fmt"
	"os"
)

func main() {
	// Read input
	report, err := Parse(os.Stdin)
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		os.Exit(1)
	}

	// Write xml
	err = JUnitReportXML(report, os.Stdout)
	if err != nil {
		fmt.Printf("Error writing XML: %s\n", err)
		os.Exit(1)
	}
}
