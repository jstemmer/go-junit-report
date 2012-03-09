package main

import (
	"fmt"
	"io"
	"os"
)

type Report struct {
}

func main() {
	// Read input
	report, err := parse(os.Stdin)
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

func parse(reader io.Reader) (Report, error) {
	return Report{}, nil
}

func (r Report) XML(io.Writer) error {
	return nil
}
