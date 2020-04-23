package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/jstemmer/go-junit-report/parser"
)

var (
	noXMLHeader   = flag.Bool("no-xml-header", false, "do not print xml header")
	packageName   = flag.String("package-name", "", "specify a package name (compiled test have no package name in output)")
	goVersionFlag = flag.String("go-version", "", "specify the value to use for the go.version property in the generated XML")
	setExitCode   = flag.Bool("set-exit-code", false, "set exit code to 1 if tests failed")
	goTestFile    = flag.String("go-test-output", "", "specify file containing the contents of gotest verbose execution")
)

func main() {
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintf(os.Stderr, "%s does not accept positional arguments\n", os.Args[0])
		flag.Usage()
		os.Exit(1)
	}

	var goTestOutputReader io.Reader
	if goTestFilePath := *goTestFile; len(goTestFilePath) > 0 {
		if content, readErr := ioutil.ReadFile(goTestFilePath); readErr != nil {
			fmt.Printf("Error reading file[path= %s]: %s", goTestFilePath, readErr)
			os.Exit(1)
		} else {
			goTestOutputReader = bytes.NewBuffer(content)
		}
	} else {
		goTestOutputReader = os.Stdin
	}
	// Read input
	report, err := parser.Parse(goTestOutputReader, *packageName)
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		os.Exit(1)
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
