package gojunitreport

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

const testDataDir = "../../testdata/"

var matchTest = flag.String("match", "", "only test testdata matching this pattern")

var testConfigs = map[int]Config{
	5:  {SkipXMLHeader: true},
	6:  {SkipXMLHeader: true},
	7:  {PackageName: "test/package"},
	39: {Properties: make(map[string]string)},
}

func TestRun(t *testing.T) {
	matchRegex := compileMatch(t)

	files, err := filepath.Glob(testDataDir + "*.txt")
	if err != nil {
		t.Fatalf("error finding files in testdata: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("no files found in %s", testDataDir)
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

		t.Run(filepath.Base(file), func(t *testing.T) {
			testRun(file, testDataDir+reportFile, conf, t)
		})
	}
}

func testRun(inputFile, reportFile string, config Config, t *testing.T) {
	input, err := os.Open(inputFile)
	if err != nil {
		t.Fatalf("error opening input file: %v", err)
	}
	defer input.Close()

	wantReport, err := ioutil.ReadFile(reportFile)
	if os.IsNotExist(err) {
		t.Skipf("Skipping test with missing report file: %s", reportFile)
	} else if err != nil {
		t.Fatalf("error loading report file: %v", err)
	}

	config.Parser = "gotest"
	if strings.HasSuffix(inputFile, ".gojson.txt") {
		config.Parser = "gojson"
	}
	config.Hostname = "hostname"
	if config.Properties == nil {
		config.Properties = map[string]string{"go.version": "1.0"}
	}
	config.TimestampFunc = func() time.Time {
		return time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	var output bytes.Buffer
	if _, err := config.Run(input, &output); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(string(wantReport), output.String()); diff != "" {
		t.Errorf("Unexpected report diff (-want, +got):\n%v", diff)
	}
}

func testFileConfig(filename string) (config Config, reportFile string, err error) {
	idx := strings.IndexByte(filename, '-')
	if idx < 0 {
		return config, "", fmt.Errorf("testdata file does not contain a dash (-); expected name `{id}-{name}.txt` got `%s`", filename)
	}

	prefix := filename[:idx]
	id, err := strconv.Atoi(prefix)
	if err != nil {
		return config, "", fmt.Errorf("testdata file did not start with a valid number: %w", err)
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
