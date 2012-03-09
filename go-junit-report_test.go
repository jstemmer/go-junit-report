package main

import (
	"strings"
	"testing"
)

const testOutputPass = `=== RUN TestOne
--- PASS: TestOne (0.06 seconds)
=== RUN TestTwo
--- PASS: TestTwo (0.10 seconds)
PASS
ok  	package/name 0.160s`

func TestOutputPass(t *testing.T) {
	_, err := parse(strings.NewReader(testOutputPass))
	if err != nil {
		t.Fatalf("error parsing: %s", err)
	}
}

const testOutputFail = `=== RUN TestOne
--- FAIL: TestOne (0.02 seconds)
	file_test.go:11: Error message
	file_test.go:11: Longer
		error
		message.
=== RUN TestTwo
--- PASS: TestTwo (0.13 seconds)
FAIL
exit status 1
FAIL	package/name 0.151s`

func TestOutputFail(t *testing.T) {
	_, err := parse(strings.NewReader(testOutputPass))
	if err != nil {
		t.Fatalf("error parsing: %s", err)
	}
}
