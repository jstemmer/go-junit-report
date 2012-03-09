package main

import (
	"fmt"
	"io"
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
	err = report.XML(os.Stdout)
	if err != nil {
		fmt.Printf("Error writing XML: %s\n", err)
		os.Exit(1)
	}
}

func (r Report) XML(io.Writer) error {
	return nil
}
