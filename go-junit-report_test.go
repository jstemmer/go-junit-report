package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jstemmer/go-junit-report/v2/pkg/junit"
	"github.com/jstemmer/go-junit-report/v2/pkg/parser/gotest"

	"github.com/google/go-cmp/cmp"
)

const testDataDir = "testdata/"

var matchTest = flag.String("match", "", "only test testdata matching this pattern")

type TestConfig struct {
	noXMLHeader bool
	packageName string
}

var testConfigs = map[int]TestConfig{
	5: {noXMLHeader: true},
	6: {noXMLHeader: true},
	7: {packageName: "test/package"},
}

func TestRun(t *testing.T) {
	matchRegex := compileMatch(t)

	files, err := filepath.Glob(testDataDir + "*.txt")
	if err != nil {
		t.Fatalf("error finding files in testdata: %v", err)
	}

	for _, file := range files {
		if !matchRegex.MatchString(file) {
			continue
		}

		conf, reportFile, err := testFileConfig(strings.TrimPrefix(file, testDataDir))
		if err != nil {
			t.Errorf("testFileConfig error: %v", err)
			continue
		}

		t.Run(file, func(t *testing.T) {
			testRun(file, testDataDir+reportFile, conf, t)
		})
	}
}

func testRun(inputFile, reportFile string, config TestConfig, t *testing.T) {
	input, err := os.Open(inputFile)
	if err != nil {
		t.Fatalf("error opening input file: %v", err)
	}
	defer input.Close()

	wantReport, err := os.ReadFile(reportFile)
	if os.IsNotExist(err) {
		t.Skipf("Skipping test with missing report file: %s", reportFile)
	} else if err != nil {
		t.Fatalf("error loading report file: %v", err)
	}

	parser := gotest.New(
		gotest.PackageName(config.packageName),
		gotest.TimestampFunc(func() time.Time {
			return time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		}),
	)

	report, err := parser.Parse(input)
	if err != nil {
		t.Fatal(err)
	}

	for i := range report.Packages {
		report.Packages[i].SetProperty("go.version", "1.0")
	}
	testsuites := junit.CreateFromReport(report, "hostname")

	var output bytes.Buffer
	if err := writeXML(&output, testsuites, config.noXMLHeader); err != nil {
		t.Fatalf("error writing XML: %v", err)
	}

	if diff := cmp.Diff(output.String(), string(wantReport)); diff != "" {
		t.Errorf("Unexpected report diff (-got, +want):\n%v", diff)
	}
}

func testFileConfig(filename string) (conf TestConfig, reportFile string, err error) {
	var prefix string
	if idx := strings.IndexByte(filename, '-'); idx < 0 {
		return conf, "", fmt.Errorf("testdata file does not contain a dash (-); expected name `{id}-{name}.txt` got `%s`", filename)
	} else {
		prefix = filename[:idx]
	}
	id, err := strconv.Atoi(prefix)
	if err != nil {
		return conf, "", fmt.Errorf("testdata file did not start with a valid number: %w", err)
	}
	return testConfigs[id], fmt.Sprintf("%s-report.xml", prefix), nil
}

func compileMatch(t *testing.T) *regexp.Regexp {
	rx, err := regexp.Compile(*matchTest)
	if err != nil {
		t.Fatalf("Error compiling -match flag %q: %v", *matchTest, err)
	}
	return rx
}
