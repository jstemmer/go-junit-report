package parser

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// struct copied from https://godoc.org/github.com/golang/go/src/cmd/test2json
type testEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

type testoutput []string

func (o *testoutput) Append(s string) {
	s = strings.TrimSuffix(s, "\n")
	s = strings.TrimPrefix(s, "\t")
	*o = append(*o, s)
}

type pkgoutput []string

func (o *pkgoutput) Append(s string) {
	statusPrefixes := []string{"FAIL", "PASS", "SKIP", "?"}
	for _, prefix := range statusPrefixes {
		if strings.HasPrefix(s, prefix) {
			return
		}
	}
	s = strings.TrimSuffix(s, "\n")
	*o = append(*o, s)
}

type jsonlTest struct {
	result  Result
	output  testoutput
	elapsed float64
}

type jsonlPackage struct {
	tests           map[string]*jsonlTest
	testsOrder      []string
	elapsed         float64
	coverage        string
	output          pkgoutput
	defaultTestName string
}

type jsonlParser struct {
	pkgName       string
	packages      map[string]*jsonlPackage
	packagesOrder []string

	currentlyBuiltPackage *jsonlPackage
}

func newJsonlParser(pkgName string) parser {
	return &jsonlParser{
		pkgName:  pkgName,
		packages: make(map[string]*jsonlPackage),
	}
}

func (p *jsonlParser) ingestJSONLine(line string) error {
	var ev testEvent
	err := json.Unmarshal([]byte(line), &ev)
	if err != nil {
		return err
	}

	act := ev.Action
	if act == "pause" || act == "cont" {
		return nil
	}

	if act == "run" {
		return p.recordTest(ev)
	}

	if act == "pass" || act == "fail" || act == "skip" {
		return p.recordResult(ev)
	}

	if act == "output" {
		return p.recordOutput(ev)
	}
	return nil
}

func (p *jsonlParser) ingestBuildLine(line string) error {
	if strings.HasPrefix(line, "FAIL ") {
		cpts := regexp.MustCompile("\\s+").Split(line, 3)
		if len(cpts) != 3 {
			return fmt.Errorf("Invalid format: %v", cpts)
		}
		pkg := p.getPackage(cpts[1])
		pkg.defaultTestName = cpts[2]
		p.cookupFailure(cpts[1])
		p.packagesOrder = append(p.packagesOrder, cpts[1])
		return nil
	}

	if strings.HasPrefix(line, "# ") {
		name := line[2:]
		p.currentlyBuiltPackage = p.getPackage(name)
	} else {
		p.currentlyBuiltPackage.output.Append(line)
	}
	return nil
}

func (p *jsonlParser) IngestLine(line string) error {
	if strings.HasPrefix(line, "{") {
		return p.ingestJSONLine(line)
	}
	return p.ingestBuildLine(line)
}

func (p *jsonlParser) Report() (*Report, error) {
	r := Report{
		Packages: make([]Package, len(p.packages)),
	}

	for i, pname := range p.packagesOrder {
		pkg := p.packages[pname]
		tests := make([]*Test, len(pkg.tests))

		for j, tname := range pkg.testsOrder {
			t := pkg.tests[tname]
			tests[j] = &Test{
				Name:   tname,
				Time:   int(t.elapsed * 1000),
				Result: t.result,
				Output: t.output,
			}
		}

		r.Packages[i] = Package{
			Name:        pname,
			Time:        int(pkg.elapsed * 1000),
			Tests:       tests,
			CoveragePct: pkg.coverage,
		}
	}
	return &r, nil
}

func (p *jsonlParser) recordTest(ev testEvent) error {
	_ = p.getTest(ev.Package, ev.Test)
	return nil
}

func (p *jsonlParser) cookupFailure(pck string) {
	pkg := p.getPackage(pck)
	t := p.getTest(pck, pkg.defaultTestName)
	t.result = FAIL
	t.output = testoutput(pkg.output)
}

func (p *jsonlParser) recordResult(ev testEvent) error {
	pkg := p.getPackage(ev.Package)
	if ev.Test == "" {
		pkg.elapsed = ev.Elapsed
		p.packagesOrder = append(p.packagesOrder, ev.Package)
		if ev.Action == "fail" && len(pkg.tests) == 0 {
			p.cookupFailure(ev.Package)
		}
		return nil
	}

	act := ev.Action
	var status Result

	switch act {
	case "pass":
		status = PASS
	case "fail":
		status = FAIL
	default:
		status = SKIP
	}

	t := p.getTest(ev.Package, ev.Test)
	t.result = status
	t.elapsed = ev.Elapsed
	return nil
}

func (p *jsonlParser) recordOutput(ev testEvent) error {
	out := ev.Output
	// skip control messages as we should have a proper corresponding event
	if strings.HasPrefix(out, "--- ") || strings.HasPrefix(out, "=== ") {
		return nil
	}
	if ev.Test == "" {
		pkg := p.getPackage(ev.Package)
		if strings.HasPrefix(out, "coverage: ") {
			cpts := strings.Split(out[10:], "%")
			pct := cpts[0]
			if !strings.Contains(pct, ".") {
				pct = pct + ".0"
			}
			pkg.coverage = pct
			return nil
		}

		pkg.output.Append(out)
		return nil
	}

	p.getTest(ev.Package, ev.Test).output.Append(out)
	return nil
}

func (p *jsonlParser) getPackage(pname string) *jsonlPackage {
	if pkg, ok := p.packages[pname]; ok {
		return pkg
	}
	pkg := &jsonlPackage{
		tests:           make(map[string]*jsonlTest),
		defaultTestName: "Failure",
	}
	p.packages[pname] = pkg
	return pkg
}

func (p *jsonlParser) getTest(pname, tname string) *jsonlTest {
	pkg := p.getPackage(pname)

	if t, ok := pkg.tests[tname]; ok {
		return t
	}
	t := &jsonlTest{}
	pkg.tests[tname] = t
	pkg.testsOrder = append(pkg.testsOrder, tname)
	return t
}
